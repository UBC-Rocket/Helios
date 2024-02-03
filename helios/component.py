from __future__ import annotations

from typing import TYPE_CHECKING
if TYPE_CHECKING:
    from helios.components.base import ComponentBase

from abc import ABC
from typing import Optional, Any


def generate_component_path(component: AbstractComponent) -> tuple[str, ...]:
    return (*generate_component_path(component.parent), component.name) if component.parent else (component.name,)


class AbstractComponent(ABC):
    def __init__(self, name: str, parent: Optional[ComponentGroup] = None):
        self.name: str = name
        self.parent: Optional[ComponentGroup] = parent

    def get_path(self) -> tuple[str, ...]:
        raise NotImplementedError()

    def print_tree(self, last=True, header=''):
        raise NotImplementedError()


class ComponentGroup(AbstractComponent):
    def __init__(self, name: str, parent: Optional[ComponentGroup] = None):
        super().__init__(name, parent)
        self.components: dict[str, AbstractComponent] = {}

    def create_component_group(self, name: str) -> ComponentGroup:
        if name in self.components:
            raise ValueError(f"Component with name {name} already exists")

        group: ComponentGroup = ComponentGroup(name, parent=self)
        self.components[name] = group
        return group

    def add_component(self, component: AbstractComponent):
        if component.name in self.components:
            raise ValueError(f"Component with name {component.name} in group {self.get_path()} already exists")

        if isinstance(component, ComponentGroup):
            self.components[component.name] = component

        elif isinstance(component, AbstractComponentManager):
            self.components[component.name] = component
            component.parent = self

        else:
            raise ValueError(f"Component with name {component.name} is not a valid component")

    def get_path(self) -> tuple[str, ...]:
        return (*self.parent.get_path(), self.name) if self.parent else (self.name,)

    def print_tree(self, last=True, header=''):
        if not self.parent:
            print("[HELIOS CORE]")
        children = list(self.components.values())
        elbow = "└──"
        pipe = "│  "
        tee = "├──"
        blank = "   "
        print(f"{header}{(elbow if last else tee)}[{self.name}] ({len(children)} children)")
        for i, c in enumerate(children):
            c.print_tree(header=header + (blank if last else pipe), last=i == len(children) - 1)


class AbstractComponentManager(AbstractComponent, ABC):
    def __init__(self, name: str, parent: Optional[ComponentGroup] = None):
        super().__init__(name, parent)
        self.running: bool = False


class ReferenceComponentManager(AbstractComponentManager):
    def __init__(self, name: str, component_object: ComponentBase):
        super().__init__(name)

        self.component_object: ComponentBase = component_object
        self.reference: Any = None

    def get_path(self) -> tuple[str, ...]:
        return (*self.parent.get_path(), self.name) if self.parent else (self.name,)

    def print_tree(self, last=True, header=''):
        elbow = "└──"
        pipe = "│  "
        tee = "├──"
        blank = "   "
        print(f"{header}{(elbow if last else tee)}<{self.name}: {self.running=}>")
        print(f"{header + (blank if last else pipe)}{elbow}({self})")
