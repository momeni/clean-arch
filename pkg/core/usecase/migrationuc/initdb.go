// Copyright (c) 2024 Behnam Momeni
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package migrationuc

import (
	"context"
	"fmt"

	"github.com/momeni/clean-arch/pkg/core/repo"
)

// InitDBUseCase represents the database initialization use case. It may
// be used to initialize database with development or production
// suitable data as asked by the InitDev and InitProd methods.
type InitDBUseCase struct {
	settings   Settings    // target settings
	schemaRepo repo.Schema // schema management repo
}

// NewInitDB creates an InitDBUseCase instance, using the `ss` schema
// settings in order to find the target database connection information
// and also create its schema initializer based on the expected database
// semantic version.
// The `repo.Schema` repo will be taken from the `ss` in order to be
// used for dropping and (re)creating an empty schema, creating normal
// role, granting it privileges on the empty schema, and renewing the
// passwords of admin and normal roles.
func NewInitDB(ss Settings) *InitDBUseCase {
	return &InitDBUseCase{
		settings:   ss,
		schemaRepo: ss.NewSchemaRepo(),
	}
}

// InitProd drops cawebN schema (if N is the relevant major version)
// and (re)creates it, assuming that it is an empty schema, using
// the admin role. It also creates the normal role (if it does not
// exist), grants privileges on the created schema to normal role so
// it can create tables, and renews passwords of both admin and normal
// roles. These operations will be performed using the admin role in a
// single transaction and coordinated with password files so they can
// be repeated in case of an abrupt failure as elaborated in docs of the
// pkg/core/usecase/migrationuc.SchemaSettings.RenewPasswords method.
// Thereafter, it connects to the target database using the normal role
// and completes its operation (in a second transaction) by creating all
// relevant tables and filling them with the production suitable data.
func (iduc *InitDBUseCase) InitProd(ctx context.Context) error {
	return iduc.initDB(
		ctx,
		func(ctx context.Context, si repo.SchemaInitializer) error {
			return si.InitProdSchema(ctx)
		},
	)
}

// InitDev drops cawebN schema (if N is the relevant major version)
// and (re)creates it, assuming that it is an empty schema, using
// the admin role. It also creates the normal role (if it does not
// exist), grants privileges on the created schema to normal role so
// it can create tables, and renews passwords of both admin and normal
// roles. These operations will be performed using the admin role in a
// single transaction and coordinated with password files so they can
// be repeated in case of an abrupt failure as elaborated in docs of the
// pkg/core/usecase/migrationuc.SchemaSettings.RenewPasswords method.
// Thereafter, it connects to the target database using the normal role
// and completes its operation (in a second transaction) by creating all
// relevant tables and filling them with the development suitable data.
func (iduc *InitDBUseCase) InitDev(ctx context.Context) error {
	return iduc.initDB(
		ctx,
		func(ctx context.Context, si repo.SchemaInitializer) error {
			return si.InitDevSchema(ctx)
		},
	)
}

func (iduc *InitDBUseCase) initDB(
	ctx context.Context,
	dbi func(ctx context.Context, si repo.SchemaInitializer) error,
) error {
	if err := iduc.dropAndCreateAgain(ctx); err != nil {
		return fmt.Errorf("dropping/recreating schema: %w", err)
	}
	p, err := iduc.settings.ConnectionPool(ctx, repo.NormalRole)
	if err != nil {
		return fmt.Errorf("creating DB pool for normal role: %w", err)
	}
	defer p.Close()
	err = p.Conn(ctx, func(ctx context.Context, c repo.Conn) error {
		return c.Tx(ctx, func(ctx context.Context, tx repo.Tx) error {
			ms, err := iduc.settings.Serialize()
			if err != nil {
				return fmt.Errorf("obtaining mutable settings: %w", err)
			}
			si, err := iduc.settings.SchemaInitializer(tx)
			if err != nil {
				return fmt.Errorf("creating SchemaInitializer: %w", err)
			}
			if err := dbi(ctx, si); err != nil {
				return fmt.Errorf("initializing schema: %w", err)
			}
			err = si.PersistSettings(ctx, ms)
			if err != nil {
				return fmt.Errorf("saving mutable settings: %w", err)
			}
			return nil
		})
	})
	if err != nil {
		return fmt.Errorf("normal connection: %w", err)
	}
	return nil
}

func (iduc *InitDBUseCase) dropAndCreateAgain(
	ctx context.Context,
) error {
	p, err := iduc.settings.ConnectionPool(ctx, repo.AdminRole)
	if err != nil {
		return fmt.Errorf("creating DB pool for admin: %w", err)
	}
	defer p.Close()
	var finalizer func() error
	err = p.Conn(ctx, func(ctx context.Context, c repo.Conn) error {
		return c.Tx(ctx, func(ctx context.Context, tx repo.Tx) error {
			q := iduc.schemaRepo.Tx(tx)
			v := iduc.settings.SchemaVersion()
			sn := SchemaName(v[0])
			if err := q.DropIfExists(ctx, sn); err != nil {
				return fmt.Errorf("dropping %q: %w", sn, err)
			}
			if err := q.CreateSchema(ctx, sn); err != nil {
				return fmt.Errorf("creating %q: %w", sn, err)
			}
			if err := q.CreateRoleIfNotExists(
				ctx, repo.NormalRole,
			); err != nil {
				return fmt.Errorf("creating normal role: %w", err)
			}
			if err := q.GrantPrivileges(
				ctx, sn, repo.NormalRole,
			); err != nil {
				return fmt.Errorf("granting normal role privs: %w", err)
			}
			if err := q.SetSearchPath(
				ctx, sn, repo.NormalRole,
			); err != nil {
				return fmt.Errorf(
					"setting search_path of normal role to %q: %w",
					sn, err,
				)
			}
			finalizer, err = iduc.settings.RenewPasswords(
				ctx, q.ChangePasswords, repo.AdminRole, repo.NormalRole,
			)
			if err != nil {
				return fmt.Errorf("RenewPasswords: %w", err)
			}
			return nil
		})
	})
	if err != nil {
		return fmt.Errorf("admin connection: %w", err)
	}
	if err := finalizer(); err != nil {
		return fmt.Errorf("finalizing passwords renewal: %w", err)
	}
	return nil
}

// SchemaName returns the target database schema name for the given
// major version. It should return cawebN for version N.
func SchemaName(major uint) string {
	return fmt.Sprintf("caweb%d", major)
}
