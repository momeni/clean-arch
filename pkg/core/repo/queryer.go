// Copyright (c) 2023-2024 Behnam Momeni
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package repo

import "context"

// Queryer interface includes methods for running SQL statements.
// There are two main types of statements. One category may affect
// multiple rows, but do not return a result set, like DDL commands
// or an UPDATE without a RETURNING clause. These statements are
// executed with the Exec method. Another category of statements
// (which may or may not modify the database contents) provide a result
// set, like SELECT or an UPDATE with a RETURNING clause. These queries
// may be executed with the Query method.
// This interface is embedded by both of Conn and Tx since they may be
// used for execution of commands, of course, with distinct isolation
// levels.
type Queryer interface {
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
	// This should not be confused with ? or @name parameters which are
	// supported by some ORM libraries via query rewriting. Those extra
	// formats may be supported by an implementation, however, users of
	// this interface should solely use the native numbered placeholders
	// in order to stay independent of the adapter/frameworks layers.
	Exec(ctx context.Context, sql string, args ...any) (count int64, err error)

	// Query runs SQL statement with given args given ctx context.
	// The result set is returned as the Rows interface, while errors
	// are returned as the second return value (if any).
	// If args is provided, sql will be prepared and args will be passed
	// separately to the DBMS in order to prevent SQL injection.
	// Nevertheless, sql must contain exactly one statement.
	//
	// Parameters in sql should be numbered like $1, $2, etc. as they
	// are supported by the PostgreSQL wire protocol natively.
	// This should not be confused with ? or @name parameters which are
	// supported by some ORM libraries via query rewriting. Those extra
	// formats may be supported by an implementation, however, users of
	// this interface should solely use the native numbered placeholders
	// in order to stay independent of the adapter/frameworks layers.
	//
	// The Query or Exec may not be called again until the Rows is
	// closed since only one ongoing statement may be used on each
	// connection. If you need to run multiple queries concurrently,
	// either use multiple connections or rewrite the query using
	// the CURSOR concept:
	// https://www.postgresql.org/docs/current/plpgsql-cursors.html
	Query(ctx context.Context, sql string, args ...any) (Rows, error)
}

// Rows represents the result set of an executed query.
// Before trying to read a row out of the Rows instance, the Next()
// must be called. As while as the Next() returns true, a row is
// obtained from the DBMS server (multiple rows may be fetched and
// cached based on the implementation details).
// The Scan or Values methods may be used in order to read columns
// of the current row. Calling the Close() method releases the result
// set and ignores any remaining unread rows. The Err() method must be
// called after the Close() in order to obtain possible errors.
// Before calling the Close, Err may return nil even if there are some
// errors since they may not be detected necessarily beforehand.
type Rows interface {
	// Close closes the result set, allowing the next query to be
	// executed. After calling this method, the Err() method may be
	// called in order to find out about possible errors.
	Close()

	// Err returns any error which was seen during the last operation.
	// It has to be checked after calling the Close() method.
	Err() error

	// Next prepares the next row to be read using the Scan or Values
	// methods. It must be called even before reading the first row.
	// If it returns false, no more rows exists. In this case, the
	// result set is closed automatically and it is not required to
	// call the Close() method. Nevertheless, calling the Close() is
	// harmless and the Err() method must be checked yet.
	Next() bool

	// Scan reads the values of all columns from the current row
	// into the given dest arguments.
	//
	// Scan converts columns read from the database into the following
	// common Go types and special types provided by the sql package:
	//
	//	*string
	//	*[]byte
	//	*int, *int8, *int16, *int32, *int64
	//	*uint, *uint8, *uint16, *uint32, *uint64
	//	*bool
	//	*float32, *float64
	//	*interface{}
	//	*RawBytes
	//	*Rows (cursor value)
	//	any type implementing Scanner (see Scanner docs)
	Scan(dest ...any) error

	// Values is like Scan, but does not need the caller to prepare
	// destination arguments with relevant types. Instead, it uses
	// *interface{} types in order to read all column values.
	// It is recommended to use the Scan in order to eliminate the
	// subsequent type checking codes.
	Values() ([]any, error)
}
