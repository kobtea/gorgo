package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// cleanCmd represents the clean command
var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Remove contents at working directory",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := os.RemoveAll(cfg.WorkingDir); err != nil {
			return err
		}
		if err := os.Mkdir(cfg.WorkingDir, 0755); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(cleanCmd)
}
