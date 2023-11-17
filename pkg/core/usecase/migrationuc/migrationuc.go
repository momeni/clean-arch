// Copyright (c) 2024 Behnam Momeni
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Package migrationuc provides the database migration use cases.
// It exposes two main use cases, namely InitDBUseCase for initializing
// of database schema with initial sample data (for development or
// production environment) and MigrateDBUseCase for migrating from a
// source database to a destination database while converting the schema
// format from one version to another version (upwards or downwards).
// This package also exposes the Settings and SchemaSettings interfaces
// which represent the version-independent expectations from any
// configuration file representation type, so all configuration types
// with different formats may be taken similarly in the use cases layer
// for sake of seamless migration.
package migrationuc
