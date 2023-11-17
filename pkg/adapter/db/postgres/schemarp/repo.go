// Copyright (c) 2024 Behnam Momeni
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Package schemarp provides a reification of the repo.Schema interface
// making it possible to create or drop different schema, foreign
// server, or manage database user roles.
package schemarp

import (
	"context"

	"github.com/momeni/clean-arch/pkg/adapter/db/postgres"
	"github.com/momeni/clean-arch/pkg/core/repo"
)

// Repo represents a schema management repository.
type Repo struct {
}

// New instantiates a schema management Repo struct. Although this New
// function does not perform complex operations, and users may use
// a &schemarp.Repo{} directly too, but this method improves the code
// readability as schemarp.New() makes the package to look alike a
// data type.
func New() *Repo {
	return &Repo{}
}

type connQueryer struct {
	*postgres.Conn
}

// Conn unwraps the given repo.Conn instance, expecting to find an
// instance of *postgres.Conn as created by this adapter layer.
// Otherwise, it will panic. Unwrapped connection will be wrapped and
// returned as an instance of repo.SchemaConnQueryer interface, so
// it can be used in the use cases layer without requiring to type
// assert again and again.
// Returned querier instance can be used to run the connection-specific
// queries in addition to queries which support connections and
// transactions.
//
// Currently, two operations mandate a connection.
// The InstallFDWExtensionIfMissing creates an extension which will be
// accessible by roles based on their granted privileges.
// That installation should be performed independent of a failed or
// successful migration and it is not required to uninstall it later
// too. Also, it requires the extension related files to be installed
// previously. So it is not possible to confine all of its steps in a
// transaction anyway.
// The DropServerIfExists drops a foreign server and its user mapping.
// Thereafter, it should be created with a distinct user role. And it
// should be attempted only after dropping the main and intermediate
// schema successfully. So, asking a connection helps to guarantee that
// other removals are completed (in their transaction or auto-commit
// transaction) beforehand. By the way, result of this operation can
// affect multiple user roles.
func (schema *Repo) Conn(c repo.Conn) repo.SchemaConnQueryer {
	cc := c.(*postgres.Conn)
	return connQueryer{Conn: cc}
}

// InstallFDWExtensionIfMissing creates the postgres_fdw extension
// assuming that its relavant .so files are available in proper
// paths. If the extension is already created, calling this method
// causes no change.
func (cq connQueryer) InstallFDWExtensionIfMissing(
	ctx context.Context,
) error {
	return InstallFDWExtensionIfMissing(ctx, cq.Conn)
}

// DropServerIfExists drops the `serverName` foreign server, if it
// exists, with cascade. That is, dependent objects such as its
// user mapping will be dropped too.
//
// Caller is responsible to pass a trusted serverName string.
func (cq connQueryer) DropServerIfExists(
	ctx context.Context, serverName string,
) error {
	return DropServerIfExists(ctx, cq.Conn, serverName)
}

// DropIfExists drops the `schema` schema without cascading if it
// exists. That is, if `schema` does not exist, a nil error will be
// returned without any change. And if `schema` exists and is empty,
// it will be dropped. But if `schema` exists and is not empty, an
// error will be returned.
//
// Caller is responsible to pass a trusted schema name string.
func (cq connQueryer) DropIfExists(
	ctx context.Context, schema string,
) error {
	return DropIfExists(ctx, cq.Conn, schema)
}

// DropCascade drops `schema` schema with cascading, dropping all
// dependent objects recursively. The `schema` must exist,
// otherwise, an error will be returned.
// This method is useful for dropping the intermediate schema
// which are created during a migration.
//
// Caller is responsible to pass a trusted schema name string.
func (cq connQueryer) DropCascade(
	ctx context.Context, schema string,
) error {
	return DropCascade(ctx, cq.Conn, schema)
}

// CreateSchema tries to create the `schema` schema.
// There must be no other schema with the `schema` name, otherwise,
// this operation will fail.
//
// Caller is responsible to pass a trusted schema name string.
func (cq connQueryer) CreateSchema(
	ctx context.Context, schema string,
) error {
	return CreateSchema(ctx, cq.Conn, schema)
}

// CreateRoleIfNotExists creates the `role` role if it does not
// exist right now. Although the login option is enabled for the
// created role, but no specific password will be set for it.
// The ChangePasswords method may be used for setting a password if
// desired. Otherwise, that user may not login effectively (but
// using the trust or local identity methods).
func (cq connQueryer) CreateRoleIfNotExists(
	ctx context.Context, role repo.Role,
) error {
	return CreateRoleIfNotExists(ctx, cq.Conn, role)
}

// GrantPrivileges grants ALL privileges on the `schema` schema
// to the `role` role, so it may create or access tables in that schema
// and run relevant queries.
func (cq connQueryer) GrantPrivileges(
	ctx context.Context, schema string, role repo.Role,
) error {
	return GrantPrivileges(ctx, cq.Conn, schema, role)
}

