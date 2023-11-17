// Copyright (c) 2024 Behnam Momeni
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package config

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/momeni/clean-arch/pkg/adapter/config/cfg1"
	"github.com/momeni/clean-arch/pkg/adapter/config/cfg2"
	"github.com/momeni/clean-arch/pkg/adapter/config/down/dnmig1"
	"github.com/momeni/clean-arch/pkg/adapter/config/down/dnmig2"
	"github.com/momeni/clean-arch/pkg/adapter/config/settings"
	"github.com/momeni/clean-arch/pkg/adapter/config/up/upmig1"
	"github.com/momeni/clean-arch/pkg/adapter/config/up/upmig2"
	"github.com/momeni/clean-arch/pkg/adapter/config/vers"
	"github.com/momeni/clean-arch/pkg/core/repo"
	"github.com/momeni/clean-arch/pkg/core/usecase/migrationuc"
)

// LoadSrcMigrator reads a config file which is stored at the given
// `path` argument, examines its stored contents format version, and
// creates a migrator object wrapping the read file contents.
// This function is similar to the LoadMigrator function with this
// difference that loaded configuration settings may be overridden by
// their corresponding values from the source database.
// In the current implementation, settings are not kept in the database
// and so, LoadSrcMigrator simply calls the LoadMigrator function.
func LoadSrcMigrator(path string) (
	repo.Migrator[migrationuc.Settings], error,
) {
	return LoadMigrator(path)
}

// LoadMigrator reads a config file which is stored at the given `path`,
// examines its stored contents format version, and creates a migrator
// object wrapping the read file contents.
// A migrator object is able to load a specific configuration settings
// version (assuming that it follows a supported format version), wrap
// them by corresponding upwards/downwards migrator objects which can
// be used to convert the loaded settings to their upper/downer format
// version, and finally settle the migration by retrieving the resulting
// settings contents (which can be marshalled and written to a file
// again).
// Since each major version requires a distinct migrator object (which
// loads different cfgN.Config structs later and wraps them with
// different upmigN.Migrator and dnmigN.Migrator objects subsequently),
// it is required to adapt these types (with help of a series of Adapter
// structs) in order to expose a common version-independent interface.
// This common interface is the repo.Migrator[migrationuc.Settings] and
// it may be passed to the use cases layer too.
func LoadMigrator(path string) (
	repo.Migrator[migrationuc.Settings], error,
) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}
	v, err := vers.Load(data)
	if err != nil {
		return nil, fmt.Errorf("loading versions: %w", err)
	}
	vc := v.Versions
	switch m := vc.Config[0]; m {
	case 1:
		return &Migrator[*cfg1.Config]{
			data:    data,
			loader:  cfg1.Load,
			upmiger: upmig1.NewUpMig,
			dnmiger: dnmig1.NewDnMig,
		}, nil
	case 2:
		return &Migrator[*cfg2.Config]{
			data:    data,
			loader:  cfg2.Load,
			upmiger: upmig2.NewUpMig,
			dnmiger: dnmig2.NewDnMig,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported major version: %d", m)
	}
}

