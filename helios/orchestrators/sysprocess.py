from helios.orchestrator import AbstractOrchestrator
from helios.component import ComponentManager

# from multiprocessing import Process, Pipe
# from typing import Type


class SystemProcessOrchestrator(AbstractOrchestrator):
    def __init__(self, grpc_host: str, grpc_port: int):
        super().__init__(grpc_host, grpc_port)

    def start(self, component: ComponentManager):
        raise NotImplementedError()
