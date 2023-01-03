package cmd

import (
	"github.com/kobtea/gorgo/log"
	"github.com/spf13/cobra"
)

var cfgFile string
var logLevel string

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
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "log level")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if err := log.InitLogger(&log.LoggerOption{Level: logLevel}); err != nil {
		panic(err)
	}
}
