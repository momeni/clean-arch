// Copyright (c) 2024 Behnam Momeni
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package repo

// Role is a string specifying a database connection role. Each role
// has a set of granted privileges which indicates which operations
// may be performed after using it for connecting to a database.
//
// The pkg/core/usecase/migrationuc.SchemaSettings.ConnectionPool needs
// one Role instance in order to connect to a database (which its
// identification information are captured from a config file and
// its authentication information are read from a passwords file).
type Role string

// These constants specify the expected database roles. At least the
// AdminRole must exist beforehand (i.e., must be created manually)
// and it must have super user privileges, so it can be used to create
// other required roles (if they are not already created).
// The authentication information of these roles are kept in pass files
// as indicated in the configuration file.
const (
	// AdminRole is an administrator (super user) role which may be used
	// for creation of other roles, granting them relevant privileges,
	// or creation of empty schema. Generally, the minimal set of
	// operations which are not required normally but may be essential
	// to start other normal use cases, may be performed by this role.
	AdminRole Role = "admin"

	// NormalRole is a normal (unprivilged) role which is used for
	// all common operations, including creation of tables in existing
	// schema and filling them during the migration use cases and also
	// database changes (for the latest schema version only) during
	// the non-migration use cases.
	NormalRole Role = "caweb"
)
