// Copyright (c) 2024 Behnam Momeni
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package command

import (
	"context"
	"fmt"

	"github.com/momeni/clean-arch/pkg/adapter/config"
	"github.com/momeni/clean-arch/pkg/adapter/db/postgres/schemarp"
	"github.com/momeni/clean-arch/pkg/core/usecase/migrationuc"
	"github.com/spf13/cobra"
)

var initProdCmd = &cobra.Command{
	Use:   "init-prod",
	Short: "Initialize database contents with production suitable data",
	Long: `Initialize database contents with production suitable data
for the database schema version which is specified in the configuration
file. The database connection information are also read from the config
file. No changes will be made to the config file itself.
` + credsRenewalMessage + `

If database schema version X.Y.Z is asked in the config file, while the
latest known minor and patch versions in the X schema major version are
equal to Y' and Z' respectively, relevant tables of version X.Y'.Z' will
be created in the cawebX schema (without updating the config file).
The cawebX schema must be either non-existent or empty. Otherwise, it
will not be modified and an error will be reported.`,
	RunE: initProd,
	Args: cobra.NoArgs,
}

func initProd(_ *cobra.Command, _ []string) error {
	ctx := context.Background()
	mig, err := config.LoadMigrator(cfgPath)
	if err != nil {
		return fmt.Errorf("config.LoadMigrator(%q): %w", cfgPath, err)
	}
	err = mig.Load(ctx)
	if err != nil {
		return fmt.Errorf("mig.Load(): %w", err)
	}
	ss, err := mig.Settler(ctx)
	if err != nil {
		return fmt.Errorf("mig.Settler(): %w", err)
	}
	muc := migrationuc.NewInitDB(ss, schemarp.New())
	err = muc.InitProd(ctx)
	if err != nil {
		return fmt.Errorf("initializing DB with prod data: %w", err)
	}
	return nil
}

func init() {
	dbCmd.AddCommand(initProdCmd)
}
