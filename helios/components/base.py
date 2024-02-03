from __future__ import annotations

from abc import ABC
from typing import Any


class ComponentBase(ABC):
    def __init__(self, *args, **kwargs):
        self.initialized: bool = False
        self.launch_args: tuple[Any, ...] = args
        self.launch_kwargs: dict[str, Any] = kwargs

    def initComponent(self, name: str, path: tuple[str, ...]):
        self.name: str = name
        self.path: tuple[str, ...] = path
        self.initialized = True

    def run(self):
        raise NotImplementedError()
