// Copyright (c) 2024 Behnam Momeni
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Package cfg2 makes it possible to load configuration settings with
// version 2.x.y since all minor and patch versions (which are known)
// with the same major version, can be loaded with one implementation.
// When trying to serialize and write out settings, the latest known
// minor and patch version will be used since older versions (with the
// same major version) can ignore the extra fields too.
package cfg2

import (
	"context"
	"fmt"
	"time"

	"github.com/momeni/clean-arch/pkg/adapter/config/cfg1"
	"github.com/momeni/clean-arch/pkg/adapter/config/settings"
	"github.com/momeni/clean-arch/pkg/adapter/config/vers"
	"github.com/momeni/clean-arch/pkg/adapter/db/postgres/migration"
	"github.com/momeni/clean-arch/pkg/core/model"
	"github.com/momeni/clean-arch/pkg/core/repo"
	"github.com/momeni/clean-arch/pkg/core/usecase/carsuc"
	"gopkg.in/yaml.v3"
)

// These constants define the major, minor, and patch version of the
// configuration settings which are supported by the Config struct.
const (
	Major = 2
	Minor = 0
	Patch = 0
)

// Version is the semantic version of Config struct.
var Version = model.SemVer{Major, Minor, Patch}

// Config contains all settings which are required by different parts
// of the project following the v2.x.y format, such as adapters or
// use cases. It is preferred to implement Config with primitive fields
// or other structs which are defined locally, not models or structs
// which are defined in lower layers, so the configuration can be
// versioned and kept intact while other layers can change freely.
// This version (when freezed and no further minor or patch release
// of it was supposed acceptable) may be embedded by the future config
// versions (if they need to copy some parts of this config version).
type Config struct {
	Database cfg1.Database // PostgreSQL database connection settings
	Gin      cfg1.Gin      // Gin-Gonic instantiation settings
	Usecases Usecases      // Supported use cases configuration settings

	// Vers contains the configuration file and database schema version
	// strings corresponding to this Config instance and its Database
	// target.
	Vers vers.Config `yaml:",inline"`
}

// ConnectionPool creates a database connection pool using the
// connection information which are kept in the `c` settings.
func (c *Config) ConnectionPool(
	ctx context.Context, r repo.Role,
) (repo.Pool, error) {
	return c.Database.ConnectionPool(ctx, r)
}

// ConnectionInfo returns the host, port, and database name of the
// connection information which are kept in this Config instance.
func (c *Config) ConnectionInfo() (dbName, host string, port int) {
	return c.Database.ConnectionInfo()
}

// SchemaMigrator creates a repo.Migrator[repo.SchemaSettler] instance
// which wraps the given `tx` transaction argument and can be used for
//  1. loading the source database schema information with this
//     assumption that tx belongs to the destination database and
//     this Config instance contains the source database connection
//     information, so it can modify the destination database within
//     a given transaction and fill a schema with tables which represent
//     the source database contents (not moving data items necessarily,
//     but may create them as a foreign data wrapper, aka FDW),
//  2. creating upwards or downwards migrator objects in order to
//     transform the loaded data into their upper/lower schema versions,
//     again with minimal data transfer and using views instead of
//     tables as far as possible, while creating tables or even loading
//     data into this Golang process if it is necessary, and at last
//  3. obtaining a repo.SchemaSettler instance for the target schema
//     major version, so it can persist the target schema version by
//     creating tables and filling them with contents of the
//     corresponding views.
func (c *Config) SchemaMigrator(tx repo.Tx) (
	repo.Migrator[repo.SchemaSettler], error,
) {
	return c.Database.SchemaMigrator(tx, c.SchemaVersion())
}

// SchemaInitializer creates a repo.SchemaInitializer instance which
// wraps the given transaction argument and can be used to initialize
// the database with development or production suitable data. The format
// of the created tables and their initial data rows are chosen based
// on the database schema version, as indicated by SchemaVersion method.
// All table creation and data insertion operations will be performed
// in the given transaction and will be persisted only if that
// transaction could commit successfully.
func (c *Config) SchemaInitializer(tx repo.Tx) (
	repo.SchemaInitializer, error,
) {
	return migration.NewInitializer(tx, c.SchemaVersion())
}

// RenewPasswords generates new secure passwords for the given roles
// and after recording them in a temporary file, will use the change
// function in order to update the passwords of those roles in the
// database too. The change function argument should perform the
// update operation in a transaction which may or may not be committed
// when RenewPasswords returns. In case of a successful commitment,
// the temporary passwords file should be moved over the main passwords
// file. The temporary passwords file is named as .pgpass.new and the
// main passwords file is named as .pgpass in this version. Keeping
// the .pgpass file (in the `c.Database.PassDir`) up-to-date, makes it
// possible to use ConnectionPool method again (both if the passwords
// are or are not updated successfully). This final file movement can
// be performed using the returned finalizer function.
func (c *Config) RenewPasswords(
	ctx context.Context,
	change func(
		ctx context.Context, roles []repo.Role, passwords []string,
	) error,
	roles ...repo.Role,
) (finalizer func() error, err error) {
	return c.Database.RenewPasswords(ctx, change, roles...)
}

