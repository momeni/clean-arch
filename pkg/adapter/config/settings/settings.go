// Copyright (c) 2024 Behnam Momeni
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Package settings provides generic interfaces which should be
// implemented by each configuration settings major version type, and
// their upwards and downwards migrator types.
// This package also provides a generic Adapter type which can adapt
// version specific cfgN.Config structs into version-independent
// migrationuc.Settings interface, so they can be passed to the
// use cases layer.
package settings

import (
	"context"
	"fmt"

	"github.com/momeni/clean-arch/pkg/core/model"
	"github.com/momeni/clean-arch/pkg/core/repo"
	"github.com/momeni/clean-arch/pkg/core/usecase/migrationuc"
	"gopkg.in/yaml.v3"
)

// Dereferencer is a generic interface which its C type parameter
// represents a wrapped type (e.g., pkg1.Config struct). This interface
// can be implemented by a wrapper type which wants to be dereferenced
// and return its wrapped object.
// The ideal place for implementing the Dereference method is on the
// wrapped object's type itself, so any type which embeds it may
// inherit the Dereference method and implement the Dereferencer
// automatically.
type Dereferencer[C any] interface {
	// Dereference returns an instance of the `C` type which is wrapped
	// by the current object, so the wrapped object may be fetched.
	//
	// Methods of a wrapped struct may need to refer to other types
	// based on some form of grouping, e.g., cfg1.Config can be merged
	// by instances of cfg1.Config but may not be merged by cfg2.Config
	// instances. Such wrapped structs have to be adapted by an abstract
	// interface, so they can be used uniformly.
	// Presence of the Dereference method allows the wrapped struct and
	// its wrapper adapter type to be used uniformly. Indeed, the raw
	// and wrapper adapter instances can be represented by Dereferencer
	// interface, so the wrapped C instance may be obtained. Note that
	// type assertion from an abstract interface (implemented by the
	// wrapper adapter type) requires pre-knowledge about the adapter
	// type, while the Dereferencer[C] interface can be provided by
	// any adapter implementation simply by embedding the C instance.
	Dereference() C
}

// Config of C describes the expected interface of Config structs.
// It contains the database-related settings by embedding SchemaSettings
// interface, controls how the Config should be serialized to YAML using
// the Marshaler interface, keeps the main Config object accessible
// even when it is embedded by the Adapter[C] struct since it embeds
// the Dereferencer[C] interface, supports merging the destination
// configuration settings (only from the same version C struct) by the
// MergeConfig method, and reports the current configuration version
// by the Version method.
// All pkg/adapter/config/cfgN.Config structs must implement this
// interface.
//
// This interface usage is twofold. First, asserting that each Config
// struct implements it (in its test.go file) ensures that all relevant
// methods are implemented with the proper types. For example, if the
// implementation of cfg1 package was copied in order to create cfg2
// package, updating the test file assertion to Config[*cfg2.Config]
// type is enough to produce a compilation error while the relevant
// MergeConfig type is not updated.
// Second, it unifies different cfgN.Config versions, so they can be
// wrapped by a generic Adapter[C] struct instead of having to implement
// a distinct Adapter type per C type.
type Config[C any] interface {
	migrationuc.SchemaSettings
	yaml.Marshaler
	Dereferencer[C]

	// Clone creates a deep copy of this configuration instance, so
	// its fields can be changed without updating this instance.
	Clone() C

	// MergeConfig overwrites all fields of this Config[C] instance
	// which are not initialized (and so have nil value) with their
	// corresponding values from the `c` argument. The version value is
	// also set to the latest known version values (keeping the major
	// version matched with the C type parameter version).
	// All database settings are copied from the `c` argument
	// unconditionally because after a migration, new tables are placed
	// in the target database which its connection information is
	// presented by the `c` argument. The database version number will
	// be set to its latest supported version too, having the same major
	// version as specified in the `c` instance.
	MergeConfig(c C) error

	// Version returns the semantic version of this Config[C] struct
	// contents. Returned version corresponds to one of the supported
	// config major versions. The minor and patch versions may
	// correspond to the Minor and Patch constants of the relevant
	// package or may describe an older version (newer/unknown minor
	// versions are rejected by the Load function). The patch version
	// has no special constraint since it has no visible effect.
	Version() model.SemVer

	// MajorVersion returns the major semantic version of this Config[C]
	// struct. This value matches with the first component of the
	// version which is returned by the Version method. However, the
	// Version method returns the complete semantic version as written
	// in a configuration file, hence, it cannot be called without
	// creating an instance of Config[C] first. In contrast, this
	// method only depends on the C type and so can be called with a nil
	// instance of the C type too.
	MajorVersion() uint
}

// Adapter of C is a generic struct which wraps and adapts a Config[C]
// for any C which exposes the Config[C] interface in order to
// provide the pkg/core/usecase/migrationuc.Settings interface.
//
// All methods of the Settings interface are provided by the Config[C]
// interface, hence, only the MergeSettings method should be implemented
// by Adapter[C] struct.
type Adapter[C Config[C]] struct {
	Config[C]
}

// Clone creates a deep copy of this Adapter[C] instance by first
// cloning `a.Config` and then wrapping it in a new Adapter[C] instance.
func (a Adapter[C]) Clone() migrationuc.Settings {
	c := a.Config.Clone()
	return Adapter[C]{c}
}

// MergeSettings expects to receive a Settings instance which has
// the same settings version as the current instance, matching with
// the C type parameter version. For this purpose, the `s` argument is
// expected to implement the Dereferencer[C] interface (which is the
// case for Config[C] and also the Adapter[C] instances). When migrating
// settings to newer or older versions, some of the target version
// settings may be left uninitialized. This method fills those items
// with their correspoding values from the given `s` argument.
// Some settings, such as the database connection information, are
// unconditionally taken from the `s` argument because they need to
// describe the destination settings values.
func (a Adapter[C]) MergeSettings(s migrationuc.Settings) error {
	c, ok := s.(Dereferencer[C])
	if !ok {
		return fmt.Errorf("unsupported settings type: %T", s)
	}
	return a.MergeConfig(c.Dereference())
}

