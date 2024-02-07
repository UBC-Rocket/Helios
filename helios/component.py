from __future__ import annotations

from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from helios.helios import Helios
    from helios.components.base import ComponentBase

from abc import ABC, abstractmethod
from typing import Optional
from typing_extensions import override
from multiprocessing import Process


class AbstractComponent(ABC):
    """Abstract class for a component.

    This is a common interface for all components in Helios, including
    component groups and component managers.

    Component groups contain zero or more other components, and are used to
    organize components into a tree structure. Component managers are used to
    manage the lifecycle of a component, including starting, stopping, and
    killing the component.

    Attributes:
        name (str): The name of the component.
        parent (ComponentGroup, optional): The parent of the current component.
            Defaults to None. If None, there is not parent
    """

    def __init__(self, name: str, parent: Optional[ComponentGroup] = None):
        """Creates a simple component object.

        Args:
            name (str): The name of the component.
            parent (ComponentGroup, optional): The parent of the current component.
                Defaults to None. If None, there is not parent
        """
        self.name: str = name
        self.parent: Optional[ComponentGroup] = parent

    @abstractmethod
    def get_path(self) -> str:
        """Returns the full path of the component in string form.

        For example, if the component is named "test" and is a child of a component named "parent",
        the return value would be "parent.test". If the component has no parent, the return value would be "test".

        The path will contain all parent components all the way up to the root component,
        or the first component that has no parent.

        Returns:
            str: The full path of the component
        """
        raise NotImplementedError()

    @abstractmethod
    def print_tree(self, last=True, header=''):
        """Prints a formatted tree of the component and its children.

        An example output is:
        └──[basic_test] (4 children)
            ├──<test_producer: self.running=False>
            │  └──(<helios.component.PyComponent object at 0x7fec375e3cd0>)
            ├──<test_consumer: self.running=False>
            │  └──(<helios.component.PyComponent object at 0x7fec375e3d90>)
            ├──<test_subscriber: self.running=False>
            │  └──(<helios.component.PyComponent object at 0x7fec375e3e50>)
            └──[test_group] (2 children)
                ├──<test_producer_two: self.running=False>
                │  └──(<helios.component.PyComponent object at 0x7fec375e3f70>)
                └──<test_consumer_two: self.running=False>
                    └──(<helios.component.PyComponent object at 0x7fec37638070>)

        Args:
            last (bool, optional): Is this the last row in the print? Defaults to True.
            header (str, optional): Header to print before formatting. Defaults to ''.
        """
        raise NotImplementedError()


