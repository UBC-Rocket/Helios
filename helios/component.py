from __future__ import annotations

from typing import TYPE_CHECKING
if TYPE_CHECKING:
    from helios.helios import Helios
    from helios.components.base import ComponentBase

from abc import ABC
from typing import Optional, Any
from multiprocessing import Process


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

    def start_all(self, server: Helios):
        for c in self.components.values():
            if isinstance(c, AbstractComponentManager):
                c.start(server)
            elif isinstance(c, ComponentGroup):
                c.start_all(server)
            else:
                raise ValueError(f"Component {c.get_path()} is not a valid component")

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

    def start(self, server: Helios):
        raise NotImplementedError()

    def stop(self, server: Helios):
        raise NotImplementedError()

    def kill(self, server: Helios):
        raise NotImplementedError()


class PyComponent(AbstractComponentManager):
    def __init__(self, name: str, component_object: ComponentBase):
        super().__init__(name)

        self.component_object: ComponentBase = component_object
        self.process: Optional[Process] = None
        self.reference: Any = None

    def start(self, server: Helios):
        if self.running:
            raise ValueError(f"Component {self.get_path()} is already running")

        self.component_object.initComponent(
            name=self.name,
            path=self.get_path(),
            grpc_host=server.grpc_host,
            grpc_port=server.grpc_port
        )

        self.reference = Process(
            target=self.component_object.run,
            args=self.component_object.launch_args,
            kwargs=self.component_object.launch_kwargs
        )

        self.running = True
        self.reference.start()

    def get_path(self) -> tuple[str, ...]:
        return (*self.parent.get_path(), self.name) if self.parent else (self.name,)

    def print_tree(self, last=True, header=''):
        elbow = "└──"
        pipe = "│  "
        tee = "├──"
        blank = "   "
        print(f"{header}{(elbow if last else tee)}<{self.name}: {self.running=}>")
        print(f"{header + (blank if last else pipe)}{elbow}({self})")
