package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/xid"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	sourceFile string
	exportDir  string
)

// processCmd represents the process command
var processCmd = &cobra.Command{
	Use:   "process",
	Short: "Process Apple Health export file",
	Long: `Process an Apple Health JSON export file and organize the data into
structured, categorized JSON files.

The command exports data by type (metrics, workouts, state of mind, ECG,
heart rate notifications, symptoms) into separate timestamped JSON files.`,
	Example: `  # Process a health export file
  apple-health-export-parser process --source health-export.json

  # Process with custom export directory
  apple-health-export-parser process --source health-export.json --export ./output/

  # Process with debug logging to file
  apple-health-export-parser process --source health-export.json --log-level debug --log-output ./logs/`,
	RunE: runProcess,
}

func init() {
	rootCmd.AddCommand(processCmd)

	// Command-specific flags
	processCmd.Flags().StringVarP(&sourceFile, "source", "s", "", "source JSON file to process (required)")
	processCmd.Flags().StringVarP(&exportDir, "export", "e", "exports", "directory to export processed data")

	// Mark required flags
	processCmd.MarkFlagRequired("source")

	// Bind flags to viper
	viper.BindPFlag("source", processCmd.Flags().Lookup("source"))
	viper.BindPFlag("export", processCmd.Flags().Lookup("export"))
}

// runProcess executes the process command
func runProcess(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Get configuration from Viper
	source := viper.GetString("source")
	export := viper.GetString("export")

	// Validate source file exists
	if _, err := os.Stat(source); os.IsNotExist(err) {
		return fmt.Errorf("source file '%s' does not exist", source)
	}

	// Generate trace ID for this execution
	traceID := xid.New().String()

	// Add trace ID and version to logger
	logger := slog.Default().With(
		"trace_id", traceID,
		"version", GetVersion().ShortString(),
		"pid", os.Getpid(),
	)
	slog.SetDefault(logger)

	// Add trace ID to context
	ctx = context.WithValue(ctx, "trace_id", traceID)

	// Log startup information
	logger.Info("Starting Apple Health Export Parser",
		"source", source,
		"export_dir", export)

	// Create the export directory if it doesn't exist
	if err := os.MkdirAll(export, 0755); err != nil {
		return fmt.Errorf("failed to create export directory '%s': %w", export, err)
	}

	// Process the source file
	if err := processHealthData(ctx, source, export); err != nil {
		return fmt.Errorf("failed to process health data: %w", err)
	}

	logger.Info("Data export completed successfully")
	return nil
}

// processHealthData reads and processes the Apple Health export file
func processHealthData(ctx context.Context, source, export string) error {
	slog.Info("Processing health data")

	// Open and parse the source file
	file, err := os.Open(source)
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
	return exportData(healthData, export)
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
