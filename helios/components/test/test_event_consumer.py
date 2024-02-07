from ..base import ComponentBase


class TestEventConsumer(ComponentBase):
    def __init__(self, timeout: int = 0):
        super().__init__()

    def run(self, timeout: int = 0):
        super().run()
        print(f"Hello TestEventConsumer from {self.path}!")
