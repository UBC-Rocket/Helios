from helios.helios import Helios

from helios.transports.pipe import PipeTransport
from helios.orchestrators.mp import MultiprocessingOrchestrator

from helios.components.test.test_component import TestEventProducer, TestEventConsumer, TestEventSubscriber


def main():
    # Create Helios base
    helios = Helios(
        "basic_test",
        friendly_name="Basic Test Profile",
        transport=PipeTransport,
        orchestrator=MultiprocessingOrchestrator
    )

    # Build component tree
    rocket = helios.get_component_tree()

    rocket.add_component("test_producer", event_producer := TestEventProducer())
    rocket.add_component("test_consumer", TestEventConsumer())
    rocket.add_component("test_subscriber", TestEventSubscriber(source=event_producer))

    group = rocket.create_component_group("test_group")
    group.add_component("test_producer_two", TestEventProducer(timeout=10))
    group.add_component("test_consumer_two", TestEventConsumer(timeout=20))

    # Print the component tree
    print()
    rocket.print_tree()
    print()

    # Start Helios
    helios.start()


if __name__ == "__main__":
    main()
