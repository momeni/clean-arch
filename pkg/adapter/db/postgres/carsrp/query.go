package carsrp

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/momeni/clean-arch/pkg/adapter/db/postgres"
	"github.com/momeni/clean-arch/pkg/core/model"
)

type flattenedCar struct {
	Name     string
	Lat, Lon float64
	Parked   bool
}

func UnparkAndMove[Q postgres.Queryer](ctx context.Context, q Q, carID uuid.UUID, c model.Coordinate) (*model.Car, error) {
	rows, err := q.Query(
		ctx,
		`UPDATE cars
SET lat=$1, lon=$2, parked=false
WHERE cid=$3
RETURNING *`,
		c.Lat, c.Lon, carID,
	)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	r := rows.(pgx.Rows)
	res, err := pgx.CollectOneRow(r, pgx.RowToAddrOfStructByName[flattenedCar])
	if err != nil {
		return nil, fmt.Errorf("collecting a row: %w", err)
	}
	return &model.Car{
		Name:       res.Name,
		Coordinate: model.Coordinate{Lat: res.Lat, Lon: res.Lon},
		Parked:     res.Parked,
	}, nil
}

func Park[Q postgres.Queryer](ctx context.Context, q Q, carID uuid.UUID) (*model.Car, error) {
	rows, err := q.Query(
		ctx,
		`UPDATE cars
SET parked=true
WHERE cid=$1
RETURNING *`,
		carID,
	)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	r := rows.(pgx.Rows)
	res, err := pgx.CollectOneRow(r, pgx.RowToAddrOfStructByName[flattenedCar])
	if err != nil {
		return nil, fmt.Errorf("collecting a row: %w", err)
	}
	return &model.Car{
		Name:       res.Name,
		Coordinate: model.Coordinate{Lat: res.Lat, Lon: res.Lon},
		Parked:     res.Parked,
	}, nil
}