// Migrator of C is a generic interface which adapts version-specific
// configuration settings data (contained in a C type instance which
// conforms to the settings.Config[C] interface) in order to provide the
// version-independent repo.Migrator[migrationuc.Settings] interface.
// Migrator[C] struct instances are created by the LoadMigrator function
// wrapping the configuration settings data as a byte slice without
// parsing its fields (beyond the settings format version).
//
// Following the logic of repo.Migrator[migrationuc.Settings] interface,
// migration follows three main steps.
//
// First, loading phase obtains the relevant data items from the source
// version. For this phase, `loader` function is taken which accepts the
// `data` byte slice, loads their contents, validates them, and finally
// normalizes them (replacing default values as appropriate). If some
// settings need to be taken from environment variables or need to be
// overwritten by values from a database, those replacements should be
// performed in this phase too. Thereafter, complete source
// configuration settings are available in memory as an instance of the
// C type parameter.
// The `loader` function may be reified by cfgN.Load functions.
//
// Second, an upwards or downwards migrator object should be created.
// This step is responsible to migrate from the current minor version
// to the latest supported minor version within the same major version.
// However, since configuration settings are loaded from YAML files,
// all minor version changes which just add optional fields with respect
// to their previous versions (in the same major version) can be loaded
// similarly and optional fields may take their default (or zero) values
// with no special migration logic. Therefore, this step just needs to
// create the proper upwards/downwards migrator objects. It is not
// possible to handle both directions in one object because each
// migrator object needs to create an instance of the upper/downer
// version and so has to import its package. So their implementation in
// one package requires all distinct major versions to be implemented
// in one package. Comparing the source/destination semantic versions,
// an appropriate migration direction can be selected. If versions are
// the same (considering their major version alone), both directions
// may be selected with no difference.
// For the second phase, upmigN.NewUpMig and dnmigN.NewDnMig functions
// can be used, as they accept a C instance, wrap it by an upwards or
// downwards Migrator object, and adapt it with a relevant Adapter
// struct exposing the repo.UpMigrator[migrationuc.Settings] or
// repo.DownMigrator[migrationuc.Settings] interface respectively.
//
// Third, after migrating settings upwards/downwards over the major
// versions and obtaining a migrator object with the destination major
// version (and its latest supported minor and patch versions), the
// repo.Settler[migrationuc.Settings] interface which is implemented by
// all upwards/downwards migrators (from the second phase) can be used
// to obtain the migration settler instance. For configuration settings,
// the settler instance has the migrationuc.Settings type. That is, it
// simply needs to return the prepared C instance (at target version)
// with proper adaptation to the migrationuc.Settings interface while
// the possibly extra fields (if an older minor version was desired) can
// be ignored. Caller is responsible to merge the migrationuc.Settings
// instance with another Settings instance which contains the default
// values at target version (using the MergeSettings method) in order to
// fill its uninitialized fields (if there were fields which could not
// find a relevant value based on their previous/next version) and also
// fill those fields which must match with their destination values
// unconditionally (such as the database connection information which
// must be renewed after each migration).
//
// Migrator[C] struct consolidates individual C type dependent
// loading and adaptation functions from the cfgN, upmigN, and dnmigN
// packages and provides repo.Migrator[migrationuc.Settings] interface
// so all supported major versions can be handled uniformly by the
// LoadMigrator function.
type Migrator[C settings.Config[C]] struct {
	data    []byte
	loader  func(data []byte) (C, error)
	upmiger func(c C) repo.UpMigrator[migrationuc.Settings]
	dnmiger func(c C) repo.DownMigrator[migrationuc.Settings]

	c *C
}

// Load uses a C loader function (which must be provided during the
// instantiation of Migrator[C] struct) in order to parse the settings
// data byte slice (which is known from the instantiation time too).
// Parsed C instance will be kept in the `m` instance. If such an
// instance was parsed/loaded previously, Load returns nil and causes
// no changes. Returned error (if any) is from the underlying `loader`
// function.
// If loading and overriding of settings from the database contents are
// desired, creation of a temporary database connection, and loading
// database information (e.g., using a LoadFromDB method) should be
// performed by this method too.
func (m *Migrator[C]) Load(ctx context.Context) error {
	if m.c != nil {
		return nil
	}
	c, err := m.loader(m.data)
	if err != nil {
		return err
	}
	m.c = &c
	return nil
}

// MajorVersion returns the major semantic version of this Migrator[C]
// instance. It reflects the major version of a configuration file and
// its value only depends on the C config type. This method may be
// used for identification of the migration versions path, passing
// through the major versions one by one.
func (m *Migrator[C]) MajorVersion() uint {
	var c C
	// MajorVersion only depends on the type of C and so can be called
	// without calling Load and obtaining a non-zero instance of C
	return c.MajorVersion()
}

// UpMigrator creates a new upwards migrator object, wrapping and
// adapting the previously loaded C instance (or returns an error if
// a C instance was not loaded previously by Load method), and returns
// it as a repo.UpMigrator[migrationuc.Settings] interface.
// This upwards migrator contains the C configuration settings at the
// source format version and may convert them to their next major
// version, moving one version forward at a time.
func (m *Migrator[C]) UpMigrator(ctx context.Context) (
	repo.UpMigrator[migrationuc.Settings], error,
) {
	if m.c == nil {
		return nil, errors.New("settings are not loaded yet")
	}
	return m.upmiger(*m.c), nil
}

// DownMigrator creates a new downwards migrator object, wrapping and
// adapting the previously loaded C instance (or returns an error if
// a C instance was not loaded previously by Load method), and returns
// it as a repo.DownMigrator[migrationuc.Settings] interface.
// This downwards migrator contains the C configuration settings at the
// source format version and may convert them to their previous major
// version, moving one version backward at a time.
func (m *Migrator[C]) DownMigrator(ctx context.Context) (
	repo.DownMigrator[migrationuc.Settings], error,
) {
	if m.c == nil {
		return nil, errors.New("settings are not loaded yet")
	}
	return m.dnmiger(*m.c), nil
}

// Settler is a helper method which creates an upwards or downwards
// migrator object and then calls its Settler method in order to obtain
// a migrationuc.Settings instance at the current major version.
// This method is useful if the source and destination versions have the
// same major version. Nevethreless, the minor and patch versions will
// be migrated to their latest supported versions.
func (m *Migrator[C]) Settler(
	ctx context.Context,
) (migrationuc.Settings, error) {
	s, err := m.UpMigrator(ctx)
	if err != nil {
		return nil, err
	}
	return s.Settler(), nil
}
