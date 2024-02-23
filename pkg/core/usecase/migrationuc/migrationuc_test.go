// Copyright (c) 2024 Behnam Momeni
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package migrationuc_test

import (
	"context"
	"crypto/rand"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/bitcomplete/sqltestutil"
	"github.com/momeni/clean-arch/internal/test/dbcontainer"
	"github.com/momeni/clean-arch/internal/test/schema"
	"github.com/momeni/clean-arch/pkg/adapter/config"
	"github.com/momeni/clean-arch/pkg/adapter/config/cfg1"
	"github.com/momeni/clean-arch/pkg/adapter/config/cfg2"
	"github.com/momeni/clean-arch/pkg/adapter/config/settings"
	"github.com/momeni/clean-arch/pkg/adapter/config/vers"
	"github.com/momeni/clean-arch/pkg/adapter/db/postgres"
	"github.com/momeni/clean-arch/pkg/adapter/db/postgres/migration/sch1v0"
	"github.com/momeni/clean-arch/pkg/adapter/db/postgres/migration/sch1v1"
	"github.com/momeni/clean-arch/pkg/adapter/hash/scram"
	"github.com/momeni/clean-arch/pkg/core/model"
	"github.com/momeni/clean-arch/pkg/core/repo"
	"github.com/momeni/clean-arch/pkg/core/usecase/migrationuc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

type MigrationUseCasesTestSuite struct {
	Ctx  context.Context
	Pg   *sqltestutil.PostgresContainer
	Pool *postgres.Pool
	Port int

	dbDir, cfgDir string
	hasher        *scram.Mechanism
}

func TestMigrationUseCasesTestSuite(t *testing.T) {
	ctx := context.Background()
	pg, pool, dfrs, ok := dbcontainer.New(ctx, 60*time.Second, t)
	for _, f := range dfrs {
		defer f()
	}
	if !ok {
		return // errors are already logged
	}
	u, err := url.Parse(pg.ConnectionString())
	if ok := assert.NoError(t, err, "parsing DB container URL"); !ok {
		return
	}
	p, err := strconv.Atoi(u.Port())
	if ok := assert.NoError(t, err, "parsing DB container port"); !ok {
		return
	}
	dbDir, err := os.MkdirTemp("", "miguc-db")
	if ok := assert.NoError(t, err, "creating temp db dir"); !ok {
		return
	}
	defer func() {
		err := os.RemoveAll(dbDir)
		assert.NoError(t, err, "removing temp db dir")
	}()
	cfgDir, err := os.MkdirTemp("", "miguc-cfg")
	if ok := assert.NoError(t, err, "creating temp configs dir"); !ok {
		return
	}
	defer func() {
		err = os.RemoveAll(cfgDir)
		assert.NoError(t, err, "removing temp configs dir")
	}()
	migucts := &MigrationUseCasesTestSuite{
		Ctx:  ctx,
		Pg:   pg,
		Pool: pool,
		Port: p,

		dbDir:  dbDir,
		cfgDir: cfgDir,
		hasher: scram.SHA256(),
	}
	t.Run("initialization and migration", migucts.TestMigrations)
}

func (migucts *MigrationUseCasesTestSuite) TestMigrations(
	t *testing.T,
) {
	for _, mode := range []string{"dev", "prod"} {
		migucts.visitCfgDBVers(t, mode, migucts.TestInitDB)
	}
}

func (migucts *MigrationUseCasesTestSuite) TestInitDB(
	t *testing.T,
	s migrationuc.Settings,
	cfgVer, dbVer model.SemVer,
	name string,
) {
	t.Parallel()
	r := require.New(t)
	dev := strings.HasSuffix(name, "dev")
	migucts.initDBAndVerifySchema(t, r, s, dev, dbVer)

	b, err := yaml.Marshal(s)
	require.NoError(t, err, "marshaling source settings; v=%v", cfgVer)
	srcCfgPath := filepath.Join(migucts.cfgDir, name)
	err = os.WriteFile(srcCfgPath, b, 0o644)
	require.NoError(t, err, "writing source settings; name=%q", name)

	migucts.visitCfgDBVers(t, name, func(
		t *testing.T,
		dstSettings migrationuc.Settings,
		dstCfgVer, dstDBVer model.SemVer,
		dstName string,
	) {
		t.Parallel()
		r := require.New(t)
		mig, err := config.LoadSrcMigrator(srcCfgPath)
		r.NoError(err, "config.LoadSrcMigrator(%q)", srcCfgPath)
		targetCfgPath := filepath.Join(migucts.cfgDir, dstName)
		mduc := migrationuc.NewMigrateDB(
			mig, dstSettings, targetCfgPath, loader,
		)
		err = mduc.Migrate(migucts.Ctx)
		r.NoError(err, "migrate from schema %v to %v", dbVer, dstDBVer)
		targetSettings, err := loader(migucts.Ctx, targetCfgPath)
		r.NoError(err, "loading target settings from %q", targetCfgPath)
		same, err := migrationuc.HasTheSameConnectionInfo(
			targetSettings, dstSettings,
		)
		r.NoError(
			err,
			"target/dst schema (version %v/%v) do not match",
			targetSettings.SchemaVersion(), dstDBVer,
		)
		tin, tih, tip := targetSettings.ConnectionInfo()
		din, dih, dip := dstSettings.ConnectionInfo()
		r.True(
			same,
			"target (%s:%d/%s) and dst (%s:%d/%s) DBs are different",
			tih, tip, tin,
			dih, dip, din,
		)
		verifySchema(
			migucts.Ctx, t, r, targetSettings, dstDBVer,
			func(ctx context.Context, v schema.Verifier, t *testing.T) {
				v.VerifySchema(ctx, t)
			},
		)
	})
}

func loader(
	ctx context.Context, path string,
) (migrationuc.Settings, error) {
	mig, err := config.LoadMigrator(path)
	if err != nil {
		return nil, fmt.Errorf("config.LoadMigrator: %w", err)
	}
	err = mig.Load(ctx)
	if err != nil {
		return nil, fmt.Errorf("mig.Load(): %w", err)
	}
	settings, err := mig.Settler(ctx)
	if err != nil {
		return nil, fmt.Errorf("mig.Settler(): %w", err)
	}
	return settings, nil
}

func (migucts *MigrationUseCasesTestSuite) initDBAndVerifySchema(
	t *testing.T,
	r *require.Assertions,
	s migrationuc.Settings,
	dev bool,
	dbVer model.SemVer,
) {
	iduc := migrationuc.NewInitDB(s)
	if dev {
		err := iduc.InitDev(migucts.Ctx)
		r.NoError(err, "initializing database with dev suitable data")
	} else {
		err := iduc.InitProd(migucts.Ctx)
		r.NoError(err, "initializing database with prod suitable data")
	}
	verifySchema(
		migucts.Ctx, t, r, s, dbVer,
		func(ctx context.Context, v schema.Verifier, t *testing.T) {
			v.VerifySchema(ctx, t)
			if dev {
				v.VerifyDevData(ctx, t)
			} else {
				v.VerifyProdData(ctx, t)
			}
		},
	)
}

func verifySchema(
	ctx context.Context,
	t *testing.T,
	r *require.Assertions,
	s migrationuc.Settings,
	dbVer model.SemVer,
	verify func(ctx context.Context, v schema.Verifier, t *testing.T),
) {
	p, err := s.ConnectionPool(ctx, repo.NormalRole)
	r.NoError(err, "creating connection pool")
	defer p.Close()
	err = p.Conn(ctx, func(ctx context.Context, c repo.Conn) error {
		v, err := schema.NewVerifier(c, dbVer)
		if err != nil {
			return fmt.Errorf("NewVerifier(%v): %w", dbVer, err)
		}
		verify(ctx, v, t)
		return nil
	})
	r.NoError(err, "verifying database schema")
}

func (migucts *MigrationUseCasesTestSuite) visitCfgDBVers(
	t *testing.T,
	suffix string,
	visit func(
		t *testing.T,
		s migrationuc.Settings,
		cfgVer, dbVer model.SemVer,
		name string,
	),
) {
	a := assert.New(t)
	for _, dbVer := range []model.SemVer{
		{sch1v0.Major, sch1v0.Minor, sch1v0.Patch},
		{sch1v1.Major, sch1v1.Minor, sch1v1.Patch},
	} {
		cfgVer := model.SemVer{cfg1.Major, cfg1.Minor, cfg1.Patch}
		d, name, rs := migucts.createEmptyDB(a, cfgVer, dbVer, suffix)
		var s migrationuc.Settings
		c1 := &cfg1.Config{
			Database: cfg1.Database{
				Host:       "127.0.0.1",
				Port:       migucts.Port,
				Name:       name,
				PassDir:    d,
				RoleSuffix: rs,
			},
			Vers: vers.Config{
				Versions: vers.Versions{
					Database: dbVer,
					Config:   cfg1.Version,
				},
			},
		}
		err := c1.ValidateAndNormalize()
		a.NoError(err, "validating *cfg1.Config instance")
		s = settings.Adapter[*cfg1.Config, cfg1.Serializable]{c1}
		t.Run(name, func(t *testing.T) {
			visit(t, s, cfgVer, dbVer, name)
		})

		cfgVer = model.SemVer{cfg2.Major, cfg2.Minor, cfg2.Patch}
		d, name, rs = migucts.createEmptyDB(a, cfgVer, dbVer, suffix)
		c2 := &cfg2.Config{
			Database: cfg1.Database{
				Host:       "127.0.0.1",
				Port:       migucts.Port,
				Name:       name,
				PassDir:    d,
				RoleSuffix: rs,
			},
			Vers: vers.Config{
				Versions: vers.Versions{
					Database: dbVer,
					Config:   cfg2.Version,
				},
			},
		}
		err = c2.ValidateAndNormalize()
		a.NoError(err, "validating *cfg2.Config instance")
		s = settings.Adapter[*cfg2.Config, cfg2.Serializable]{c2}
		t.Run(name, func(t *testing.T) {
			visit(t, s, cfgVer, dbVer, name)
		})
	}
}

func (migucts *MigrationUseCasesTestSuite) createEmptyDB(
	a *assert.Assertions,
	cfgVer, dbVer model.SemVer, suffix string,
) (dbDir, dbName string, roleSuffix repo.Role) {
	name := fmt.Sprintf(
		"cfg%d_%d_%d_sch%d_%d_%d_%s",
		cfgVer[0], cfgVer[1], cfgVer[2],
		dbVer[0], dbVer[1], dbVer[2],
		suffix,
	)
	roleSuffix = repo.Role("_" + name)
	u := repo.AdminRole + roleSuffix
	p := migucts.randPass(a)
	err := migucts.Pool.Conn(
		migucts.Ctx, func(ctx context.Context, c repo.Conn) error {
			// The database and role creation DDL statements do not
			// support parameterized queries, nevertheless, the `name`
			// and `u` variables are trusted.
			if _, err := c.Exec(
				ctx, "CREATE DATABASE "+name,
			); err != nil {
				return fmt.Errorf("creating %q database: %w", name, err)
			}
			// The `p` password is hashed before being sent to DBMS, so
			// it may not leak even if it is recorded in some log file.
			hp, err := migucts.hasher.Hash(p, "", 15000)
			if err != nil {
				return fmt.Errorf(
					"computing scram hash of password: %w", err,
				)
			}
			// SUPERUSER is required for CREATE EXTENSION
			if _, err := c.Exec(
				ctx,
				fmt.Sprintf(
					`CREATE ROLE %s
WITH SUPERUSER LOGIN PASSWORD '%s';
GRANT ALL PRIVILEGES ON DATABASE %s TO %[1]s`,
					u, hp, name,
				),
			); err != nil {
				return fmt.Errorf("creating %q role: %w", u, err)
			}
			return nil
		},
	)
	if !a.NoError(err, "main connection error") {
		a.FailNow("failed to get a connection with superuser role")
	}
	d := filepath.Join(migucts.dbDir, name)
	err = os.Mkdir(d, 0o700)
	if !a.NoError(err, "creating %q dir", d) {
		a.FailNow("cannot create top database dir")
	}
	line := fmt.Sprintf(
		"127.0.0.1:%d:%s:%s:%s\n", migucts.Port, name, u, p,
	)
	pgpass := filepath.Join(d, ".pgpass")
	err = os.WriteFile(pgpass, []byte(line), 0o600)
	if !a.NoError(err, "writing %q file", pgpass) {
		a.FailNow("cannot write .pgpass file")
	}
	return d, name, roleSuffix
}

func (migucts *MigrationUseCasesTestSuite) randPass(
	a *assert.Assertions,
) string {
	b := make([]byte, 8)
	_, err := rand.Read(b)
	if !a.NoError(err, "generating a random password") {
		a.FailNow("cannot read random bytes")
	}
	return fmt.Sprintf("%x", b)
}
