// Package routes contains all resource packages and facilitates
// instantiation and registration of all repo, use case, and resource
// packages based on the user provided configuration settings.
package routes

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/momeni/clean-arch/pkg/adapter/config"
	"github.com/momeni/clean-arch/pkg/adapter/db/postgres"
	"github.com/momeni/clean-arch/pkg/adapter/db/postgres/carsrp"
	"github.com/momeni/clean-arch/pkg/adapter/restful/gin/carsrs"
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
func Register(e *gin.Engine, p *postgres.Pool, c config.Usecases) error {
	carsRepo := carsrp.New()

	carsUseCase, err := c.Cars.NewUseCase(p, carsRepo)
	if err != nil {
		return fmt.Errorf("creating cars use case: %w", err)
	}

	r := e.Group("/api/caweb/v1")
	carsrs.Register(r, carsUseCase)
	return nil
}
