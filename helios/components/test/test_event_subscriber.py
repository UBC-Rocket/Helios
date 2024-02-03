from ..base import ComponentBase


class TestEventSubscriber(ComponentBase):
    def __init__(self, source: tuple[str, ...]):
        super().__init__(source)

    def run(self, source: tuple[str, ...]):
        print(f"Hello TestEventSubscriber from {self.path}! -> I am subscribing to: {source}")
