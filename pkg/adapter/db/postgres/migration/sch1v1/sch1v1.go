// Copyright (c) 2024 Behnam Momeni
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Package sch1v1 provides the top-level Migrator type for database
// schema version 1.1.x which can be used for starting a multi-database
// migration operation. This package contains the main logic for
// querying of v1.1 schema and converting them to the latest supported
// minor version within the major version 1 series.
//
// Since schXvY packages only depend on their highest minor version
// implementation for creation of their corresponding upwards/downwards
// migrators and settlers, they can be adapted to a version-independent
// interface using a common Adapter interface which is provided by the
// schi.Adapter generic type (in contrast to the upwards/downwards
// migrator types in upmigN/dnmigN packages which have to ship their
// distinct Adapter types).
package sch1v1

import (
	"context"

	"github.com/momeni/clean-arch/pkg/adapter/db/postgres/migration/down/dnmig1"
	"github.com/momeni/clean-arch/pkg/adapter/db/postgres/migration/schi"
	"github.com/momeni/clean-arch/pkg/adapter/db/postgres/migration/settle/stlmig1"
	"github.com/momeni/clean-arch/pkg/adapter/db/postgres/migration/up/upmig1"
	"github.com/momeni/clean-arch/pkg/core/repo"
)

// These constants define the major, minor, and patch version of the
// database schema which is managed by Migrator struct in this package.
const (
	Major = 1
	Minor = 1
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
// high-level database schema migration logic for the v1.1 schema.
// It may be created with an open transaction of the destination
// database and a URL containing the source database connection info.
// The migration logic starts by calling the Load method which makes
// source database schema (having v1.1 format) accessible from the
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
// remote tables in a schema such as fdw1_1 while the settler object
// expects a schema such as mig1 for its data persistence queries.
func (s1v1 *Migrator) Settler(
	ctx context.Context,
) (*stlmig1.Settler, error) {
	if err := s1v1.migrateToLastMinorVersion(ctx); err != nil {
		return nil, err
	}
	return stlmig1.New(s1v1.tx), nil
}

// Load creates a Foreign Data Wrapper (FDW) link from the destination
// database to the source database (having the connection information
// of the source database) and imports the source database schema into
// a local schema. Thereafter, queries in the destination database
// transaction may access the source database contents.
// This method must be called (and returned without error) before it is
// possible to call any other method of the Migrator struct.
func (s1v1 *Migrator) Load(ctx context.Context) error {
	panic("Not implemented yet") // TODO: Implement
}

// UpMigrator expects the fdw1_1 schema to contain the source database
// contents (created by the Load method) and it fills the mig1 local
// schema using a series of views, keeping the v1.1 schema unchanged.
// Finally, it returns an instance of the upwards migrator object
// (having the U type) which can be used to migrate schema to the next
// major versions (if any) or obtain the settler object.
func (s1v1 *Migrator) UpMigrator(
	ctx context.Context,
) (*upmig1.Migrator, error) {
	if err := s1v1.migrateToLastMinorVersion(ctx); err != nil {
		return nil, err
	}
	return &upmig1.Migrator{s1v1.tx}, nil
}

// DownMigrator expects the fdw1_1 schema to contain the source database
// contents (created by the Load method) and it fills the mig1 local
// schema using a series of views, keeping the v1.1 schema unchanged.
// Finally, it returns an instance of the downwards migrator object
// (having the D type) which can be used to migrate schema to the
// previous major versions (if any) or obtain the settler object.
func (s1v1 *Migrator) DownMigrator(
	ctx context.Context,
) (*dnmig1.Migrator, error) {
	if err := s1v1.migrateToLastMinorVersion(ctx); err != nil {
		return nil, err
	}
	return &dnmig1.Migrator{s1v1.tx}, nil
}

// migrateToLastMinorVersion expects the fdw1_1 schema to contain the
// source database contents and it fills the mig1 local schema using a
// series of views, keeping the v1.1 schema unchanged.
func (s1v1 *Migrator) migrateToLastMinorVersion(
	ctx context.Context,
) error {
	panic("Not implemented yet") // TODO: Implement
}
