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
		out, err := cmd.Flags().GetString("output")
		if err != nil {
			return err
		}
		return check.Check(ctx, cfg, out)
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)
	checkCmd.Flags().StringP("output", "o", "stdout", "output format for results (stdout, json, tap, table, junit, github)")
}
