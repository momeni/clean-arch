package repo

import "context"

type ConnHandler func(context.Context, Conn) error

type Pool interface {
	Conn(ctx context.Context, handler ConnHandler) error
}
