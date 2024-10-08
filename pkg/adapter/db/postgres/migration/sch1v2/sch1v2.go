// Copyright (c) 2024 Behnam Momeni
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Package sch1v2 provides the top-level Migrator type for database
// schema version 1.2.x which can be used for starting a multi-database
// migration operation. This package contains the main logic for
// querying of v1.2 schema and converting them to the latest supported
// minor version within the major version 1 series.
//
// Since schXvY packages only depend on their highest minor version
// implementation for creation of their corresponding upwards/downwards
// migrators and settlers, they can be adapted to a version-independent
// interface using a common Adapter interface which is provided by the
// schi.Adapter generic type (in contrast to the upwards/downwards
// migrator types in upmigN/dnmigN packages which have to ship their
// distinct Adapter types).
//
// Each schema minor-version specific package contains (and embeds) a
// file, namely lmv.sql, standing for the last-minor-version which
// contains the required DDL statements in order to create views in an
// intermediate migration schema, representing the last supported minor
// version within the same major version, based on the current minor
// version views which are prepared by the Load method. That is, lmv.sql
// specifies how we may migrate upwards from this minor version to the
// last supported minor version without switching the major version.
package sch1v2

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/momeni/clean-arch/pkg/adapter/db/postgres/migration/down/dnmig1"
	"github.com/momeni/clean-arch/pkg/adapter/db/postgres/migration/schi"
	"github.com/momeni/clean-arch/pkg/adapter/db/postgres/migration/settle/stlmig1"
	"github.com/momeni/clean-arch/pkg/adapter/db/postgres/migration/up/upmig1"
	"github.com/momeni/clean-arch/pkg/core/repo"
	"github.com/momeni/clean-arch/pkg/core/usecase/migrationuc"
)

// These constants define the major, minor, and patch version of the
// database schema which is managed by Migrator struct in this package.
const (
	Major = 1
	Minor = 2
	Patch = 0
)

// Following type aliases represent the version-dependent migrator
// related types. The U and D types represent the upwards and downwards
// migrator types. As Migrator.UpMigrator and Migrator.DownMigrator
// methods migrate a database schema from Minor minor version to the
// latest supported minor version (having the Major major version),
// they create an instance of U and D respectively which can be used
// for migrating to next/previous major versions. The S type represents
// the schema settler type. If a migration reaches to the Major major
// version, an instance of S may be used for persisting the migration
// result. The Type combines these type aliases using the schi.Migrator
// generic interface. Ensuring that Migrator implements the Type
// interface helps to receive a compilation error in case of a missing
// method or having the wrong method type (as enforced by a test file).
type (
	// S is the schema settler type.
	S = *stlmig1.Settler
	// U is an upwards migrator type.
	U = *upmig1.Migrator
	// D is downwards migrator type.
	D = *dnmig1.Migrator
	// Type is provided by Migrator type.
	Type = schi.Migrator[S, U, D]
)

// New creates a Migrator struct wrapping the given `tx` transaction
// from the destination database and `url` URL representing the source
// database connection information. The created Migrator instance is
// then wrapped by a schi.Adapter in order to adapt its version
// dependent interface (see Type type alias) to a version independent
// repo.Migrator[repo.SchemaSettler] interface.
func New(tx repo.Tx, url string) repo.Migrator[repo.SchemaSettler] {
	m := &Migrator{tx, url}
	return schi.Adapter[S, U, D]{m}
}

// Migrator implements Type generic interface in order to provide
// high-level database schema migration logic for the v1.2 schema.
// It may be created with an open transaction of the destination
// database and a URL containing the source database connection info.
// The migration logic starts by calling the Load method which makes
// source database schema (having v1.2 format) accessible from the
// destination database. Then UpMigrator or DownMigrator method should
// be called in order to convert them (by creating relevant views and
// without actual transfer of data items as far as possible) into the
// latest available minor version within the v1 major version.
// Obtained upwards/downwards migrator object (having the U/D type)
// may be used for changing the major version (keeping the minor version
// at its latest supported version in each major version).
// Finally, the settler object is used to persist the migration. If
// the ultimate major version is v1 too, the Settler method may be used
// as a shortcut for migrating from the current minor version to the
// latest supported minor version and returning an instance of the
// settler object (having the S type) for persisting it.
type Migrator struct {
	tx  repo.Tx // an open transaction for the destination database
	url string  // connection information for the source database
}

