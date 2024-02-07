from __future__ import annotations

import helios.protocol.generated.protocol_pb2 as protocol_pb2
import helios.protocol.generated.protocol_pb2_grpc as protocol_pb2_grpc

import grpc
from abc import ABC
from typing import Any


class ComponentBase(ABC):
    def __init__(self, *args, **kwargs):
        self.initialized: bool = False
        self.launch_args: tuple[Any, ...] = args
        self.launch_kwargs: dict[str, Any] = kwargs

    def init_component(self, name: str, path: str, grpc_host: str, grpc_port: int):
        self.name: str = name
        self.path: str = path
        self.grpc_host: str = grpc_host
        self.grpc_port: int = grpc_port
        self.initialized = True

    def run(self):
        # Check if the component has been initialized
        if not self.initialized:
            raise ValueError("Component has not been initialized")

        # Connect to the gRPC server
        self.channel = grpc.insecure_channel(f"{self.grpc_host}:{self.grpc_port}")

        # Send Handshake
        self.stub = protocol_pb2_grpc.HeliosProtocolStub(self.channel)
        response = self.stub.initial_handshake(protocol_pb2.HandshakeRequest())
        print(f"Handshake response: {response.success}")
