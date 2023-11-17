// Copyright (c) 2024 Behnam Momeni
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Package down provides the common aspects among all downwards database
// schema migrators. This package may be imported by its sub-packages
// in order to provide the Migrator[S, D] generic interface for them.
package down

import "context"

// Migrator of S and D is a generic interface which presents
// expectations from a version-dependent database schema downwards
// migrator object.
// Each downwards migrator needs a MigrateDown method which is
// responsible to migrate one major version downwards (staying at the
// latest supported minor version). The D type parameter indicates the
// previous downwards migrator type and S type parameter indicates the
// settler object type. After reaching to the target major version,
// Settler method can be used to obtain the corresponding settler
// object with type S and persist the migration results.
//
// This generic interface must be implemented by all dnmigN.Migrator
// types. It is useful for causing compile-time errors and catching
// programming errors when an old dnmigN package is copied, but its
// methods are not updated properly in order to reflect the new version.
type Migrator[S any, D any] interface {
	// Settler returns a settler object (with S type) without performing
	// any migration action (so, no error condition may arise). Returned
	// settler object may be employed to persist the migration results.
	Settler() S

	// MigrateDown migrates from current major version to the previous
	// major version by creating relevant views in a schema such as migN
	// based on the views in a schema such as migM where N=M-1
	// considering the latest supported minor versions of those N and M
	// major versions. It may be necessary to create tables and
	// materialize data items or load them in the Golang process memory
	// in more complex scenarios, but using views is preferred as it
	// postpones the actual data copying till the last schema and
	// the settlement phase.
	// It then creates another downwards migrator object (with D type)
	// which can be used for continuing the downwards migration from
	// the previous major version similarly.
	MigrateDown(ctx context.Context) (D, error)
}
