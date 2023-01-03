package cmd

import (
	"context"

	"github.com/kobtea/gorgo/check"
	"github.com/kobtea/gorgo/config"
	"github.com/spf13/cobra"
)

// checkCmd represents the check command
var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Test policies",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		cfg, err := config.ParseFromFile(cfgFile)
		if err != nil {
			return err
		}
		return check.Check(ctx, cfg)
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)
}
