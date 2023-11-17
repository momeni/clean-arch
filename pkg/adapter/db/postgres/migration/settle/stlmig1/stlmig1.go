// Copyright (c) 2024 Behnam Momeni
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Package stlmig1 provides Settler type for database schema major
// version 1 with two main usages. First, it can be used to initialize
// a database with major version 1 schema, having development or
// production suitable sample data. Second, it can be used to settle
// a multi-database schema migration operation by reading views from
// an intermediate schema and filling tables in the target schema
// without converting their format (e.g., column names) in order to
// persist the migration results.
package stlmig1

import (
	"context"

	"github.com/momeni/clean-arch/pkg/core/repo"
)

// These constants indicate the major, minor, and patch components of
// the database schema migration settler implementation. Each major
// version has a separate stlmigN package and the Minor is the latest
// supported minor version within the Major major version series.
const (
	Major = 1
	Minor = 1
	Patch = 0
)

// Settler struct provides the database schema migration settlement
// logic for the major version 1. Settlement is the last phase of
// migration which persists the prepared database views contents into
// their corresponding tables. See the SettleSchema method for this
// use case. It can also be used for creation and filling of tables
// with the development and production suitable initial data. Check
// the InitDevSchema and InitProdSchema methods for this purpose.
//
// Each instance of Settler wraps and uses a single transaction of the
// destination database, but the caller is responsible to commit that
// transaction in order to finalize the settlement or initialization
// operation results.
type Settler struct {
	tx repo.Tx // destination database transaction
}

// New creates a new Settler instance, wrapping the given `tx` database
// transaction. The settler object expects the database schema to exist
// and only tries to create relevant tables in that schema.
func New(tx repo.Tx) *Settler {
	return &Settler{
		tx: tx,
	}
}

// SettleSchema creates major version 1 tables in the caweb1 schema
// (representing the v1.x schema) and fills them with the contents of
// those database views which are prepared in the mig1 schema.
// The mig1 schema and caweb1 schema have views and tables with the
// same format, so no conversion will happen. Ideally, this is the
// only operation which copies all user data from the source database
// and converts them by passing through the fdwN, migN, ..., migM, to
// this mig1 and then persists them in tables of caweb1 schema.
func (sm1 *Settler) SettleSchema(ctx context.Context) error {
	panic("not implemented yet") // TODO: Implement
}

// InitDevSchema creates major version 1 tables in caweb1 schema and
// fills them with the development suitable initial data.
func (sm1 *Settler) InitDevSchema(ctx context.Context) error {
	panic("not implemented yet") // TODO: Implement
}

// InitProdSchema creates major version 1 tables in caweb1 schema and
// fills them with the production suitable initial data.
func (sm1 *Settler) InitProdSchema(ctx context.Context) error {
	panic("not implemented yet") // TODO: Implement
}

// MajorVersion returns the major semantic version of this Settler
// instance. This value matches with the Major constant which is defined
// in this package. Indeed, this method can be called with a nil
// instance too because it only depends on the Settler type (not its
// instance).
func (sm1 *Settler) MajorVersion() uint {
	return Major
}
