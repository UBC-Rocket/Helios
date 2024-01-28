from helios.components.component import AbstractComponent


class TestEventProducer(AbstractComponent):
    def __init__(self, timeout: int = 0):
        super().__init__()

    def run(self, timeout: int = 0):
        print(f"Hello TestEventProducer from {self.path}!")


class TestEventConsumer(AbstractComponent):
    def __init__(self, timeout: int = 0):
        super().__init__()

    def run(self, timeout: int = 0):
        print(f"Hello TestEventConsumer from {self.path}!")


class TestEventSubscriber(AbstractComponent):
    def __init__(self, source: TestEventProducer):
        super().__init__(source)

    def run(self, source: tuple[str, ...]):
        print(f"Hello TestEventSubscriber from {self.path}! -> I am subscribing to: {source}")
