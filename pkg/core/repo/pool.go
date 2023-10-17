// Copyright (c) 2023 Behnam Momeni
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Package repo specifies the expected interfaces for management
// of repositories, including a database connections pool which can
// be used concurrently by several goroutines, how individual
// connections and transactions may be obtained from it, which
// repositories exist, and which actions may be performed on them.
package repo

import "context"

// ConnHandler is a handler function which takes a context and a
// database connection which should be used solely from the current
// goroutine (or by proper synchronization). When it returns, the
// connection may be released and reused by other routines.
type ConnHandler func(context.Context, Conn) error

// Pool represents a database connection pool.
// It may be used concurrently from different goroutines.
type Pool interface {
	// Conn acquires a database connection, passes it into the
	// handler function, and when it returns will release the connection
	// so it may be used by other callers.
	// This method may be blocked (as while as the ctx allows it)
	// until a connection is obtained. That connection will not be
	// used by any other handler concurrently.
	// Returned errors from the handler will be returned by this
	// method after possible wrapping.
	// The ctx which is used for acquisition of a connection is also
	// passed to the handler function.
	Conn(ctx context.Context, handler ConnHandler) error
}
