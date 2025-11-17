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
	// Initialize manifest
	manifest := &ExportManifest{
		GeneratedAt: time.Now(),
		Version:     GetVersion().ShortString(),
		Metrics:     []string{},
		Workouts:    []string{},
		StateOfMind: []string{},
	}

	// Get trace ID from context if available
	if ctx := context.Background(); ctx.Value("trace_id") != nil {
		manifest.TraceID = ctx.Value("trace_id").(string)
	}

	if err := exportMetrics(healthData.Data.Metrics, exportDir, manifest); err != nil {
		return fmt.Errorf("exporting metrics: %w", err)
	}

	if err := exportWorkouts(healthData.Data.Workouts, exportDir, manifest); err != nil {
		return fmt.Errorf("exporting workouts: %w", err)
	}

	if err := exportStateOfMind(healthData.Data.StateOfMind, exportDir, manifest); err != nil {
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

	// Set summary counts
	manifest.Summary.TotalMetrics = len(healthData.Data.Metrics)
	manifest.Summary.TotalWorkouts = len(healthData.Data.Workouts)
	manifest.Summary.TotalStateOfMind = len(healthData.Data.StateOfMind)

	// Export manifest
	manifestFile := filepath.Join(exportDir, "manifest.json")
	if err := exportToJSON(manifest, manifestFile); err != nil {
		return fmt.Errorf("exporting manifest: %w", err)
	}

	slog.Info("Exported manifest", "file", manifestFile)
	return nil
}

func exportMetrics(metrics []Metric, exportDir string, manifest *ExportManifest) error {
	metricsDir := filepath.Join(exportDir, "metrics")
	if err := os.MkdirAll(metricsDir, 0755); err != nil {
		return fmt.Errorf("creating metrics directory: %w", err)
	}

	for _, metric := range metrics {
		if len(metric.Data) > 0 {
			timestamp := metric.Data[0].Date.Format("2006-01-02_15-04-05")
			relFilename := fmt.Sprintf("metrics/%s_%s.json", timestamp, sanitizeFilename(metric.Name))
			filename := filepath.Join(exportDir, relFilename)

			if err := exportToJSON(metric, filename); err != nil {
				return fmt.Errorf("exporting metric %s: %w", metric.Name, err)
			}
			manifest.Metrics = append(manifest.Metrics, relFilename)
		}
	}

	slog.Info("Exported metrics", "count", len(metrics))
	return nil
}

func exportWorkouts(workouts []Workout, exportDir string, manifest *ExportManifest) error {
	workoutsDir := filepath.Join(exportDir, "workouts")
	workoutDetailsDir := filepath.Join(exportDir, "workout_details")

	// Create base directories
	if err := os.MkdirAll(workoutsDir, 0755); err != nil {
		return fmt.Errorf("creating workouts directory: %w", err)
	}
	if err := os.MkdirAll(workoutDetailsDir, 0755); err != nil {
		return fmt.Errorf("creating workout_details directory: %w", err)
	}

	for _, workout := range workouts {
		timestamp := workout.Start.Format("2006-01-02_15-04-05")
		baseFilename := fmt.Sprintf("%s_%s", timestamp, sanitizeFilename(workout.Name))

		// Create workout summary
		summary := createWorkoutSummary(workout)
		relSummaryFilename := fmt.Sprintf("workouts/%s_summary.json", baseFilename)
		summaryFilename := filepath.Join(exportDir, relSummaryFilename)
		if err := exportToJSON(summary, summaryFilename); err != nil {
			return fmt.Errorf("exporting workout summary %s: %w", workout.Name, err)
		}
		manifest.Workouts = append(manifest.Workouts, relSummaryFilename)

		// Export detail files for time-series data
		detailsSubdir := filepath.Join(workoutDetailsDir, baseFilename)
		if err := os.MkdirAll(detailsSubdir, 0755); err != nil {
			return fmt.Errorf("creating workout details directory: %w", err)
		}

		// Export heart rate data if present
		if len(workout.HeartRateData) > 0 {
			relHrFile := fmt.Sprintf("workout_details/%s/heart_rate.json", baseFilename)
			hrFile := filepath.Join(exportDir, relHrFile)
			if err := exportToJSON(workout.HeartRateData, hrFile); err != nil {
				return fmt.Errorf("exporting heart rate data: %w", err)
			}
			manifest.WorkoutDetails.HeartRate = append(manifest.WorkoutDetails.HeartRate, relHrFile)
			slog.Debug("Exported heart rate data", "workout", workout.Name, "points", len(workout.HeartRateData))
		}

		// Export heart rate recovery data if present
		if len(workout.HeartRateRecovery) > 0 {
			relHrrFile := fmt.Sprintf("workout_details/%s/heart_rate_recovery.json", baseFilename)
			hrrFile := filepath.Join(exportDir, relHrrFile)
			if err := exportToJSON(workout.HeartRateRecovery, hrrFile); err != nil {
				return fmt.Errorf("exporting heart rate recovery data: %w", err)
			}
			manifest.WorkoutDetails.HeartRateRecovery = append(manifest.WorkoutDetails.HeartRateRecovery, relHrrFile)
			slog.Debug("Exported heart rate recovery data", "workout", workout.Name, "points", len(workout.HeartRateRecovery))
		}

		// Export active energy data if present
		if len(workout.ActiveEnergy) > 0 {
			relEnergyFile := fmt.Sprintf("workout_details/%s/active_energy.json", baseFilename)
			energyFile := filepath.Join(exportDir, relEnergyFile)
			if err := exportToJSON(workout.ActiveEnergy, energyFile); err != nil {
				return fmt.Errorf("exporting active energy data: %w", err)
			}
			manifest.WorkoutDetails.Energy = append(manifest.WorkoutDetails.Energy, relEnergyFile)
			slog.Debug("Exported active energy data", "workout", workout.Name, "points", len(workout.ActiveEnergy))
		}

		// Export step count data if present
		if len(workout.StepCount) > 0 {
			relStepsFile := fmt.Sprintf("workout_details/%s/step_count.json", baseFilename)
			stepsFile := filepath.Join(exportDir, relStepsFile)
			if err := exportToJSON(workout.StepCount, stepsFile); err != nil {
				return fmt.Errorf("exporting step count data: %w", err)
			}
			manifest.WorkoutDetails.Steps = append(manifest.WorkoutDetails.Steps, relStepsFile)
			slog.Debug("Exported step count data", "workout", workout.Name, "points", len(workout.StepCount))
		}
	}

	slog.Info("Exported workouts", "count", len(workouts))
	return nil
}

