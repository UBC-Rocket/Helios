from helios.components.component import ComponentGroup, ComponentManager

from typing import Union
from abc import ABC


class AbstractOrchestrator(ABC):
    def __init__(self, grpc_host: str, grpc_port: int):
        self.grpc_host: str = grpc_host
        self.grpc_port: int = grpc_port

    def start_all(self, component_tree: ComponentGroup):
        def _start_all(comp: Union[ComponentGroup, ComponentManager]):
            if isinstance(comp, ComponentGroup):
                for _, subcomponent in comp.components.items():
                    _start_all(subcomponent)
            else:
                self.start(comp)

        _start_all(component_tree)

    def start(self, manager: ComponentManager):
        raise NotImplementedError()

    def stop(self, manager: ComponentManager):
        raise NotImplementedError()

    def kill(self, manager: ComponentManager):
        raise NotImplementedError()
