from helios.component import ComponentGroup
from helios.protocol.generated.protocol_pb2_grpc import add_HeliosProtocolServicer_to_server
from helios.protocol.impl import HeliosProtocol

from concurrent import futures
from typing import Optional
import grpc


class Helios:
    def __init__(
        self,
        base_name: str,
        friendly_name: str,
        grpc_host: str = "[::]",
        grpc_port: int = 50051,
    ):
        # Argument-defined attributes
        self.base_name: str = base_name
        self.grpc_host: str = grpc_host
        self.grpc_port: int = grpc_port
        self.friendly_name: Optional[str] = friendly_name

        # Other attributes
        self.component_tree: ComponentGroup = ComponentGroup(self.base_name)

    def get_component_tree(self) -> ComponentGroup:
        # Return the component tree
        return self.component_tree

    def init_grpc_server(self):
        # Create the gRPC server instance
        self.grpc_server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))

        # Add protocol handler
        add_HeliosProtocolServicer_to_server(HeliosProtocol(), self.grpc_server)

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
