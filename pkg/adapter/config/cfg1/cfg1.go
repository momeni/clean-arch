// Copyright (c) 2024 Behnam Momeni
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Package cfg1 makes it possible to load configuration settings with
// version 1.x.y since all minor and patch versions (which are known)
// with the same major version, can be loaded with one implementation.
// When trying to serialize and write out settings, the latest known
// minor and patch version will be used since older versions (with the
// same major version) can ignore the extra fields too.
package cfg1

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/momeni/clean-arch/pkg/adapter/config/settings"
	"github.com/momeni/clean-arch/pkg/adapter/config/vers"
	"github.com/momeni/clean-arch/pkg/adapter/db/postgres"
	"github.com/momeni/clean-arch/pkg/adapter/db/postgres/migration"
	"github.com/momeni/clean-arch/pkg/adapter/restful/gin"
	"github.com/momeni/clean-arch/pkg/core/model"
	"github.com/momeni/clean-arch/pkg/core/repo"
	"github.com/momeni/clean-arch/pkg/core/usecase/carsuc"
	"gopkg.in/yaml.v3"
)

// These constants define the major, minor, and patch version of the
// configuration settings which are supported by the Config struct.
const (
	Major = 1
	Minor = 0
	Patch = 0
)

// Version is the semantic version of Config struct.
var Version = model.SemVer{Major, Minor, Patch}

// Config contains all settings which are required by different parts
// of the project following the v1.x.y format, such as adapters or
// use cases. It is preferred to implement Config with primitive fields
// or other structs which are defined locally, not models or structs
// which are defined in lower layers, so the configuration can be
// versioned and kept intact while other layers can change freely.
// This version (when freezed and no further minor or patch release
// of it was supposed acceptable) may be embedded by the future config
// versions (if they need to copy some parts of this config version).
type Config struct {
	Database Database // PostgreSQL database connection settings
	Gin      Gin      // Gin-Gonic instantiation settings
	Usecases Usecases // Configuration settings for supported use cases

	// Vers contains the configuration file and database schema version
	// strings corresponding to this Config instance and its Database
	// target.
	Vers vers.Config `yaml:",inline"`
}

// Database contains the database related configuration settings.
type Database struct {
	Host    string // domain name or IP address of the DBMS server
	Port    int    // port number of the DBMS server
	Name    string // database name, like caweb1_0_0
	PassDir string `yaml:"pass-dir"` // path of the passwords dir
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

// ConnectionPool creates a database connection pool using the
// connection information which are kept in the `d` settings.
// Initially, the .pgpass file in the d.PassDir folder is checked
// which should conform with the pgpass format with lines like this:
//
//	host:port:dbname:role:password
//
// If a database connection could be established, created pool and nil
// error will be returned. Otherwise, passwords might have been updated
// during a previous incomplete migration operation. So the .pgpass.new
// file in the same d.PassDir folder is checked too. If a connection
// could be established successfully, the .pgpass.new will be moved to
// the .pgpass file, so the .pgpass.new file may be overwritten safely
// by the subsequent migration operations.
func (d Database) ConnectionPool(
	ctx context.Context, r repo.Role,
) (repo.Pool, error) {
	path := filepath.Join(d.PassDir, ".pgpass")
	u, err := d.ConnectionURL(r, path)
	if err != nil {
		return nil, fmt.Errorf("using %q pass-file: %w", path, err)
	}
	p, err := postgres.NewPool(ctx, u)
	if err == nil {
		return p, nil
	}
	fmt.Printf("failed to connect with %q: %v\n", path, err)
	newPath := filepath.Join(d.PassDir, ".pgpass.new")
	fmt.Printf("now, trying to connect with %q\n", newPath)
	u, err = d.ConnectionURL(r, newPath)
	if err != nil {
		return nil, fmt.Errorf("using %q pass-file: %w", newPath, err)
	}
	p, err = postgres.NewPool(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("can use neither pass-file: %w", err)
	}
	if err = os.Rename(newPath, path); err != nil {
		p.Close()
		return nil, fmt.Errorf("os.Rename: %w", err)
	}
	return p, nil
}

// ConnectionURL returns the database connection URL embedding the host,
// port, role name, database name, and password value. These items are
// directly taken from the `d` settings, but the role name which is
// specified by the `r` argument and the password value which is read
// from the given `path` file. Returned URL has the postgresql scheme.
// The `path` file may contain empty or `#`-commented lines in addition
// to the password specifying lines which should conform with the pgpass
// files format with lines like this:
//
//	host:port:dbname:role:password
//
// If the `path` file could be read and a password for the asked `r`
// role could be identified, a URL and a nil error will be returned.
// Otherwise, returned string will be empty and error will describe the
// wrapped error condition.
func (d Database) ConnectionURL(
	r repo.Role, path string,
) (string, error) {
	passLines, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("reading pass-file: %w", err)
	}
	prfx := fmt.Sprintf("%s:%d:%s:%s:", d.Host, d.Port, d.Name, r)
	var pass string
	for _, line := range strings.Split(string(passLines), "\n") {
		if line == "" || line[0] == '#' {
			continue
		}
		if strings.HasPrefix(line, prfx) {
			pass = line[len(prfx):]
			break
		}
	}
	if pass == "" {
		return "", fmt.Errorf("no matching password line")
	}
	u := url.URL{
		Scheme: "postgresql",
		User:   url.UserPassword(string(r), pass),
		Host:   fmt.Sprintf("%s:%d", d.Host, d.Port),
		Path:   d.Name,
	}
	return u.String(), nil
}

