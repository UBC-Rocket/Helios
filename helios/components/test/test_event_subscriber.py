from ..base import ComponentBase


class TestEventSubscriber(ComponentBase):
    def __init__(self, source: str):
        super().__init__(source)

    def run(self, source: str):
        super().run()
        print(f"Hello TestEventSubscriber from {self.path}! -> I am subscribing to: {source}")
