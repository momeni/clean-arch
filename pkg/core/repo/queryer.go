package repo

import "context"

type Queryer interface {
	Exec(ctx context.Context, sql string, args ...any) (count int64, err error)
	Query(ctx context.Context, sql string, args ...any) (Rows, error)
}

type Rows interface {
	Close()
	Err() error
	Next() bool
	Scan(dest ...any) error
	Values() ([]any, error)
}
