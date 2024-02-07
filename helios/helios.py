from helios.component import ComponentGroup

from concurrent import futures
from typing import Optional
import grpc


class Helios:
    def __init__(
        self,
        base_name: str,
        grpc_host: str = "[::]",
        grpc_port: int = 50051,
        *,
        friendly_name: str
    ):
        # For those unaware, the asterisk in the function signature indicates that the following parameters are keyword-only.
        # This means that they can only be passed by name, and not by position.

        # Argument-defined attributes
        self.base_name: str = base_name
        self.grpc_host: str = grpc_host
        self.grpc_port: int = grpc_port
        self.friendly_name: Optional[str] = friendly_name

        # Other attributes
        self.component_tree: ComponentGroup = ComponentGroup(self.base_name)

    def get_component_tree(self) -> ComponentGroup:
        return self.component_tree

    def init_grpc_server(self):
        # Create the gRPC server instance
        self.grpc_server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))

        # Add the bind info
        self.grpc_server.add_insecure_port(f"{self.grpc_host}:{self.grpc_port}")

    def start(self):
        # Initialize the gRPC server
        self.init_grpc_server()

        # Start the gRPC server
        self.grpc_server.start()

        # Start the components
        self.component_tree.start_all(self)

        # Wait until gRPC server is terminated
        self.grpc_server.wait_for_termination()
