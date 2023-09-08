package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/momeni/clean-arch/pkg/core/repo"
)

type Pool struct {
	*pgxpool.Pool
}

func NewPool(ctx context.Context, url string) (*Pool, error) {
	p, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("pgxpool.New: %w", err)
	}
	pool := &Pool{Pool: p}
	err = pool.Conn(ctx, NoOpConnHandler)
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("testing connection: %w", err)
	}
	return pool, nil
}

type ConnHandler = repo.ConnHandler

func NoOpConnHandler(context.Context, repo.Conn) error {
	return nil
}

func (p *Pool) Conn(ctx context.Context, f ConnHandler) error {
	return p.Pool.AcquireFunc(ctx, func(c *pgxpool.Conn) error {
		return f(ctx, &Conn{Conn: c})
	})
}
