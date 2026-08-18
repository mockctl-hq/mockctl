package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// Build metadata populated by GoReleaser via -ldflags
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"

	cfgFile string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "mockctl",
	Short: "Mock:ctl - Offline API Simulation Engine",
	Long: `Mock:ctl is an advanced API simulation engine designed to give frontend developers 
a highly realistic, offline, and instant backend environment.

This CLI is the internal developer harness for testing the Mock:ctl Go Engine.`,
	// PersistentPreRunE is run for all subcommands
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		setupLogger()
		return nil
	},
}

// ExecuteContext runs the root command and passes the context to all subcommands.
func ExecuteContext(ctx context.Context) error {
	return rootCmd.ExecuteContext(ctx)
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.mockctl.yaml)")
	rootCmd.PersistentFlags().Bool("debug", false, "Enable debug structured logging")

	// Bind viper explicitly to ignore errors
	_ = viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory or current directory with name ".mockctl" (without extension).
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName(".mockctl")
	}

	viper.SetEnvPrefix("MOCKCTL")
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	_ = viper.ReadInConfig()
}

// setupLogger initializes the global structured logger using log/slog.
func setupLogger() {
	level := slog.LevelInfo
	if viper.GetBool("debug") {
		level = slog.LevelDebug
	}

	// Use JSON format for structured logging
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})

	logger := slog.New(handler)
	slog.SetDefault(logger)
}
