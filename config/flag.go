package config

import "github.com/spf13/cobra"

func AddIndexerFlag(cmd *cobra.Command) {
	cmd.Flags().Bool(flagIndexerEnable, false, "enable indexer")
}
