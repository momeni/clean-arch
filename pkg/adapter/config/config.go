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

type Database struct {
	Host     string
	Port     int
	Name     string
	Role     string
	PassFile string `yaml:"pass-file"`
}

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

type Gin struct {
	Logger   bool
	Recovery bool
}

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

type Usecases struct {
	Cars Cars
}

type Cars struct {
	OldParkingDelay *time.Duration `yaml:"old-parking-method-delay"`
}

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
