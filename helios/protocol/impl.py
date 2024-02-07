import helios.protocol.generated.protocol_pb2 as protocol_pb2
import helios.protocol.generated.protocol_pb2_grpc as protocol_pb2_grpc

import grpc

class HeliosProtocol(protocol_pb2_grpc.HeliosProtocolServicer):
    def __init__(self):
        self.component_tree = None

    def InitialHandshake(self, request, context):
        print("Received InitialHandshake")
        return protocol_pb2.HandshakeResponse()
