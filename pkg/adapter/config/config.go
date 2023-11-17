// Copyright (c) 2023-2024 Behnam Momeni
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Package config is an adapter which accepts yaml formatted config
// files from its users and allows the caweb to instantiate different
// components, from the adapter or use cases layers, using those loaded
// configuration settings.
// These settings may be versioned and maintained by sub-packages.
// However, the parsed and validated configurations should be passed
// to their ultimate components as a series of individual params (for
// the mandatory items) and a series of functional options (for
// the optional items), so they may be accumulated and validated
// in another (possibly non-exported) config struct (or directly in the
// relevant end-component such as a UseCase instance). This design
// decision causes a bit of redundancy in favor of a defensive solution.
package config

import (
	"fmt"
	"os"

	"github.com/momeni/clean-arch/pkg/adapter/config/cfg2"
	"github.com/momeni/clean-arch/pkg/adapter/config/vers"
	"github.com/momeni/clean-arch/pkg/adapter/db/postgres"
)

// Load function loads, validates, and normalizes the configuration
// file and returns its settings as an instance of the Config struct.
// Given path must belong to a configuration file which conforms with
// the latest known configuration settings format.
// The corresponding database schema version must also match with the
// latest known database schema version.
func Load(path string) (*cfg2.Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}
	v, err := vers.Load(data)
	if err != nil {
		return nil, fmt.Errorf("loading versions: %w", err)
	}
	vc := v.Versions
	switch {
	case vc.Config != cfg2.Version:
		return nil, fmt.Errorf(
			"unexpected config version: %s", vc.Config.String(),
		)
	case vc.Database != postgres.Version:
		return nil, fmt.Errorf(
			"unexpected database schema version: %s",
			vc.Database.String(),
		)
	}
	c, err := cfg2.Load(data)
	if err != nil {
		return nil, fmt.Errorf("loading cfg2.Config: %w", err)
	}
	return c, nil
}
