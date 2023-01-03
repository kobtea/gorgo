package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/kobtea/gorgo/config"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var cfgFile string
var cfg *config.Config
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

	loggerConfig := zap.NewProductionConfig()
	lv, err := zap.ParseAtomicLevel(logLevel)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	loggerConfig.Level = lv
	loggerConfig.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339)
	logger, err := loggerConfig.Build()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer func() {
		if er := logger.Sync(); er != nil {
			// see: https://github.com/uber-go/zap/issues/880
		}
	}()
	zap.ReplaceGlobals(logger)
}
