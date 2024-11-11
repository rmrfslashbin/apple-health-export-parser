package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// config holds the application configuration
type config struct {
	sourceFile string
	exportDir  string
	verbose    bool
	version    bool
}

// Validate returns error if config is invalid
func (c *config) Validate() error {
	if c.sourceFile == "" {
		return fmt.Errorf("source file is required")
	}
	return nil
}

// Add input validation for exportDir
func (c *config) validateExportDir() error {
	if c.exportDir == "" {
		return fmt.Errorf("export directory cannot be empty")
	}
	// Check if directory exists
	if _, err := os.Stat(c.exportDir); os.IsNotExist(err) {
		return fmt.Errorf("export directory %s does not exist", c.exportDir)
	}
	return nil
}

// parseFlags parses command-line flags and returns a config struct
func parseFlags() (*config, error) {
	cfg := &config{}

	// Define flags
	flag.StringVar(&cfg.sourceFile, "source", "", "Source JSON file to process (required)")
	flag.StringVar(&cfg.exportDir, "export", "exports", "Directory to export processed data")
	flag.BoolVar(&cfg.verbose, "verbose", false, "Enable verbose logging")
	flag.BoolVar(&cfg.version, "version", false, "Display version information")

	// Override usage message
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "\nUsage of %s:\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nApple Health Export Parser processes Apple Health JSON export files and organizes the data into separate JSON files by type.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
	}

	// Parse flags
	flag.Parse()

	return cfg, cfg.Validate()
}

// validateConfig validates the configuration and returns an error if invalid
func validateConfig(cfg *config) error {
	// Handle version flag
	if cfg.version {
		vInfo := GetVersion()
		fmt.Println(vInfo.String())
		os.Exit(0)
	}

	if cfg.sourceFile == "" {
		return fmt.Errorf("source file is required")
	}

	if _, err := os.Stat(cfg.sourceFile); os.IsNotExist(err) {
		return fmt.Errorf("source file '%s' does not exist", cfg.sourceFile)
	}

	return nil
}

// setupLogging configures the logging based on the verbose flag
func setupLogging(cfg *config) {
	var logLevel slog.Level
	if cfg.verbose {
		logLevel = slog.LevelDebug
	} else {
		logLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: logLevel,
	}

	handler := slog.NewTextHandler(os.Stderr, opts)
	logger := slog.New(handler)
	slog.SetDefault(logger)
}

func main() {
	// Parse command line flags
	cfg, err := parseFlags()
	if err != nil {
		slog.Error("Invalid configuration", "error", err)
		flag.Usage()
		os.Exit(1)
	}

	// Handle version flag
	if cfg.version {
		vInfo := GetVersion()
		fmt.Println(vInfo.String())
		os.Exit(0)
	}

	// Set up logging
	setupLogging(cfg)

	// Validate configuration
	if err := validateConfig(cfg); err != nil {
		slog.Error("Invalid configuration", "error", err)
		flag.Usage()
		os.Exit(1)
	}

	// Log startup information
	slog.Info("Starting Apple Health Export Parser",
		"version", GetVersion().ShortString())

	// Create the export directory if it doesn't exist
	if err := os.MkdirAll(cfg.exportDir, 0755); err != nil {
		slog.Error("Failed to create export directory",
			"dir", cfg.exportDir,
			"error", err)
		os.Exit(1)
	}

	// Process the source file
	if err := processHealthData(cfg); err != nil {
		slog.Error("Failed to process health data", "error", err)
		os.Exit(1)
	}

	slog.Info("Data export completed successfully")
}

// processHealthData reads and processes the Apple Health export file
func processHealthData(cfg *config) error {
	slog.Info("Processing health data",
		"source", cfg.sourceFile,
		"export_dir", cfg.exportDir)

	// Open and parse the source file
	file, err := os.Open(cfg.sourceFile)
	if err != nil {
		return fmt.Errorf("opening source file: %w", err)
	}
	defer file.Close()

	var healthData HealthData
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&healthData); err != nil {
		return fmt.Errorf("decoding JSON: %w", err)
	}

	// Process and export the data
	return exportData(healthData, cfg.exportDir)
}

