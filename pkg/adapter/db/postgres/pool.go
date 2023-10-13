// Package postgres provides an adapter for a PostgreSQL database
// in order to expose the interfaces which are required in the
// github.com/momeni/clean-arch/pkg/core/repo package.
// The actual implementation uses github.com/jackc/pgx/v5 for the
// connections and gorm.io/gorm for the models mapping and ORM.
package postgres

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/momeni/clean-arch/pkg/core/repo"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Pool represents a database connection pool.
// It may be used concurrently from different goroutines.
type Pool struct {
	*gorm.DB
}

// NewPool instances a connection pool using the url connection string.
func NewPool(ctx context.Context, url string) (*Pool, error) {
	gdb, err := gorm.Open(postgres.Open(url), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("gorm.Open: %w", err)
	}
	gdb = gdb.Session(&gorm.Session{
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{
				SlowThreshold:             200 * time.Millisecond,
				LogLevel:                  logger.Warn,
				IgnoreRecordNotFoundError: false,
				Colorful:                  true,
				// Set to false in order to log with replaced vars
				ParameterizedQueries: true,
			}),
	})
	pool := &Pool{DB: gdb}
	err = pool.Conn(ctx, NoOpConnHandler)
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("testing connection: %w", err)
	}
	return pool, nil
}

// ConnHandler is a handler function which takes a context and a
// database connection which should be used solely from the current
// goroutine (or by proper synchronization). When it returns, the
// connection may be released and reused by other routines.
type ConnHandler = repo.ConnHandler

// NoOpConnHandler is a connection handler which performs no operation.
// It is used in order to test a connection establishment during the
// creation of a new connection pool.
func NoOpConnHandler(context.Context, repo.Conn) error {
	return nil
}

// Conn acquires a database connection, passes it into the f handler
// function, and when it returns will release the connection so it may
// be used by other callers.
// This method may be blocked (as while as the ctx allows it) until a
// connection is obtained. That connection will not be used by any other
// handler concurrently. Returned errors from the f will be returned by
// this method after possible wrapping. The ctx which is used for
// acquisition of a connection is also passed to the f function.
func (p *Pool) Conn(ctx context.Context, f ConnHandler) error {
	return p.DB.WithContext(ctx).Connection(func(c *gorm.DB) error {
		cc := &Conn{DB: c}
		return f(ctx, cc)
	})
}

// Close closes all connections of this connection pool and returns
// any occurred error.
func (p *Pool) Close() error {
	db, err := p.DB.DB()
	if err != nil {
		return err
	}
	return db.Close()
}
