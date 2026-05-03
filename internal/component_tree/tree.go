package component_tree

import (
	"fmt"
	"strings"
	"sync"

	"helios/internal/logger"
)

/*
Tree invariants:
  - node is package-private; all shape changes must go through ComponentTree methods.
  - A node is either a branch or a leaf. component == nil means branch, component != nil means leaf.
  - Addresses are derived from parent links and node names. They are not stored on nodes.
  - branches and components index nodes by current address so lookups do not traverse the tree.
  - Only leaves may own Component runtime state.
*/
type node struct {
	parent    *node
	name      string
	component *Component
	children  []*node
}

func (n *node) address() string {
	if n == nil {
		return ""
	}
	if n.parent == nil {
		return n.name
	}
	return n.parent.address() + "." + n.name
}

func (n *node) isBranch() bool {
	return n != nil && n.component == nil
}

type ComponentTree struct {
	mu sync.RWMutex

	rootNode   *node
	branches   map[string]*node
	components map[string]*node
}

/*
 * ========================================================
 * = Public API
 * ========================================================
 */

func New() *ComponentTree {
	return &ComponentTree{
		branches:   make(map[string]*node),
		components: make(map[string]*node),
	}
}

func (t *ComponentTree) ComponentFromAddress(address string) (*Component, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	node, ok := t.components[address]
	if !ok {
		return nil, false
	}
	return node.component, true
}

func (t *ComponentTree) AddRootComponent(name string, component *Component) (string, error) {
	return t.addComponent(nil, name, component)
}

func (t *ComponentTree) AddRootComponentGroup(name string) (string, error) {
	return t.addComponentGroup(nil, name)
}

func (t *ComponentTree) AddComponent(parentAddress string, name string, component *Component) (string, error) {
	parent, err := t.branchAtAddress(parentAddress)
	if err != nil {
		return "", err
	}
	return t.addComponent(parent, name, component)
}

func (t *ComponentTree) AddComponentGroup(parentAddress string, name string) (string, error) {
	parent, err := t.branchAtAddress(parentAddress)
	if err != nil {
		return "", err
	}
	return t.addComponentGroup(parent, name)
}

func (t *ComponentTree) Close() error {
	return nil
}

/*
 * ========================================================
 * = Internal shape mutations
 * ========================================================
 */

// Good luck to the next maintainer!

func (t *ComponentTree) addComponent(parent *node, name string, component *Component) (string, error) {
	logger.Debugw("Adding component", "parent", parent, "name", name, "component", component)
	name = strings.TrimSpace(name)
	if name == "" {
		return "", fmt.Errorf("component name is empty")
	}
	if component == nil {
		return "", fmt.Errorf("component %q is nil", name)
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	// All shape validation happens here so builders cannot desynchronize the tree.
	if err := t.canAttach(parent); err != nil {
		return "", err
	}
	if t.hasChildNamed(parent, name) {
		return "", duplicateChildError(parent, name)
	}

	node := &node{
		parent:    parent,
		name:      name,
		component: component,
	}
	address := node.address()
	logger.Debugw("Component address", "address", address)
	if t.addressExists(address) {
		return "", fmt.Errorf("duplicate node address %q", address)
	}

	if parent == nil {
		t.rootNode = node
	} else {
		parent.children = append(parent.children, node)
	}
	// Register after linking so address() observes the final parent chain.
	t.components[address] = node
	return address, nil
}

func (t *ComponentTree) addComponentGroup(parent *node, name string) (string, error) {
	logger.Debugw("Adding component group", "parent", parent, "name", name)
	name = strings.TrimSpace(name)
	if name == "" {
		return "", fmt.Errorf("component group name is empty")
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	// Branch construction is the only place children can be appended.
	if err := t.canAttach(parent); err != nil {
		return "", err
	}
	if t.hasChildNamed(parent, name) {
		return "", duplicateChildError(parent, name)
	}

	node := &node{
		parent:   parent,
		name:     name,
		children: []*node{},
	}
	address := node.address()
	logger.Debugw("Component group address", "address", address)
	if t.addressExists(address) {
		return "", fmt.Errorf("duplicate node address %q", address)
	}
	if parent == nil {
		t.rootNode = node
	} else {
		parent.children = append(parent.children, node)
	}
	// Branches are indexed too because they can be future parents.
	t.branches[address] = node
	return address, nil
}

/*
 * ========================================================
 * = Internal validation and lookup helpers
 * ========================================================
 */

func (t *ComponentTree) branchAtAddress(address string) (*node, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	parent := t.branches[address]
	if parent == nil {
		return nil, fmt.Errorf("unknown parent component group %q", address)
	}
	return parent, nil
}

func (t *ComponentTree) canAttach(parent *node) error {
	if parent == nil {
		if t.rootNode != nil {
			return fmt.Errorf("root node is already set")
		}
		return nil
	}
	if !parent.isBranch() {
		return fmt.Errorf("parent node %q is not a component group", parent.address())
	}
	return nil
}

func (t *ComponentTree) hasChildNamed(parent *node, name string) bool {
	if parent == nil {
		return t.rootNode != nil && t.rootNode.name == name
	}
	for _, child := range parent.children {
		if child.name == name {
			return true
		}
	}
	return false
}

func (t *ComponentTree) addressExists(address string) bool {
	if _, exists := t.branches[address]; exists {
		return true
	}
	_, exists := t.components[address]
	return exists
}

func duplicateChildError(parent *node, name string) error {
	parentAddress := "<root>"
	if parent != nil {
		parentAddress = parent.address()
	}
	return fmt.Errorf("component tree already has a child named %q under %q", name, parentAddress)
}
