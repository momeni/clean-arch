// Copyright (c) 2024 Behnam Momeni
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package migrationuc

import (
	"context"
	"fmt"

	"github.com/momeni/clean-arch/pkg/core/model"
	"github.com/momeni/clean-arch/pkg/core/repo"
	"gopkg.in/yaml.v3"
)

// SchemaSettings represents the database-related settings which should
// be provided by a configuration file. It allows a database connection
// pool to be established for an asked role using the ConnectionPool
// method, reports the database schema version which is required for
// querying the stored tables, may be used as a factory for the
// repo.Migrator (in order to migrate the database to an upper or downer
// version), for changing passwords of a set of database roles and
// storing new passwords in relevant files (with the atomic updating
// considerations), or as a factory for repo.SchemaInitializer (in order
// to initialize an empty database with development or production
// suitable data).
type SchemaSettings interface {
	// ConnectionPool creates a database connection pool using the
	// connection information which are kept in this SchemaSettings
	// instance. The `r` argument specifies the role name for the
	// created connection pool.
	//
	// Password values are kept in files in a specific password dir
	// and creation of a connection pool depends on identification of
	// a valid password for the given role and the database host, port,
	// and name which are taken from this SchemaSettings instance.
	// Each non-empty and non-commented line of the passwords file
	// should conform with this format:
	//
	//	host:port:dbname:role:password
	//
	// For sake of atomic passwords updating operations (during a DB
	// migration), a second temporary passwords file may be created
	// in order to hold the new values of passwords. Therefore, even in
	// case of a failed migration operation, either old or new passwords
	// from the main or temporary passwords file may be used to connect
	// to the database. If such a temporary passwords file was used for
	// establishment of a connection pool, it will be moved to the main
	// passwords file before returning (so the temporary file may be
	// overwritten safely by the subsequent migration operations).
	ConnectionPool(ctx context.Context, r repo.Role) (repo.Pool, error)

	// ConnectionInfo returns the database name, host, and port of the
	// connection information which are kept in this SchemaSettings
	// instance. The ConnectionPool method can be used to employ these
	// information and connect to a database.
	ConnectionInfo() (dbName, host string, port int)

	// SchemaMigrator creates a repo.Migrator[repo.SchemaSettler]
	// instance which wraps the given transaction argument and can be
	// used for (1) loading the source database schema information
	// with this assumption that tx belongs to the destination database
	// and this SchemaSettings contains the source database connection
	// information, so it can modify the destination database within a
	// given transaction and fill a schema with tables which represent
	// the source database contents (not moving data items necessarily,
	// but may create them as a foreign data wrapper, aka FDW),
	// (2) creating upwards or downwards migrator objects in order to
	// transform the loaded data into their upper/lower schema versions
	// (again with minimal data transfer and using views instead of
	// tables as far as possible, while creating tables or even loading
	// data into this Golang process if it is necessary), and at last
	// (3) obtaining a repo.SchemaSettler instance for the target schema
	// major version, so it can persist the target schema version by
	// creating tables and filling them with contents of corresponding
	// views.
	SchemaMigrator(tx repo.Tx) (
		repo.Migrator[repo.SchemaSettler], error,
	)

	// SchemaInitializer creates a repo.SchemaInitializer instance
	// which wraps the given transaction argument and can be used to
	// initialize the database with development or production suitable
	// data. The format of the created tables and their initial data
	// rows are chosen based on the database schema version, as
	// indicated by SchemaVersion method. All table creation and data
	// insertion operations will be performed in the given transaction
	// and will be persisted only if the `tx` could commit successfully.
	SchemaInitializer(tx repo.Tx) (repo.SchemaInitializer, error)

	// RenewPasswords generates new secure passwords for the given roles
	// and after recording them in a temporary file, will use the change
	// function in order to update the passwords of those roles in the
	// database too. The change function argument should perform the
	// update operation in a transaction which may or may not be
	// committed when RenewPasswords returns. In case of a successful
	// commitment, the temporary passwords file should be moved over
	// the main passwords file, as known in the current SchemaSettings
	// instance (so it may be used for the future calls to the
	// ConnectionPool method). This final file movement can be performed
	// using the returned finalizer function.
	RenewPasswords(
		ctx context.Context,
		change func(
			ctx context.Context,
			roles []repo.Role,
			passwords []string,
		) error,
		roles ...repo.Role,
	) (finalizer func() error, err error)

	// SchemaVersion returns the semantic version of the database schema
	// which its connection information are kept by this SchemaSettings.
	SchemaVersion() model.SemVer

	// SetSchemaVersion updates the semantic version of the database
	// schema as recorded in this schema settings and reported by the
	// SchemaVersion method.
	SetSchemaVersion(sv model.SemVer)
}

// Settings interface represents the expectations of the migration
// use cases from the configuration files contents. Each Config struct
// version has to be adapted in order to provide this interface before
// being passed to the migration use cases.
//
// The Config structs and their adapter-layer migrators may use the
// concrete struct types for sake of type-safety. For example, migrating
// a cfg1.Config instance up leads to creation of cfg2.Config struct.
// However, these adapter-layer structs differences among a series of
// versions are masked out (by their Adapter implementations), so they
// can be managed uniformly in the use cases layer.
type Settings interface {
	// Marshaler interface customizes the YAML serialization of a
	// configuration file contents, so it can replace specific settings
	// such as a slices of numbers in a vers.Config with alternative
	// data types and have control on the final serilization result.
	//
	// See the Marshal function of any Config struct for the reification
	// details and how marshaling logic can be distributed among nested
	// Config structs.
	yaml.Marshaler

	// SchemaSettings represents the database-related parts of Settings.
	SchemaSettings

	// Clone creates a deep copy of this Settings instance.
	Clone() Settings

	// MergeSettings expects to receive a Settings instance which has
	// the same settings version as the current instance. When migrating
	// settings to newer or older versions, some of the target version
	// settings may be left uninitialized. This method fills those items
	// with their correspoding values from the given `s` argument.
	// Some settings, such as the database connection information, are
	// unconditionally taken from the `s` argument because they need to
	// describe the destination settings values.
	MergeSettings(s Settings) error

	// Version returns the semantic version of this Settings format.
	Version() model.SemVer
}

// HasTheSameConnectionInfo returns true if and only if both of the
// `s1` and `s2` schema settings contain the connection information
// for a common database. That is, their host, port, and database name
// do match. If they described the same database, they must also have
// the same database schema semantic major version and the minor version
// of `s1` must be equal to or greater than the `s2` minor version.
// Otherwise, a non-nil error will be returned too.
//
// This method is useful for finding out if we have two distinct
// databases, hence, one can be used as a read-only source database
// while writing to the other database as a migration destination.
func HasTheSameConnectionInfo(s1, s2 SchemaSettings) (bool, error) {
	n1, h1, p1 := s1.ConnectionInfo()
	n2, h2, p2 := s2.ConnectionInfo()
	if h1 != h2 || p1 != p2 || n1 != n2 {
		return false, nil
	}
	v1 := s1.SchemaVersion()
	v2 := s2.SchemaVersion()
	if v1[0] == v2[0] && v1[1] >= v2[1] {
		return true, nil
	}
	return true, fmt.Errorf("version mismatch: %v vs. %v", v1, v2)
}