// Settler returns a settler object for the database schema v1 major
// version. Beforehand, it migrates the database schema from its
// current minor version (represented by Minor const) to the latest
// available minor version. This upwards migration is also applicable to
// the latest supported minor version itself, because the Load method
// (which must be called before calling Settler method) will put the
// remote tables in a schema such as fdw1_2 while the settler object
// expects a schema such as mig1 for its data persistence queries.
func (s1v2 *Migrator) Settler(
	ctx context.Context,
) (*stlmig1.Settler, error) {
	if err := s1v2.migrateToLastMinorVersion(ctx); err != nil {
		return nil, err
	}
	return stlmig1.New(s1v2.tx), nil
}

// Load creates a Foreign Data Wrapper (FDW) link from the destination
// database to the source database (having the connection information
// of the source database) and imports the source database schema into
// a local schema. Thereafter, queries in the destination database
// transaction may access the source database contents.
// This method must be called (and returned without error) before it is
// possible to call any other method of the Migrator struct.
func (s1v2 *Migrator) Load(ctx context.Context) error {
	if err := schi.LoadFDW(
		ctx, Major, Minor, s1v2.tx, s1v2.url,
	); err != nil {
		return fmt.Errorf(
			"schi.LoadFDW(major=%d, minor=%d, srcURL=%q): %w",
			Major, Minor, s1v2.url, err,
		)
	}
	return nil
}

// UpMigrator expects the fdw1_2 schema to contain the source database
// contents (created by the Load method) and it fills the mig1 local
// schema using a series of views, keeping the v1.2 schema unchanged.
// Finally, it returns an instance of the upwards migrator object
// (having the U type) which can be used to migrate schema to the next
// major versions (if any) or obtain the settler object.
func (s1v2 *Migrator) UpMigrator(
	ctx context.Context,
) (*upmig1.Migrator, error) {
	if err := s1v2.migrateToLastMinorVersion(ctx); err != nil {
		return nil, err
	}
	return &upmig1.Migrator{s1v2.tx}, nil
}

// DownMigrator expects the fdw1_2 schema to contain the source database
// contents (created by the Load method) and it fills the mig1 local
// schema using a series of views, keeping the v1.2 schema unchanged.
// Finally, it returns an instance of the downwards migrator object
// (having the D type) which can be used to migrate schema to the
// previous major versions (if any) or obtain the settler object.
func (s1v2 *Migrator) DownMigrator(
	ctx context.Context,
) (*dnmig1.Migrator, error) {
	if err := s1v2.migrateToLastMinorVersion(ctx); err != nil {
		return nil, err
	}
	return &dnmig1.Migrator{s1v2.tx}, nil
}

// lastMinorVersionStatements embeds the lmv.sql file contents which are
// supposed to create database schema tables (or preferably just views)
// in the mig1 schema for the last supported minor version in the major
// version 1 and fill them (or in case of the views, just specify the
// rule which can be used for computation of the corresponding columns
// values) with this assumption that the current minor version tables
// are accessible in the fdw1_2 schema as prepared by the Load method.
//
//go:embed lmv.sql
var lastMinorVersionStatements string

// migrateToLastMinorVersion expects the fdw1_2 schema to contain the
// source database contents and it fills the mig1 local schema using a
// series of views, keeping the v1.2 schema unchanged.
func (s1v2 *Migrator) migrateToLastMinorVersion(
	ctx context.Context,
) error {
	if _, err := s1v2.tx.Exec(
		ctx, lastMinorVersionStatements,
	); err != nil {
		fdwSchema := migrationuc.ForeignSchemaName(Major, Minor)
		migSchema := migrationuc.MigrationSchemaName(Major)
		return fmt.Errorf(
			"migrating from %q schema to %q schema: %w",
			fdwSchema, migSchema, err,
		)
	}
	return nil
}
