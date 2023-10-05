package postgres

import (
	"context"

	"github.com/momeni/clean-arch/pkg/core/repo"
	"gorm.io/gorm"
)

type Tx struct {
	*gorm.DB
}

func (tx *Tx) Exec(ctx context.Context, sql string, args ...any) (int64, error) {
	tt := tx.DB.WithContext(ctx).Exec(sql, args...)
	if err := tt.Error; err != nil {
		return 0, err
	}
	return tt.RowsAffected, nil
}

func (tx *Tx) Query(ctx context.Context, sql string, args ...any) (repo.Rows, error) {
	rows, err := tx.DB.WithContext(ctx).Raw(sql, args...).Rows()
	return rowsAdapter{rows}, err
}

func (tx *Tx) IsTx() {
}

func (tx *Tx) GORM(ctx context.Context) *gorm.DB {
	return tx.DB.WithContext(ctx)
}
