package carsuc

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/momeni/clean-arch/pkg/core/cerr"
	"github.com/momeni/clean-arch/pkg/core/model"
	"github.com/momeni/clean-arch/pkg/core/repo"
)

type UseCase struct {
	pool   repo.Pool
	carsrp repo.Cars
}

func New(p repo.Pool, c repo.Cars) *UseCase {
	return &UseCase{pool: p, carsrp: c}
}

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

func (cars *UseCase) Park(ctx context.Context, cid uuid.UUID, mode model.ParkingMode) (car *model.Car, err error) {
	err = mode.Validate()
	if err != nil {
		return nil, cerr.BadRequest(err)
	}
	if mode == model.ParkingModeOld {
		time.Sleep(1 * time.Minute) // old method is slow :)
	}
	err = cars.pool.Conn(ctx, func(ctx context.Context, c repo.Conn) error {
		q := cars.carsrp.Conn(c)
		car, err = q.Park(ctx, cid)
		return err
	})
	if err != nil {
		car = nil
	}
	return
}
