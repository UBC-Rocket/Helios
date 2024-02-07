import helios.fix_imports as import_fix  # noqa # Must be the first import!
from helios.helios import Helios
from helios.component import PyComponent
from helios.components.test.test_event_consumer import TestEventConsumer
from helios.components.test.test_event_producer import TestEventProducer
from helios.components.test.test_event_subscriber import TestEventSubscriber


def main():
    """Example of a basic Helios profile.

    This profile creates a simple component tree with a few test components and starts the Helios instance.
    """
    # Create Helios base
    helios = Helios(
        "basic_test",
        friendly_name="Basic Test Profile"
    )

    # Build component tree
    rocket = helios.get_component_tree()

    # Add test components to the component tree
    rocket.add_component(event_producer := PyComponent("test_producer", TestEventProducer()))
    rocket.add_component(PyComponent("test_consumer", TestEventConsumer()))
    rocket.add_component(PyComponent("test_subscriber", TestEventSubscriber(source=event_producer.get_path())))

    # Create a component sub-group and add two test components to it
    group = rocket.create_component_group("test_group")
    group.add_component(PyComponent("test_producer_two", TestEventProducer(timeout=10)))
    group.add_component(PyComponent("test_consumer_two", TestEventConsumer(timeout=20)))

    # Print the component tree
    print()
    rocket.print_tree()
    print()

    # Start Helios
    helios.start()


# Run the main function if this script is executed
if __name__ == "__main__":
    main()
