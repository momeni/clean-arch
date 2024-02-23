// Copyright (c) 2024 Behnam Momeni
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package repo

import "context"

// SchemaSettler interface specifies the expectations from the settler
// objects for a database schema migration operation.
//
// An efficient implementation of database migration may use views
// in order to convert data format without copying data items themselves
// as far as possible. For this purpose, migration may begin by creating
// a Foreign Data Wrapper (FDW) link to the source database from within
// a transaction of the destination database. This link allows source
// tables to be accessible like a local database schema within the
// destination database. For migrating to the latest minor version from
// the same major version, new views may be created in another schema
// by queries over the FDW imported schema's foreign tables. Thereafter,
// views may be migrated to one upper/downer version by creating a set
// of other views in another schema again and again, changing column
// names and/or computing them based on other views. If some conversions
// are more complex, they may need to create intermediate tables or
// even need to load data and process them in the Golang process.
// Finally, views (or tables) in an ultimate local schema must be
// materialized into a series of persistent tables in the expected
// schema. And this is the point that SchemaSettler interface becomes
// relevant. The SettleSchema method is supposed to perform this
// final database schema settlement operation.
type SchemaSettler interface {
	// SettingsPersister interface indicates that each SchemaSettler,
	// in addition to the database schema settlement, can persist the
	// independently migrated mutable settings in the database.
	SettingsPersister

	// SettleSchema settles the database schema migration operation.
	//
	// If the migration operation was implemented by creating a set of
	// views, the settlement can be performed by creating the expected
	// database tables and filling them by querying their corresponding
	// views.
	// In absence of errors, settlement persists the migration operation
	// results logically. However, if the entire migration was performed
	// in a transaction, caller is responsible to commit that
	// transaction (not the SchemaSettler interface).
	SettleSchema(ctx context.Context) error

	// MajorVersion returns the major semantic version of this schema
	// settler instance. Each schema settler supports exactly one major
	// version which is also included in its corresponding schema name.
	// For example, caweb1 schema name may be filled by major version 1
	// schema settler.
	MajorVersion() uint
}

// SettingsPersister interface specifies that how mutable settings may
// be persisted in a database, after being serialized as a byte slice.
// Each instance of this interface shall embed the relevant database
// transaction instance, so its lifetime is entangled with a single
// migration process (just like the SchemaSettler and other migration
// objects), allowing them to store migration process states (if any).
//
// When a multi-database migration operation is carried, an instance
// of SchemaSettler will be obtained in the last phase which can be
// used in order to persist all migrated tables contents.
// The SettingsPersister interface is embedded by SchemaSettler, so it
// can also persist the migrated mutable settings in the same
// transaction. When a uni-database migration operation is carried,
// that is, only the configuration file format is being changed, then
// the migrationuc.SchemaSettings.SettingsPersister method may be used
// in order to obtain a SettingsPersister for storing the settings in
// the src/dst database.
type SettingsPersister interface {
	// PersistSettings persists the given mutableSettings byte slice as
	// the serialized form of the system mutable configuration settings,
	// using the transaction which is hold by this interface and may be
	// used for persistence of other tables contents too. The
	// persistence applies whenever the caller commits its transaction.
	PersistSettings(ctx context.Context, mutableSettings []byte) error
}

// SchemaInitializer interface is exposed by each schema version
// implementation. It provides two methods of InitDevSchema and
// InitProdSchema in order to create new tables and fill an existing
// schema with them, using the development and production suitable
// initial data rows respectively.
// Each implementation (for settlement of the latest minor version of
// a specific major version) should contain the relevant information
// for finding the destination database (such as a database transaction)
// so the SchemaInitializer does not need to take any argument.
type SchemaInitializer interface {
	// InitDevSchema creates tables in an existing database schema
	// and fills them with the development suitable initial data.
	// The database connection, target schema name, tables format, and
	// their semantic version are known since the SchemaInitializer
	// interface instantiation time.
	InitDevSchema(ctx context.Context) error

	// InitProdSchema creates tables in an existing database schema
	// and fills them with the production suitable initial data.
	// The database connection, target schema name, tables format, and
	// their semantic version are known since the SchemaInitializer
	// interface instantiation time.
	InitProdSchema(ctx context.Context) error
}

// Schema interface presents expectations from a repository which allows
// database schema and roles management. This repository creates schema
// and grant relevant privileges on them, so they may be filled by
// tables during a migration or queried during other use cases.
type Schema interface {
	// Conn takes a Conn interface instance, unwraps it as required,
	// and returns a SchemaConnQueryer interface which (with access to
	// the implementation-dependent connection object) can create or
	// drop schema or manage database roles.
	Conn(Conn) SchemaConnQueryer

	// Tx takes a Tx interface instance, unwraps it as required,
	// and returns a SchemaTxQueryer interface which (with access to the
	// implementation-dependent transaction object) can manage database
	// roles, change their passwords, or perform schema-level management
	// operations.
	Tx(Tx) SchemaTxQueryer
}

