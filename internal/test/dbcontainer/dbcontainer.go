// Copyright (c) 2023 Behnam Momeni
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Package dbcontainer is an internal helper for the test packages.
// This packages facilitates creation of a temporary postgres:16
// podman container and connecting to it, using a *postgres.Pool
// connection pool.
// It may be used in all integration-level test suites which require
// a real PostgreSQL DBMS server.
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

// New creates and starts up a postgres podman container.
// The podman.service needs to be started and the DOCKER_HOST
// environment variable needs to be initialized beforehand like
// DOCKER_HOST=unix://$XDG_RUNTIME_DIR/podman/podman.sock
// in order to be identified by this function properly.
// The ctx will be used during the container start up and shutdown,
// while the timeout will be considered only during the start up phase.
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