class ComponentGroup(AbstractComponent):
    """Component group class.

    This class is used to organize components into a tree structure. It can contain
    zero or more other components, including other component groups and component managers.

    Attributes:
        components (dict[str, AbstractComponent]): A dictionary of components, where the key is the name of the component
            and the value is the component object.
    """

    @override
    def __init__(self, name: str, parent: Optional[ComponentGroup] = None):
        super().__init__(name, parent)
        self.components: dict[str, AbstractComponent] = {}

    def create_component_group(self, name: str) -> ComponentGroup:
        """Creates a new component group and adds it to the components dictionary.

        Args:
            name (str): Name of the component group to create.

        Raises:
            ValueError: If a component with the same name already exists.

        Returns:
            ComponentGroup: The newly created component group.
        """
        # Check if component with name already exists
        if name in self.components:
            raise ValueError(f"Component with name {name} already exists")

        # Create the component group and add it to the components dictionary
        group: ComponentGroup = ComponentGroup(name, parent=self)
        self.components[name] = group
        return group

    def add_component(self, component: AbstractComponent):
        """Adds a component to the component group.

        Args:
            component (AbstractComponent): The component to add.

        Raises:
            ValueError: If a component with the same name already exists.
            ValueError: If the component is not a valid component.
        """
        # Check if component with name already exists
        if component.name in self.components:
            raise ValueError(f"Component with name {component.name} in group {self.get_path()} already exists")

        # Check what type of component it is
        if isinstance(component, ComponentGroup):
            # If it's a component group, add it to the components dictionary
            self.components[component.name] = component
        elif isinstance(component, AbstractComponentManager):
            # If it's a component manager, add it to the components dictionary
            # and set the parent of the component to the current component group
            self.components[component.name] = component
            component.parent = self
        else:
            # This should not happen, but raise an error if the component is not valid
            raise ValueError(f"Component with name {component.name} is not a valid component")

    def start_all(self, server: Helios):
        """Recursively starts all components in the component group.

        Args:
            server (Helios): The Helios server instance. Used to pass to the start method of the components.

        Raises:
            ValueError: If the component is not a valid component.
        """
        # Loop through all components and start them
        for c in self.components.values():
            if isinstance(c, AbstractComponentManager):
                # If the component is a component manager, start it
                c.start(server)
            elif isinstance(c, ComponentGroup):
                # If the component is a component group, recursively start all of its children
                c.start_all(server)
            else:
                # This should not happen, but raise an error if the component is not valid
                raise ValueError(f"Component {c.get_path()} is not a valid component")

    @override
    def get_path(self) -> str:
        return f"{self.parent.get_path()}.{self.name}" if self.parent else self.name

    @override
    def print_tree(self, last=True, header=''):
        if not self.parent:
            print("[HELIOS CORE]")
        children = list(self.components.values())
        elbow = "└──"
        pipe = "│  "
        tee = "├──"
        blank = "   "
        print(f"{header}{(elbow if last else tee)}[{self.name}] ({len(children)} children)")
        for i, c in enumerate(children):
            c.print_tree(header=header + (blank if last else pipe), last=i == len(children) - 1)


class AbstractComponentManager(AbstractComponent, ABC):
    """Abstract class for a component manager.

    This class is used to manage the lifecycle of a component, including starting, stopping, and killing the component.

    Attributes:
        running (bool): Is the component currently running?
    """
    @override
    def __init__(self, name: str, parent: Optional[ComponentGroup] = None):
        super().__init__(name, parent)
        self.running: bool = False

    @abstractmethod
    def start(self, server: Helios):
        """Starts the component.

        Args:
            server (Helios): The Helios server instance. Used to pass to the start method of the components.

        Raises:
            ValueError: If the component is already running.
        """
        raise NotImplementedError()

    @abstractmethod
    def stop(self):
        """Stops the component.

        Raises:
            ValueError: If the component is not running.
        """
        raise NotImplementedError()

    @abstractmethod
    def kill(self):
        """Kills the component.

        Raises:
            ValueError: If the component is not running.
        """
        raise NotImplementedError()


class PyComponent(AbstractComponentManager):
    def __init__(self, name: str, component_object: ComponentBase, parent: Optional[ComponentGroup] = None):
        super().__init__(name, parent)

        self.component_object: ComponentBase = component_object
        self.process: Optional[Process] = None

    @override
    def start(self, server: Helios):
        # Check if component is already running
        if self.running:
            raise ValueError(f"Component {self.get_path()} is already running")

        # Populate the name, path, and gRPC information for the component
        self.component_object.init_component(
            name=self.name,
            path=self.get_path(),
            grpc_host=server.grpc_host,
            grpc_port=server.grpc_port
        )

        # Create the python process
        self.process = Process(
            target=self.component_object.run,
            args=self.component_object.launch_args,
            kwargs=self.component_object.launch_kwargs,
            daemon=True
        )

        # Start the process
        self.running = True
        self.process.start()

    @override
    def stop(self, server: Helios):
        raise NotImplementedError()

    @override
    def kill(self, server: Helios):
        raise NotImplementedError()

    @override
    def get_path(self) -> str:
        return f"{self.parent.get_path()}.{self.name}" if self.parent else self.name

    @override
    def print_tree(self, last=True, header=''):
        elbow = "└──"
        pipe = "│  "
        tee = "├──"
        blank = "   "
        print(f"{header}{(elbow if last else tee)}<{self.name}: {self.running=}>")
        print(f"{header + (blank if last else pipe)}{elbow}({self})")
