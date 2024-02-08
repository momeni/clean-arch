// Copyright (c) 2024 Behnam Momeni
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package command

import (
	"context"
	"fmt"

	"github.com/momeni/clean-arch/pkg/adapter/config"
	"github.com/momeni/clean-arch/pkg/core/usecase/migrationuc"
	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate a source database contents to a new empty database",
	Long: `Migrate a source database contents to a new empty database
while changing the format of source schema contents in order to match
with the expected destination schema version.
For this purpose, three configuration file paths are required.
First, a source config file which indicates how the source database may
be located and which schema semantic version it contains.
Second, a destination config file which indicates how an empty database
may be located as the destination of this migration and which schema
version must be created there. The migration process is responsible to
create required tables there and transfer all data (with proper format
conversion) from the source database to the destination database.
The configuration file format is converted similarly (based on the
semantic version of the source and destination configuration files).
Third (optional) argument (which is passed by the -c flag) indicates
the path of the target configuration file. This is the path which will
be overwritten.

The semantic versions of the source and destination configuration files
and database schema indicate that an upgrade or downgrade is asked.
The target configuration file will be updated only at the end and by an
atomic move operation in order to facilitate recovery from an incomplete
migration attempt. The destination database must be empty, otherwise,
its contents will not be updated. Also, relevant roles will be created.
` + credsRenewalMessage + `

If the database schema version X.Y.Z is asked in the destination config
file, while the latest known minor and patch versions in the X schema
major version are equal to Y' and Z' respectively, relevant tables of
version X.Y'.Z' will be created in the cawebX schema in the destination
database (other temporary schema may be created and dropped during the
migration in the destination database; all accesses to the source
database are read-only).`,
	RunE: migrate,
	Args: cobra.ExactArgs(2),
}

func migrate(_ *cobra.Command, args []string) error {
	ctx := context.Background()
	srcCfgPath := args[0]    // read-only source configs and DB
	dstCfgPath := args[1]    // default configs and destination DB
	targetCfgPath := cfgPath // destination path to be overwritten
	mig, err := config.LoadSrcMigrator(srcCfgPath)
	if err != nil {
		return fmt.Errorf(
			"config.LoadSrcMigrator(%q): %w", srcCfgPath, err,
		)
	}
	dstSettings, err := loadConfigFile(ctx, dstCfgPath)
	if err != nil {
		return fmt.Errorf("loading %q config file: %w", dstCfgPath, err)
	}
	muc := migrationuc.NewMigrateDB(
		mig, dstSettings, targetCfgPath, loadConfigFile,
	)
	err = muc.Migrate(ctx)
	if err != nil {
		return fmt.Errorf("migrating DB: %w", err)
	}
	return nil
}

func loadConfigFile(
	ctx context.Context, path string,
) (migrationuc.Settings, error) {
	mig, err := config.LoadMigrator(path)
	if err != nil {
		return nil, fmt.Errorf("config.LoadMigrator: %w", err)
	}
	err = mig.Load(ctx)
	if err != nil {
		return nil, fmt.Errorf("mig.Load(): %w", err)
	}
	settings, err := mig.Settler(ctx)
	if err != nil {
		return nil, fmt.Errorf("mig.Settler(): %w", err)
	}
	return settings, nil
}

func init() {
	dbCmd.AddCommand(migrateCmd)
}
