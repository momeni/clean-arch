package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/momeni/clean-arch/pkg/adapter/db/postgres"
	"github.com/momeni/clean-arch/pkg/adapter/db/postgres/carsrp"
	"github.com/momeni/clean-arch/pkg/adapter/restful/gin/carsrs"
	"github.com/momeni/clean-arch/pkg/core/usecase/carsuc"
)

func Register(e *gin.Engine, p *postgres.Pool) {
	carsRepo := carsrp.New()

	carsUseCase := carsuc.New(p, carsRepo)

	r := e.Group("/api/caweb/v1")
	carsrs.Register(r, carsUseCase)
}
