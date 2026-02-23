"""
HeliosClient — connects to Helios and exposes send/recv helpers.
"""

import socket
import time
import logging
from google.protobuf.message import Message
from generated.python.transport.packet_pb2 import TransportPacket

from .utils.connection import ProtoConnection

logger = logging.getLogger(__name__)


class HeliosClient:
    """
    TCP client for protobuf messaging.

    Parameters
    ----------
    host : str
        Server hostname or IP (Default HELIOS).
    port : int
        Server port (Default 5000).
    timeout : float | None
        Socket timeout in seconds (None = blocking).
    retry : int
        Number of connection attempts before raising (default 999, 1 = no retry).
    retry_delay : float
        Seconds to wait between retries (default 1.0).
    """

    def __init__(
        self,
        host: str = "Helios",
        port: int = 5000,
        timeout: float | None = None,
        retry: int = 999,
        retry_delay: float = 1.0,
    ):
        self.host = host
        self.port = port
        self.timeout = timeout
        self.retry = retry
        self.retry_delay = retry_delay

        self._conn: ProtoConnection | None = None

    def connect(self) -> "HeliosClient":
        """Open the TCP connection. Returns self for chaining."""
        last_err = None
        for attempt in range(1, self.retry + 1):
            try:
                sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
                sock.settimeout(self.timeout)
                sock.connect((self.host, self.port))
                self._conn = ProtoConnection(sock)
                logger.info("Connected to %s:%s", self.host, self.port)
                return self
            except OSError as e:
                last_err = e
                logger.warning(
                    "Connection attempt %d/%d failed: %s", attempt, self.retry, e
                )
                if attempt < self.retry:
                    time.sleep(self.retry_delay)

        raise ConnectionError(
            f"Could not connect to {self.host}:{self.port} "
            f"after {self.retry} attempt(s)"
        ) from last_err

    def disconnect(self) -> None:
        if self._conn:
            self._conn.close()
            self._conn = None
            logger.info("Disconnected from %s:%s", self.host, self.port)

    def send(self, msg: TransportPacket) -> None:
        """Send a protobuf message to the server."""
        self._require_connected()
        self._conn.send(msg)

    def recv(self, msg_type: type[TransportPacket]) -> TransportPacket | None:
        """Receive one protobuf message from the server."""
        self._require_connected()
        return self._conn.recv(msg_type)

    def send_recv(self, msg: TransportPacket, reply_type: type[TransportPacket]) -> TransportPacket | None:
        """Send a message and immediately wait for a reply."""
        self.send(msg)
        return self.recv(reply_type)

    def __enter__(self):
        return self.connect()

    def __exit__(self, *_):
        self.disconnect()

    def _require_connected(self) -> None:
        if self._conn is None:
            raise RuntimeError("Not connected. Call connect() or use as a context manager.")