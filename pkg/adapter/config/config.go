// Copyright (c) 2023 Behnam Momeni
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Package config is an adapter which allows users to write a yaml
// configuration file and allow the caweb to instantiate different
// components, from the adapter or use cases layers, using those
// configuration settings.
// These settings may be versioned and maintained by migrations later.
// However, the parsed and validated configurations should be passed
// to their ultimate components as a series of individual params (for
// the mandatory items) and a series of functional options (for
// the optional items), so they may be accumulated and validated
// in another (possibly non-exorted) config struct (or directly in the
// relevant end-component such as a UseCase instance). This design
// decision causes a bit of redundancy in favor of a defensive solution.
package config

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/momeni/clean-arch/pkg/adapter/db/postgres"
	"github.com/momeni/clean-arch/pkg/adapter/restful/gin"
	"github.com/momeni/clean-arch/pkg/core/repo"
	"github.com/momeni/clean-arch/pkg/core/usecase/carsuc"
	"gopkg.in/yaml.v3"
)

// Config contains all settings which are required by different parts
// of the project such as adapters or use cases. It is preferred to
// implement Config with primitive fields or other structs which are
// defined in this package, not models or structs which are defined in
// other layers, so the configuration can be versioned and kept intact
// while other layers can change freely. If two versions of the Config
// struct were implemented, the newer version may embed/depend on the
// older version (which is freezed).
type Config struct {
	Database Database
	Gin      Gin
	Usecases Usecases
}

// Database contains the database related configuration settings.
type Database struct {
	Host     string // domain name or IP address of the DBMS server
	Port     int    // port number of the DBMS server
	Name     string // database name, like caweb1_0_0
	Role     string // role/username for connecting to the database
	PassFile string `yaml:"pass-file"` // path of the password file
}

// NewPool instantiates a new database connection pool based on the
// connection information which are stored in d instance.
func (d Database) NewPool(ctx context.Context) (*postgres.Pool, error) {
	pass, err := os.ReadFile(d.PassFile)
	if err != nil {
		return nil, fmt.Errorf("reading pass-file: %w", err)
	}
	u := url.URL{
		Scheme: "postgresql",
		User:   url.UserPassword(d.Role, string(pass)),
		Host:   fmt.Sprintf("%s:%d", d.Host, d.Port),
		Path:   d.Name,
	}
	p, err := postgres.NewPool(ctx, u.String())
	if err != nil {
		return nil, fmt.Errorf("pool creation: %w", err)
	}
	return p, nil
}

// Gin contains the gin-gonic related configuration settings.
type Gin struct {
	Logger   bool // Whether to register the gin.Logger() middleware
	Recovery bool // Whether to register the gin.Recovery() middleware
}

// NewEngine instantiates a new gin-gonic engine instance based on
// the g settings.
func (g Gin) NewEngine() *gin.Engine {
	middlewares := make([]gin.HandlerFunc, 0, 2)
	if g.Logger {
		middlewares = append(middlewares, gin.Logger())
	}
	if g.Recovery {
		middlewares = append(middlewares, gin.Recovery())
	}
	return gin.New(middlewares...)
}

// Usecases contains the configuration settings for all use cases.
type Usecases struct {
	Cars Cars // cars use cases related settings
}

// Cars contains the configuration settings for the cars use cases.
type Cars struct {
	// OldParkingDelay indicates the amount of delay that an old
	// parking method should incur.
	OldParkingDelay *time.Duration `yaml:"old-parking-method-delay"`
}

// NewUseCase instantiates a new cars use case based on the settings
// in the c struct.
func (c Cars) NewUseCase(
	p repo.Pool, r repo.Cars,
) (*carsuc.UseCase, error) {
	opts := make([]carsuc.Option, 0, 1)
	if c.OldParkingDelay != nil {
		opts = append(
			opts,
			carsuc.WithOldParkingMethodDelay(*c.OldParkingDelay),
		)
	}
	return carsuc.New(p, r, opts...)
}

// Load function loads, validates, and normalizes the configuration
// file and returns its settings as an instance of the Config struct.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}
	c := &Config{}
	if err = yaml.Unmarshal(data, c); err != nil {
		return nil, fmt.Errorf("unmarshalling yaml: %w", err)
	}
	if err = c.ValidateAndNormalize(); err != nil {
		return nil, fmt.Errorf("validating configs: %w", err)
	}
	return c, nil
}

// ValidateAndNormalize validates the configuration settings and
// returns an error if they were not acceptable. It can also modify
// settings in order to normalize them or replace some zero values with
// their expected default values (if any).
func (c *Config) ValidateAndNormalize() error {
	return nil
}
