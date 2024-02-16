# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).


## [1.1.0] - 2024-02-16

### Added

- Support versioning of configuration files and database schema
- Demonstrate a multi-database atomic migration scheme with up/down migrations


## [1.0.0] - 2023-10-19

### Added

- Demonstrate the Clean Architecture with a cars riding/parking example
- Support Gin Gonic framework for the REST APIs implementation in adapters layer
- Support GORM (+pgx) for interaction with a PostgreSQL DBMS server in adapters layer
- Support parsing a yaml configuration file and instantiating other components based on the configuration settings from the adapters layer
- Add Makefile targets for linting the project with staticcheck and reive
- Add integration tests using podman-based postgres containers
- Document Clean Architecture layers in README and main ideas as godocs
