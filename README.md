# Clean Architecture

[![Go Report Card](https://goreportcard.com/badge/github.com/momeni/clean-arch)](https://goreportcard.com/report/github.com/momeni/clean-arch)
[![Go Reference](https://pkg.go.dev/badge/github.com/momeni/clean-arch.svg)](https://pkg.go.dev/github.com/momeni/clean-arch)
[![Release](https://img.shields.io/github/release/momeni/clean-arch.svg)](https://github.com/momeni/clean-arch/releases/latest)
[![License: MPL 2.0](https://img.shields.io/badge/License-MPL_2.0-brightgreen.svg)](http://mozilla.org/MPL/2.0/)

This project demonstrates an example Go (Golang) realization of
the _Clean Architecture_ using the [Gin Gonic](https://gin-gonic.com/docs/)
web framework. The [GORM](https://gorm.io/docs/) library and a
[PostgreSQL](https://www.postgresql.org/docs/current/) DBMS are used for
data persistence and [podman](https://github.com/containers/podman) based
testing codes are provided too.

Four main layers are seen as depicted in the following component diagram.
These layers are usually drawn as
[co-centered](https://miro.medium.com/v2/resize:fit:1280/1*yTDpfIqqAdeKRhbHwfhrYQ.png)
[circles](https://miro.medium.com/v2/resize:fit:3136/1*43uKsXfq35PrJKJTBq1MJw.png).
The inner-most layer is the **Entities** layer which is also known as
the *Models* or *Domain* layer. This layer contains the entity types
which are defined in the business domain and may be used by a high-level
implementation of interesting use cases, without technology-specific
implementation details. These models are expected to have the least
foreseeable changes since their changes are likely to propagate to
other layers. As an example entity in this project, we can mention the
[model.Car](pkg/core/model/car.go) type.
These models may be used by both of the **Use Cases** and **Adapters**
layers. The standard language libraries are treated similarly too.

The **Use Cases** or **Use Case Interactors** layer contains the core
business logic of a system. This code may use the defined models for
different purposes such as data passing, requests/responses formatting,
or data persistence. This layer can be seen as a facade for the entire
system too as it defines the acceptable public use cases of a system.
The high-level implementation of this layer makes it relatively stable
as it only depends on the domain entities and does not need to change
whenever a third-party library is updated or when a new technology is
supposed to be adapted. As an example type from this layer, take a
look at the [carsuc.UseCase](pkg/core/usecase/carsuc/carsuc.go) type.

![Clean Architecture Layers](assets/uml/clean-arch.components-diagram.png "Four Layers of the Clean Architecture")

A concrete method is required in order to wire up the use cases of a
system to the use cases of other systems, allowing them to interact
based on their expected and exposed public interfaces.
For example, a web framework such as [Gin Gonic](https://gin-gonic.com/docs/)
requires a series of [HandlerFunc](https://pkg.go.dev/github.com/gin-gonic/gin#HandlerFunc)
functions in order to call them when a web request is recevied which
may be different from how they are managed in [Echo](https://echo.labstack.com/docs/request).
Conversely, use cases require to employ functionalities of an ORM like
the [GORM](https://gorm.io/docs/) in order to store or search among
models, having a database management system server.
The **Adapters** layer which is also know as *Controllers* or *Gateways*
is responsible to fill these gaps without making the frameworks and
third-party libraries on one hand and the use case interactors on the
other hand dependent on each other.
The use cases layer contains [repo.Cars](pkg/core/repo/carsrp.go) example
interface in order to show its expectations from a cars repository. This
interface is realized by [carsrp.Repo](pkg/adapter/db/postgres/carsrp/repo.go)
from the adapters layer, where it uses [GORM](https://gorm.io/docs/) and
[Pgx](https://github.com/jackc/pgx) libraries for interaction with a
[PostgreSQL](https://www.postgresql.org/docs/current/) DBMS server.
The adapters layer depends on both of our high-level logics and other
libraries provided/required interfaces in order to adapt them together.
To be precise, the *Controllers* is not a suitable alias for this layer
because adapters should be thin and focus on converting interfaces or
simple serialization/deserialization tasks, while the use cases layer
contains the more complex business-level flow controls.
Parsing configuration such as [config.Config](pkg/adapter/config/config.go)
and instantiating components from other layers during the system startup
are also a part of this layer.

The outmost layer contains **Frameworks** and *Libraries* which are
usually implemented independent of the main project. Their codes are
also independent of our system and may change from time to time, or we
may need to replace them with alternatives as new technologies are
introduced. Because it is desired to keep our codes immune to changes
in the APIs of those libraries, adapters layer has to hide their details
by realizing the use cases layers interfaces and implementing any
missing functionality or mismatched expectations.
Indeed, these frameworks are parallels to our use cases layer, but in
their own project.

## Read More

For further reading, you may check

  * Martin, R. C., Grenning J., & Brown S. (2017). [Clean architecture](https://plefebvre91.github.io/resources/clean-architecture.pdf).
  * Lano, K., & Yassipour Tehrani, S. (2023). [Introduction to Clean Architecture Concepts](https://link.springer.com/chapter/10.1007/978-3-031-44143-1_2). In Introduction to Software Architecture: Innovative Design using Clean Architecture and Model-Driven Engineering (pp. 35-49). Cham: Springer Nature Switzerland.
