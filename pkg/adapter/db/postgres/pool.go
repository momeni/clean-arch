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

type Pool struct {
	*gorm.DB
}

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

type ConnHandler = repo.ConnHandler

func NoOpConnHandler(context.Context, repo.Conn) error {
	return nil
}

func (p *Pool) Conn(ctx context.Context, f ConnHandler) error {
	return p.DB.WithContext(ctx).Connection(func(c *gorm.DB) error {
		cc := &Conn{DB: c}
		return f(ctx, cc)
	})
}

func (p *Pool) Close() error {
	db, err := p.DB.DB()
	if err != nil {
		return err
	}
	return db.Close()
}
