package routes

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/momeni/clean-arch/pkg/adapter/config"
	"github.com/momeni/clean-arch/pkg/adapter/db/postgres"
	"github.com/momeni/clean-arch/pkg/adapter/db/postgres/carsrp"
	"github.com/momeni/clean-arch/pkg/adapter/restful/gin/carsrs"
)

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
