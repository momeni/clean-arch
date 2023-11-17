// Copyright (c) 2024 Behnam Momeni
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Package up provides the common aspects among all upwards database
// schema migrators. This package may be imported by its sub-packages
// in order to provide the Migrator[S, U] generic interface for them.
package up

import "context"

// Migrator of S and U is a generic interface which presents
// expectations from a version-dependent database schema upwards
// migrator object.
// Each upwards migrator needs a MigrateUp method which is responsible
// to migrate one major version upwards (staying at the latest supported
// minor version). The U type parameter indicates the next upwards
// migrator type and S type parameter indicates the settler object type.
// After reaching to the target major version, Settler method can be
// used to obtain the corresponding settler object with type S and
// persist the migration results.
//
// This generic interface must be implemented by all upmigN.Migrator
// types. It is useful for causing compile-time errors and catching
// programming errors when an old upmigN package is copied, but its
// methods are not updated properly in order to reflect the new version.
type Migrator[S any, U any] interface {
	// Settler returns a settler object (with S type) without performing
	// any migration action (so, no error condition may arise). Returned
	// settler object may be employed to persist the migration results.
	Settler() S

	// MigrateUp migrates from current major version to the next major
	// version by creating relevant views in a schema such as migN
	// based on the views in a schema such as migM where N=M+1
	// considering the latest supported minor versions of those N and M
	// major versions. It may be necessary to create tables and
	// materialize data items or load them in the Golang process memory
	// in more complex scenarios, but using views is preferred as it
	// postpones the actual data copying till the last schema and
	// the settlement phase.
	// It then creates another upwards migrator object (with U type)
	// which can be used for continuing the upwards migration from
	// the next major version similarly.
	MigrateUp(ctx context.Context) (U, error)
}
