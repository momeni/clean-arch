package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/momeni/clean-arch/pkg/core/repo"
)

type Tx struct {
	pgx.Tx
}

func (tx *Tx) Exec(ctx context.Context, sql string, args ...any) (int64, error) {
	tag, err := tx.Tx.Exec(ctx, sql, args...)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

func (tx *Tx) Query(ctx context.Context, sql string, args ...any) (repo.Rows, error) {
	rows, err := tx.Tx.Query(ctx, sql, args...)
	return rows, err
}