// ConnectionInfo returns the host, port, and database name of the
// connection information which are kept in this Database instance.
func (d Database) ConnectionInfo() (dbName, host string, port int) {
	return d.Name, d.Host, d.Port
}

// SchemaMigrator creates a repo.Migrator[repo.SchemaSettler] instance
// which wraps the given `tx` transaction argument and can be used for
//  1. loading the source database schema information with this
//     assumption that tx belongs to the destination database and
//     this `d` instance contains the Database connection information
//     for the source database, so it can modify the destination
//     database within a transaction and fill a schema with tables and
//     views which represent the source database contents (not moving
//     data items necessarily, but may create them as a foreign data
//     wrapper, aka FDW too),
//  2. creating upwards or downwards migrator objects in order to
//     transform the loaded data into their upper/lower schema versions,
//     again with minimal data transfer and using views instead of
//     tables as far as possible, while creating tables or even loading
//     data into this Golang process if it is necessary, and at last
//  3. obtaining a repo.SchemaSettler instance for the target schema
//     major version, so it can persist the target schema version by
//     creating tables and filling them with contents of the
//     corresponding views.
func (d Database) SchemaMigrator(tx repo.Tx, srcDBVer model.SemVer) (
	repo.Migrator[repo.SchemaSettler], error,
) {
	path := filepath.Join(d.PassDir, ".pgpass")
	u, err := d.ConnectionURL(repo.NormalRole, path)
	if err != nil {
		return nil, fmt.Errorf("using %q pass-file: %w", path, err)
	}
	return migration.New(tx, srcDBVer, u)
}