// SetSearchPath alters the given database role and sets its default
// search_path to the given schema name alone.
func (cq connQueryer) SetSearchPath(
	ctx context.Context, schema string, role repo.Role,
) error {
	return SetSearchPath(ctx, cq.Conn, cq.roleSuffix, schema, role)
}

// GrantFDWUsage grants the USAGE privilege on the postgres_fdw
// extension to the `role` role. Thereafter, that `role` role can use
// the postgres_fdw extension in order to create a foreign server or
// create a user mapping for it.
func (cq connQueryer) GrantFDWUsage(
	ctx context.Context, role repo.Role,
) error {
	return GrantFDWUsage(ctx, cq.Conn, role)
}

type txQueryer struct {
	*postgres.Tx
}

// Tx unwraps the given repo.Tx instance, expecting to find an instance
// of *postgres.Tx as created by this adapter layer. Otherwise, it will
// panic. Unwrapped transaction will be wrapped and returned as an
// instance of repo.SchemaTxQueryer interface, so it can be used in
// the use cases layer without requiring to type assert again and again.
// Returned querier instance can be used to run the transaction-specific
// queries in addition to queries which support connections and
// transactions.
//
// Currently, one operation mandate a transaction.
// ChangePasswords updates passwords of some roles. When creating roles
// for the first time, it is desired to change/set their passwords
// before making them visible by committing the transaction. Also, it
// may be desired to call this method multiple times if all roles and
// passwords may not be identified as the same time. So, a transaction
// is required since there are scenarios that other operation must be
// performed in the same transaction and caller must specify the proper
// point of commitment.
func (schema *Repo) Tx(tx repo.Tx) repo.SchemaTxQueryer {
	tt := tx.(*postgres.Tx)
	return txQueryer{Tx: tt}
}

// DropIfExists drops the `schema` schema without cascading if it
// exists. That is, if `schema` does not exist, a nil error will be
// returned without any change. And if `schema` exists and is empty,
// it will be dropped. But if `schema` exists and is not empty, an
// error will be returned.
func (tq txQueryer) DropIfExists(
	ctx context.Context, schema string,
) error {
	return DropIfExists(ctx, tq.Tx, schema)
}

// DropCascade drops `schema` schema with cascading, dropping all
// dependent objects recursively. The `schema` must exist,
// otherwise, an error will be returned.
// This method is useful for dropping the intermediate schema
// which are created during a migration.
//
// Caller is responsible to pass a trusted schema name string.
func (tq txQueryer) DropCascade(
	ctx context.Context, schema string,
) error {
	return DropCascade(ctx, tq.Tx, schema)
}

// CreateSchema tries to create the `schema` schema.
// There must be no other schema with the `schema` name, otherwise,
// this operation will fail.
func (tq txQueryer) CreateSchema(
	ctx context.Context, schema string,
) error {
	return CreateSchema(ctx, tq.Tx, schema)
}

// CreateRoleIfNotExists creates the `role` role if it does not
// exist right now. Although the login option is enabled for the
// created role, but no specific password will be set for it.
// The ChangePasswords method may be used for setting a password if
// desired. Otherwise, that user may not login effectively (but
// using the trust or local identity methods).
func (tq txQueryer) CreateRoleIfNotExists(
	ctx context.Context, role repo.Role,
) error {
	return CreateRoleIfNotExists(ctx, tq.Tx, role)
}

// GrantPrivileges grants ALL privileges on the `schema` schema
// to the `role` role, so it may create or access tables in that schema
// and run relevant queries.
func (tq txQueryer) GrantPrivileges(
	ctx context.Context, schema string, role repo.Role,
) error {
	return GrantPrivileges(ctx, tq.Tx, schema, role)
}

// SetSearchPath alters the given database role and sets its default
// search_path to the given schema name alone.
func (tq txQueryer) SetSearchPath(
	ctx context.Context, schema string, role repo.Role,
) error {
	return SetSearchPath(ctx, tq.Tx, tq.roleSuffix, schema, role)
}

// GrantFDWUsage grants the USAGE privilege on the postgres_fdw
// extension to the `role` role. Thereafter, that `role` role can use
// the postgres_fdw extension in order to create a foreign server or
// create a user mapping for it.
func (tq txQueryer) GrantFDWUsage(
	ctx context.Context, role repo.Role,
) error {
	return GrantFDWUsage(ctx, tq.Tx, role)
}

// ChangePasswords updates the passwords of the given roles in the
// current transaction. The roles and passwords slices must have the
// same number of entries, so they can be used in pair.
// These fields are not combined as a struct with two role and
// password fields because passing items separately ensures that
// all items are initialized explicitly in constrast to a struct
// which its fields can be zero-initialized and are more suitable
// to pass a set of optional fields.
func (tq txQueryer) ChangePasswords(
	ctx context.Context, roles []repo.Role, passwords []string,
) error {
	return ChangePasswords(ctx, tq.Tx, roles, passwords)
}
