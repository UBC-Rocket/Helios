from __future__ import annotations
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from helios.components.base import ComponentBase

from typing import Union, Optional, Any


def generate_component_path(component: Union[ComponentGroup, ComponentManager]) -> tuple[str, ...]:
    return (*generate_component_path(component.parent), component.name) if component.parent else (component.name,)


class ComponentGroup():
    def __init__(self, name: str, parent: Optional[ComponentGroup] = None):
        self.name: str = name
        self.parent: Optional[ComponentGroup] = parent
        self.components: dict[str, Union[ComponentGroup, ComponentManager]] = {}

    def create_component_group(self, name: str) -> ComponentGroup:
        if name in self.components:
            raise ValueError(f"Component with name {name} already exists")

        group: ComponentGroup = ComponentGroup(name, parent=self)
        self.components[name] = group
        return group

    def add_component(self, name: str, component: Union[ComponentGroup, ComponentBase]):
        if name in self.components:
            raise ValueError(f"Component with name {name} already exists")

        if isinstance(component, ComponentGroup):
            self.components[name] = component

        else:
            self.components[name] = ComponentManager(name, component, parent=self)
            component.path = self.components[name].get_path()

    def get_path(self) -> tuple[str, ...]:
        return (*self.parent.get_path(), self.name) if self.parent else (self.name,)

    def print_tree(self, last=True, header=''):
        children = list(self.components.values())
        elbow = "└──"
        pipe = "│  "
        tee = "├──"
        blank = "   "
        print(f"{header}{(elbow if last else tee)}[{self.name}] ({len(children)} children)")
        for i, c in enumerate(children):
            c.print_tree(header=header + (blank if last else pipe), last=i == len(children) - 1)


class ComponentManager():
    def __init__(self, name: str, component_object: ComponentBase, parent: Optional[ComponentGroup] = None):
        self.running: bool = False
        self.name: str = name
        self.component_object: ComponentBase = component_object
        self.reference: Any = None
        self.parent: Optional[ComponentGroup] = parent

    def get_path(self) -> tuple[str, ...]:
        return (*self.parent.get_path(), self.name) if self.parent else (self.name,)

    def print_tree(self, last=True, header=''):
        elbow = "└──"
        pipe = "│  "
        tee = "├──"
        blank = "   "
        print(f"{header}{(elbow if last else tee)}<{self.name}: {self.running=}>")
        print(f"{header + (blank if last else pipe)}{elbow}({self})")
