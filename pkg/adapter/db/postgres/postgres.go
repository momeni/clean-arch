// Copyright (c) 2024 Behnam Momeni
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package postgres

import (
	"github.com/momeni/clean-arch/pkg/adapter/db/postgres/migration/settle/stlmig1"
	"github.com/momeni/clean-arch/pkg/core/model"
)

// These constants represent the major, minor, and patch components of
// the current database schema semantic version. Since each schema major
// version is backed by one stlmigN package for its migration settlement
// and intitialization operations, the latest version can be taken from
// that package (for the largest supported N major version) too.
//
// The v1.1.0 is the latest supported database schema version.
const (
	Major = stlmig1.Major // latest supported schema major version
	Minor = stlmig1.Minor // latest schema minor version in Major series
	Patch = stlmig1.Patch // latest schema patch version in Minor series
)

// Version is the latest supported database schema semantic version.
var Version = model.SemVer{Major, Minor, Patch}
