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
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		return check.Check(ctx, cfg)
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)
}
