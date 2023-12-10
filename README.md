![Helios Logo](assets/helios_dark.svg)  
# Helios - UBCRocket's Ground Support Station

Helios is UBCRocket's Experimental Next-Generation Groundstation & Ground Support System. Built with a  modular, event-driven, and microservice-based architecture, Helios aims to have an all-in-one logging, visualization, and mission control software for all of UBC Rocket's projects. 

> WARNING: This document is WIP as the Helios proof-of-concept is still being developed! 

## Table of Contents
- [Installation](#installation)
- [Architecture](#architecture)
- [Motivation](#motivation)
- [Mission Goals](#mission-goals)
- [Developing](#developing)

## Installation
> This serves as a quickstart guide for installing Helios. For more detailed instructions, see the [INSTALL.md](INSTALL.md) file.

Installation TODO...

## Architecture
![Helios High Level Architecture Diagram](meta/architecture.svg)
### Read more on Helios' architecture in the [ARCHITECTURE.md](ARCHITECTURE.md) file.

## Motivation
The idea of a fully modular and event-driven ground support system emerged when improving and building upon the [existing ground station project](https://github.com/UBC-Rocket/UBCRocketGroundStation). Because of its monolithic architecture, the existing system made adding new components and debugging and testing individual components difficult. Adding additional features, such as support for multiple stages, COTS sensors like APRS, or building additional UI features like mapping or graphing, required a lot of hacks and workarounds to the existing system. The systems were also highly coupled, meaning that a failure in any of the sub-components would result in the failure of the entire system.

Additionally, the existing system heavily depended on QT (and subsequently, PyQt). This dependency made it challenging to run the system on headless systems, such as a Raspberry Pi, or systems without a display, such as on a server. Since the UI is tightly coupled with the backend, it increased developmental and testing complexity. As a result of the tight coupling, it was difficult to test the UI and the backend separately. It also meant that any UI bugs or crashes would result in the entire system crashing and losing valuable data.

Helios aims to solve these problems through a modular, event-driven, microservice-based architecture. By decoupling all the individual components of the ground station, we aim to achieve greater flexibility, modularity, and reliability for all of UBCRocket's needs. Refer to [Mission Goals](#mission-goals) for more information.

## Mission Goals
Helios aims to provide a flexible, modular, and reliable ground support system for all of UBC Rocket's projects. The system is designed to process and ingest data from any data source, whether from SRAD radios, off-the-shelf components, or simulations. With this information, Helios can display and render this data on any device, anywhere - meaning that sub-teams from halfway across the world can monitor and fetch real-time telemetry from the rocket on the pad.

### > Flexibility
The central philosophy behind Helios is not to make any assumptions about anything. The modular design allows for any sounding rocket configuration on any communication protocol, size, or fuel type. Helios' modular design is also flexible enough to extend to work with other use cases like satellites, propulsion testing, rovers, F1 cars, or even planes.

<!-- ![Profiles Diagram](meta/profiles.svg) -->

### > Modularity
Helios employs a modular architecture where processes function independently from each other. Event emitters are oblivious to the existence of event consumers, and event consumers are oblivious to the existence of event emitters. This isolation allows swapping components easily, so long as the new module sends or receives identical events. 

Helios' modularity allows for similar components across multiple profiles, speeding up R&D and development time. The modular architecture also enables runtime hot-swapping of individual components, meaning teams can enable and disable features on the fly without having to restart the entire system. Such modularity for what teams can build and do with Helios is what makes this system so powerful. 

### > Reliability
The goal of helios is to have a fault-tolerant mechanism for collecting telemetry from over two dozen individual sensors and to handle the loss of any of those processes without losing the entire mission. With Helios' modular design, each process is independent of the others and can be started, stopped, and restarted without affecting the rest of the system.

![Helios Fault Tolerance Diagram](meta/fault_tolerance.svg)

Each component/subsystem can be individually validated and tested with a modular component-based architecture. This level of decoupling between the separate components ensures greater reliability during mission-critical operations and easier debugging and testing during development. 

![Component Testing Diagram](meta/component_testing.svg)

## Developing
> This serves as a quickstart guide for developing Helios. For more detailed instructions, see the [DEVELOPMENT.md](DEVELOPMENT.md) file.

Development TODO...