func exportData(healthData HealthData, exportDir string) error {
	if err := exportMetrics(healthData.Data.Metrics, exportDir); err != nil {
		return fmt.Errorf("exporting metrics: %w", err)
	}

	if err := exportWorkouts(healthData.Data.Workouts, exportDir); err != nil {
		return fmt.Errorf("exporting workouts: %w", err)
	}

	if err := exportStateOfMind(healthData.Data.StateOfMind, exportDir); err != nil {
		return fmt.Errorf("exporting state of mind data: %w", err)
	}

	if err := exportOtherData(healthData.Data.ECG, "ecg", exportDir); err != nil {
		return fmt.Errorf("exporting ECG data: %w", err)
	}

	if err := exportOtherData(healthData.Data.HeartRateNotifications, "heart_rate_notifications", exportDir); err != nil {
		return fmt.Errorf("exporting heart rate notifications: %w", err)
	}

	if err := exportOtherData(healthData.Data.Symptoms, "symptoms", exportDir); err != nil {
		return fmt.Errorf("exporting symptoms data: %w", err)
	}

	return nil
}

func exportMetrics(metrics []Metric, exportDir string) error {
	metricsDir := filepath.Join(exportDir, "metrics")
	if err := os.MkdirAll(metricsDir, 0755); err != nil {
		return fmt.Errorf("creating metrics directory: %w", err)
	}

	for _, metric := range metrics {
		if len(metric.Data) > 0 {
			timestamp := metric.Data[0].Date.Format("2006-01-02_15-04-05")
			filename := filepath.Join(metricsDir, fmt.Sprintf("%s_%s.json", timestamp, sanitizeFilename(metric.Name)))

			if err := exportToJSON(metric, filename); err != nil {
				return fmt.Errorf("exporting metric %s: %w", metric.Name, err)
			}
		}
	}

	slog.Info("Exported metrics", "count", len(metrics))
	return nil
}

func exportWorkouts(workouts []Workout, exportDir string) error {
	workoutsDir := filepath.Join(exportDir, "workouts")
	if err := os.MkdirAll(workoutsDir, 0755); err != nil {
		return fmt.Errorf("creating workouts directory: %w", err)
	}

	for _, workout := range workouts {
		timestamp := workout.Start.Format("2006-01-02_15-04-05")
		filename := filepath.Join(workoutsDir, fmt.Sprintf("%s_%s.json", timestamp, sanitizeFilename(workout.Name)))

		if err := exportToJSON(workout, filename); err != nil {
			return fmt.Errorf("exporting workout %s: %w", workout.Name, err)
		}
	}

	slog.Info("Exported workouts", "count", len(workouts))
	return nil
}

func exportStateOfMind(stateOfMind []StateOfMind, exportDir string) error {
	somDir := filepath.Join(exportDir, "state_of_mind")
	if err := os.MkdirAll(somDir, 0755); err != nil {
		return fmt.Errorf("creating state of mind directory: %w", err)
	}

	for _, som := range stateOfMind {
		timestamp := som.Start.Format("2006-01-02_15-04-05")
		filename := filepath.Join(somDir, fmt.Sprintf("%s_%s.json", timestamp, sanitizeFilename(som.Kind)))

		if err := exportToJSON(som, filename); err != nil {
			return fmt.Errorf("exporting state of mind record: %w", err)
		}
	}

	slog.Info("Exported state of mind records", "count", len(stateOfMind))
	return nil
}

func exportOtherData(data []interface{}, name string, exportDir string) error {
	if len(data) == 0 {
		return nil
	}

	dataDir := filepath.Join(exportDir, name)
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("creating %s directory: %w", name, err)
	}

	for i, item := range data {
		timestamp := time.Now().Format("2006-01-02_15-04-05")
		filename := filepath.Join(dataDir, fmt.Sprintf("%s_%s_%03d.json", timestamp, name, i+1))

		if err := exportToJSON(item, filename); err != nil {
			return fmt.Errorf("exporting %s record %d: %w", name, i+1, err)
		}
	}

	slog.Info("Exported records", "type", name, "count", len(data))
	return nil
}

func exportToJSON(data interface{}, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("encoding JSON: %w", err)
	}

	slog.Debug("Exported file", "filename", filename)
	return nil
}

func sanitizeFilename(name string) string {
	// Replace spaces with underscores and remove any non-alphanumeric characters
	name = strings.ReplaceAll(name, " ", "_")
	return strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			return r
		}
		return -1
	}, name)
}
