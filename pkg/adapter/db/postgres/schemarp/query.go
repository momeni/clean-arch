// Copyright (c) 2024 Behnam Momeni
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package schemarp

import (
	"context"
	"errors"
	"fmt"

	"github.com/momeni/clean-arch/pkg/adapter/db/postgres"
	"github.com/momeni/clean-arch/pkg/core/repo"
	"github.com/momeni/clean-arch/pkg/core/scram"
)

// InstallFDWExtensionIfMissing creates the postgres_fdw extension
// assuming that its relavant .so files are available in proper
// paths. If the extension is already created, calling this method
// causes no change.
func InstallFDWExtensionIfMissing(
	ctx context.Context, c *postgres.Conn,
) error {
	_, err := c.Exec(ctx, `CREATE EXTENSION IF NOT EXISTS postgres_fdw`)
	return err
}

// DropServerIfExists drops the `serverName` foreign server, if it
// exists, with cascade. That is, dependent objects such as its
// user mapping will be dropped too.
//
// Caller is responsible to pass a trusted serverName string.
func DropServerIfExists(
	ctx context.Context, c *postgres.Conn, serverName string,
) error {
	// Although this DDL statement does not support to take serverName
	// as a parameterized query, it is a trusted string.
	_, err := c.Exec(
		ctx,
		fmt.Sprintf(`DROP SERVER IF EXISTS %s CASCADE`, serverName),
	)
	return err
}

// DropIfExists drops the `schema` schema without cascading if it
// exists. That is, if `schema` does not exist, a nil error will be
// returned without any change. And if `schema` exists and is empty,
// it will be dropped. But if `schema` exists and is not empty, an
// error will be returned.
//
// Caller is responsible to pass a trusted schema name string.
func DropIfExists[Q postgres.Queryer](
	ctx context.Context, q Q, schema string,
) error {
	// Although this DDL statement does not support to take schema name
	// as a parameterized query, it is a trusted string.
	_, err := q.Exec(
		ctx,
		fmt.Sprintf(`DROP SCHEMA IF EXISTS %s RESTRICT`, schema),
	)
	return err
}

// DropCascade drops `schema` schema with cascading, dropping all
// dependent objects recursively. The `schema` must exist,
// otherwise, an error will be returned.
// This method is useful for dropping the intermediate schema
// which are created during a migration.
//
// Caller is responsible to pass a trusted schema name string.
func DropCascade[Q postgres.Queryer](
	ctx context.Context, q Q, schema string,
) error {
	// Although this DDL statement does not support to take schema name
	// as a parameterized query, it is a trusted string.
	_, err := q.Exec(
		ctx,
		fmt.Sprintf(`DROP SCHEMA %s CASCADE`, schema),
	)
	return err
}

// CreateSchema tries to create the `schema` schema.
// There must be no other schema with the `schema` name, otherwise,
// this operation will fail.
//
// Caller is responsible to pass a trusted schema name string.
func CreateSchema[Q postgres.Queryer](
	ctx context.Context, q Q, schema string,
) error {
	// Although this DDL statement does not support to take schema name
	// as a parameterized query, it is a trusted string.
	_, err := q.Exec(ctx, fmt.Sprintf(`CREATE SCHEMA %s`, schema))
	return err
}

// CreateRoleIfNotExists creates the `role` role if it does not
// exist right now. Although the login option is enabled for the
// created role, but no specific password will be set for it.
// The ChangePasswords method may be used for setting a password if
// desired. Otherwise, that user may not login effectively (but
// using the trust or local identity methods).
//
// The `role` role name may be suffixed by `roleSuffix` if it is not
// empty. This is useful to have distinct role names if repo.Role
// predefined constants are not desirable.
func CreateRoleIfNotExists[Q postgres.Queryer](
	ctx context.Context, q Q, roleSuffix repo.Role, role repo.Role,
) error {
	r := role + roleSuffix
	// Although CREATE ROLE DDL statement does not support to take
	// role name as a parameterized query, it is a trusted string.
	_, err := q.Exec(ctx, fmt.Sprintf(`DO
$body$
BEGIN
    IF NOT EXISTS (
            SELECT *
            FROM pg_catalog.pg_roles
            WHERE rolname = '%s'
    ) THEN
        BEGIN
            CREATE ROLE %[1]s WITH LOGIN;
        EXCEPTION WHEN duplicate_object THEN
            RAISE NOTICE '%%, skipping', SQLERRM USING ERRCODE=SQLSTATE;
        END;
    END IF;
END
$body$`, r))
	return err
}

