// Copyright (c) 2023-2024 Behnam Momeni
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package repo

import "context"

// TxHandler is a handler function which takes a context and an ongoing
// transaction. If an error is returned, caller will rollback the
// transaction and in absence of errors, it will be committed.
type TxHandler func(context.Context, Tx) error

// Conn represents a database connection.
// It is unsafe to be used concurrently. A connection may be used
// in order to execute one or more SQL statements or start transactions
// one at a time.
// For statement execution methods, see the Queryer interface.
type Conn interface {
	Queryer

	// Tx begins a new transaction in this connection, calls the handler
	// with the ctx (which was used for beginning the transactions) and
	// the fresh transaction, and commits the transaction ultimately.
	// In case of errors, the transaction will be rolled back and the
	// error will be returned too.
	Tx(ctx context.Context, handler TxHandler) error

	// IsConn method prevents a non-Conn object to mistakenly implement
	// the Conn interface.
	IsConn()
}
