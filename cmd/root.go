package cmd

import (
	"fmt"
	"os"

	"github.com/kobtea/gorgo/config"
	"github.com/spf13/cobra"
)

var cfgFile string
var cfg *config.Config

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gorgo",
	Short: "GitHub Organization Organizer",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "./gorgo.yaml", "config file")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	b, err := os.ReadFile(cfgFile)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	cfg, err = config.Parse(b)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
