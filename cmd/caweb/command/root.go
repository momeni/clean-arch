// Copyright (c) 2024 Behnam Momeni
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Package command provides the root and sub-commands for the clean-arch
// web project. Commands are organized using the cobra library.
// The root command starts the web server itself while the "db"
// sub-command can be used for the database migration actions.
// Three migration actions are supported. The init-dev and init-prod
// actions for initialization of the database with the development or
// production suitable data records and the migrate action for
// converting from one config and database version to another version.
//
//	./caweb [-c /path/of/main/config.yaml]           # start web server
//	./caweb db init-dev [-c /path/of/main/config.yaml]
//	./caweb db init-prod [-c /path/of/main/config.yaml]
//	./caweb db migrate
//	    /path/of/src/config.yaml
//	    /path/of/dst/config.yaml
//	    [-c /path/of/main/config.yaml]
package command

import (
	"context"
	"fmt"
	"os"

	"github.com/momeni/clean-arch/pkg/adapter/config"
	"github.com/momeni/clean-arch/pkg/adapter/restful/gin"
	"github.com/momeni/clean-arch/pkg/adapter/restful/gin/routes"
	"github.com/momeni/clean-arch/pkg/core/repo"
	"github.com/spf13/cobra"
)

var cfgPath string

var rootCmd = &cobra.Command{
	Use:   "caweb",
	Short: "A clean-architecture web project implementation pattern",
	Long: `A clean-architecture web project implementation pattern
which demonstrates how the core use cases and models layers can be
kept independent of the third-party dependent adapters layer while
interacting with them through a series of interfaces.
It also exemplifies usage of GORM and Pgx for database interactions,
the Gin Gonic web framework for the REST API implementation, how the
common requests/responses (de)serialization may be performed, how
each use case object may be instantiated, distinguishing between the
mandatory parameters and optional ones (with help of the functional
options), and how database repositories may be tested using temporary
PostgreSQL DBMS servers (created as podman containers).
It documents relevant core ideas in a README file.
It also provides a reification of a multi-database migration scheme
which can manage versioned database schema and config files.`,
	RunE: startWebServer,
}

func startWebServer(_ *cobra.Command, _ []string) error {
	ctx := context.Background()
	c, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("config.Load(%q): %w", cfgPath, err)
	}
	fmt.Printf("configs: %v\n", c)
	p, err := c.ConnectionPool(ctx, repo.NormalRole)
	if err != nil {
		return fmt.Errorf("creating DB pool: %w", err)
	}
	defer p.Close()
	var e *gin.Engine = c.Gin.NewEngine()
	if err = routes.Register(ctx, e, p, c); err != nil {
		return fmt.Errorf("registering routes: %w", err)
	}
	if err = e.Run(); err != nil {
		return fmt.Errorf("running Gin engine: %w", err)
	}
	return nil
}

// Execute runs the rootCmd which in turn parses CLI arguments and
// flags and runs the most specific cobra command. The exit code may
// be a boolean (zero for success and non-zero for failure) or may be
// chosen based on the error condition (if it is desired to report
// several error conditions in the CLI of this program).
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(fixConfigPath)
	rootCmd.PersistentFlags().StringVarP(
		&cfgPath, "config", "c", "", "config file path",
	)
}

// fixConfigPath ensures that cfgPath is set respectively by either the
// CLI args, the CONFIG_FILE environment variable, or its default value.
// By the way, default value is not necessarily a single path and may
// check several paths sequentially and take the highest priority one
// among the existing paths. For example, a user-specific path may take
// precedence over a file in /etc which is selected over a file in /usr.
func fixConfigPath() {
	if cfgPath != "" {
		return
	}
	var found bool
	if cfgPath, found = os.LookupEnv("CONFIG_FILE"); !found {
		// the default path should usually be in the /etc directory
		cfgPath = "configs/sample-config.yaml"
	}
}
