import helios.protocol.generated.protocol_pb2 as protocol_pb2
import helios.protocol.generated.protocol_pb2_grpc as protocol_pb2_grpc


class HeliosProtocol(protocol_pb2_grpc.HeliosProtocolServicer):
    def __init__(self):
        self.component_tree = None

    def initial_handshake(self, request, context):
        print("Received InitialHandshake")
        return protocol_pb2.HandshakeResponse(success=True)
