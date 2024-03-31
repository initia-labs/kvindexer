package config

import "github.com/spf13/cobra"

func AddIndexerFlag(cmd *cobra.Command) {
	cmd.Flags().Bool(flagIndexerEnable, false, "enable indexer")
	cmd.Flags().StringSlice(flagIndexerEnabledSubmodules, []string{}, "enabled submodules for handler")
	cmd.Flags().StringSlice(flagIndexerEnabledCronjobs, []string{}, "enabled cronjobs")
}