// SchemaConnQueryer interface lists all operations which may be taken
// with regards to database schema having an open connection with the
// auto-committed transactions.
// Those operations which must be executed in a connection (and may not
// be executed in an ongoing transaction which may keep running other
// statements after this one) must be listed here, while other
// operations which do not strictly require an open connection (and may
// use an open transaction too) must be defined in the embedded
// SchemaQueryer interface. This design allows a unified implementation,
// while forcing developers to think about the consequences of having
// one or multiple transactions.
type SchemaConnQueryer interface {
	SchemaQueryer

	// InstallFDWExtensionIfMissing creates the postgres_fdw extension
	// assuming that its relavant .so files are available in proper
	// paths. If the extension is already created, calling this method
	// causes no change.
	InstallFDWExtensionIfMissing(ctx context.Context) error

	// DropServerIfExists drops the `serverName` foreign server, if it
	// exists, with cascade. That is, dependent objects such as its
	// user mapping will be dropped too.
	//
	// Caller is responsible to pass a trusted serverName string.
	DropServerIfExists(ctx context.Context, serverName string) error
}

// SchemaTxQueryer interface lists all operations which may be taken
// with regards to database schema having an ongoing transaction.
// Those operations which must be executed in a transaction (and may not
// be executed with a connection) must be listed here, while other
// operations which do not strictly require an open transaction (and
// can use their own auto-committed transaction too) must be defined
// in the embedded SchemaQueryer interface. This design allows a unified
// implementation, while forcing developers to think about the
// consequences of having one or multiple transactions.
type SchemaTxQueryer interface {
	SchemaQueryer

	// ChangePasswords updates the passwords of the given roles
	// in the current transaction. The roles and passwords slices must
	// have the same number of entries, so they can be used in pair.
	// These fields are not combined as a struct with two role and
	// password fields because passing items separately ensures that
	// all items are initialized explicitly in constrast to a struct
	// which its fields can be zero-initialized and are more suitable
	// to pass a set of optional fields.
	// The given roles may be suffixed automatically too, based on
	// this transaction queryer settings.
	ChangePasswords(
		ctx context.Context, roles []Role, passwords []string,
	) error
}

// SchemaQueryer interface lists common operations which may be taken
// with regards to database schema having either a connection or open
// transaction at hand. This interface is embedded by both of the
// SchemaConnQueryer and the SchemaTxQueryer in order to avoid
// redundant implementation.
type SchemaQueryer interface {
	// DropIfExists drops the `schema` schema without cascading if it
	// exists. That is, if `schema` does not exist, a nil error will be
	// returned without any change. And if `schema` exists and is empty,
	// it will be dropped. But if `schema` exists and is not empty, an
	// error will be returned.
	//
	// Caller is responsible to pass a trusted schema name string.
	DropIfExists(ctx context.Context, schema string) error

	// DropCascade drops `schema` schema with cascading, dropping all
	// dependent objects recursively. The `schema` must exist,
	// otherwise, an error will be returned.
	// This method is useful for dropping the intermediate schema
	// which are created during a migration.
	//
	// Caller is responsible to pass a trusted schema name string.
	DropCascade(ctx context.Context, schema string) error

	// CreateSchema tries to create the `schema` schema.
	// There must be no other schema with the `schema` name, otherwise,
	// this operation will fail.
	//
	// Caller is responsible to pass a trusted schema name string.
	CreateSchema(ctx context.Context, schema string) error

	// CreateRoleIfNotExists creates the `role` role if it does not
	// exist right now. Although the login option is enabled for the
	// created role, but no specific password will be set for it.
	// The ChangePasswords method may be used for setting a password if
	// desired. Otherwise, that user may not login effectively (but
	// using the trust or local identity methods).
	//
	// The `role` role name may be suffixed automatically based on
	// this schema queryer settings.
	CreateRoleIfNotExists(ctx context.Context, role Role) error

	// GrantPrivileges grants ALL privileges on the `schema` schema
	// to the `role` role, so it may create or access tables in that
	// schema and run relevant queries.
	//
	// The `role` role name may be suffixed automatically based on
	// this schema queryer settings.
	GrantPrivileges(ctx context.Context, schema string, role Role) error

	// SetSearchPath alters the given database role and sets its default
	// search_path to the given schema name alone.
	//
	// Updated search_path will be used by default in all future
	// transactions by that role, but it may be changed using the SET
	// statement as needed.
	SetSearchPath(ctx context.Context, schema string, role Role) error

	// GrantFDWUsage grants the USAGE privilege on the postgres_fdw
	// extension to the `role` role. Thereafter, that `role` role can
	// use the postgres_fdw extension in order to create a foreign
	// server or create a user mapping for it.
	//
	// The `role` role name may be suffixed automatically based on
	// this schema queryer settings.
	GrantFDWUsage(ctx context.Context, role Role) error
}
