from helios.components.component import ComponentGroup, ComponentManager
from helios.transports.transport import AbstractTransport

from typing import Union, Type
from abc import ABC


class AbstractOrchestrator(ABC):
    def __init__(self, transport: Type[AbstractTransport]):
        self.transport: Type[AbstractTransport] = transport

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
