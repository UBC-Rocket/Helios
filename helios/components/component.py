from __future__ import annotations

from abc import ABC
from typing import Union, Optional, Any


def generate_component_path(component: Union[ComponentGroup, ComponentManager]) -> tuple[str, ...]:
    return (*generate_component_path(component.parent), component.name) if component.parent else (component.name,)


class BaseComponent(ABC):
    def __init__(self):
        pass

    def get_path(self) -> tuple[str, ...]:
        raise NotImplementedError()

    def print_tree(self, last=True, header=''):
        raise NotImplementedError()


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

    def add_component(self, name: str, component: Union[ComponentGroup, AbstractComponent]):
        if name in self.components:
            raise ValueError(f"Component with name {name} already exists")

        if isinstance(component, ComponentGroup):
            self.components[name] = component

        elif isinstance(component, AbstractComponent):
            self.components[name] = ComponentManager(name, component, parent=self)
            component.path = self.components[name].get_path()

        else:
            raise TypeError(f"Unsupported component type: {type(component).__name__}")

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
    def __init__(self, name: str, component_object: AbstractComponent, parent: Optional[ComponentGroup] = None):
        self.running: bool = False
        self.name: str = name
        self.component_object: AbstractComponent = component_object
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
        self.component_object.print_tree(header=header + (blank if last else pipe))


class AbstractComponent(BaseComponent, ABC):
    def __init__(self, *args, **kwargs):
        self.initialized: bool = False
        self.launch_args: tuple[Any, ...] = args
        self.launch_kwargs: dict[str, Any] = kwargs

    def initComponent(self, name: str, path: tuple[str, ...]):
        self.name: str = name
        self.path: tuple[str, ...] = path
        self.initialized = True

    def run(self):
        raise NotImplementedError()

    def get_path(self) -> tuple[str, ...]:
        return self.path

    def print_tree(self, last=True, header=''):
        elbow = "└──"
        # pipe = "│  "
        tee = "├──"
        # blank = "   "
        print(f"{header}{(elbow if last else tee)}({self})")
