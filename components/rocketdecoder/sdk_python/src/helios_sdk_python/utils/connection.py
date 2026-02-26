"""
ProtoConnection wraps a connected socket and provides typed send/recv methods.
"""

import socket
from typing import Callable
from google.protobuf.message import Message

from .helpers import send_message, recv_message


class ProtoConnection:
    """
    Wraps a connected TCP socket for protobuf messaging.
    """

    def __init__(self, sock: socket.socket):
        self._sock = sock

    def send(self, msg: Message) -> None:
        """Send a protobuf message."""
        send_message(self._sock, msg)

    def recv(self, msg_type: type[Message]) -> Message | None:
        """
        Receive and parse one protobuf message.

        Returns None if the peer closed the connection.
        """
        return recv_message(self._sock, msg_type)

    def recv_loop(
        self,
        msg_type: type[Message],
        handler: Callable[[Message, "ProtoConnection"], None],
    ) -> None:
        """
        Continuously receive messages and call *handler(msg, conn)* for each.
        Returns when the connection is closed or handler raises StopIteration.
        """
        while True:
            msg = self.recv(msg_type)
            if msg is None:
                break
            try:
                handler(msg, self)
            except StopIteration:
                break

    def close(self) -> None:
        self._sock.close()

    def __enter__(self):
        return self

    def __exit__(self, *_):
        self.close()