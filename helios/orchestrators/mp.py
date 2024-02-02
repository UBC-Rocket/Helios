from helios.orchestrators.orchestrator import AbstractOrchestrator
from helios.components.component import ComponentManager

from multiprocessing import Process, Pipe
from typing import Type


class MultiprocessingOrchestrator(AbstractOrchestrator):
    def __init__(self, grpc_host: str, grpc_port: int):
        super().__init__(grpc_host, grpc_port)

    def start(self, component: ComponentManager):
        if component.running:
            raise RuntimeError(f"Component {component.name} is already running")

        # Initialize the component
        component.component_object.initComponent(component.name, component.get_path())

        # Create the component
        component.reference = Process(
            target=component.component_object.run,
            args=component.component_object.launch_args,
            kwargs=component.component_object.launch_kwargs
        )

        # Start the component
        component.running = True
        component.reference.start()
