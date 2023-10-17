// Copyright (c) 2023 Behnam Momeni
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Package carsrp is the adapter for the cars repository.
// It exposes the carsrp.Repo type in order to allow use cases
// to manage car instances.
package carsrp

import (
	"context"

	"github.com/google/uuid"
	"github.com/momeni/clean-arch/pkg/adapter/db/postgres"
	"github.com/momeni/clean-arch/pkg/core/model"
	"github.com/momeni/clean-arch/pkg/core/repo"
)

// Repo represents the cars repository instance.
type Repo struct {
}

// New instantiates a cars Repo struct. Although this New does not
// perform complex operations, and users may use &carsrp.Repo{} directly
// too, but this method improves the code readability as carsrp.New()
// making the package to look alike a data type.
func New() *Repo {
	return &Repo{}
}

type connQueryer struct {
	*postgres.Conn
}

// Conn takes a Conn interface instance, unwraps it as required,
// and returns a CarsConnQueryer interface which (with access to
// the implementation-dependent connection object) can run different
// permitted operations on cars.
// The connQueryer itself is not mentioned as the return value since
// it is not exported. Otherwise, the general rule is to take interfaces
// as arguments and return exported structs.
func (cars *Repo) Conn(c repo.Conn) repo.CarsConnQueryer {
	cc := c.(*postgres.Conn)
	return connQueryer{Conn: cc}
}

// UnparkAndMove example operation unparks a car with carID UUID,
// and moves it to the c destination coordinate. Updated car model
// and possible errors are returned.
// This method calls a generic function, so the actual implementation
// can be coded at one place for both of the connection and transaction
// receiving methods.
func (cq connQueryer) UnparkAndMove(ctx context.Context, carID uuid.UUID, c model.Coordinate) (*model.Car, error) {
	return UnparkAndMove(ctx, cq.Conn, carID, c)
}

// Park example operation parks the car with carID UUID without
// changing its current location. It returns the updated can model
// and possible errors.
// This method calls a generic function, so the actual implementation
// can be coded at one place for both of the connection and transaction
// receiving methods.
func (cq connQueryer) Park(ctx context.Context, carID uuid.UUID) (*model.Car, error) {
	return Park(ctx, cq.Conn, carID)
}

type txQueryer struct {
	*postgres.Tx
}

// Tx takes a Tx interface instance, unwraps it as required,
// and returns a CarsTxQueryer interface which (with access to the
// implementation-dependent transaction object) can run different
// permitted operations on cars.
// The txQueryer itself is not mentioned as the return value since
// it is not exported. Otherwise, the general rule is to take interfaces
// as arguments and return exported structs.
func (cars *Repo) Tx(tx repo.Tx) repo.CarsTxQueryer {
	tt := tx.(*postgres.Tx)
	return txQueryer{Tx: tt}
}

// UnparkAndMove example operation unparks a car with carID UUID,
// and moves it to the c destination coordinate. Updated car model
// and possible errors are returned.
// This method calls a generic function, so the actual implementation
// can be coded at one place for both of the connection and transaction
// receiving methods.
func (tq txQueryer) UnparkAndMove(ctx context.Context, carID uuid.UUID, c model.Coordinate) (*model.Car, error) {
	return UnparkAndMove(ctx, tq.Tx, carID, c)
}

// Park example operation parks the car with carID UUID without
// changing its current location. It returns the updated can model
// and possible errors.
// This method calls a generic function, so the actual implementation
// can be coded at one place for both of the connection and transaction
// receiving methods.
func (tq txQueryer) Park(ctx context.Context, carID uuid.UUID) (*model.Car, error) {
	return Park(ctx, tq.Tx, carID)
}