// UpMigrator of C and U describes the expected interface of a Config
// upwards migrator implementation. The C type parameter specifies a
// Config implementation like *cfg1.Config and so it exposes the
// Config[C] generic interface itself.
// The upwards migrators are recognized by one main function, MigrateUp,
// which converts the contained config object (with type C) to its next
// version (if any) and then wraps it with another upwards migrator type
// which is represented by the U type parameter and returns it, so the
// config object can be migrated one major version upwards at a time.
// If there is no upper/next version, an error will be returned.
// All pkg/adapter/config/up/upmigN.Migrator structs must implement this
// interface. It also embeds the repo.Settler[C] interface, so after
// migrating upwards enough and when the target major version was
// achieved, the Settler method can be called and the ultimate C config
// instance can be fetched.
//
// Similar to the Config[C] which was useful to assert that each Config
// struct implements the relevant methods based on the C type parameter,
// the UpMigrator[C, U] is useful to assert that a Settler and MigrateUp
// methods with proper return value types are implemented by the
// corresponding upmigN.Migrator structs (which may be copied from the
// previous migrator versions, whenever a newer major version is
// required, and should get a compilation error as while as its methods
// are not updated properly).
//
// However, despite the Config[C] interface, it is not possible to use
// the UpMigrator[C, U] for unification of all upmigN.Migrator structs.
// The reason is in the limited generic types support in the Golang as
// it is not possible to deduce a list of type parameters at compile
// time or provide a variety list of type parameters. Therefore, for
// each upmigN.Migrator type which has X subsequently released major
// versions, it is required to define U using X nested type parameters.
// The appropriate C type and the nested UpMigrator[C, U] (for a U type
// parameter which depends on the X subsequent UpMigrator types) are
// defined as "C" and "Type" type aliases in each upmigN package.
//
// The goal is to adapt each UpMigrator[C, U] in order to provide the
// pkg/core/repo.UpMigrator[pkg/core/usecase/migrationuc.Settings]
// interface, so it can be passed to the use cases layer independent
// of its configuration settings format version.
// Because a U instance (with "any" constraint) cannot be adapted
// without knowing about the inner types of U type parameter, a
// distinct Adapter struct is implemented in each upmigN package.
type UpMigrator[C Config[C], U any] interface {
	repo.Settler[C]

	// MigrateUp migrates the contained config object, with C type,
	// into its next/upper major version (if any) and wraps it with
	// another upwards migrator implementation, with U type.
	// If there is no upper major version, an error will be returned.
	// After reaching to the target major version, the Settler method
	// from the repo.Settler[C] interface can be used to obtain the C
	// config instance.
	MigrateUp(ctx context.Context) (U, error)
}

// DownMigrator of C and D describes the expected interface of a Config
// downwards migrator implementation. The C type parameter specifies
// a Config implementation like *cfg1.Config and so it exposes the
// Config[C] generic interface itself.
// The downwards migrators are recognized by one main function,
// namely MigrateDown, which converts the contained config object (with
// type C) to its previous version (if any) and then wraps it with
// another downwards migrator type which is represented by the D type
// parameter and returns it, so the config object can be migrated one
// major version downwards at a time.
// If there is no downer/previous version, an error will be returned.
// All pkg/adapter/config/down/dnmigN.Migrator structs must implement
// this interface. It also embeds the repo.Settler[C] interface, so after
// migrating downwards enough and when the target major version was
// achieved, the Settler method can be called and the ultimate C config
// instance can be fetched.
//
// Similar to the Config[C] which was useful to assert that each Config
// struct implements the relevant methods based on the C type parameter,
// the DownMigrator[C, D] is useful to assert that a Settler and
// MigrateDown methods with proper return value types are implemented
// by the corresponding dnmigN.Migrator structs (which may be copied
// from the previous migrator versions, whenever a newer major version
// is required, and should get a compilation error as while as its
// methods are not updated properly).
//
// However, despite the Config[C] interface, it is not possible to use
// the DownMigrator[C, D] for unification of all dnmigN.Migrator
// structs. The reason is in the limited generic types support in the
// Golang as it is not possible to deduce a list of type parameters at
// compile time or provide a variety list of type parameters. Therefore,
// for each dnmigN.Migrator type which has X older major versions, it
// is required to define D using X nested type parameters.
// The appropriate C type and the nested DownMigrator[C, D] (for a D
// type parameter which depends on the X older DownMigrator types) are
// defined as "C" and "Type" type aliases in each dnmigN package.
//
// The goal is to adapt each DownMigrator[C, D] in order to provide the
// pkg/core/repo.DownMigrator[pkg/core/usecase/migrationuc.Settings]
// interface, so it can be passed to the use cases layer independent
// of its configuration settings format version.
// Because a D instance (with "any" constraint) cannot be adapted
// without knowing about the inner types of D type parameter, a
// distinct Adapter struct is implemented in each dnmigN package.
type DownMigrator[C Config[C], D any] interface {
	repo.Settler[C]

	// MigrateDown migrates the contained config object, with C type,
	// into its previous/downer major version (if any) and wraps it
	// with another downwards migrator implementation, with the D type.
	// If there is no older major version, an error will be returned.
	// After reaching to the target major version, the Settler method
	// from the repo.Settler[C] interface can be used to obtain the C
	// config instance.
	MigrateDown(ctx context.Context) (D, error)
}