// RenewPasswords generates new secure passwords for the given roles
// and after recording them in a temporary file (i.e., .pgpass.new file
// in the `d.PassDir` directory), will use the `change` function in
// order to update the passwords of those `roles` in the database too.
// The `change` function argument should perform the update operation
// in a transaction which may or may not be committed when the
// RenewPasswords function returns. In case of a successful commitment,
// the temporary passwords file should be moved over the main passwords
// file (i.e., .pgpass file in the `d.PassDir` directory). Keeping the
// .pgpass file up-to-date, makes it possible to use ConnectionPool
// method again (both if the passwords are or are not updated
// successfully). This final file movement can be performed using the
// returned finalizer function.
func (d Database) RenewPasswords(
	ctx context.Context,
	change func(
		ctx context.Context, roles []repo.Role, passwords []string,
	) error,
	roles ...repo.Role,
) (finalizer func() error, err error) {
	passwords := make([]string, len(roles))
	b := make([]byte, 16) // 128 bits
	enc := base64.RawStdEncoding
	p := make([]byte, enc.EncodedLen(len(b))) // for each password
	prfx := fmt.Sprintf("%s:%d:%s", d.Host, d.Port, d.Name)
	lines := make([]string, len(passwords))
	for i, r := range roles {
		if _, err = rand.Read(b); err != nil {
			return nil, fmt.Errorf("rand.Read for i=%d: %w", i, err)
		}
		enc.Encode(p, b)
		passwords[i] = string(p)
		lines[i] = fmt.Sprintf("%s:%s:%s\n", prfx, r, passwords[i])
	}
	orgPath := filepath.Join(d.PassDir, ".pgpass")
	newPath := filepath.Join(d.PassDir, ".pgpass.new")
	finalizer = func() error {
		return os.Rename(newPath, orgPath)
	}
	err = os.WriteFile(newPath, []byte(strings.Join(lines, "")), 0o600)
	if err != nil {
		return nil, fmt.Errorf("writing %q file: %w", newPath, err)
	}
	if err = change(ctx, roles, passwords); err != nil {
		return nil, fmt.Errorf("passwords change callback: %w", err)
	}
	return finalizer, nil
}

// Gin contains the gin-gonic related configuration settings.
// Fields are defined as pointers, so it is possible to detect if they
// are or are not initialized. After migrating from some configuration
// settings version, some settings may be left uninitialized because
// they may have no corresponding items in the source settings version.
// Those items can be detected as nil pointers and filled by their
// default values using the MergeConfig method.
type Gin struct {
	Logger   *bool // Whether to register the gin.Logger() middleware
	Recovery *bool // Whether to register the gin.Recovery() middleware
}

// NewEngine instantiates a new gin-gonic engine instance based on
// the `g` settings.
func (g Gin) NewEngine() *gin.Engine {
	middlewares := make([]gin.HandlerFunc, 0, 2)
	if *g.Logger {
		middlewares = append(middlewares, gin.Logger())
	}
	if *g.Recovery {
		middlewares = append(middlewares, gin.Recovery())
	}
	return gin.New(middlewares...)
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
	// OldParkingDelay indicates the amount of delay that an old
	// parking method should incur.
	// A nil value indicates that delay is left uninitialized, so the
	// use cases layer may select a default value.
	OldParkingDelay *settings.Duration `yaml:"old-parking-method-delay"`
}

// NewUseCase instantiates a new cars use case based on the settings
// in the `c` struct.
func (c Cars) NewUseCase(
	p repo.Pool, r repo.Cars,
) (*carsuc.UseCase, error) {
	opts := make([]carsuc.Option, 0, 1)
	if c.OldParkingDelay != nil {
		d := time.Duration(*c.OldParkingDelay)
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
// with number 1 which is the major version of this config package).
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
	// No need to check for c.Usecases.Cars.OldParkingDelay == nil
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
	Database Database
	Gin      Gin
	Usecases struct {
		Cars struct {
			Delay *string `yaml:"old-parking-method-delay,omitempty"`
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
	m.Usecases.Cars.Delay = c.Usecases.Cars.OldParkingDelay.Marshal()
	m.Vers = c.Vers.Marshal()
	return m
}

// Dereference returns the `c` Config instance itself.
//
// Methods of the Config struct refer to other types based on this
// package Major version for complete type-safety. For example, the
// MergeConfig only accepts an instance of Config from this package
// and passing a cfg2.Config instance will be rejected at compile time.
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
	if c.Usecases.Cars.OldParkingDelay != nil {
		opd := *c.Usecases.Cars.OldParkingDelay
		cc.Usecases.Cars.OldParkingDelay = &opd
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
		&c.Usecases.Cars.OldParkingDelay,
		c2.Usecases.Cars.OldParkingDelay,
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
// which its major version is equal to 1, while its minor and patch
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
