from helios.orchestrators.orchestrator import AbstractOrchestrator
from helios.components.component import ComponentGroup
from helios.transports.transport import AbstractTransport

from typing import Optional, Type


class Helios:
    def __init__(
        self,
        base_name: str,
        *,
        friendly_name: str,
        transport: Type[AbstractTransport],
        orchestrator: Type[AbstractOrchestrator]
    ):
        # For those unaware, the asterisk in the function signature indicates that the following parameters are keyword-only.
        # This means that they can only be passed by name, and not by position.

        # Argument-defined attributes
        self.base_name: str = base_name
        self.friendly_name: Optional[str] = friendly_name
        self.transport: Type[AbstractTransport] = transport
        self.orchestrator: AbstractOrchestrator = orchestrator(transport)

        # Other attributes
        self.component_tree: ComponentGroup = ComponentGroup(self.base_name)

    def get_component_tree(self) -> ComponentGroup:
        return self.component_tree

    def start(self):
        # Start the components
        self.orchestrator.start_all(self.component_tree)
