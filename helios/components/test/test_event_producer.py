from ..base import ComponentBase


class TestEventProducer(ComponentBase):
    def __init__(self, timeout: int = 0):
        super().__init__()

    def run(self, timeout: int = 0):
        print(f"Hello TestEventProducer from {self.path}!")
