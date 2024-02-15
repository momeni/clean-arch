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
	"errors"
	"testing"

	"github.com/momeni/clean-arch/pkg/adapter/db/postgres/migration/settle/stlmig1"
	"github.com/momeni/clean-arch/pkg/core/repo"
	"github.com/stretchr/testify/assert"
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

// errRollback will be returned from the test transactions after
// running all test queries successfully in order to roll back that
// transaction.
var errRollback = errors.New("rollback test transaction")

// VerifySchema uses the corresponding database connection of `v` in
// order to create and query temporary records in the database (e.g., in
// an uncommitted transaction), ensuring that the expected database
// schema with Major major version and Minor minor version is in place.
// If a more recent minor version was settled, this verification will
// pass too, so it is important to update this implementation whenever
// a new minor version is released.
// This process failures are reported using the `t` testing argument.
func (v *Verifier) VerifySchema(ctx context.Context, t *testing.T) {
	a := assert.New(t)
	err := v.c.Tx(ctx, func(ctx context.Context, tx repo.Tx) error {
		_, err := tx.Exec(ctx, `SET search_path TO caweb1;
INSERT INTO cars(cid, name, lat, lon, parked, parking_mode)
VALUES (
        '00000000-0000-0000-0000-000000000000',
        'test-name',
        1.1111,
        2.2222,
        true,
        'new'
    );
DO
$body$
BEGIN
    IF 'test-name' != (
            SELECT name
            FROM cars
            WHERE cid='00000000-0000-0000-0000-000000000000'
    ) THEN
        RAISE EXCEPTION 'cannot find the inserted record by PK';
    END IF;
END
$body$;`)
		if !a.NoError(err, "schema verification transaction failed") {
			return err
		}
		return errRollback
	})
	a.ErrorIs(err, errRollback, "unexpected transaction error")
}

// VerifyDevData checks for presence of the development suitable initial
// data and marks possible issues using the `t` testing argument.
// Presence of extra rows is acceptable.
func (v *Verifier) VerifyDevData(ctx context.Context, t *testing.T) {
	a := assert.New(t)
	err := v.c.Tx(ctx, func(ctx context.Context, tx repo.Tx) error {
		// Checking some records and ensuring that dev data are inserted
		// is enough (i.e., no comprehensive checking is required).
		rows, err := tx.Query(ctx, `SELECT
    cid, name,
    lat, lon,
    parked, parking_mode
FROM caweb1.cars
WHERE cid='e4f6b292-5dfe-4877-9cd2-7575d95825a8'
    OR cid='7c7d505d-2181-4352-90a8-c1426ed19159'
ORDER BY name ASC`)
		if !a.NoError(err, "selecting two sample dev rows") {
			return err
		}
		defer rows.Close()
		i := 0
		old := "old"
		expectedRows := [][]any{
			{
				"e4f6b292-5dfe-4877-9cd2-7575d95825a8", "Bugatti",
				26.239947, 55.147466,
				true, &old,
			},
			{
				"7c7d505d-2181-4352-90a8-c1426ed19159", "Nissan",
				25.880152, 55.023427,
				false, (*string)(nil),
			},
		}
		for rows.Next() {
			var (
				cid, name   string
				lat, lon    float64
				parked      bool
				parkingMode *string
			)
			err := rows.Scan(
				&cid, &name, &lat, &lon, &parked, &parkingMode,
			)
			if !a.NoError(err, "scanning row (i=%d)", i) {
				return err
			}
			a.Equal(
				expectedRows[i],
				[]any{cid, name, lat, lon, parked, parkingMode},
				"mismatch in row %d", i,
			)
			i++
		}
		err = rows.Err()
		if !a.NoError(err, "after scanning the result set completely") {
			return err
		}
		return errRollback
	})
	a.ErrorIs(err, errRollback, "unexpected transaction error")
}

// VerifyProdData checks for presence of the production suitable initial
// data and marks possible issues using the `t` testing argument.
// Presence of extra rows is acceptable.
func (v *Verifier) VerifyProdData(ctx context.Context, t *testing.T) {
	a := assert.New(t)
	err := v.c.Tx(ctx, func(ctx context.Context, tx repo.Tx) error {
		// Since no production rows are inserted, the VerifySchema
		// method is enough and there are no more applicable checks.
		return errRollback
	})
	a.ErrorIs(err, errRollback, "unexpected transaction error")
}
