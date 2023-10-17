// Copyright (c) 2023 Behnam Momeni
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package repo

import (
	"context"

	"github.com/google/uuid"
	"github.com/momeni/clean-arch/pkg/core/model"
)

// CarsConnQueryer interface lists all operations which may be executed
// in a Cars repository having an open connection with auto-committed
// transactions.
// Those operations which must be executed in a connection (and may not
// be executed in an ongoing transaction which may keep running other
// statements after this one) must be listed here, while other
// operations which do not strictly require an open connection (and may
// use an open transaction too) must be defined in the embedded
// CarsQueryer interface. This design allows a unified implementation,
// while forcing developers to think about the consequences of having
// one or multiple transactions.
type CarsConnQueryer interface {
	CarsQueryer
}

// CarsTxQueryer interface lists all operations which may be executed
// in a Cars repository having an ongoing transaction.
// Those operations which must be executed in a transaction (and may not
// be executed with a connection) must be listed here, while other
// operations which do not strictly require an open transaction (and
// can use their own auto-committed transaction too) must be defined
// in the embedded CarsQueryer interface. This design allows a unified
// implementation, while forcing developers to think about the
// consequences of having one or multiple transactions.
type CarsTxQueryer interface {
	CarsQueryer
}

// CarsQueryer interface lists common operations which may be executed
// in a Cars repository having either a connection or transaction at
// hand. This interface is embedded by both of CarsConnQueryer and
// CarsTxQueryer in order to avoid redundant implementation.
type CarsQueryer interface {
	// UnparkAndMove example operation unparks a car with carID UUID,
	// and moves it to the c destination coordinate. Updated car model
	// and possible errors are returned.
	UnparkAndMove(ctx context.Context, carID uuid.UUID, c model.Coordinate) (*model.Car, error)

	// Park example operation parks the car with carID UUID without
	// changing its current location. It returns the updated can model
	// and possible errors.
	Park(ctx context.Context, carID uuid.UUID) (*model.Car, error)
}

// Cars interface represents an example repository for management of
// the car instances. A repository interface should provide two methods
// of Conn and Tx in order to encourage developer to explicitly decide
// that a connection or a transaction is required for execution of a
// SQL statement. Each of those two methods will take a Conn/Tx
// interface which was provided by the repository implementation (from
// the adapter layer) beforehand. Implementation of these Conn()/Tx()
// methods may safely unwrap these interfaces and access the underlying
// structs if needed, hence, the unwrapping is performed just once.
type Cars interface {
	// Conn takes a Conn interface instance, unwraps it as required,
	// and returns a CarsConnQueryer interface which (with access to
	// the implementation-dependent connection object) can run different
	// permitted operations on cars.
	Conn(Conn) CarsConnQueryer

	// Tx takes a Tx interface instance, unwraps it as required,
	// and returns a CarsTxQueryer interface which (with access to the
	// implementation-dependent transaction object) can run different
	// permitted operations on cars.
	Tx(Tx) CarsTxQueryer
}
