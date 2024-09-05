# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).


## [1.3.0] - 2024-09-05.

### Added

- Accept minimum/maximum bounds for configuration settings from the configuration file
- Store configuration settings minimum/maximum bounds in the database, so other updaters may respect them
- Respect the target version settings bounds during a migration operation
- Report configuration settings bounds whenever the settings are reported themselves
- Demonstrate how the standard `log/slog` structured logging may be used

### Changed

- Avoid updating settings in the database if they cannot be used for instantiation of new use cases


## [1.2.0] - 2024-03-07

### Added

- Maintain mutable settings in the database, so configuration file settings may be overridden dynamically
- Add two REST APIs for querying visible (including immutable) settings and updating the mutable (including visible, but not immutable) settings
- Support in-database mutable settings during the multi-database migration operation
- Reload instantiated use case objects whenever the mutable settings are updated
- Preserve comments in the configuration YAML files during the migration operation

### Fixed

- Return a bool ok flag from the DserXReq methods


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
