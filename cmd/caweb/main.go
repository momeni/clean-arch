// Package main is the main entry point of the clean-arch web project.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/momeni/clean-arch/pkg/adapter/config"
	"github.com/momeni/clean-arch/pkg/adapter/restful/gin/routes"
)

func main() {
	ctx := context.Background()
	cfgPath, found := os.LookupEnv("CONFIG_FILE")
	if !found {
		cfgPath = "configs/config.yaml"
	}
	c, err := config.Load(cfgPath)
	if err != nil {
		fmt.Printf("config.Load(%q): %v\n", cfgPath, err)
		os.Exit(-1)
	}
	fmt.Printf("configs: %v\n", c)
	p, err := c.Database.NewPool(ctx)
	if err != nil {
		fmt.Printf("Failed to connect to database: %v\n", err)
		os.Exit(-1)
	}
	defer p.Close()
	var e *gin.Engine = c.Gin.NewEngine()
	if err = routes.Register(e, p, c.Usecases); err != nil {
		fmt.Printf("Failed to register routes: %v\n", err)
		os.Exit(-1)
	}
	e.Run()
}
