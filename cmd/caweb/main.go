package main

import (
	"context"
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/momeni/clean-arch/pkg/adapter/db/postgres"
	"github.com/momeni/clean-arch/pkg/adapter/restful/gin/routes"
)

func main() {
	ctx := context.Background()
	p, err := postgres.NewPool(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Printf("Failed to connect to database: %v", err)
		os.Exit(-1)
	}
	defer p.Close()
	var e *gin.Engine = gin.New()
	e.Use(gin.Logger(), gin.Recovery())
	routes.Register(e, p)
	e.Run()
}
