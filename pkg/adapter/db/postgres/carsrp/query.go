// Copyright (c) 2023-2024 Behnam Momeni
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package carsrp

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/momeni/clean-arch/pkg/adapter/db/postgres"
	"github.com/momeni/clean-arch/pkg/core/cerr"
	"github.com/momeni/clean-arch/pkg/core/model"
	"gorm.io/gorm/clause"
)

type gCar struct {
	CID         uuid.UUID `gorm:"primaryKey;type:uuid;column:cid"`
	Name        string
	Coordinate  model.Coordinate `gorm:"embedded"`
	Parked      bool
	ParkingMode *string
}

func (gc *gCar) TableName() string {
	return "cars"
}

func (gc *gCar) Model() *model.Car {
	return &model.Car{
		Name:       gc.Name,
		Coordinate: gc.Coordinate,
		Parked:     gc.Parked,
	}
}

// UnparkAndMove example operation unparks a car with carID UUID,
// and moves it to the c destination coordinate. Updated car model
// and possible errors are returned.
// This generic function allows a unified implementation to be used
// for both of the connection and transaction receiving methods.
func UnparkAndMove[Q postgres.Queryer](ctx context.Context, q Q, carID uuid.UUID, c model.Coordinate) (*model.Car, error) {
	gdb := q.GORM(ctx)
	var gc []gCar
	gdb.Model(&gc).Clauses(clause.Returning{}).Select(
		"lat", "lon", "parked", "parking_mode",
	).Where(
		"cid=?", carID,
	).Updates(gCar{
		Coordinate:  c,
		Parked:      false,
		ParkingMode: nil,
	})
	if err := gdb.Error; err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	if n := len(gc); n != 1 {
		return nil, cerr.NotFound(
			fmt.Errorf("expected one row, but got %d", n),
		)
	}
	return gc[0].Model(), nil
}

// Park example operation parks the car with carID UUID without
// changing its current location. It returns the updated car model
// and possible errors. The parking mode is recorded too.
// This generic function allows a unified implementation to be used
// for both of the connection and transaction receiving methods.
func Park[Q postgres.Queryer](
	ctx context.Context, q Q, carID uuid.UUID, mode model.ParkingMode,
) (*model.Car, error) {
	gdb := q.GORM(ctx)
	var gc []gCar
	modeStr := mode.String()
	gdb.Model(&gc).Clauses(clause.Returning{}).Select(
		"parked", "parking_mode",
	).Where(
		"cid=?", carID,
	).Updates(gCar{
		Parked:      true,
		ParkingMode: &modeStr,
	})
	if err := gdb.Error; err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	if n := len(gc); n != 1 {
		return nil, cerr.NotFound(
			fmt.Errorf("expected one row, but got %d", n),
		)
	}
	return gc[0].Model(), nil
}
