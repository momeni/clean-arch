package repo

import (
	"context"

	"github.com/google/uuid"
	"github.com/momeni/clean-arch/pkg/core/model"
)

type CarsConnQueryer interface {
	CarsQueryer
}

type CarsTxQueryer interface {
	CarsQueryer
}

type CarsQueryer interface {
	UnparkAndMove(ctx context.Context, carID uuid.UUID, c model.Coordinate) (*model.Car, error)
	Park(ctx context.Context, carID uuid.UUID) (*model.Car, error)
}

type Cars interface {
	Conn(Conn) CarsConnQueryer
	Tx(Tx) CarsTxQueryer
}
