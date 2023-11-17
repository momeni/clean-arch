// Copyright (c) 2024 Behnam Momeni
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Package schi provides the schema migrator interfaces and is imported
// by schXvY packages. Since the main migrator package has to import
// all schema migrator implementation packages, it cannot provide these
// internal interfaces too (so a distinct package is necessary to avoid
// import cycle).
//
// A series of Migrator, UpMigAdapter, and DnMigAdapter generic
// interfaces are defined which must be implemented by top-level
// schema migrators, and their corresponding upwards/downwards
// migrators, so they can be adapted using the generic Adapter
// implementation to the version-independent
// repo.Migrator[repo.SchemaSettler],
// repo.UpMigrator[repo.SchemaSettler], and
// repo.DownMigrator[repo.SchemaSettler] interfaces.
package schi

import (
	"context"

	"github.com/momeni/clean-arch/pkg/core/repo"
)

// Migrator of S, U, and D is a generic interface which is created for
// two main purposes. First, it can be implemented by database schema
// migrator implementations with proper version-dependent type params
// in order to cause compile-time errors if they miss any required
// method or have inappropriate method types. So, when a new version
// is implemented by copying an older package and updating its methods,
// it is easier to guarantee that all methods are updated properly.
// Second, this interface may be used in order to unify adaptation of
// the migrator objects to a version-independent interface (which is
// required for passing migrators to the use cases layer) using a
// generic Adapter[S, U, D] struct implementation.
//
// For example implementations of this interface, see the schXvY
// packages such as pkg/adapter/db/postgres/migration/sch1v0 for v1.0.
// For the corresponding version-independent interface,
// see pkg/core/repo/migrator.go file.
//
// The S, U, and D type parameters represent the settler, upwards
// migrator, and downwards migrator types respectively, all for the
// source database schema version. The Load method may be used to make
// tables of a specific major.minor version accessible remotely in a
// destination database, UpMigrator and DownMigrator methods may be used
// to convert them by creation of proper views to the expected schema
// at the same major version, but having its latest supported minor
// version, returned objects with the U and D types may be used for
// migrating one major version upwards/downwards at a time, and finally
// the settler object (which in the source specific major version has
// the S type) may be used for persisting the migration by creating the
// relevant tables and filling them based on their corresponding views.
//
// All methods may only be called after the Load method was called and
// returned nil error.
type Migrator[
	S repo.SchemaSettler, U UpMigAdapter, D DnMigAdapter,
] interface {
	// Settler migrates from the current minor version to the latest
	// supported version, creating the relevant views in a local schema
	// such as migN based on the tables in a schema such as fdwN_M,
	// where the N represents the common major version and M represents
	// some supported minor version. Thereafter, it returns a settler
	// object (with S type) which can persist the migration results
	// (from the migN schema) by creating relevant tables and filling
	// them (in a schema such as cawebN).
	Settler(ctx context.Context) (S, error)

	// Load loads the source database schema, so they may be used
	// in the destination database. Loaded data are as limited as
	// possible, restricted to metadata like a FDW link if possible.
	Load(ctx context.Context) error

	// UpMigrator migrates from the current minor version to the latest
	// supported version, creating the relevant views in a local schema
	// such as migN based on the tables in a schema such as fdwN_M,
	// where the N represents the common major version and M represents
	// some supported minor version. Thereafter, it returns an upwards
	// migrator object (with U type) which can be used for migrating to
	// the next versions (if any).
	UpMigrator(ctx context.Context) (U, error)

	// DownMigrator migrates from the current minor version to the
	// latest supported version, creating the relevant views in a local
	// schema such as migN based on the tables in a schema such as
	// fdwN_M, where the N represents the common major version and M
	// represents some supported minor version. Thereafter, it returns
	// a downwards migrator object (with D type) which can be used for
	// migrating to the previous versions (if any).
	DownMigrator(ctx context.Context) (D, error)
}

// UpMigAdapter interface represents an object which can be adapted to
// the repo.UpMigrator[repo.SchemaSettler] version-independent upwards
// migrator interface.
//
// The Migrator[S, U, D] generic interface can create and return an
// upwards migrator object with U type. The Adapter[S, U, D] struct
// needs to wrap a Migrator[S, U, D] instance and adapt it to a
// version-independent interface and so it needs a way to adapt
// the U instances to repo.UpMigrator[repo.SchemaSettler] interface.
// Requring the U type to implement UpMigAdapter interface provides that
// way by asking each database schema upwards migrator implementation
// to implement its relevant Adapt method.
type UpMigAdapter interface {
	// Adapt adapts this version-dependent upwards migrator object to
	// the repo.UpMigrator[repo.SchemaSettler] version-independent
	// interface.
	Adapt() repo.UpMigrator[repo.SchemaSettler]
}