// createWorkoutSummary generates a summary view of a workout with aggregated statistics.
func createWorkoutSummary(w Workout) WorkoutSummary {
	summary := WorkoutSummary{
		ID:                w.ID,
		Name:              w.Name,
		Start:             w.Start,
		End:               w.End,
		Duration:          w.Duration,
		Temperature:       w.Temperature,
		Humidity:          w.Humidity,
		Intensity:         w.Intensity,
		TotalEnergyBurned: w.ActiveEnergyBurned,
		Metadata:          w.Metadata,

		ActiveEnergyCount:      len(w.ActiveEnergy),
		HeartRateDataCount:     len(w.HeartRateData),
		HeartRateRecoveryCount: len(w.HeartRateRecovery),
		StepCountDataCount:     len(w.StepCount),
	}

	// Calculate active energy statistics
	if len(w.ActiveEnergy) > 0 {
		summary.ActiveEnergyStats = calculateEnergyStats(w.ActiveEnergy)
	}

	// Calculate heart rate statistics
	if len(w.HeartRateData) > 0 {
		summary.HeartRateStats = calculateHeartRateStats(w.HeartRateData)
	}

	// Calculate heart rate recovery statistics
	if len(w.HeartRateRecovery) > 0 {
		summary.HeartRateRecoveryStats = calculateHeartRateStats(w.HeartRateRecovery)
	}

	// Calculate step count statistics
	if len(w.StepCount) > 0 {
		summary.StepCountStats = calculateStepStats(w.StepCount)
	}

	return summary
}

// calculateEnergyStats computes statistics for energy records.
func calculateEnergyStats(records []EnergyRecord) *Statistics {
	if len(records) == 0 {
		return nil
	}

	stats := &Statistics{
		Count: len(records),
		Min:   records[0].Qty,
		Max:   records[0].Qty,
		First: records[0].Qty,
		Last:  records[len(records)-1].Qty,
	}

	var sum float64
	for _, r := range records {
		sum += r.Qty
		if r.Qty < stats.Min {
			stats.Min = r.Qty
		}
		if r.Qty > stats.Max {
			stats.Max = r.Qty
		}
	}

	stats.Avg = sum / float64(len(records))
	stats.Total = sum

	return stats
}

// calculateHeartRateStats computes statistics for heart rate data.
func calculateHeartRateStats(records []HeartRateData) *Statistics {
	if len(records) == 0 {
		return nil
	}

	stats := &Statistics{
		Count: len(records),
		Min:   records[0].Avg,
		Max:   records[0].Avg,
		First: records[0].Avg,
		Last:  records[len(records)-1].Avg,
	}

	var sum float64
	for _, r := range records {
		sum += r.Avg
		if r.Avg < stats.Min {
			stats.Min = r.Avg
		}
		if r.Avg > stats.Max {
			stats.Max = r.Avg
		}
	}

	stats.Avg = sum / float64(len(records))

	return stats
}

// calculateStepStats computes statistics for step count data.
func calculateStepStats(records []StepRecord) *Statistics {
	if len(records) == 0 {
		return nil
	}

	stats := &Statistics{
		Count: len(records),
		Min:   records[0].Qty,
		Max:   records[0].Qty,
		First: records[0].Qty,
		Last:  records[len(records)-1].Qty,
	}

	var sum float64
	for _, r := range records {
		sum += r.Qty
		if r.Qty < stats.Min {
			stats.Min = r.Qty
		}
		if r.Qty > stats.Max {
			stats.Max = r.Qty
		}
	}

	stats.Avg = sum / float64(len(records))
	stats.Total = sum

	return stats
}

func exportStateOfMind(stateOfMind []StateOfMind, exportDir string, manifest *ExportManifest) error {
	somDir := filepath.Join(exportDir, "state_of_mind")
	if err := os.MkdirAll(somDir, 0755); err != nil {
		return fmt.Errorf("creating state of mind directory: %w", err)
	}

	for _, som := range stateOfMind {
		timestamp := som.Start.Format("2006-01-02_15-04-05")
		relFilename := fmt.Sprintf("state_of_mind/%s_%s.json", timestamp, sanitizeFilename(som.Kind))
		filename := filepath.Join(exportDir, relFilename)

		if err := exportToJSON(som, filename); err != nil {
			return fmt.Errorf("exporting state of mind record: %w", err)
		}
		manifest.StateOfMind = append(manifest.StateOfMind, relFilename)
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
