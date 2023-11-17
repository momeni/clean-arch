// Copyright (c) 2023-2024 Behnam Momeni
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Package carsuc contains the cars UseCase which supports the
// cars related use cases. Currently, two uses cases are supported:
//  1. Riding a car,
//  2. Parking a car.
package carsuc

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/momeni/clean-arch/pkg/core/cerr"
	"github.com/momeni/clean-arch/pkg/core/model"
	"github.com/momeni/clean-arch/pkg/core/repo"
)

// UseCase represents a cars use case. It holds a database connection
// pool, the cars repository instance (to be guided with the DB pool),
// and the cars use case specific settings.
type UseCase struct {
	pool   repo.Pool
	carsrp repo.Cars

	oldParkingMethodDelay time.Duration
}

// New instantiates a cars use case.
// Required parameters are passed individually, so caller has to
// provision them and whenever they change, caller will notice and fix
// them due to a compilation error.
// Optional parameters are passed as a series of functional options
// in order to facilitate their validation and flexibility.
func New(p repo.Pool, c repo.Cars, opts ...Option) (*UseCase, error) {
	uc := &UseCase{pool: p, carsrp: c}
	for _, opt := range opts {
		if err := opt(uc); err != nil {
			return nil, fmt.Errorf("invalid option: %w", err)
		}
	}
	// now, deal with defaults
	if uc.oldParkingMethodDelay == 0 {
		uc.oldParkingMethodDelay = 10 * time.Second
	}
	return uc, nil
}

// Ride use case unparks the cid car and moves it to the given
// destination geographical location. Updated car model and
// possible errors are returned.
func (cars *UseCase) Ride(ctx context.Context, cid uuid.UUID, destination model.Coordinate) (car *model.Car, err error) {
	err = cars.pool.Conn(ctx, func(ctx context.Context, c repo.Conn) error {
		q := cars.carsrp.Conn(c)
		car, err = q.UnparkAndMove(ctx, cid, destination)
		return err
	})
	if err != nil {
		car = nil
	}
	return
}

// Park use case tries to park the cid car using the mode parking mode.
// The new parking mode works quickly while the old method incurs delay
// based on the configuration. It returns the updated car model and
// possible errors.
func (cars *UseCase) Park(ctx context.Context, cid uuid.UUID, mode model.ParkingMode) (car *model.Car, err error) {
	err = mode.Validate()
	if err != nil {
		return nil, cerr.BadRequest(err)
	}
	if mode == model.ParkingModeOld {
		time.Sleep(cars.oldParkingMethodDelay) // old method is slow :)
	}
	err = cars.pool.Conn(ctx, func(ctx context.Context, c repo.Conn) error {
		q := cars.carsrp.Conn(c)
		car, err = q.Park(ctx, cid, mode)
		return err
	})
	if err != nil {
		car = nil
	}
	return
}
