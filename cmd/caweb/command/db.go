// Copyright (c) 2024 Behnam Momeni
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package command

import "github.com/spf13/cobra"

var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "Database management actions",
	Long: `Database management actions can be chosen by sub-commands.
For fresh installation in a development or production environment,
the init-dev or init-prod may be used and for upgrade or downgrade
from an existing installation, the migrate may be used.`,
}

func init() {
	rootCmd.AddCommand(dbCmd)
}
