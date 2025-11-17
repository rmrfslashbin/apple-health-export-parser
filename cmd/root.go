package main

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile   string
	logLevel  string
	logFormat string
	logOutput string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "apple-health-export-parser",
	Short: "Process Apple Health export files",
	Long: `Apple Health Export Parser processes JSON export files from HealthyApps.dev
and organizes the data into structured, categorized JSON files.

The parser handles multiple types of health data including metrics, workouts,
state of mind, ECG, heart rate notifications, and symptoms.`,
	PersistentPreRunE: setupLoggingCmd,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.apple-health-export-parser.yaml)")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().StringVar(&logFormat, "log-format", "text", "log format (json, text)")
	rootCmd.PersistentFlags().StringVar(&logOutput, "log-output", "stderr", "log output (stderr, /path/to/file, or /path/to/dir/)")

	// Bind flags to viper
	viper.BindPFlag("log-level", rootCmd.PersistentFlags().Lookup("log-level"))
	viper.BindPFlag("log-format", rootCmd.PersistentFlags().Lookup("log-format"))
	viper.BindPFlag("log-output", rootCmd.PersistentFlags().Lookup("log-output"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error finding home directory: %v\n", err)
			os.Exit(1)
		}

		// Search config in home directory with name ".apple-health-export-parser" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".apple-health-export-parser")
	}

	// Read in environment variables that match
	viper.SetEnvPrefix("AHEP") // Apple Health Export Parser
	viper.AutomaticEnv()       // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		slog.Debug("Using config file", "file", viper.ConfigFileUsed())
	}
}

// setupLoggingCmd configures logging based on global flags
func setupLoggingCmd(cmd *cobra.Command, args []string) error {
	logger, err := setupLoggingFromViper()
	if err != nil {
		return fmt.Errorf("failed to setup logger: %w", err)
	}
	slog.SetDefault(logger)
	return nil
}

// setupLoggingFromViper configures logging based on Viper configuration
func setupLoggingFromViper() (*slog.Logger, error) {
	// Get configuration from Viper
	level := viper.GetString("log-level")
	format := viper.GetString("log-format")
	output := viper.GetString("log-output")

	// Parse log level
	var slogLevel slog.Level
	switch strings.ToLower(level) {
	case "debug":
		slogLevel = slog.LevelDebug
	case "info":
		slogLevel = slog.LevelInfo
	case "warn", "warning":
		slogLevel = slog.LevelWarn
	case "error":
		slogLevel = slog.LevelError
	default:
		return nil, fmt.Errorf("invalid log level: %s (valid: debug, info, warn, error)", level)
	}

	// Setup output writer
	var writer io.Writer
	switch {
	case output == "" || output == "stderr":
		writer = os.Stderr
	case strings.HasSuffix(output, "/"):
		// Directory - create dated log file
		if err := os.MkdirAll(output, 0755); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %w", err)
		}
		logFile := filepath.Join(output,
			fmt.Sprintf("apple-health-export-parser-%s.log", time.Now().Format("2006-01-02")))
		f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}
		// Use MultiWriter to write to both stderr and file
		writer = io.MultiWriter(os.Stderr, f)
	default:
		// Specific file path
		dir := filepath.Dir(output)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %w", err)
		}
		f, err := os.OpenFile(output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}
		// Use MultiWriter to write to both stderr and file
		writer = io.MultiWriter(os.Stderr, f)
	}

	// Create handler based on format
	opts := &slog.HandlerOptions{Level: slogLevel}
	var handler slog.Handler

	if strings.ToLower(format) == "json" {
		handler = slog.NewJSONHandler(writer, opts)
	} else {
		handler = slog.NewTextHandler(writer, opts)
	}

	return slog.New(handler), nil
}
