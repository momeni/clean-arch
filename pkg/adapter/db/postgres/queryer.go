package postgres

import "github.com/momeni/clean-arch/pkg/core/repo"

type Queryer interface {
	*Conn | *Tx
	repo.Queryer
}
