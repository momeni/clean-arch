package carsrp

import (
	"context"

	"github.com/google/uuid"
	"github.com/momeni/clean-arch/pkg/adapter/db/postgres"
	"github.com/momeni/clean-arch/pkg/core/model"
	"github.com/momeni/clean-arch/pkg/core/repo"
)

type Repo struct {
}

func New() *Repo {
	return &Repo{}
}

type connQueryer struct {
	*postgres.Conn
}

func (cars *Repo) Conn(c repo.Conn) repo.CarsConnQueryer {
	cc := c.(*postgres.Conn)
	return connQueryer{Conn: cc}
}

func (cq connQueryer) UnparkAndMove(ctx context.Context, carID uuid.UUID, c model.Coordinate) (*model.Car, error) {
	return UnparkAndMove(ctx, cq.Conn, carID, c)
}

func (cq connQueryer) Park(ctx context.Context, carID uuid.UUID) (*model.Car, error) {
	return Park(ctx, cq.Conn, carID)
}

type txQueryer struct {
	*postgres.Tx
}

func (cars *Repo) Tx(tx repo.Tx) repo.CarsTxQueryer {
	tt := tx.(*postgres.Tx)
	return txQueryer{Tx: tt}
}

func (tq txQueryer) UnparkAndMove(ctx context.Context, carID uuid.UUID, c model.Coordinate) (*model.Car, error) {
	return UnparkAndMove(ctx, tq.Tx, carID, c)
}

func (tq txQueryer) Park(ctx context.Context, carID uuid.UUID) (*model.Car, error) {
	return Park(ctx, tq.Tx, carID)
}
