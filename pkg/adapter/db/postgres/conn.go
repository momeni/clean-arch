package postgres

import (
	"context"
	"fmt"

	"github.com/momeni/clean-arch/pkg/core/repo"
	"gorm.io/gorm"
)

// Conn represents a database connection.
// It is unsafe to be used concurrently. A connection may be used
// in order to execute one or more SQL statements or start transactions
// one at a time.
// For statement execution methods, see the Queryer interface.
// Conn embeds the *gorm.DB, hence, may be used like GORM from within
// the repository packages (which can depend on frameworks).
type Conn struct {
	*gorm.DB
}

// TxHandler is a handler function which takes a context and an ongoing
// transaction. If an error is returned, caller will rollback the
// transaction and in absence of errors, it will be committed.
type TxHandler = repo.TxHandler

// Tx begins a new transaction in this connection, calls the f handler
// with the ctx (which was used for beginning the transactions) and
// the fresh transaction, and commits the transaction ultimately.
// In case of errors, the transaction will be rolled back and the
// error will be returned too.
func (c *Conn) Tx(ctx context.Context, f TxHandler) (err error) {
	tx := c.DB.WithContext(ctx).Begin()
	if err = tx.Error; err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if r := recover(); r != nil {
			err = tx.Rollback().Error
			if err == nil {
				err = fmt.Errorf("panicked: %v", r)
				return
			}
			err = fmt.Errorf("panicked: %v, rollback: %w", r, err)
			return
		}
		if err != nil {
			if err2 := tx.Rollback().Error; err2 != nil {
				err = fmt.Errorf("handler: %w, rollback: %w", err, err2)
				return
			}
			err = fmt.Errorf("handler: %w", err)
			return
		}
		err = tx.Commit().Error
		if err != nil {
			err = fmt.Errorf("commit: %w", err)
		}
	}()
	tt := &Tx{DB: tx}
	return f(ctx, tt)
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
func (c *Conn) Exec(ctx context.Context, sql string, args ...any) (int64, error) {
	tt := c.DB.WithContext(ctx).Exec(sql, args...)
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
func (c *Conn) Query(ctx context.Context, sql string, args ...any) (repo.Rows, error) {
	rows, err := c.DB.WithContext(ctx).Raw(sql, args...).Rows()
	return rowsAdapter{rows}, err
}

// IsConn method prevents a non-Conn object to mistakenly implement
// the Conn interface.
func (c *Conn) IsConn() {
}

// GORM returns the embedded *gorm.DB instance, configuring it
// to operate on the given ctx context (in a gorm.Session).
func (c *Conn) GORM(ctx context.Context) *gorm.DB {
	return c.DB.WithContext(ctx)
}
