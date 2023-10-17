// Copyright (c) 2023 Behnam Momeni
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package postgres

import (
	"context"

	"github.com/momeni/clean-arch/pkg/core/repo"
	"gorm.io/gorm"
)

// Tx represents a database transaction.
// It is unsafe to be used concurrently. A transaction may be used
// in order to execute one or more SQL statements one at a time.
// For statement execution methods, see the Queryer interface.
// All statements which are in a single transaction observe the
// ACID properties. The exact amount of isolation between transactions
// depends on their types. By default, a READ-COMMITTED transaction is
// expected from a PostgreSQL DBMS server. For details, read
// https://www.postgresql.org/docs/current/transaction-iso.html#XACT-READ-COMMITTED
// Tx embeds the *gorm.DB, hence, may be used like GORM from within
// the repository packages (which can depend on frameworks).
type Tx struct {
	*gorm.DB
}

// Exec runs SQL statements with given args given ctx context.
// Number of affected rows and possible errors will be returned.
// If args is provided, sql will be prepared and args will be passed
// separately to the DBMS in order to prevent SQL injection.
// In this case, sql must contain exactly one statement.
// In absence of args, sql may contain multiple semi-colon separated
// statements too.
//
// Parameters in sql should be numbered like $1, $2, etc. as they
// are supported by the PostgreSQL wire protocol natively.
// This implementation additionally supports the ? and @name parameter
// placeholders using the GORM framework.
func (tx *Tx) Exec(ctx context.Context, sql string, args ...any) (int64, error) {
	tt := tx.DB.WithContext(ctx).Exec(sql, args...)
	if err := tt.Error; err != nil {
		return 0, err
	}
	return tt.RowsAffected, nil
}

// Query runs SQL statement with given args given ctx context.
// The result set is returned as the Rows interface, while errors
// are returned as the second return value (if any).
// If args is provided, sql will be prepared and args will be passed
// separately to the DBMS in order to prevent SQL injection.
// Nevertheless, sql must contain exactly one statement.
//
// Parameters in sql should be numbered like $1, $2, etc. as they
// are supported by the PostgreSQL wire protocol natively.
// This implementation additionally supports the ? and @name parameter
// placeholders using the GORM framework.
//
// The Query or Exec may not be called again until the Rows is
// closed since only one ongoing statement may be used on each
// connection. If you need to run multiple queries concurrently,
// either use multiple connections or rewrite the query using
// the CURSOR concept:
// https://www.postgresql.org/docs/current/plpgsql-cursors.html
func (tx *Tx) Query(ctx context.Context, sql string, args ...any) (repo.Rows, error) {
	rows, err := tx.DB.WithContext(ctx).Raw(sql, args...).Rows()
	return rowsAdapter{rows}, err
}

// IsTx method prevents a non-Tx object (such as a Conn) to
// mistakenly implement the Tx interface.
func (tx *Tx) IsTx() {
}

// GORM returns the embedded *gorm.DB instance, configuring it
// to operate on the given ctx context (in a gorm.Session).
func (tx *Tx) GORM(ctx context.Context) *gorm.DB {
	return tx.DB.WithContext(ctx)
}
