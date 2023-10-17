// Copyright (c) 2023 Behnam Momeni
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package repo

// Tx represents a database transaction.
// It is unsafe to be used concurrently. A transaction may be used
// in order to execute one or more SQL statements one at a time.
// For statement execution methods, see the Queryer interface.
// All statements which are in a single transaction observe the
// ACID properties. The exact amount of isolation between transactions
// depends on their types. By default, a READ-COMMITTED transaction is
// expected from a PostgreSQL DBMS server. For details, read
// https://www.postgresql.org/docs/current/transaction-iso.html#XACT-READ-COMMITTED
type Tx interface {
	Queryer

	// IsTx method prevents a non-Tx object (such as a Conn) to
	// mistakenly implement the Tx interface.
	IsTx()
}
