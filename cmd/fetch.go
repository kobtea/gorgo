package cmd

import (
	"context"

	"github.com/kobtea/gorgo/fetch"
	"github.com/spf13/cobra"
)

// fetchCmd represents the fetch command
var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Retrieve repository metadata",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		if err := fetch.Fetch(ctx, cfg); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(fetchCmd)
}
