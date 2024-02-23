// Copyright (c) 2024 Behnam Momeni
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Package vers contains the common versions parsing which is required
// by all config versions. Two versions are tracked here, namely the
// configuration files and database schema (and other aspects which may
// need migration support can be added later). The idea is that versions
// should be known before trying to obtain and parse the actual data,
// so the actual data format can be known and verified when loading them
// and although the format of keeping versions may change too, but it is
// less likely to change over time.
package vers

import (
	"fmt"

	"github.com/momeni/clean-arch/pkg/core/model"
	"gopkg.in/yaml.v3"
)

// Config contains the versions of those system components which should
// be supported in the migration operations. It may be embedded with
// inline format in the released config struct versions in order to
// indicate their versions and relevant items format.
type Config struct {
	Versions Versions `yaml:"versions"`
}

// Versions contains the configuration file and database schema versions
// which are used for detecting their relevant formats.
// Although different versions may be supported by system migration
// operations, each binary only supports the latest version which is
// known to that binary for its non-migration operations.
type Versions struct {
	Database model.SemVer `yaml:"database"`
	Config   model.SemVer `yaml:"config"`
}

// Marshalled is an alternative form of Config struct which replaces
// its model.SemVer inner fields by their string representation.
// The Marshalled is used during the YAML serialization operation
// instead of the main Config struct.
//
// The yaml.Marshaler interface with MarshalYAML method may be used
// by a struct in order to provide an alternative value to be serialized
// instead of the main value at hand. However, MarshalYAML is only used
// for replacing the most top level value, where the serialization begin
// from. The Marshalled struct (returned by the Marshal method) allows
// multiple nested structs to participate in this value replacement
// during a serialization operation.
type Marshalled struct {
	Versions struct {
		Database string
		Config   string
	}
}

// Marshal creates and returns a Marshalled instance representing the vc
// Config instance. It may be serialized instead of vc to YAML format.
func (vc *Config) Marshal() *Marshalled {
	m := &Marshalled{}
	m.Versions.Database = vc.Versions.Database.Marshal()
	m.Versions.Config = vc.Versions.Config.Marshal()
	return m
}

// Load deserializes the data byte slice into a new instance of Config
// struct. Of course, data may contain extra fields which will be
// ignored. The deserialized version fields (in the returned Config)
// can be used to detect the format of other settings in the data and
// complete deserialization of the remaining fields.
func Load(data []byte) (*Config, error) {
	vc := &Config{}
	if err := yaml.Unmarshal(data, vc); err != nil {
		return nil, err
	}
	return vc, nil
}

// Validate returns an error if the configuration settings version which
// is stored in the `vc` Config instance is not supported by the given
// major and minor version arguments. That is, stored major version
// must match with the major argument and the stored minor version must
// be at most equal with the given minor version (not newer than it).
func (vc *Config) Validate(major, minor uint) error {
	v := vc.Versions.Config
	if v[0] != major {
		return fmt.Errorf("incompatible major version: %d", v[0])
	}
	if v[1] > minor {
		return fmt.Errorf("unsupported minor version: %d", v[1])
	}
	return nil
}
