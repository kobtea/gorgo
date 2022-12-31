package cmd

import (
	"context"

	"github.com/kobtea/gorgo/check"
	"github.com/spf13/cobra"
)

// checkCmd represents the check command
var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Test policies",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		check.Check(ctx, cfg)
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)
}
