import time
import sdk_python.src.helios_sdk_python as helios
from generated.python.transport.packet_pb2 import TransportPacket
from google.protobuf.timestamp_pb2 import Timestamp

print("Hello from rocketdecoder container!")

with helios.HeliosClient(retry=999, retry_delay=5.0) as client:
    print("Connected to Helios!")

    packet = TransportPacket(
        id=1,
        timestamp=Timestamp(seconds=int(time.time())),
        address="rocketdecoder",
        data=b"my raw payload bytes",
    )
    reply = client.send_recv(packet, TransportPacket)
    print(f"rocketdecoder - Received from server: {reply.data}")

    while True:
        msg = client.recv(TransportPacket)
        if msg is None:
            print("Connection closed by peer.")
            break

        print(f"rocketdecoder - Received: {msg.data}")
        time.sleep(5)