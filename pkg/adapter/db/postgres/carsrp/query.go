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
	CID        uuid.UUID `gorm:"primaryKey;type:uuid;column:cid"`
	Name       string
	Coordinate model.Coordinate `gorm:"embedded"`
	Parked     bool
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
		"lat", "lon", "parked",
	).Where(
		"cid=?", carID,
	).Updates(gCar{
		Coordinate: c,
		Parked:     false,
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
// changing its current location. It returns the updated can model
// and possible errors.
// This generic function allows a unified implementation to be used
// for both of the connection and transaction receiving methods.
func Park[Q postgres.Queryer](ctx context.Context, q Q, carID uuid.UUID) (*model.Car, error) {
	gdb := q.GORM(ctx)
	var gc []gCar
	gdb.Model(&gc).Clauses(clause.Returning{}).Select(
		"parked",
	).Where(
		"cid=?", carID,
	).Updates(gCar{
		Parked: true,
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