// GrantPrivileges grants ALL privileges on the `schema` schema
// to the `role` role, so it may create or access tables in that schema
// and run relevant queries.
//
// The `role` role name may be suffixed by `roleSuffix` if it is not
// empty. This is useful to have distinct role names if repo.Role
// predefined constants are not desirable.
func GrantPrivileges[Q postgres.Queryer](
	ctx context.Context,
	q Q,
	roleSuffix repo.Role,
	schema string,
	role repo.Role,
) error {
	r := role + roleSuffix
	// Although this DDL statement does not support to take schema and
	// role names as a parameterized query, they are trusted strings.
	_, err := q.Exec(ctx, fmt.Sprintf(
		`GRANT ALL PRIVILEGES ON SCHEMA %s TO %s`, schema, r,
	))
	return err
}

// SetSearchPath alters the given database role and sets its default
// search_path to the given schema name alone.
func SetSearchPath[Q postgres.Queryer](
	ctx context.Context,
	q Q,
	roleSuffix repo.Role,
	schema string,
	role repo.Role,
) error {
	r := role + roleSuffix
	// Although this DDL statement does not support to take schema and
	// role names as a parameterized query, they are trusted strings.
	_, err := q.Exec(
		ctx,
		fmt.Sprintf(`ALTER ROLE %s SET search_path = %s`, r, schema),
	)
	return err
}

// GrantFDWUsage grants the USAGE privilege on the postgres_fdw
// extension to the `role` role. Thereafter, that `role` role can use
// the postgres_fdw extension in order to create a foreign server or
// create a user mapping for it.
//
// The `role` role name may be suffixed by `roleSuffix` if it is not
// empty. This is useful to have distinct role names if repo.Role
// predefined constants are not desirable.
func GrantFDWUsage[Q postgres.Queryer](
	ctx context.Context, q Q, roleSuffix repo.Role, role repo.Role,
) error {
	r := role + roleSuffix
	// Although this DDL statement does not support to take role name
	// as a parameterized query, it is a trusted string.
	_, err := q.Exec(ctx, fmt.Sprintf(`GRANT USAGE
ON FOREIGN DATA WRAPPER postgres_fdw
TO %s`, r))
	return err
}

// ChangePasswords updates the passwords of the given roles in the
// current transaction. The roles and passwords slices must have the
// same number of entries, so they can be used in pair.
// These fields are not combined as a struct with two role and
// password fields because passing items separately ensures that
// all items are initialized explicitly in constrast to a struct
// which its fields can be zero-initialized and are more suitable
// to pass a set of optional fields.
//
// The `roles` role names may be suffixed by `roleSuffix` if it is not
// empty. This is useful to have distinct role names if repo.Role
// predefined constants are not desirable.
// The `hasher` will be used for hashing of the `passwords` before
// sending them to the DBMS (so they may not leak in plaintext).
// This SCRAM hasher format must conform with the DBMS expected format.
func ChangePasswords(
	ctx context.Context,
	tx *postgres.Tx,
	roleSuffix repo.Role,
	hasher scram.Hasher,
	roles []repo.Role,
	passwords []string,
) error {
	switch len(roles) {
	default:
		return errors.New("number of roles/passwords must be equal")
	case 0:
		return errors.New("at least one role name must be provided")
	case len(passwords):
	}
	salt := "" // empty string asks for a fresh random salt
	for i, r := range roles {
		r = r + roleSuffix
		h, err := hasher.Hash(passwords[i], salt, 16000)
		if err != nil {
			return fmt.Errorf("hashing password of %q role: %w", r, err)
		}
		err = alterRolePass(ctx, tx, r, h)
		if err != nil {
			return fmt.Errorf(
				"alterRolePass(role=%q, hashed-pass=%q): %w", r, h, err,
			)
		}
	}
	return nil
}

func alterRolePass(
	ctx context.Context,
	tx *postgres.Tx,
	suffixedRole repo.Role,
	hashedPassword string,
) error {
	// Although this DDL statement does not support to take
	// role name and password as a parameterized query,
	// the role name is trusted and password is hashed and so
	// it follows a well-known format.
	_, err := tx.Exec(ctx, fmt.Sprintf(`ALTER ROLE %s
WITH PASSWORD '%s'`, suffixedRole, hashedPassword))
	return err
}