// DnMigAdapter interface represents an object which can be adapted to
// the repo.DownMigrator[repo.SchemaSettler] version-independent
// downwards migrator interface.
//
// The Migrator[S, U, D] generic interface can create and return a
// downwards migrator object with D type. The Adapter[S, U, D] struct
// needs to wrap a Migrator[S, U, D] instance and adapt it to a
// version-independent interface and so it needs a way to adapt
// the D instances to repo.DownMigrator[repo.SchemaSettler] interface.
// Requring the D type to implement DnMigAdapter interface provides that
// way by asking each database schema downwards migrator implementation
// to implement its relevant Adapt method.
type DnMigAdapter interface {
	// Adapt adapts this version-dependent downwards migrator object to
	// the repo.DownMigrator[repo.SchemaSettler] version-independent
	// interface.
	Adapt() repo.DownMigrator[repo.SchemaSettler]
}

// Adapter of S, U, and D wraps and adapts an instance of the
// Migrator[S, U, D] in order to provide the version-independent
// repo.Migrator[repo.SchemaSettler] interface.
//
// The S, U, and D type parameters represent the settler, upwards
// migrator, and downwards migrator types respectively, all for the
// source database schema version.
type Adapter[
	S repo.SchemaSettler, U UpMigAdapter, D DnMigAdapter,
] struct {
	// M is the version-dependent migrator object which is taken
	// as a generic interface in order to be adapted.
	M Migrator[S, U, D]
}

// Settler migrates from the current minor version to the latest
// supported version, creating the relevant views in a local schema such
// as migN based on the tables in a schema such as fdwN_M, where the N
// represents the common major version and M represents some supported
// minor version. Thereafter, it obtains a settler object (with S type)
// which can persist the migration results (from the migN schema) by
// creating relevant tables and filling them (in a schema such as
// cawebN) and returns that object as a repo.SchemaSettler interface.
func (a Adapter[S, U, D]) Settler(
	ctx context.Context,
) (repo.SchemaSettler, error) {
	return a.M.Settler(ctx)
}

// Load loads the source database schema, so they may be used in the
// destination database. Loaded data are as limited as possible,
// restricted to metadata like a FDW link if possible.
// Since Load has no type-parameter dependent argument or return value,
// it simply delegates to the `a.M.Load` method.
func (a Adapter[S, U, D]) Load(ctx context.Context) error {
	return a.M.Load(ctx)
}

// MajorVersion returns the major semantic version of this `a.M`
// Migrator instance. It reflects the major version of a database schema
// and its value only depends on the S settler type. This method may be
// used for identification of the migration versions path, passing
// through the major versions one by one.
func (a Adapter[S, U, D]) MajorVersion() uint {
	var s S
	// MajorVersion only depends on the type of S because each settler
	// type support exactly one major version and so MajorVersion method
	// can be independent of the actual a.M instance.
	return s.MajorVersion()
}

// UpMigrator migrates from the current minor version to the latest
// supported version, creating the relevant views in a local schema
// such as migN based on the tables in a schema such as fdwN_M,
// where the N represents the common major version and M represents
// some supported minor version. Thereafter, it obtains an upwards
// migrator object (with U type) which can be used for migrating to
// the next versions (if any) and adapts it to the version-independent
// repo.UpMigrator[repo.SchemaSettler] interface using its Adapt method.
// In case of errors, first return value will be nil.
func (a Adapter[S, U, D]) UpMigrator(ctx context.Context) (
	repo.UpMigrator[repo.SchemaSettler], error,
) {
	uma, err := a.M.UpMigrator(ctx)
	if err != nil {
		return nil, err
	}
	return uma.Adapt(), nil
}

// DownMigrator migrates from the current minor version to the latest
// supported version, creating the relevant views in a local schema
// such as migN based on the tables in a schema such as fdwN_M,
// where the N represents the common major version and M represents
// some supported minor version. Thereafter, it obtains a downwards
// migrator object (with D type) which can be used for migrating to
// the previous versions (if any) and adapts it to the
// version-independent repo.DownMigrator[repo.SchemaSettler] interface
// using its Adapt method.
// In case of errors, first return value will be nil.
func (a Adapter[S, U, D]) DownMigrator(ctx context.Context) (
	repo.DownMigrator[repo.SchemaSettler], error,
) {
	dma, err := a.M.DownMigrator(ctx)
	if err != nil {
		return nil, err
	}
	return dma.Adapt(), nil
}
