// Copyright (c) 2024 Behnam Momeni
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Package upmig1 provides configuration file settings upwards Migrator
// for settings major version 1 and its Adapter type for the version
// independent repo.UpMigrator[migrationuc.Settings] interface.
//
// This package provides the main logic for converting settings with
// major version 1 format to major version 2.
//
// The settings.UpMigrator generic interface is employed in order to
// ensure that this version-specific implementation uses consistent
// types as its method return types.
package upmig1

import (
	"context"

	"github.com/momeni/clean-arch/pkg/adapter/config/cfg1"
	"github.com/momeni/clean-arch/pkg/adapter/config/cfg2"
	"github.com/momeni/clean-arch/pkg/adapter/config/settings"
	"github.com/momeni/clean-arch/pkg/adapter/config/up/upmig2"
	"github.com/momeni/clean-arch/pkg/adapter/config/vers"
	"github.com/momeni/clean-arch/pkg/core/repo"
	"github.com/momeni/clean-arch/pkg/core/usecase/migrationuc"
)

// These type aliases specify the underlying Config struct (with major
// version 1) as C and the parameterized settings.UpMigrator interface
// which is supposed to be implemented by the Migrator struct as Type.
// The Type uses the *upmig2.Migrator as its newer/upper counterpart.
type (
	// C is the underlying Config type
	C = *cfg1.Config
	// Type is implemented by the Migrator type
	Type = settings.UpMigrator[C, *upmig2.Migrator]
)

// Adapter wraps and adapts an instance of Type in order to provide
// the repo.UpMigrator[migrationuc.Settings] interface.
type Adapter struct {
	T Type
}

// NewUpMig creates a Migrator struct wrapping the given Config instance
// and then uses the Adapt function in order to adapt it to the version
// independent repo.UpMigrator[migrationuc.Settings] interface.
//
// The Migrator struct is exported and users which need a concrete type
// can create it directly and wrap the `c` instance. This helper New
// function is provided in order to combine these two steps (of creation
// and adaptation) together.
func NewUpMig(c *cfg1.Config) repo.UpMigrator[migrationuc.Settings] {
	m := &Migrator{c}
	return Adapt(m)
}

// Adapt creates an instance of Adapter struct wraping the `m` argument.
// Because Adapter expects to wrap a Type instance, it asserts that
// Migrator struct implements the Type interface, its implementation is
// correct (considering the expected return types), and provides the
// repo.UpMigrator[migrationuc.Settings] interface.
func Adapt(m *Migrator) repo.UpMigrator[migrationuc.Settings] {
	return Adapter{m}
}

// Settler calls the wrapped Type Settler method, obtains a C instance,
// and wraps it by settings.Adapter[C] in order to expose an instance
// of migrationuc.Settings interface.
func (a Adapter) Settler() migrationuc.Settings {
	c := a.T.Settler()
	return settings.Adapter[C]{c}
}

// MigrateUp calls the wrapped Type MigrateUp method, obtains the next
// upwards migrator object, and adapts it to the version-independent
// repo.UpMigrator[migrationuc.Settings] interface using the Adapt
// function.
func (a Adapter) MigrateUp(ctx context.Context) (
	repo.UpMigrator[migrationuc.Settings], error,
) {
	m, err := a.T.MigrateUp(ctx)
	if err != nil {
		return nil, err
	}
	return upmig2.Adapt(m), nil
}

// Migrator is an upwards Config migrator for *cfg1.Config instances.
// It wraps a Config struct (with major version 1) and implements
// the repo.Settler and pkg/adapter/config/settings.UpMigrator
// generic interfaces.
type Migrator struct {
	*cfg1.Config
}

// MigrateUp creates a *cfg2.Config instance and fills it with the
// settings which are kept in `m.Config` fields. Any field which its
// value may not be computed based on the settings which are available
// in this version will be left uninitialized. The computed Config
// instance with major version 2 will be wrapped by its corresponding
// upwards migrator before being returned.
func (m *Migrator) MigrateUp(
	_ context.Context,
) (*upmig2.Migrator, error) {
	c := &cfg2.Config{
		Database: m.Config.Database,
		Vers: vers.Config{
			Versions: vers.Versions{
				Database: m.Config.Vers.Versions.Database,
				Config:   cfg2.Version,
			},
		},
	}
	settings.OverwriteNil(&c.Gin.Logger, m.Config.Gin.Logger)
	settings.OverwriteNil(&c.Gin.Recovery, m.Config.Gin.Recovery)
	settings.OverwriteNil(
		&c.Usecases.Cars.DelayOfOPM,
		m.Config.Usecases.Cars.OldParkingDelay,
	)
	return &upmig2.Migrator{c}, nil
}

// Settler returns the wrapped Config object. After migrating from a
// source Config version upwards and reaching to an ultimate version,
// this method reveals the final migrated Config object.
// This object may have some uninitialized settings too. The MergeConfig
// method may be used in order to fill them from another Config instance
// containing the default settings for major version 1.
func (m *Migrator) Settler() *cfg1.Config {
	return m.Config
}
