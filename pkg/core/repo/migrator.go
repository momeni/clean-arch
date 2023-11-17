// Copyright (c) 2024 Behnam Momeni
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package repo

import "context"

// Migrator of S is the core migration interface which defines how a
// semantically versioned resource may be migrated from a S1.S2.S3
// source version to a D1.D2.D3 destination version.
// The S type parameter represents the migration settler type of the
// versioned resource.
// For example, pkg/core/usecase/migrationuc.Settings interface can be
// used as S parameter for configuration settings migration and the
// SchemaSettler interface can be used as S for DB schema migration.
//
// Migration of a given resource consists of three main phases.
//  1. Source version of the given resource should be loaded. For simple
//     resources which can be held completely in memory, such as config
//     settings, this step can load and hold the source version resource
//     as an struct instance in memory. While for more complex resources
//     such as a DB schema, it may use a transaction in the target DB
//     in order to connect to and load the source database schema as
//     a Foreign Data Wrapper (FDW) without actually transferring data
//     items. Relevant extra information, such as an open transaction
//     in the destination database or source database connection URL or
//     a config file data may be fetched all while instantiating the
//     Migrator[S]. The Load function is responsible for this phase.
//  2. Comparing the source and destination semantic versions indicates
//     that an upwards or downwards migration is required. If those
//     versions are equal, either direction may be used. The UpMigrator
//     and DownMigrator methods create an upwards and downwards
//     migrator object with UpMigrator[S] and DownMigrator[S] types
//     respectively. For simple scenarios such as configuration files
//     where all minor versions of each major version can be loaded
//     similarly, these methods just have to create a migrator object.
//     However, for more complex scenarios such as DB schema migration
//     where each minor version requires a distinct schema definition
//     and loading logic, the UpMigrator and DownMigrator have to
//     migrate the source version to its latest supported minor version
//     (in the same major version) too. Thereafter, migrator objects
//     can be used to create new resources at the next/previous version
//     moving over one major version at a time. A DB schema migration
//     implementation may create new database views based on the views
//     or tables which were defined for the previous/next version, so
//     they can complete without transferring actual data rows. More
//     complex changes may mandate a migrator object to create a table
//     and materialize its changes or even load data rows in the Golang
//     process memory. Such details will remain transparent at interface
//     level.
//  3. After migrating a resource upwards/downwards over its major
//     version and obtaining a migrator object with the destination
//     major version (and its latest supported minor and patch versions)
//     using the UpMigrator[S] or DownMigrator[S] instances, the
//     Settler[S] interface (which is embedded by both of them) can be
//     used to obtain the migration settler instance. For simple
//     resources like configuration settings, the settler instance is
//     used to marshal and save the ultimate resource (or merge it with
//     the target settings version default values). For more complex
//     resources like schema migration which had proceeded the migration
//     without actually fetching and converting data items (but just
//     kept some metadata about them by creating a series of views),
//     the settler instance may be used to persist the migration results
//     by creating the ultimate resources (e.g., tables which their rows
//     are obtained by querying the previous/next version views).
//     For more details, check the SchemaSettler interface.
//
// By the way, Migrator[S] provides a Settler method, so the special
// scenario which has the same source and destination major version
// may be simplified and the settler instance can be fetched directly
// instead of taking an upwards/downwards migrator object and using it
// to fetch the settler instance afterwards, although, an implementation
// is free to employ this sequence of method calls (in order to realize
// the Settler[S] requirements). This Settler method, despite the
// generic Settler[S] interface, may return an error too because this
// Settler method may need to migrate from the current minor version to
// the latest supported minor version (as expected by the database
// schema migration settlers) and so may fail. However, the Settler[S]
// only has to create a settler object without attempting any real
// migration (which is the duty of the returned settler object itself).
type Migrator[S any] interface {
	// Settler creates and returns a settler object (with S type) which
	// can be used to settle the migration results.
	// For example, a configuration settler object may be used for
	// merging the default configuration settings or a database schema
	// settler object may be used in order to persist tables in a local
	// schema. This method is a shortcut when the source and destination
	// major versions are the same. It may migrate from the current
	// minor version to the latest supported minor version before
	// creating the settler object, hence, it may return an error
	// despite the Settler[S] interface which is embedded by the
	// UpMigrator[S] and the DownMigrator[S] interfaces and may not
	// return an error.
	Settler(ctx context.Context) (S, error)

	// Load tries to load the source version of the migrating resource
	// and return an error if it failed to load it completely.
	// The loading may be performed by creating an in-memory
	// representation of that resource in simple scenarios such as
	// loading a configuration file or may be performed by setting up
	// some metadata alone in order to make the migrating resource
	// accessible, such as a FDW provided schema for migrating a DB.
	// All extra information which may be required for finding out
	// the source version of the migrating resource should be kept by
	// the Migrator[S] instance since its instantiation time.
	// Calling Load method multiple times causes undefined behavior and
	// an implementation is free to ignore extra calls, load data again,
	// or return an error.
	Load(ctx context.Context) error

	// MajorVersion returns the major semantic version of this Migrator
	// instance. It may reflect the major version of a configuration
	// file or a database schema for example. This method may be used
	// for identification of the migration versions path, passing
	// through the major versions one by one.
	MajorVersion() uint

	// UpMigrator creates a new upwards migrator object. It may use a
	// version-specific struct for its implementation, however, it has
	// to adapt the created struct to the version-independent UpMigrator
	// interface, so it can be used uniformly in the use cases layer.
	// This upwards migrator may contain extra information such as a
	// loaded configuration settings struct or an open database
	// transaction as appropriate. Such information are transparent at
	// the interface level.
	//
	// Before calling UpMigrator, it is necessary to call Load method
	// in order to obtain a resource at its source major and minor
	// version (or at a more recent minor version in simplest cases).
	// After calling UpMigrator, the resource will be migrated to the
	// latest supported minor version within the source major version.
	// Obtained UpMigrator[S] instance may be used to keep migrating
	// to the next major versions (and their latest supported minor
	// versions) one at a time.
	UpMigrator(ctx context.Context) (UpMigrator[S], error)

	// DownMigrator creates a new downwards migrator object. It may use
	// a version-specific struct for its implementation, however, it has
	// to adapt the created struct to the version-independent
	// DownMigrator interface, so it can be used uniformly in the
	// use cases layer. This downwards migrator may contain extra
	// information such as a loaded configuration settings struct or
	// an open database transaction as appropriate. Such information
	// are transparent at the interface level.
	//
	// Before calling DownMigrator, it is necessary to call Load method
	// in order to obtain a resource at its source major and minor
	// version (or at a more recent minor version in simplest cases).
	// After calling DownMigrator, the resource will be migrated to the
	// latest supported minor version within the source major version.
	// Obtained DownMigrator[S] instance may be used to keep migrating
	// to the previous major versions (and their latest supported minor
	// versions) one at a time.
	DownMigrator(ctx context.Context) (DownMigrator[S], error)
}

