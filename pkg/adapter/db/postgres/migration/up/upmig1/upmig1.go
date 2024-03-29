// Copyright (c) 2024 Behnam Momeni
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Package upmig1 provides an upwards database schema Migrator type for
// major version 1 and its corresponding Adapter type which can adapt
// it to the version independent repo.UpMigrator[repo.SchemaSettler]
// interface.
//
// Each upwards schema migrator package, but the last major version
// which has no newer major version, may contain (and embed) an up.sql
// file in order to migrate from its own major version to the next
// major version, creating the relevant views.
package upmig1

import (
	"context"
	"errors"

	"github.com/momeni/clean-arch/pkg/adapter/db/postgres/migration/settle/stlmig1"
	"github.com/momeni/clean-arch/pkg/adapter/db/postgres/migration/up"
	"github.com/momeni/clean-arch/pkg/core/repo"
)

// These type aliases specify the settler type (with major version 1)
// as S and the parameterized up.Migrator interface which is supposed
// to be implemented by the Migrator struct as Type.
// The Type uses the *Migrator from the same package/version because
// there is no newer/upper version.
type (
	// S is the schema settler type.
	S = *stlmig1.Settler
	// Type is provided by Migrator type.
	Type = up.Migrator[S, *Migrator]
)

// Migrator provides a database schema upwards migrator which wraps
// a transaction of the destination database, assumes that it contains
// a filled schema, namely mig1, which contains the major version 1
// database schema views (or tables), and allows upwards major version
// migrations by implementing the Type generic interface. It also
// implements the pkg/adapter/db/postgres/migration/schi.UpMigAdapter
// interface, so it may be adapted to a version-independent interface.
// This adaptation may not be implemented by a generic adapter struct
// because each upwards migrator Type may depend on a long sequence of
// migrator types belonging to the subsequent major versions, while
// Go language does not support variadic type parameters.
type Migrator struct {
	Tx repo.Tx // a transaction of the destination database
}

// Adapt creates an instance of the Adapter struct and wraps `upmig1`
// object in order to adapt it to the version-independent
// repo.UpMigrator[repo.SchemaSettler] interface.
//
// This method makes the Migrator to implement the
// pkg/adapter/db/postgres/migration/schi.UpMigAdapter interface.
// It is not required to explicitly test that Migrator implements
// the UpMigAdapter interface (in a test file) because upwards migrator
// types are only used by high-level migrator types (in schXvY packages)
// which implement the pkg/adapter/db/postgres/migration/schi.Migrator
// generic interface themselves and so guarantee that upwards migrators
// actually implement the schi.UpMigAdapter interface.
// Generally, it is preferred to have an actual code in order to ensure
// that an interface is implemented instead of a syntactic test code.
// Ensuring such an implementation in test codes is only justified when
// even a wrong implementation may be taken at compile-time (since its
// type is so general which may be passed instead of another struct) and
// so a test code is required for catching bugs before running the main
// non-test code. For example, asserting that Migrator implements the
// Type type is required because its users only work with the version
// independent interface which is returned by this function. By the way,
// we do not have to write that test code since the Adapter type is
// wrapping the Migrator as an instance of the Type type. If we
// detected a performance issue (by runtime profiling or benchmarks),
// it is possible to update Adapter in order to wrap a *Migrator
// directly and move its Type implementation assertion to a test file.
func (upmig1 *Migrator) Adapt() repo.UpMigrator[repo.SchemaSettler] {
	return Adapter{upmig1}
}

// Adapter wraps a Type interface instance and adapts it to the
// version-independent repo.UpMigrator[repo.SchemaSettler] interface.
type Adapter struct {
	T Type // wrapped version-dependent upwards migrator object
}

// Settler adapts the Settler method of the wrapped `a.T` instance, so
// its return value can be provided as a version-independent
// repo.SchemaSettler interface instead of the version-dependent S type.
func (a Adapter) Settler() repo.SchemaSettler {
	return a.T.Settler()
}

// MigrateUp adapts the MigrateUp method of the wrapped `a.T` instance,
// so its return value can be provided as a version-independent
// repo.UpMigrator[repo.SchemaSettler] interface instead of the
// version-dependent *Migrator type. Any returned error is produced
// by the underlying MigrateUp method.
func (a Adapter) MigrateUp(ctx context.Context) (
	repo.UpMigrator[repo.SchemaSettler], error,
) {
	m, err := a.T.MigrateUp(ctx)
	if err != nil {
		return nil, err
	}
	return m.Adapt(), nil
}

// Settler returns a settler object (with S type) without performing
// any migration action (so, no error condition may arise). Returned
// settler object may be employed to persist the migration results.
// See the stlmig1.Settler type for more details.
func (upmig1 *Migrator) Settler() *stlmig1.Settler {
	return stlmig1.New(upmig1.Tx)
}

// MigrateUp migrates from the major version 1 to the next major
// version by creating relevant views in a schema such as mig2
// based on the views in a schema such as mig1 considering their
// latest supported minor versions. However, version 1 is the latest
// supported major version. Therefore, this method always returns
// an error (and a nil migrator as the first return value).
func (upmig1 *Migrator) MigrateUp(
	_ context.Context,
) (*Migrator, error) {
	return nil, errors.New("v1 is the latest schema major version")
}
