// Copyright (c) 2024 Behnam Momeni
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Package schema is a facade for database schema verifiers which can
// be used for testing purposes. Because database migrations start by
// migrating each schema from each major version upwards to its latest
// supported minor version before continuing to migrate its major
// version upwards or downwards, ultimate migrated schema will have
// the latest minor version of some major version. Therefore, the code
// which verifies the migration result depends on the major version and
// not the specific minor version. By the way, such verification only
// checks the schema itself because the schema contents may be lost
// during a migration based on the available or missing tables and
// columns in the migration path.
// Extra verifications (considering the schema contents) are only useful
// when the expected rows can be gussed unambiguously. For example,
// after a direct database initialization (instead of a multi-database
// migration), inserted contents may be checked too.
package schema

import (
	"context"
	"fmt"
	"testing"

	"github.com/momeni/clean-arch/internal/test/schema/sch1"
	"github.com/momeni/clean-arch/pkg/core/model"
	"github.com/momeni/clean-arch/pkg/core/repo"
)

// Verifier interface presents the database schema verifier expectations
// as they are provided by each major version specific implementation.
type Verifier interface {
	// VerifySchema verifies the database schema (such as tables and
	// their columns) using its wrapped database connection.
	// The `t` argument is marked as failed if the schema was invalid.
	// The schema contents (i.e., existing rows) are not checked.
	VerifySchema(ctx context.Context, t *testing.T)

	// VerifyDevData verifies the schema contents assuming that the
	// development suitable data items were inserted there. Only
	// presence of development data and not the absence of extra data
	// rows will be checked.
	VerifyDevData(ctx context.Context, t *testing.T)

	// VerifyProdData verifies the schema contents assuming that the
	// production suitable data items were inserted there. Only
	// presence of production data and not the absence of extra data
	// rows will be checked.
	VerifyProdData(ctx context.Context, t *testing.T)
}

// NewVerifier creates a new schema Verifier instance based on the
// given `v` semantic version. Since each major version has a distinct
// implementation and this function needs to return anyone of those
// struct instances (from the schN sub-packages), they are combined
// as the Verifier interface. If the major or minor versions of `v` are
// not supported, an error will be returned.
// Returned Verifier instance will wrap the `c` database connection.
func NewVerifier(c repo.Conn, v model.SemVer) (Verifier, error) {
	switch major := v[0]; major {
	case 1:
		if minor := v[1]; minor > sch1.Minor {
			return nil, fmt.Errorf("unsupported minor: %d", minor)
		}
		return sch1.New(c), nil
	default:
		return nil, fmt.Errorf("unsupported major: %d", major)
	}
}
