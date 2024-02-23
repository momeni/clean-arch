// Copyright (c) 2024 Behnam Momeni
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Package migration is the top-level database migration package which
// acts as a facade for all supported database schema versions.
//
// The New function is a facade for creation of migrator objects which
// depend on the major and minor semantic versions, while NewInitializer
// and LatestVersion functions can be used to find out the latest
// supported minor version for each major version and create its schema
// initializer object.
// This package depends on its sub-packages and returns the relevant
// types as version-independent interfaces.
package migration

import (
	"context"
	"fmt"

	"github.com/momeni/clean-arch/pkg/adapter/db/postgres/migration/sch1v0"
	"github.com/momeni/clean-arch/pkg/adapter/db/postgres/migration/sch1v1"
	"github.com/momeni/clean-arch/pkg/adapter/db/postgres/migration/settle/stlmig1"
	"github.com/momeni/clean-arch/pkg/core/model"
	"github.com/momeni/clean-arch/pkg/core/repo"
)

// LatestVersion returns the latest supported database schema version
// within the major version of the given `v` semantic version.
// If the minor version of `v` argument is beyond the supported database
// schema versions, an error will be returned.
//
// This function is useful for finding the target database schema
// version before performing the actual database schema migration.
// Obtained version may be written in the target temporary configuration
// file in order to describe the database schema which is being created.
func LatestVersion(v model.SemVer) (lv model.SemVer, err error) {
	switch major := v[0]; major {
	case 1:
		if minor := v[1]; minor > stlmig1.Minor {
			err = fmt.Errorf("unsupported minor: %d", minor)
			return
		}
		lv = model.SemVer{1, stlmig1.Minor, stlmig1.Patch}

	default:
		err = fmt.Errorf("unsupported major: %d", major)
	}
	return
}

// NewInitializer creates a database schema initializer instance for the
// given `v` semantic version. A repo.SchemaInitializer can be used for
// filling an existing database schema with the development or
// production suitable initial data. Since new tables and columns may
// be introduced by each minor version (within a major version), but
// they may not be removed or renamed, the older codes (which may query
// fewer tables/columns of an older schema) may connect to and query a
// schema with newer minor version too (ignoring the extra tables and
// columns). Therefore, the major version of the given `v` semantic
// version is examined and an instance of its corresponding database
// schema settler is created. Each schema settler is implemented in its
// separate package, namely stlmigN (for the N major version) and is
// able to create tables and fill them with initial data for the N major
// version and its latest supported minor version.
//
// If the minor version of `v` argument is beyond the supported versions
// of stlmigN package, an error will be returned because that package
// may create tables which are consumable by its own minor version and
// older minor versions, while newer minor versions expect to see more
// tables or columns which are not going to be created.
//
// The returned instance wraps the `tx` transaction argument and
// uses it for creation and initialization of tables. The caller remains
// responsible to commit that transaction.
func NewInitializer(tx repo.Tx, v model.SemVer) (
	repo.SchemaInitializer, error,
) {
	switch major := v[0]; major {
	case 1:
		if minor := v[1]; minor > stlmig1.Minor {
			return nil, fmt.Errorf("unsupported minor: %d", minor)
		}
		return stlmig1.New(tx), nil
	default:
		return nil, fmt.Errorf("unsupported major: %d", major)
	}
}

// New creates a repo.Migrator[repo.SchemaSettler] instance based on
// the given `v` semantic version. A repo.SchemaSettler specifies how
// a given database schema tables may be created and a repo.Migrator
// indicates how we may begin from a source version of SchemaSettler,
// migrate its minor version to its latest supported one within the
// same major version, migrate major versions upwards or downwards
// one version at a time (staying at the latest supported minor version
// of each major version), and finally settle the migration result.
// Read the repo.SchemaSettler and repo.Migrator documentation for more
// details. Since new tables and columns may be introduced by each minor
// version (within a major version), it is not possible to use a code
// which queries a newer minor version for loading data from an older
// minor version (as it queries tables/columns which do not exist).
// Therefore, the major and minor versions of the given `v` semantic
// version are examined and an instance of their corresponding database
// schema migrator is created. Each schema migrator is implemented in
// its separate package, namely schXvY (for the schema with X major
// version and Y minor version), and uses concrete version-dependent
// types for its loading, upwards/downwards migration, and settlement
// operations. However, each schXvY package is also accompanied by an
// Adapter struct which helps it in realization of a version-independent
// interface, aka repo.Migrator[repo.SchemaSettler], which is returned
// by this method.
//
// If the major or minor versions of `v` argument are not supported
// specifically, an error will be returned. The patch version is
// irrelevant. The returned instance also wraps the `tx` transaction
// and the `url` database connection URL arguments. The former must be
// an open transaction to the destination database (which will be
// updated in order to create tables and views for a series of versions
// starting from the source database version to the expected destination
// version) and the latter must contain the connection information of
// the source database (so it can be used from within the `tx`
// destination transaction in order to connect to the source database
// and start accessing it with a FDW link). The caller remains
// responsible to commit that transaction.
func New(tx repo.Tx, v model.SemVer, url string) (
	repo.Migrator[repo.SchemaSettler], error,
) {
	switch major := v[0]; major {
	case 1:
		switch minor := v[1]; minor {
		case 0:
			return sch1v0.New(tx, url), nil
		case 1:
			return sch1v1.New(tx, url), nil
		default:
			return nil, fmt.Errorf("unsupported minor: %d", minor)
		}
	default:
		return nil, fmt.Errorf("unsupported major: %d", major)
	}
}

// LoadSettings loads the serialized mutable settings from the database
// using the given `c` connection, assuming that the database schema
// version is equal with the given `v` argument. Loading depends on
// both of the major and minor versions just like the migration loading
// phase.
func LoadSettings(ctx context.Context, c repo.Conn, v model.SemVer) (
	[]byte, error,
) {
	switch major := v[0]; major {
	case 1:
		switch minor := v[1]; minor {
		case 0:
			return sch1v0.LoadSettings(ctx, c)
		case 1:
			return sch1v1.LoadSettings(ctx, c)
		default:
			return nil, fmt.Errorf("unsupported minor: %d", minor)
		}
	default:
		return nil, fmt.Errorf("unsupported major: %d", major)
	}
}

// NewSettingsPersister instantiates a repo.SettingsPersister which
// wraps the given `tx` transaction and supports storage of a serialized
// version of mutable settings in the database.
//
// For the underlying implementation, see stlmigN packages which depend
// on the schema major version (since migration needs to fill tables
// for the latest supported minor version of a desired major version).
func NewSettingsPersister(tx repo.Tx, v model.SemVer) (
	repo.SettingsPersister, error,
) {
	switch major := v[0]; major {
	case 1:
		if minor := v[1]; minor > stlmig1.Minor {
			return nil, fmt.Errorf("unsupported minor: %d", minor)
		}
		return stlmig1.New(tx), nil
	default:
		return nil, fmt.Errorf("unsupported major: %d", major)
	}
}
