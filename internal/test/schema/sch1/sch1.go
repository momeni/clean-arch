// Copyright (c) 2024 Behnam Momeni
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Package sch1 provides database schema major version 1 verification
// logic. This implementation may be instantiated indirectly using
// the github.com/momeni/clean-arch/internal/test/schema package.
package sch1

import (
	"context"
	"testing"

	"github.com/momeni/clean-arch/pkg/adapter/db/postgres/migration/settle/stlmig1"
	"github.com/momeni/clean-arch/pkg/core/repo"
)

// These constants present the relevant major, minor, and patch semantic
// versions of this schema verifier package. They are initialized based
// on the stlmig1 package because whenever a new minor version is
// released, the stlmig1 has to be updated based on it and this verifier
// needs to verify its updated changes too.
// Also, the stlmig1 constants are not used directly, so users of this
// package do not need to import it just for checking the provided
// semantic version components.
const (
	Major = stlmig1.Major
	Minor = stlmig1.Minor
	Patch = stlmig1.Patch
)

// Verifier implements the schema major version 1 verification logic. It
// implements github.com/momeni/clean-arch/internal/test/schema.Verifier
// interface and wraps a database connection as noted in New function.
type Verifier struct {
	c repo.Conn // database connection which is used for testing
}

// New instantiates a Verifier struct, wrapping the `c` database
// connection. Since Verifier fields are not exported, the New function
// is required for its initialization.
func New(c repo.Conn) *Verifier {
	return &Verifier{c}
}

// VerifySchema uses the corresponding database connection of `v` in
// order to create and query temporary records in the database (e.g., in
// an uncommitted transaction), ensuring that the expected database
// schema with Major major version and Minor minor version is in place.
// If a more recent minor version was settled, this verification will
// pass too, so it is important to update this implementation whenever
// a new minor version is released.
// This process failures are reported using the `t` testing argument.
func (v *Verifier) VerifySchema(ctx context.Context, t *testing.T) {
	panic("Not implemented yet") // TODO: Implement
}

// VerifyDevData checks for presence of the development suitable initial
// data and marks possible issues using the `t` testing argument.
// Presence of extra rows is acceptable.
func (v *Verifier) VerifyDevData(ctx context.Context, t *testing.T) {
	panic("Not implemented yet") // TODO: Implement
}

// VerifyProdData checks for presence of the production suitable initial
// data and marks possible issues using the `t` testing argument.
// Presence of extra rows is acceptable.
func (v *Verifier) VerifyProdData(ctx context.Context, t *testing.T) {
	panic("Not implemented yet") // TODO: Implement
}