// SchemaVersion returns the semantic version of the database schema
// which its connection information are kept by this Config struct.
// There is no direct dependency between the configuration file and
// database schema versions.
func (c *Config) SchemaVersion() model.SemVer {
	return c.Vers.Versions.Database
}

// SetSchemaVersion updates the semantic version of the database
// schema as recorded in this Config instance and reported by the
// SchemaVersion method.
func (c *Config) SetSchemaVersion(sv model.SemVer) {
	c.Vers.Versions.Database = sv
}

// Usecases contains the configuration settings for all use cases.
type Usecases struct {
	Cars Cars // cars use cases related settings
}

// Cars contains the configuration settings for the cars use cases.
// Fields are defined as pointers, so it is possible to detect if they
// are or are not initialized. After migrating from some configuration
// settings version, some settings may be left uninitialized because
// they may have no corresponding items in the source settings version.
// Those items can be detected as nil pointers and filled by their
// default values using the MergeConfig method.
type Cars struct {
	// DelayOfOPM indicates the amount of delay that an
	// old parking method should incur.
	DelayOfOPM *settings.Duration `yaml:"delay-of-old-parking-method"`
}

// NewUseCase instantiates a new cars use case based on the settings
// in the c struct.
func (c Cars) NewUseCase(
	p repo.Pool, r repo.Cars,
) (*carsuc.UseCase, error) {
	opts := make([]carsuc.Option, 0, 1)
	if c.DelayOfOPM != nil {
		d := time.Duration(*c.DelayOfOPM)
		opts = append(opts, carsuc.WithOldParkingMethodDelay(d))
	}
	return carsuc.New(p, r, opts...)
}

// Load unmarshals the data byte slice and loads a Config instance
// assuming that it contains the Config settings. Extra items in the
// data will be ignored and missing items will take their default
// values. Thereafter, loaded Config will be validated and normalized
// in order to ensure that provided settings are acceptable (for example
// the major version which is reported by data settings must match
// with number 2 which is the major version of this config package).
//
// If some settings should be overridden by environment variables,
// this method is the proper place for that replacement. However, if
// settings should be overridden by some information from the database,
// they must not be replaced here because the Load method provides
// those settings which are fixed by each execution (while the database
// contents may change continually and their loading must be performed
// by a separate method, such as LoadFromDB).
func Load(data []byte) (*Config, error) {
	c := &Config{}
	if err := yaml.Unmarshal(data, c); err != nil {
		return nil, fmt.Errorf("unmarshalling yaml: %w", err)
	}
	if err := c.ValidateAndNormalize(); err != nil {
		return nil, fmt.Errorf("validating configs: %w", err)
	}
	return c, nil
}

// ValidateAndNormalize validates the configuration settings and
// returns an error if they were not acceptable. It can also modify
// settings in order to normalize them or replace some zero values with
// their expected default values (if any).
func (c *Config) ValidateAndNormalize() error {
	v := c.Vers.Versions.Config
	if v[0] != Major {
		return fmt.Errorf(
			"major version is %d instead of %d", v[0], Major,
		)
	}
	if v[1] > Minor {
		return fmt.Errorf("minor version %d is not supported", v[1])
	}
	settings.Nil2Zero(&c.Gin.Logger)
	settings.Nil2Zero(&c.Gin.Recovery)
	// No need to check for c.Usecases.Cars.DelayOfOPM == nil
	// because it has no default in adapters layer.
	return nil
}

// Marshalled struct contains a field for each one of the Config struct
// fields. The field names may be different for simplicity, but the
// yaml tag of fields are chosen to have consistent names after the
// serialization operation. The types of those fields are the same if
// their default serialization format is acceptable, otherwise, they
// will be serialized manually using the Marshal method and their
// target primitive types will be used in the Marshalled struct.
type Marshalled struct {
	Database cfg1.Database
	Gin      cfg1.Gin
	Usecases struct {
		Cars struct {
			Delay *string `yaml:"delay-of-old-parking-method,omitempty"`
		}
	}
	Vers *vers.Marshalled `yaml:",inline"`
}

// MarshalYAML returns an instance of the Marshalled struct, as created
// by the Marshal method, so it may be marshalled instead of the `c`
// Config instance. This replacement makes it possible to substitute
// specific settings such as a slices of numbers in a vers.Config with
// their alternative primitive data types and have control on the final
// serialization result.
//
// See the Marshal function for the reification details and how
// marshaling logic can be distributed among nested Config structs.
func (c *Config) MarshalYAML() (interface{}, error) {
	return c.Marshal(), nil
}

