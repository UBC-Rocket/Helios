from helios.component import AbstractComponent, ComponentGroup, AbstractComponentManager

from abc import ABC


class AbstractOrchestrator(ABC):
    def __init__(self, grpc_host: str, grpc_port: int):
        self.grpc_host: str = grpc_host
        self.grpc_port: int = grpc_port

    def start_all(self, component_tree: ComponentGroup):
        def _start_all(comp: AbstractComponent):
            if isinstance(comp, ComponentGroup):
                for _, subcomponent in comp.components.items():
                    _start_all(subcomponent)
            elif isinstance(comp, AbstractComponentManager):
                self.start(comp)

        _start_all(component_tree)

    def start(self, manager: AbstractComponentManager):
        raise NotImplementedError()

    def stop(self, manager: AbstractComponentManager):
        raise NotImplementedError()

    def kill(self, manager: AbstractComponentManager):
        raise NotImplementedError()
