from helios.orchestrators.orchestrator import AbstractOrchestrator
from helios.components.component import ComponentManager, generate_component_path
from helios.transports.transport import AbstractTransport
from helios.transports.pipe import PipeTransport

from multiprocessing import Process, Pipe
from typing import Type


class MultiprocessingOrchestrator(AbstractOrchestrator):
    def __init__(self, transport: Type[AbstractTransport]):
        # Check if the transport is of a supported type
        if transport not in (PipeTransport,):
            raise TypeError(f"Unsupported transport type: {transport.__name__}")

        super().__init__(transport)

    def start(self, component: ComponentManager):
        if component.running:
            raise RuntimeError(f"Component {component.name} is already running")

        # Create the transport
        if self.transport == PipeTransport:
            parent_pipe, child_pipe = Pipe(duplex=True)
            parent_transport: PipeTransport = PipeTransport(transmit=parent_pipe, receive=parent_pipe)
            child_transport: PipeTransport = PipeTransport(transmit=child_pipe, receive=child_pipe)
        else:
            raise TypeError(f"Unsupported transport type: {self.transport.__name__}")

        # Initialize the component
        component.component_object.initComponent(component.name, generate_component_path(component), child_transport)

        # Create the component
        component.reference = Process(target=component.component_object.run, args=())

        # Set the manager transport
        component.transport = parent_transport

        # Start the component
        component.running = True
        component.reference.start()