// Marshal creates an instance of the Marshalled struct and fills it
// with the `c` Config instance contents. The Marshalled and Config
// fields do correspond with each other with one difference. Any field
// which requires a specific MarshalYAML logic (and its default encoding
// logic into YAML format is not suitable) is replaced by a primitive
// data type, so it can contain the properly serialized version of that
// field.
//
// This Marshal method encodes and replaces fields which are defined in
// this package and recursively calls Marshal method on those fields
// which are defined in other packages. Therefore, the marshaling logic
// can be distributed among packages, near to the relevant data types
// (while MarshalYAML from the yaml.Marshaler interface is only called
// for the top-most object and is ignored for nested types).
func (c *Config) Marshal() *Marshalled {
	m := &Marshalled{}
	m.Database = c.Database
	m.Gin = c.Gin
	m.Usecases.Cars.Delay = c.Usecases.Cars.DelayOfOPM.Marshal()
	m.Vers = c.Vers.Marshal()
	return m
}

// Dereference returns the `c` Config instance itself.
//
// Methods of the Config struct refer to other types based on this
// package Major version for complete type-safety. For example, the
// MergeConfig only accepts an instance of Config from this package
// and passing a cfg1.Config instance will be rejected at compile time.
// However, the use cases layer which does not know about the config
// version at compile time has to receive Config as an abstract
// interface which is common among all config versions. That abstract
// interface is defined as pkg/core/usecase/migrationuc.Settings which
// provides MergeSettings method instead of MergeConfig and accepts
// an instance of Settings interface instead of the Config instance.
// The pkg/adapter/config/settings.Adapter[Config] is defined in order
// to wrap a Config instance and implement the Settings instance.
//
// Presence of the Dereference method allows users of the Config struct
// and the Adapter[Config] struct to use them uniformly. Indeed, both
// of the raw Config and its wrapper Adapter[Config] instances can be
// represented by pkg/adapter/config/settings.Dereferencer[Config]
// interface and so the wrapped Config instance may be obtained from
// them using the Dereference method. Note that a type assertion from
// the Settings interface to the Adapter instance requires pre-knowledge
// about the Adapter (and a Settings interface which is provided by some
// other adapter implementation may not be supported), while the
// Dereferencer[Config] interface can be provided by any adapter
// implementation simply by embedding the Config instance.
func (c *Config) Dereference() *Config {
	return c
}

// Clone creates a new instance of Config and initializes its fields
// based on the `c` fields. Pointers are renewed too, so changes in
// the returned Config instance and `c` stay independent.
func (c *Config) Clone() *Config {
	cc := &Config{
		Database: c.Database,
		Vers:     c.Vers,
	}
	if c.Gin.Logger != nil {
		l := *c.Gin.Logger
		cc.Gin.Logger = &l
	}
	if c.Gin.Recovery != nil {
		r := *c.Gin.Recovery
		cc.Gin.Recovery = &r
	}
	if c.Usecases.Cars.DelayOfOPM != nil {
		doo := *c.Usecases.Cars.DelayOfOPM
		cc.Usecases.Cars.DelayOfOPM = &doo
	}
	return cc
}

// MergeConfig overwrites all fields of `c` which are not initialized
// (and have nil value) with their corresponding values from `c2` arg.
// The `c` config version will be set to the latest known version values
// as specified by Major, Minor, and Patch constants in this package.
// All database settings in `c` are overwritten by the `c2` values
// unconditionally. The database version number will be set to its
// latest supported version too, having the same major version as
// specified in `c2` instance.
func (c *Config) MergeConfig(c2 *Config) error {
	c.Database = c2.Database
	settings.OverwriteNil(&c.Gin.Logger, c2.Gin.Logger)
	settings.OverwriteNil(&c.Gin.Recovery, c2.Gin.Recovery)
	settings.OverwriteNil(
		&c.Usecases.Cars.DelayOfOPM, c2.Usecases.Cars.DelayOfOPM,
	)
	c.Vers.Versions.Config = model.SemVer{Major, Minor, Patch}
	sv, err := migration.LatestVersion(c2.SchemaVersion())
	if err != nil {
		return err
	}
	c.Vers.Versions.Database = sv
	return nil
}

// Version returns the semantic version of this Config struct contents
// which its major version is equal to 2, while its minor and patch
// versions may correspond to the Minor and Patch constants or may
// describe an older version (if the minor version of the returned
// semantic version was more recent than Minor constant, it could not
// be loaded by the Load function). By the way, no constraint exists on
// the patch version because it has no visible effect.
func (c *Config) Version() model.SemVer {
	return c.Vers.Versions.Config
}

// MajorVersion returns the major semantic version of this Config
// instance. This value matches with the first component of the version
// which is returned by the Version method. However, the Version method
// returns the complete semantic version as written in a configuration
// file, hence, it cannot be called without creating an instance of
// Config first. In contrast, this method only depends on the Config
// type and so can be called with a nil instance too.
func (c *Config) MajorVersion() uint {
	return Major
}
