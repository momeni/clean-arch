// Copyright (c) 2024 Behnam Momeni
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Package settingsrp is the adapter for the settings repository.
// It exposes the settingsrp.Repo type in order to allow use cases
// to update mutable settings or query them from the database.
package settingsrp

import (
	"context"

	"github.com/momeni/clean-arch/pkg/adapter/config/cfg2"
	"github.com/momeni/clean-arch/pkg/adapter/db/postgres"
	"github.com/momeni/clean-arch/pkg/core/model"
	"github.com/momeni/clean-arch/pkg/core/repo"
	"github.com/momeni/clean-arch/pkg/core/usecase/appuc"
)

// Repo represents the settings repository instance.
type Repo struct {
	baseConfs *cfg2.Config
}

// New instantiates a settings Repo struct. Created instance wraps
// the given configuration instance as its base configuration items, so
// whenever it needs to update the mutable settings or reload them from
// the database, it can apply them on a fresh clone of this base confs.
func New(c *cfg2.Config) *Repo {
	return &Repo{
		baseConfs: c,
	}
}

type connQueryer struct {
	*postgres.Conn
	baseConfs *cfg2.Config
}

// Conn takes a Conn interface instance, unwraps it as required,
// and returns a SettingsConnQueryer interface which (with access to
// the implementation-dependent connection object) can run different
// permitted operations on settings.
// The connQueryer itself is not mentioned as the return value since
// it is not exported. Otherwise, the general rule is to take interfaces
// as arguments and return exported structs.
func (settings *Repo) Conn(c repo.Conn) appuc.SettingsConnQueryer {
	cc := c.(*postgres.Conn)
	return connQueryer{Conn: cc, baseConfs: settings.baseConfs}
}

// Fetch queries the mutable settings from the settings repository,
// deserializes them, merges them into a clone of the base settings
// (representing the configuration file and environment variables state
// when the settings repository instance was created), and returns the
// fresh configuration instance as an appuc.Builder interface in
// addition to its visible settings (as an instance of the
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
func (cq connQueryer) Fetch(ctx context.Context) (
	b appuc.Builder,
	vs *model.VisibleSettings,
	minb, maxb *model.Settings,
	err error,
) {
	return Fetch(ctx, cq.Conn, cq.baseConfs)
}

type txQueryer struct {
	*postgres.Tx
	baseConfs *cfg2.Config
}

// Tx takes a Tx interface instance, unwraps it as required,
// and returns a SettingsTxQueryer interface which (with access to the
// implementation-dependent transaction object) can run different
// permitted operations on settings.
// The txQueryer itself is not mentioned as the return value since
// it is not exported. Otherwise, the general rule is to take interfaces
// as arguments and return exported structs.
func (settings *Repo) Tx(tx repo.Tx) appuc.SettingsTxQueryer {
	tt := tx.(*postgres.Tx)
	return txQueryer{Tx: tt, baseConfs: settings.baseConfs}
}

// Update converts the version-independent mutable model.Settings
// instance into a version-dependent serializable settings instance
// for the last supported version, serializes them as JSON, and
// then stores them in the settings repository. Given mutable settings
// are also used in order to update a clone of the base settings.
// Updated configuration settings will be returned as an instance of
// the appuc.Builder interface in addition to its visible settings
// (which are provided as an instance of the version-independent
// model.VisibleSettings struct).
//
// The settings boundary values are also returned as `minb` and
// `maxb` instances (of the version-independent model.Settings
// struct), taken from the base settings. The argument `s` settings
// must fall in this acceptable range of values, otherwise, an error
// will be returned and settings will be kept unchanged.
// When updating the database with new settings, the boundary values
// will be serialized and stored alongside them too.
func (tq txQueryer) Update(ctx context.Context, s *model.Settings) (
	b appuc.Builder,
	vs *model.VisibleSettings,
	minb, maxb *model.Settings,
	err error,
) {
	return Update(ctx, tq.Tx, tq.baseConfs, s)
}
