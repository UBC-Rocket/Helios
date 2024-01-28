from helios.transports.transport import AbstractTransport

from multiprocessing.connection import Connection


class PipeTransport(AbstractTransport):
    def __init__(self, transmit: Connection, receive: Connection):
        self.transmit: Connection = transmit
        self.receive: Connection = receive
