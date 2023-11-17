// Copyright (c) 2024 Behnam Momeni
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Package dnmig1 provides a downwards database schema Migrator type for
// major version 1 and its corresponding Adapter type which can adapt
// it to the version independent repo.DownMigrator[repo.SchemaSettler]
// interface.
package dnmig1

import (
	"context"
	"errors"

	"github.com/momeni/clean-arch/pkg/adapter/db/postgres/migration/down"
	"github.com/momeni/clean-arch/pkg/adapter/db/postgres/migration/settle/stlmig1"
	"github.com/momeni/clean-arch/pkg/core/repo"
)

// These type aliases specify the settler type (with major version 1)
// as S and the parameterized down.Migrator interface which is supposed
// to be implemented by the Migrator struct as Type.
// The Type uses the *Migrator from the same package/version because
// there is no older/downer version.
type (
	// S is the schema settler type.
	S = *stlmig1.Settler
	// Type is provided by Migrator type.
	Type = down.Migrator[S, *Migrator]
)

// Migrator provides a database schema downwards migrator which wraps
// a transaction of the destination database, assumes that it contains
// a filled schema, namely mig1, which contains the major version 1
// database schema views (or tables), and allows downwards major version
// migrations by implementing the Type generic interface. It also
// implements the pkg/adapter/db/postgres/migration/schi.DnMigAdapter
// interface, so it may be adapted to a version-independent interface.
// This adaptation may not be implemented by a generic adapter struct
// because each downwards migrator Type may depend on a long sequence of
// migrator types belonging to the older major versions, while Golang
// does not support variadic type parameters.
type Migrator struct {
	Tx repo.Tx // a transaction of the destination database
}

// Adapt creates an instance of the Adapter struct and wraps `dnmig1`
// object in order to adapt it to the version-independent
// repo.DownMigrator[repo.SchemaSettler] interface.
//
// This method makes the Migrator to implement the
// pkg/adapter/db/postgres/migration/schi.DnMigAdapter interface. It is
// not required to explicitly test that Migrator implements the
// DnMigAdapter interface (in a test file) because downwards migrator
// types are only used by high-level migrator types (in schXvY packages)
// which implement the pkg/adapter/db/postgres/migration/schi.Migrator
// generic interface themselves and so guarantee that downwards
// migrators actually implement the schi.DnMigAdapter interface.
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
func (dnmig1 *Migrator) Adapt() repo.DownMigrator[repo.SchemaSettler] {
	return Adapter{dnmig1}
}

// Adapter wraps a Type interface instance and adapts it to the
// version-independent repo.DownMigrator[repo.SchemaSettler] interface.
type Adapter struct {
	T Type
}

// Settler adapts the Settler method of the wrapped `a.T` instance, so
// its return value can be provided as a version-independent
// repo.SchemaSettler interface instead of the version-dependent S type.
func (a Adapter) Settler() repo.SchemaSettler {
	return a.T.Settler()
}

// MigrateDown adapts the MigrateDown method of the wrapped `a.T`
// instance, so its return value can be provided as a
// version-independent repo.DownMigrator[repo.SchemaSettler] interface
// instead of the version-dependent *Migrator type. Any returned error
// is produced by the underlying MigrateDown method.
func (a Adapter) MigrateDown(ctx context.Context) (
	repo.DownMigrator[repo.SchemaSettler], error,
) {
	m, err := a.T.MigrateDown(ctx)
	if err != nil {
		return nil, err
	}
	return m.Adapt(), nil
}

// Settler returns a settler object (with S type) without performing
// any migration action (so, no error condition may arise). Returned
// settler object may be employed to persist the migration results.
// See the stlmig1.Settler type for more details.
func (dnmig1 *Migrator) Settler() *stlmig1.Settler {
	return stlmig1.New(dnmig1.Tx)
}

// MigrateDown migrates from the major version 1 to the previous major
// version by creating relevant views in a schema such as mig0
// based on the views in a schema such as mig1 considering their
// latest supported minor versions. However, version 0 changes are not
// maintained and major version 1 is the first version which its schema
// changes are tracked by migrations. Therefore, this method always
// returns an error (and a nil migrator as the first return value).
func (dnmig1 *Migrator) MigrateDown(
	ctx context.Context,
) (*Migrator, error) {
	return nil, errors.New("v1 is the foremost schema major version")
}
