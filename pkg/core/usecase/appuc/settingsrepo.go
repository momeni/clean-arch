// Copyright (c) 2024 Behnam Momeni
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package appuc

import (
	"context"

	"github.com/momeni/clean-arch/pkg/core/model"
	"github.com/momeni/clean-arch/pkg/core/repo"
)

// SettingsRepo specifies the settings repository expectations, allowing
// version-independent mutable settings to be converted, serialized, and
// stored in the database or queried from the database, deserialized as
// a version-dependent mutable settings struct, and converted back to a
// version-independent struct again. In both paths of memory-database
// updating/fetching paths, a clone of the base configuration settings
// (which must be passed to and kept in the settings repository instance
// during its instantiation) must be updated by application of the new
// mutable settings. The base settings are usually read from a
// configuration file and some of them may be overridden by relevant
// environment variables. When base settings are overridden by the
// in-database mutable settings, result is returned to the use cases
// layer as an instance of the Builder interface, so they may be used
// for creation of new use case objects.
type SettingsRepo interface {
	// Conn wraps the provided connection instance and creates a new
	// settings repository connection-based queryer.
	Conn(repo.Conn) SettingsConnQueryer

	// Tx wraps the provided transaction instance and creates a new
	// settings repository transaction-based queryer.
	Tx(repo.Tx) SettingsTxQueryer
}

// SettingsConnQueryer interface indicates queries which require a
// database connection for their execution specifically (i.e., they need
// to indicate the transactions boundaries themselves and may not
// tolerate execution of extra queries before or after their own queries
// in a single transaction) and also queries which may be executed with
// either a connection or an ongoing transaction (by embedding the
// common SettingsQueryer interface).
type SettingsConnQueryer interface {
	SettingsQueryer

	// Fetch queries the mutable settings from the settings repository,
	// deserializes them, merges them into a clone of the base settings
	// (representing the configuration file and environment variables
	// state when the settings repository instance was created), and
	// returns the fresh configuration instance as a Builder interface
	// in addition to its visible settings (as an instance of the
	// version-independent model.VisibleSettings struct).
	//
	// The settings boundary values are also returned as `minb` and
	// `maxb` instances (of the version-independent model.Settings
	// struct), taken from the base settings. In order to sync these
	// boundary values from the base settings to the database (so they
	// can be queried by other components from the database), it is
	// required to perform a migration or use the Update method of the
	// SettingsTxQueryer interface instead. In other words, Fetch
	// neither updates the database nor verifies it beyound the version
	// of the persisted configuration settings.
	// If the database settings were out of the acceptable range of
	// values, they will take the nearest (minimum or maximum) boundary
	// value and that adjustment will be logged as a warning.
	//
	// Fetch requires a connection because the new Builder instance
	// may not be created but after a successful commit of the mutable
	// settings fetching query. It is also reified by the minor-version
	// specific schema loading packages (i.e., some schXvY package)
	// which expects a connection instance as they are used before the
	// migration process may begin (which depends on an updated instance
	// of the configuration settings).
	Fetch(ctx context.Context) (
		b Builder,
		vs *model.VisibleSettings,
		minb, maxb *model.Settings,
		err error,
	)
}

// SettingsTxQueryer interface indicates queries which require an open
// database transaction for their execution specifically (i.e., they
// expect other queries to be executed before or after their own queries
// and so must take an open transaction which is committed or rolled
// back by their caller as appropriate) and also queries which may be
// executed with either an ongoing transaction or a connection (by
// embedding the common SettingsQueryer interface).
type SettingsTxQueryer interface {
	SettingsQueryer

	// Update converts the version-independent mutable model.Settings
	// instance into a version-dependent serializable settings instance
	// for the last supported version, serializes them as JSON, and
	// then stores them in the settings repository. Given mutable
	// settings are also used in order to update a clone of the base
	// settings. Updated configuration settings will be returned as an
	// instance of the Builder interface in addition to its visible
	// settings (which are provided as an instance of the
	// version-independent model.VisibleSettings struct).
	//
	// The settings boundary values are also returned as `minb` and
	// `maxb` instances (of the version-independent model.Settings
	// struct), taken from the base settings. The argument `s` settings
	// must fall in this acceptable range of values, otherwise, an error
	// will be returned and settings will be kept unchanged.
	// When updating the database with new settings, the boundary values
	// will be serialized and stored alongside them too.
	//
	// Update requires an open transaction because it is supposed to
	// be reified by some major-version specific schema migration
	// settler object (i.e., some stlmigN package) which expects a
	// transaction instance (as they are used for the last phase of
	// a multi-database migration operation).
	Update(ctx context.Context, s *model.Settings) (
		b Builder,
		vs *model.VisibleSettings,
		minb, maxb *model.Settings,
		err error,
	)
}

// SettingsQueryer interface indicates queries which can be executed
// on a settings repository either with a connection or an ongoing
// transaction. This interface is embedded by both of SettingsTxQueryer
// and SettingsConnQueryer interfaces.
// There is no such methods in this version.
type SettingsQueryer interface {
}
