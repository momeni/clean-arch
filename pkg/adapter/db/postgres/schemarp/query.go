// Copyright (c) 2024 Behnam Momeni
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package schemarp

import (
	"context"

	"github.com/momeni/clean-arch/pkg/adapter/db/postgres"
	"github.com/momeni/clean-arch/pkg/core/repo"
)

// InstallFDWExtensionIfMissing creates the postgres_fdw extension
// assuming that its relavant .so files are available in proper
// paths. If the extension is already created, calling this method
// causes no change.
func InstallFDWExtensionIfMissing(
	ctx context.Context, c *postgres.Conn,
) error {
	panic("not implemented yet") // TODO: Implement
}

// DropServerIfExists drops the `serverName` foreign server, if it
// exists, with cascade. That is, dependent objects such as its
// user mapping will be dropped too.
//
// Caller is responsible to pass a trusted serverName string.
func DropServerIfExists(
	ctx context.Context, c *postgres.Conn, serverName string,
) error {
	panic("not implemented yet") // TODO: Implement
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
	panic("not implemented yet") // TODO: Implement
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
	panic("not implemented yet") // TODO: Implement
}

// CreateSchema tries to create the `schema` schema.
// There must be no other schema with the `schema` name, otherwise,
// this operation will fail.
//
// Caller is responsible to pass a trusted schema name string.
func CreateSchema[Q postgres.Queryer](
	ctx context.Context, q Q, schema string,
) error {
	panic("not implemented yet") // TODO: Implement
}

// CreateRoleIfNotExists creates the `role` role if it does not
// exist right now. Although the login option is enabled for the
// created role, but no specific password will be set for it.
// The ChangePasswords method may be used for setting a password if
// desired. Otherwise, that user may not login effectively (but
// using the trust or local identity methods).
func CreateRoleIfNotExists[Q postgres.Queryer](
	ctx context.Context, q Q, role repo.Role,
) error {
	panic("not implemented yet") // TODO: Implement
}

// GrantPrivileges grants ALL privileges on the `schema` schema
// to the `role` role, so it may create or access tables in that schema
// and run relevant queries.
func GrantPrivileges[Q postgres.Queryer](
	ctx context.Context, q Q, schema string, role repo.Role,
) error {
	panic("not implemented yet") // TODO: Implement
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
	panic("not implemented yet") // TODO: Implement
}

// GrantFDWUsage grants the USAGE privilege on the postgres_fdw
// extension to the `role` role. Thereafter, that `role` role can use
// the postgres_fdw extension in order to create a foreign server or
// create a user mapping for it.
func GrantFDWUsage[Q postgres.Queryer](
	ctx context.Context, q Q, role repo.Role,
) error {
	panic("not implemented yet") // TODO: Implement
}

// ChangePasswords updates the passwords of the given roles in the
// current transaction. The roles and passwords slices must have the
// same number of entries, so they can be used in pair.
// These fields are not combined as a struct with two role and
// password fields because passing items separately ensures that
// all items are initialized explicitly in constrast to a struct
// which its fields can be zero-initialized and are more suitable
// to pass a set of optional fields.
func ChangePasswords(
	ctx context.Context,
	tx *postgres.Tx,
	roles []repo.Role,
	passwords []string,
) error {
	panic("not implemented yet") // TODO: Implement
}