// UpMigrator of S interface specifies the upwards migrator objects
// requirements for a resource with S settler type. It embeds the
// Settler[S] interface and has one main method, the MigrateUp method.
// The S type parameter which represents the migration settler type can
// be filled, for example, by pkg/core/usecase/migrationuc.Settings for
// configuration settings migration or the SchemaSettler interface from
// this package for database schema migration.
// The MigrateUp method may be used to migrate one major version upwards
// and obtain the next relevant UpMigrator[S] instance, while the
// Settler method from the Settler[S] interface may be used to obtain
// an instance of S in order to settle the migration operation at its
// ultimate major version. The actual logic of migration settlement
// depends on the S type and is not defined by UpMigrator[S] itself.
type UpMigrator[S any] interface {
	Settler[S]

	// MigrateUp creates the next/upper version of the migrating
	// resource either physically (e.g., by creating a new settings
	// struct instance in order to maintain settings items in memory)
	// or logically (e.g., by creating some views in the upper version
	// format based on queries over views or tables from the current
	// version in order to proceed with a database schema migration).
	// Finally, another UpMigrator[S] instance will be created and
	// returned which describes the migrating resource at its upper
	// version, so the migration may continue.
	// If there is no more upper major version, an error will be
	// returned.
	MigrateUp(ctx context.Context) (UpMigrator[S], error)
}

// DownMigrator of S interface specifies the downwards migrator objects
// requirements for a resource with S settler type. It embeds the
// Settler[S] interface and has one main method, the MigrateDown method.
// The S type parameter which represents the migration settler type can
// be filled, for example, by pkg/core/usecase/migrationuc.Settings for
// configuration settings migration or the SchemaSettler interface from
// this package for database schema migration.
// The MigrateDown method may be used to migrate one major version
// downwards and obtain the previous relevant DownMigrator[S] instance,
// while the Settler method from the Settler[S] interface may be used
// to obtain an instance of S in order to settle the migration
// operation at its ultimate major version. The actual logic of
// migration settlement depends on the S type and is not defined
// by DownMigrator[S] itself.
type DownMigrator[S any] interface {
	Settler[S]

	// MigrateDown creates the previous/downer version of the migrating
	// resource either physically (e.g., by creating a new settings
	// struct instance in order to maintain settings items in memory)
	// or logically (e.g., by creating some views in the downer version
	// format based on queries over views or tables from the current
	// version in order to proceed with a database schema migration).
	// Finally, another DownMigrator[S] instance will be created and
	// returned which describes the migrating resource at its downer
	// version, so the migration may continue.
	// If there is no more downer major version, an error will be
	// returned.
	MigrateDown(ctx context.Context) (DownMigrator[S], error)
}

// Settler of S interface specifies the migration settler objects
// requirements for a resource with S settler type. This interface
// has one method, namely Settler, and it returns an instance of S
// type, so it may be used to settle the migration operation.
// It should be called when the migrating resource has reached to its
// destination major version and needs to persist its migration results.
// The S type parameter which represents the migration settler type can
// be filled, for example, by pkg/core/usecase/migrationuc.Settings for
// configuration settings migration or the SchemaSettler interface from
// this package for database schema migration.
//
// For simple resources like configuration settings, the settler
// instance is used to marshal and save the ultimate resource (or
// merge it with the target settings version default values as seen in
// the MergeSettings method of pkg/core/usecase/migrationuc.Settings
// interface). For more complex resources like database schema migration
// which had proceeded the migration without actually fetching and
// converting data items (but just kept some metadata about them by
// creating a series of views), the settler instance may be used to
// persist the migration results by creating the ultimate resources
// (e.g., tables which their rows are obtained by querying the
// previous/next version views). For more details, check
// the SchemaSettler interface.
type Settler[S any] interface {
	// Settler method returns an instance of S type which represents
	// the migration settler instance. This method should be called
	// after reaching to the ultimate major version. The S instance
	// may be used to finalize migration operation and persist its
	// results.
	Settler() S
}
