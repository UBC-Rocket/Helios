from helios.component import ComponentGroup
from helios.protocol.generated.protocol_pb2_grpc import add_HeliosProtocolServicer_to_server
from helios.protocol.impl import HeliosProtocol

from concurrent import futures
from typing import Optional
import grpc


class Helios:
    """High level server class for Helios.

    Attributes:
        base_name (str): Base name of the server.
        friendly_name (str): Print-friendly name of the server.
        grpc_host (int): Host of the gRPC server.
        grpc_port (int): Port of the gRPC server.
        component_tree (ComponentGroup): Root-level component group for all components.
    """
    def __init__(
        self,
        base_name: str,
        friendly_name: Optional[str] = None,
        grpc_host: str = "[::]",
        grpc_port: int = 50051,
    ):
        """Creates Helios server instance.

        Args:
            base_name (str): Base name of the server.
            friendly_name (str): Print-friendly name of the server.
            grpc_host (int, optional): The address for which to open a port. Defaults to "[::]".
            grpc_port (int, optional): Integer port on which server will accept RPC requests. Defaults to 50051.
        """
        # Argument-defined attributes
        self.base_name: str = base_name
        self.friendly_name: str = friendly_name if friendly_name else base_name
        self.grpc_host: str = grpc_host
        self.grpc_port: int = grpc_port

        # Other attributes
        self.component_tree: ComponentGroup = ComponentGroup(self.base_name)

    def get_component_tree(self) -> ComponentGroup:
        """Returns the component tree.

        Returns:
            ComponentGroup: The base component group.
        """
        return self.component_tree

    def create_grpc_server(self) -> grpc.Server:
        """Generates a gRPC server instance.

        This method initializes the gRPC server and adds the protocol handler.

        Returns:
            grpc.Server: The gRPC server instance.
        """
        # Create the gRPC server instance
        grpc_server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))

        # Add protocol handler
        add_HeliosProtocolServicer_to_server(HeliosProtocol(), grpc_server)

        # Add the bind info
        grpc_server.add_insecure_port(f"{self.grpc_host}:{self.grpc_port}")

        return grpc_server

    def start(self):
        """Starts the Helios server.

        This creates, initializes, and starts the gRPC server,
        and recursively starts all components.

        This method will block until the gRPC server is terminated.
        """
        # Initialize the gRPC server
        self.grpc_server = self.create_grpc_server()

        # Start the gRPC server
        self.grpc_server.start()

        # Start the components
        self.component_tree.start_all(self)

        # Wait until gRPC server is terminated
        self.grpc_server.wait_for_termination()
