package postgres

import (
	"context"

	"github.com/momeni/clean-arch/pkg/core/repo"
	"gorm.io/gorm"
)

type Queryer interface {
	*Conn | *Tx
	repo.Queryer
	GORM(ctx context.Context) *gorm.DB
}
