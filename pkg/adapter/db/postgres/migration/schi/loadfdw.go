// Copyright (c) 2024 Behnam Momeni
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package schi

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"

	"github.com/momeni/clean-arch/pkg/core/repo"
	"github.com/momeni/clean-arch/pkg/core/usecase/migrationuc"
)

// LoadFDW creates a foreign server, describing a PostgreSQL server
// which is accessible using the srcURL connection URL and postgres_fdw
// extension from within the `tx` transaction (in the destination
// database). It also creates a user mapping for the current database
// role, so it may employ the username and password (which must be
// present in the srcURL) whenever a connection to that foreign server
// is required. Ultimately, the cawebN schema from that foreign server
// will be imported into the fdwN_M local schema, creating the relevant
// foreign tables, where N and M are the given major and minor semantic
// schema version numbers.
// All of these items must be non-existing and they must be created by
// this call of LoadFDW in the `tx` transaction. Otherwise, an error
// will be returned.
func LoadFDW(
	ctx context.Context,
	major, minor uint,
	tx repo.Tx,
	srcURL string,
) error {
	dbName, uname, pass, h, p, err := extractSrcURLComponents(srcURL)
	if err != nil {
		return fmt.Errorf("extractSrcURLComponents: %w", err)
	}
	server := migrationuc.ForeignServerName(major, minor)
	remoteSchema := migrationuc.SchemaName(major)
	localSchema := migrationuc.ForeignSchemaName(major, minor)
	// Although it is not possible to use parameterized queries in
	// these DDL statements, but the server name and local and remote
	// schema names are all trusted strings.
	//
	// The h, dbName, uname, and pass are not trusted and may contain
	// any characters (specially the password), but they are put in
	// single quotes and their own single quotes (if any) are doubled.
	// When standard_conforming_strings is on (which is the default
	// since PostgreSQL 9.1), backslash characters may not escape
	// characters and doubling the single quotes is only escaping
	// method which works in regular string constants.
	if h == "127.0.0.1" {
		// Because destination database is running in a container,
		// it cannot see the localhost ports which are accessible
		// from out of the container using the 127.0.0.1 address.
		h = "host.containers.internal"
	}
	if _, err = tx.Exec(
		ctx,
		fmt.Sprintf(
			`CREATE SERVER %s
FOREIGN DATA WRAPPER postgres_fdw
OPTIONS (host '%s', port '%d', dbname '%s')`,
			server, escapeQuotes(h), p, escapeQuotes(dbName),
		),
	); err != nil {
		return fmt.Errorf("creating %q foreign server: %w", server, err)
	}
	if _, err = tx.Exec(
		ctx,
		fmt.Sprintf(
			`CREATE USER MAPPING FOR CURRENT_ROLE
SERVER %s
OPTIONS (user '%s', password '%s')`,
			server, escapeQuotes(uname), escapeQuotes(pass),
		),
	); err != nil {
		return fmt.Errorf(
			"creating user mapping for %q user in %q server: %w",
			uname, server, err,
		)
	}
	if _, err = tx.Exec(
		ctx,
		fmt.Sprintf(`IMPORT FOREIGN SCHEMA %s
FROM SERVER %s
INTO %s`, remoteSchema, server, localSchema),
	); err != nil {
		return fmt.Errorf(
			"importing %q foreign schema of %q server into %q: %w",
			remoteSchema, server, localSchema, err,
		)
	}
	return nil
}

func extractSrcURLComponents(
	srcURL string,
) (dbName, user, pass, host string, port int, err error) {
	mkErr := func(format string, args ...any) (
		dbName, user, pass, host string, port int, e error,
	) {
		return "", "", "", "", 0, fmt.Errorf(format, args...)
	}
	u, err := url.Parse(srcURL)
	if err != nil {
		return mkErr("parsing source URL address: %w", err)
	}
	host, p, err := net.SplitHostPort(u.Host)
	var nae *net.AddrError
	switch {
	case errors.As(err, &nae) && nae.Err == "missing port in address":
		host = u.Host
		port = 5432
	case err != nil:
		return mkErr("net.SplitHostPort(%q): %w", u.Host, err)
	default:
		port, err = strconv.Atoi(p)
		if err != nil {
			return mkErr("parsing %q as port number: %w", p, err)
		}
	}
	dbName = u.Path
	if len(dbName) < 2 || dbName[0] != '/' {
		return mkErr("unexpected database name: %q", dbName)
	}
	dbName = dbName[1:]
	user = u.User.Username()
	pass, set := u.User.Password()
	if user == "" || !set {
		return mkErr("source URL has no username or password")
	}
	return dbName, user, pass, host, port, nil
}

func escapeQuotes(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}
