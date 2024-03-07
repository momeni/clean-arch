// Copyright (c) 2023-2024 Behnam Momeni
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Package routes contains all resource packages and facilitates
// instantiation and registration of all repo, use case, and resource
// packages based on the user provided configuration settings.
package routes

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/momeni/clean-arch/pkg/adapter/config/cfg2"
	"github.com/momeni/clean-arch/pkg/adapter/db/postgres/carsrp"
	"github.com/momeni/clean-arch/pkg/adapter/db/postgres/settingsrp"
	"github.com/momeni/clean-arch/pkg/adapter/restful/gin/carsrs"
	"github.com/momeni/clean-arch/pkg/adapter/restful/gin/settingsrs"
	"github.com/momeni/clean-arch/pkg/core/repo"
)

// Register instantiates relevant repositories and use cases based on
// the c configuration settings. The p connections pool is passed to
// the use case instances, so they may acquire/release connections
// and transactions on demand. These connections/transactions will be
// passed to the repositories later in order to run relevant queries on
// them and accomplish those use cases. Each use case package is named
// like carsuc and each repository package is named like carsrp.
// Register instantiates a series of "resource" structs, from packages
// which are named like carsrs, in order to adapt the use cases
// interfaces with the REST APIs. These resources are registered as
// request handlers using the e gin-gonic engine instance.
// Possible errors will be returned after possible wrapping.
// Actual instantiation of use case objects are delegated to the
// c Config instance and the appuc use case.
func Register(
	ctx context.Context, e *gin.Engine, p repo.Pool, c *cfg2.Config,
) error {
	settingsRepo := settingsrp.New(c)
	carsRepo := carsrp.New()

	appUseCase, err := c.NewAppUseCase(p, settingsRepo, carsRepo)
	if err != nil {
		return fmt.Errorf("creating application use case: %w", err)
	}
	err = appUseCase.Reload(ctx)
	if err != nil {
		return fmt.Errorf("reloading use cases based on DB: %w", err)
	}
	r := e.Group("/api/caweb/v1")
	settingsrs.Register(r, appUseCase)
	carsrs.Register(r, appUseCase.CarsUseCase)
	return nil
}
