package postgres

import (
	"context"
	"fmt"

	"github.com/momeni/clean-arch/pkg/core/repo"
	"gorm.io/gorm"
)

type Conn struct {
	*gorm.DB
}

type TxHandler = repo.TxHandler

func (c *Conn) Tx(ctx context.Context, f TxHandler) (err error) {
	tx := c.DB.WithContext(ctx).Begin()
	if err = tx.Error; err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if r := recover(); r != nil {
			err = tx.Rollback().Error
			if err == nil {
				err = fmt.Errorf("panicked: %v", r)
				return
			}
			err = fmt.Errorf("panicked: %v, rollback: %w", r, err)
			return
		}
		if err != nil {
			if err2 := tx.Rollback().Error; err2 != nil {
				err = fmt.Errorf("handler: %w, rollback: %w", err, err2)
				return
			}
			err = fmt.Errorf("handler: %w", err)
			return
		}
		err = tx.Commit().Error
		if err != nil {
			err = fmt.Errorf("commit: %w", err)
		}
	}()
	tt := &Tx{DB: tx}
	return f(ctx, tt)
}

func (c *Conn) Exec(ctx context.Context, sql string, args ...any) (int64, error) {
	tt := c.DB.WithContext(ctx).Exec(sql, args...)
	if err := tt.Error; err != nil {
		return 0, err
	}
	return tt.RowsAffected, nil
}

func (c *Conn) Query(ctx context.Context, sql string, args ...any) (repo.Rows, error) {
	rows, err := c.DB.WithContext(ctx).Raw(sql, args...).Rows()
	return rowsAdapter{rows}, err
}

func (c *Conn) IsConn() {
}

func (c *Conn) GORM(ctx context.Context) *gorm.DB {
	return c.DB.WithContext(ctx)
}
