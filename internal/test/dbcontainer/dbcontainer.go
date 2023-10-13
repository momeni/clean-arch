package dbcontainer

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/bitcomplete/sqltestutil"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/momeni/clean-arch/pkg/adapter/db/postgres"
	"github.com/stretchr/testify/assert"
)

func New(ctx context.Context, timeout time.Duration, t *testing.T) (
	pg *sqltestutil.PostgresContainer,
	pool *postgres.Pool,
	dfrs []func(),
	ok bool,
) {
	ctx2, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	dbmsVer := "16"
	pg, err := sqltestutil.StartPostgresContainer(ctx2, dbmsVer)
	ok = assert.NoError(t, err, "failed to set up a test database")
	if !ok {
		return
	}
	dfrs = append(dfrs, func() {
		err := pg.Shutdown(ctx)
		assert.NoError(t, err, "failed to shutdown test database")
	})
	u := pg.ConnectionString()
	for pool == nil {
		pool, err = postgres.NewPool(ctx2, u)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.SQLState() == "57P03" {
			continue // the database system is starting up
		}
		var netErr net.Error
		if ctx2.Err() == nil && errors.As(err, &netErr) {
			continue // tolerate network errors until a timeout
		}
		ok = assert.NoError(t, err, "cannot connect to test database")
		if !ok {
			return
		}
	}
	dfrs = append(dfrs, func() {
		err := pool.Close()
		assert.NoError(t, err, "failed to close the connections pool")
	})
	return
}
