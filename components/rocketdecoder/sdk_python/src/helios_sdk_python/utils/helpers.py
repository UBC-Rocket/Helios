"""
Low-level framing helpers.

Wire format:  [ 4-byte big-endian length ][ protobuf bytes ]
"""

import struct
import socket
from google.protobuf.message import Message, DecodeError


_HEADER = struct.Struct("!I")  # network byte order (big-endian), 4-byte length
_MAX_MSG_SIZE = 2^30           # 1 GB


def _recvall(sock: socket.socket, n: int) -> bytes | None:
    """Read exactly *n* bytes from *sock*. Returns None if the connection closed."""
    buf = bytearray()
    while len(buf) < n:
        chunk = sock.recv(n - len(buf))
        if not chunk:
            return None
        buf.extend(chunk)
    return bytes(buf)


def send_message(sock: socket.socket, msg: Message) -> None:
    """
    Serialize *msg* and send it with a big-endian length prefix.

    Raises ValueError if the serialized message exceeds the max message size
    """
    data = msg.SerializeToString()
    if len(data) > _MAX_MSG_SIZE:
        raise ValueError(
            f"Serialized message is {len(data)} bytes, which exceeds the "
            f"4-byte header limit of {_MAX_MSG_SIZE} bytes."
        )
    sock.sendall(_HEADER.pack(len(data)) + data)


def recv_message(sock: socket.socket, msg_type: type[Message]) -> Message | None:
    """
    Receive one length-prefixed message from *sock*.

    Returns a parsed instance of *msg_type*, or None if the connection closed.
    Raises DecodeError if the bytes cannot be parsed.
    """
    raw_len = _recvall(sock, _HEADER.size)
    if raw_len is None:
        return None

    (msg_len,) = _HEADER.unpack(raw_len)
    data = _recvall(sock, msg_len)
    if data is None:
        return None

    msg = msg_type()
    msg.ParseFromString(data)
    return msg