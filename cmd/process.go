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
	sourceFile         string
	exportDir          string
	targetCollections  []string
	batchSizeWorkouts  int
	batchSizeSOM       int
	batchSizeMetrics   int
	generateImportScript bool
	memoryBinaryPath   string
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

	// MCP import configuration
	processCmd.Flags().StringSliceVarP(&targetCollections, "collections", "c", []string{}, "target collections for MCP import (comma-separated)")
	processCmd.Flags().IntVar(&batchSizeWorkouts, "batch-size-workouts", 20, "batch size for workout records")
	processCmd.Flags().IntVar(&batchSizeSOM, "batch-size-som", 20, "batch size for state of mind records")
	processCmd.Flags().IntVar(&batchSizeMetrics, "batch-size-metrics", 10, "batch size for metric records")

	// Import script generation
	processCmd.Flags().BoolVar(&generateImportScript, "generate-import-script", false, "generate MCP Memory import script (import.sh)")
	processCmd.Flags().StringVar(&memoryBinaryPath, "memory-binary", "memory", "path to memory CLI binary (default: memory in PATH)")

	// Mark required flags
	processCmd.MarkFlagRequired("source")

	// Bind flags to viper
	viper.BindPFlag("source", processCmd.Flags().Lookup("source"))
	viper.BindPFlag("export", processCmd.Flags().Lookup("export"))
	viper.BindPFlag("collections", processCmd.Flags().Lookup("collections"))
	viper.BindPFlag("batch-size-workouts", processCmd.Flags().Lookup("batch-size-workouts"))
	viper.BindPFlag("batch-size-som", processCmd.Flags().Lookup("batch-size-som"))
	viper.BindPFlag("batch-size-metrics", processCmd.Flags().Lookup("batch-size-metrics"))
	viper.BindPFlag("generate-import-script", processCmd.Flags().Lookup("generate-import-script"))
	viper.BindPFlag("memory-binary", processCmd.Flags().Lookup("memory-binary"))
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
	type contextKey string
	const traceIDKey contextKey = "trace_id"
	ctx = context.WithValue(ctx, traceIDKey, traceID)

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

	// Generate import batches for MCP Memory server
	if err := generateImportBatches(healthData.Data, exportDir); err != nil {
		slog.Warn("Failed to generate import batches", "error", err)
		// Don't fail the entire export if batch generation fails
	}

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
		TotalDistance:     w.Distance,
		ElevationUp:       w.ElevationUp,
		HasLocation:       w.Location != nil,
		HasRoute:          w.Route != nil,
		Metadata:          w.Metadata,

		ActiveEnergyCount:      len(w.ActiveEnergy),
		HeartRateDataCount:     len(w.HeartRateData),
		HeartRateRecoveryCount: len(w.HeartRateRecovery),
		StepCountDataCount:     len(w.StepCount),
		DistanceDataCount:      len(w.WalkingAndRunningDistance),
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

	// Calculate distance statistics
	if len(w.WalkingAndRunningDistance) > 0 {
		summary.DistanceStats = calculateDistanceStats(w.WalkingAndRunningDistance)
	}

	// Generate import metadata
	summary.ImportMetadata = generateImportMetadata(
		w.Start,
		w.Duration,
		len(w.HeartRateData) > 0,
		len(w.StepCount) > 0,
		len(w.HeartRateRecovery) > 0,
	)

	// Generate memory content (must be done after all stats are calculated)
	summary.MemoryContent = MemoryContent{
		Title:    fmt.Sprintf("%s - %s", w.Name, w.Start.Format("January 2, 2006")),
		Summary:  generateWorkoutSummaryText(summary),
		Markdown: generateWorkoutMarkdown(summary),
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

// calculateDistanceStats computes statistics for distance data.
func calculateDistanceStats(records []DistanceRecord) *Statistics {
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

// generateImportMetadata creates import-ready metadata for a workout.
func generateImportMetadata(start time.Time, duration float64, hasHeartRate, hasSteps, hasRecovery bool) ImportMetadata {
	return ImportMetadata{
		Date:            start.Format("2006-01-02"),
		Time:            start.Format("15:04:05"),
		DayOfWeek:       start.Weekday().String(),
		TimeOfDay:       getTimeOfDay(start),
		DurationMinutes: duration / 60.0,
		HasHeartRate:    hasHeartRate,
		HasSteps:        hasSteps,
		HasRecovery:     hasRecovery,
	}
}

// getTimeOfDay returns the period of day for a given time.
func getTimeOfDay(t time.Time) string {
	hour := t.Hour()
	switch {
	case hour >= 5 && hour < 12:
		return "morning"
	case hour >= 12 && hour < 17:
		return "afternoon"
	case hour >= 17 && hour < 21:
		return "evening"
	default:
		return "night"
	}
}

// generateWorkoutMarkdown creates full markdown content for a workout.
func generateWorkoutMarkdown(summary WorkoutSummary) string {
	var md strings.Builder

	// Title
	md.WriteString(fmt.Sprintf("# %s - %s\n\n", summary.Name, summary.Start.Format("January 2, 2006")))

	// Duration and timing
	durationMin := summary.Duration / 60.0
	md.WriteString(fmt.Sprintf("**Duration:** %.1f minutes\n", durationMin))
	md.WriteString(fmt.Sprintf("**Start:** %s\n", summary.Start.Format("15:04:05")))
	md.WriteString(fmt.Sprintf("**End:** %s\n\n", summary.End.Format("15:04:05")))

	// Environmental conditions
	md.WriteString("## Environmental Conditions\n")
	md.WriteString(fmt.Sprintf("- Temperature: %.1f%s\n", summary.Temperature.Qty, summary.Temperature.Units))
	md.WriteString(fmt.Sprintf("- Humidity: %.0f%s\n", summary.Humidity.Qty, summary.Humidity.Units))
	md.WriteString("\n")

	// Performance summary
	md.WriteString("## Performance Summary\n")
	if summary.TotalDistance.Qty > 0 {
		md.WriteString(fmt.Sprintf("- Distance: %.2f %s\n", summary.TotalDistance.Qty, summary.TotalDistance.Units))
	}
	if summary.ElevationUp.Qty > 0 {
		md.WriteString(fmt.Sprintf("- Elevation Gain: %.1f %s\n", summary.ElevationUp.Qty, summary.ElevationUp.Units))
	}
	md.WriteString(fmt.Sprintf("- Total Energy: %.1f %s\n", summary.TotalEnergyBurned.Qty, summary.TotalEnergyBurned.Units))
	md.WriteString(fmt.Sprintf("- Intensity: %.2f %s\n", summary.Intensity.Qty, summary.Intensity.Units))
	md.WriteString("\n")

	// Heart rate
	if summary.HeartRateStats != nil {
		md.WriteString("## Heart Rate\n")
		md.WriteString(fmt.Sprintf("- Average: %.0f bpm\n", summary.HeartRateStats.Avg))
		md.WriteString(fmt.Sprintf("- Range: %.0f-%.0f bpm\n", summary.HeartRateStats.Min, summary.HeartRateStats.Max))
		md.WriteString(fmt.Sprintf("- Data Points: %d\n", summary.HeartRateStats.Count))
		md.WriteString("\n")
	}

	// Heart rate recovery
	if summary.HeartRateRecoveryStats != nil && summary.HeartRateRecoveryStats.Count > 0 {
		md.WriteString("## Heart Rate Recovery\n")
		md.WriteString(fmt.Sprintf("- Average: %.0f bpm\n", summary.HeartRateRecoveryStats.Avg))
		md.WriteString(fmt.Sprintf("- Range: %.0f-%.0f bpm\n", summary.HeartRateRecoveryStats.Min, summary.HeartRateRecoveryStats.Max))
		md.WriteString(fmt.Sprintf("- Data Points: %d\n", summary.HeartRateRecoveryStats.Count))
		md.WriteString("\n")
	}

	// Active energy
	if summary.ActiveEnergyStats != nil {
		md.WriteString("## Active Energy\n")
		md.WriteString(fmt.Sprintf("- Total: %.1f kcal\n", summary.ActiveEnergyStats.Total))
		md.WriteString(fmt.Sprintf("- Average: %.3f kcal/point\n", summary.ActiveEnergyStats.Avg))
		md.WriteString(fmt.Sprintf("- Data Points: %d\n", summary.ActiveEnergyStats.Count))
		md.WriteString("\n")
	}

	// Steps
	if summary.StepCountStats != nil && summary.StepCountStats.Count > 0 {
		md.WriteString("## Steps\n")
		md.WriteString(fmt.Sprintf("- Total: %.0f steps\n", summary.StepCountStats.Total))
		md.WriteString(fmt.Sprintf("- Average: %.2f steps/point\n", summary.StepCountStats.Avg))
		md.WriteString(fmt.Sprintf("- Data Points: %d\n", summary.StepCountStats.Count))
		md.WriteString("\n")
	}

	// Distance
	if summary.DistanceStats != nil && summary.DistanceStats.Count > 0 {
		md.WriteString("## Distance\n")
		md.WriteString(fmt.Sprintf("- Total: %.2f %s\n", summary.TotalDistance.Qty, summary.TotalDistance.Units))
		md.WriteString(fmt.Sprintf("- Average: %.4f %s/point\n", summary.DistanceStats.Avg, summary.TotalDistance.Units))
		md.WriteString(fmt.Sprintf("- Data Points: %d\n", summary.DistanceStats.Count))
		md.WriteString("\n")
	}

	// Footer
	md.WriteString("---\n")
	md.WriteString(fmt.Sprintf("*Source: Apple Health (ID: %s)*\n", summary.ID))

	return md.String()
}

// generateWorkoutSummaryText creates a one-line summary of a workout.
func generateWorkoutSummaryText(summary WorkoutSummary) string {
	durationMin := summary.Duration / 60.0
	parts := []string{
		fmt.Sprintf("%.1f minute %s", durationMin, strings.ToLower(summary.Name)),
	}

	if summary.TotalDistance.Qty > 0 {
		parts = append(parts, fmt.Sprintf("covering %.2f %s", summary.TotalDistance.Qty, summary.TotalDistance.Units))
	}

	if summary.HeartRateStats != nil {
		parts = append(parts, fmt.Sprintf("average heart rate of %.0f bpm", summary.HeartRateStats.Avg))
	}

	parts = append(parts, fmt.Sprintf("burning %.1f kcal", summary.TotalEnergyBurned.Qty))

	return strings.Join(parts, " with ")
}

// generateImportBatches creates MCP Memory import batch files from all health data types.
// This generates import-ready JSON files that can be directly used with the Memory MCP server,
// avoiding the need for bash script string manipulation that can introduce formatting issues.
func generateImportBatches(data Data, exportDir string) error {
	slog.Info("Generating MCP import batches",
		"workouts", len(data.Workouts),
		"state_of_mind", len(data.StateOfMind),
		"metrics", len(data.Metrics),
		"collections", targetCollections)

	// Validate collections
	if len(targetCollections) == 0 {
		slog.Warn("No target collections specified - memories will need collections added before import")
	}

	// Create import directory
	importDir := filepath.Join(exportDir, "import")
	if err := os.MkdirAll(importDir, 0755); err != nil {
		return fmt.Errorf("creating import directory: %w", err)
	}

	// Track batch statistics
	batchStats := BatchSummary{
		TotalRecords:      len(data.Workouts) + len(data.StateOfMind) + len(data.Metrics),
		WorkoutRecords:    len(data.Workouts),
		StateOfMindRecords: len(data.StateOfMind),
		MetricRecords:     len(data.Metrics),
		TargetCollections: targetCollections,
		Timestamp:         time.Now(),
	}

	// Generate workout batches
	if len(data.Workouts) > 0 {
		count, err := generateWorkoutBatches(data.Workouts, importDir)
		if err != nil {
			return fmt.Errorf("generating workout batches: %w", err)
		}
		batchStats.WorkoutBatches = count
	}

	// Generate state of mind batches
	if len(data.StateOfMind) > 0 {
		count, err := generateStateOfMindBatches(data.StateOfMind, importDir)
		if err != nil {
			return fmt.Errorf("generating state of mind batches: %w", err)
		}
		batchStats.StateOfMindBatches = count
	}

	// Generate metric batches
	if len(data.Metrics) > 0 {
		count, err := generateMetricBatches(data.Metrics, importDir)
		if err != nil {
			return fmt.Errorf("generating metric batches: %w", err)
		}
		batchStats.MetricBatches = count
	}

	// Generate summary report
	if err := generateBatchSummary(batchStats, importDir); err != nil {
		slog.Warn("Failed to generate batch summary", "error", err)
	}

	// Generate import script if requested
	if generateImportScript {
		if err := generateMCPImportScript(batchStats, importDir); err != nil {
			slog.Warn("Failed to generate import script", "error", err)
		} else {
			slog.Info("Generated import script", "path", filepath.Join(importDir, "import.sh"))
		}
	}

	slog.Info("Import batch generation complete",
		"total_batches", batchStats.WorkoutBatches+batchStats.StateOfMindBatches+batchStats.MetricBatches,
		"total_records", batchStats.TotalRecords)
	return nil
}

// generateWorkoutBatches creates batch files for workout data.
// Returns the number of batches created.
func generateWorkoutBatches(workouts []Workout, importDir string) (int, error) {
	// Convert workouts to summaries with import metadata
	summaries := make([]WorkoutSummary, 0, len(workouts))
	for _, workout := range workouts {
		summaries = append(summaries, createWorkoutSummary(workout))
	}

	// Use configured batch size
	batchNum := 1

	for i := 0; i < len(summaries); i += batchSizeWorkouts {
		end := i + batchSizeWorkouts
		if end > len(summaries) {
			end = len(summaries)
		}

		batch := summaries[i:end]
		memories := make([]Memory, 0, len(batch))

		for _, summary := range batch {
			metadata := map[string]interface{}{
				"workout_type":     summary.Name,
				"date":             summary.ImportMetadata.Date,
				"time":             summary.ImportMetadata.Time,
				"day_of_week":      summary.ImportMetadata.DayOfWeek,
				"time_of_day":      summary.ImportMetadata.TimeOfDay,
				"duration_minutes": summary.ImportMetadata.DurationMinutes,
				"data_source":      "apple_health",
				"apple_health_id":  summary.ID,
				"review_status":    "unreviewed",
				"privacy_level":    "private",
			}

			// Add distance data if available
			if summary.TotalDistance.Qty > 0 {
				metadata["distance"] = summary.TotalDistance.Qty
				metadata["distance_units"] = summary.TotalDistance.Units
			}
			if summary.ElevationUp.Qty > 0 {
				metadata["elevation_gain"] = summary.ElevationUp.Qty
				metadata["elevation_units"] = summary.ElevationUp.Units
			}
			if summary.HasLocation {
				metadata["has_location"] = true
			}
			if summary.HasRoute {
				metadata["has_route"] = true
			}

			memory := Memory{
				Type:        "workout_log",
				Content:     summary.MemoryContent.Markdown,
				Metadata:    metadata,
				Collections: targetCollections,
			}
			memories = append(memories, memory)
		}

		batchFilename := filepath.Join(importDir, fmt.Sprintf("batch_%d_workouts.json", batchNum))
		if err := exportToJSON(memories, batchFilename); err != nil {
			return 0, fmt.Errorf("exporting workout batch %d: %w", batchNum, err)
		}

		slog.Info("Generated workout batch",
			"batch", batchNum,
			"file", batchFilename,
			"count", len(memories))
		batchNum++
	}

	return batchNum - 1, nil
}

// generateStateOfMindBatches creates batch files for state of mind data.
// Returns the number of batches created.
func generateStateOfMindBatches(stateOfMind []StateOfMind, importDir string) (int, error) {
	// Convert to summaries with import metadata
	summaries := make([]StateOfMindSummary, 0, len(stateOfMind))
	for _, som := range stateOfMind {
		summaries = append(summaries, createStateOfMindSummary(som))
	}

	// Use configured batch size
	batchNum := 1

	for i := 0; i < len(summaries); i += batchSizeSOM {
		end := i + batchSizeSOM
		if end > len(summaries) {
			end = len(summaries)
		}

		batch := summaries[i:end]
		memories := make([]Memory, 0, len(batch))

		for _, summary := range batch {
			memory := Memory{
				Type:    "mental_health_log",
				Content: summary.MemoryContent.Markdown,
				Metadata: map[string]interface{}{
					"kind":                    summary.Kind,
					"valence":                 summary.Valence,
					"valence_classification":  summary.ValenceClassification,
					"date":                    summary.ImportMetadata.Date,
					"time":                    summary.ImportMetadata.Time,
					"day_of_week":             summary.ImportMetadata.DayOfWeek,
					"time_of_day":             summary.ImportMetadata.TimeOfDay,
					"data_source":             "apple_health",
					"apple_health_id":         summary.ID,
					"review_status":           "unreviewed",
					"privacy_level":           "private",
				},
				Collections: targetCollections,
			}
			memories = append(memories, memory)
		}

		batchFilename := filepath.Join(importDir, fmt.Sprintf("batch_%d_state_of_mind.json", batchNum))
		if err := exportToJSON(memories, batchFilename); err != nil {
			return 0, fmt.Errorf("exporting state of mind batch %d: %w", batchNum, err)
		}

		slog.Info("Generated state of mind batch",
			"batch", batchNum,
			"file", batchFilename,
			"count", len(memories))
		batchNum++
	}

	return batchNum - 1, nil
}

// generateMetricBatches creates batch files for metric data.
// Returns the number of batches created.
func generateMetricBatches(metrics []Metric, importDir string) (int, error) {
	// Convert to summaries with import metadata
	summaries := make([]MetricSummary, 0, len(metrics))
	for _, metric := range metrics {
		summary := createMetricSummary(metric)
		if summary.DataPoints > 0 { // Only include metrics with data
			summaries = append(summaries, summary)
		}
	}

	// Use configured batch size
	batchNum := 1

	for i := 0; i < len(summaries); i += batchSizeMetrics {
		end := i + batchSizeMetrics
		if end > len(summaries) {
			end = len(summaries)
		}

		batch := summaries[i:end]
		memories := make([]Memory, 0, len(batch))

		for _, summary := range batch {
			memory := Memory{
				Type:    "health_metric",
				Content: summary.MemoryContent.Markdown,
				Metadata: map[string]interface{}{
					"metric_name":      summary.Name,
					"units":            summary.Units,
					"data_points":      summary.DataPoints,
					"average":          summary.Average,
					"minimum":          summary.Min,
					"maximum":          summary.Max,
					"start_date":       summary.StartDate.Format("2006-01-02"),
					"end_date":         summary.EndDate.Format("2006-01-02"),
					"date":             summary.ImportMetadata.Date,
					"time":             summary.ImportMetadata.Time,
					"day_of_week":      summary.ImportMetadata.DayOfWeek,
					"data_source":      "apple_health",
					"review_status":    "unreviewed",
					"privacy_level":    "private",
				},
				Collections: targetCollections,
			}
			memories = append(memories, memory)
		}

		batchFilename := filepath.Join(importDir, fmt.Sprintf("batch_%d_metrics.json", batchNum))
		if err := exportToJSON(memories, batchFilename); err != nil {
			return 0, fmt.Errorf("exporting metric batch %d: %w", batchNum, err)
		}

		slog.Info("Generated metric batch",
			"batch", batchNum,
			"file", batchFilename,
			"count", len(memories))
		batchNum++
	}

	return batchNum - 1, nil
}

// createStateOfMindSummary generates a summary view of a state of mind record with import metadata.
func createStateOfMindSummary(som StateOfMind) StateOfMindSummary {
	summary := StateOfMindSummary{
		ID:                    som.ID,
		Kind:                  som.Kind,
		Start:                 som.Start,
		End:                   som.End,
		Valence:               som.Valence,
		ValenceClassification: som.ValenceClassification,
		Labels:                som.Labels,
		Associations:          som.Associations,
	}

	// Generate import metadata (no duration for state of mind entries)
	summary.ImportMetadata = ImportMetadata{
		Date:            som.Start.Format("2006-01-02"),
		Time:            som.Start.Format("15:04:05"),
		DayOfWeek:       som.Start.Weekday().String(),
		TimeOfDay:       getTimeOfDay(som.Start),
		DurationMinutes: 0,
		HasHeartRate:    false,
		HasSteps:        false,
		HasRecovery:     false,
	}

	// Generate memory content
	summary.MemoryContent = MemoryContent{
		Title:    generateStateOfMindTitle(summary),
		Summary:  generateStateOfMindSummaryText(summary),
		Markdown: generateStateOfMindMarkdown(summary),
	}

	return summary
}

// generateStateOfMindTitle creates a title for a state of mind record.
func generateStateOfMindTitle(summary StateOfMindSummary) string {
	kindLabel := strings.ReplaceAll(summary.Kind, "_", " ")
	// Simple title case - capitalize first letter of each word
	words := strings.Fields(kindLabel)
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + word[1:]
		}
	}
	kindLabel = strings.Join(words, " ")
	return fmt.Sprintf("%s - %s", kindLabel, summary.Start.Format("January 2, 2006"))
}

// generateStateOfMindSummaryText creates a one-line summary of a state of mind record.
func generateStateOfMindSummaryText(summary StateOfMindSummary) string {
	return fmt.Sprintf("%s mood recorded as %s (valence: %.2f)",
		strings.ReplaceAll(summary.Kind, "_", " "),
		summary.ValenceClassification,
		summary.Valence)
}

// generateStateOfMindMarkdown creates full markdown content for a state of mind record.
func generateStateOfMindMarkdown(summary StateOfMindSummary) string {
	var md strings.Builder

	// Title
	md.WriteString(fmt.Sprintf("# %s\n\n", generateStateOfMindTitle(summary)))

	// Time
	md.WriteString(fmt.Sprintf("**Time:** %s\n\n", summary.Start.Format("15:04:05")))

	// Classification
	md.WriteString("## Classification\n")
	md.WriteString(fmt.Sprintf("- **Valence:** %.3f\n", summary.Valence))
	md.WriteString(fmt.Sprintf("- **Classification:** %s\n\n", summary.ValenceClassification))

	// Labels (if any)
	if len(summary.Labels) > 0 {
		md.WriteString("## Labels\n")
		for _, label := range summary.Labels {
			md.WriteString(fmt.Sprintf("- %v\n", label))
		}
		md.WriteString("\n")
	}

	// Associations (if any)
	if len(summary.Associations) > 0 {
		md.WriteString("## Associations\n")
		for _, assoc := range summary.Associations {
			md.WriteString(fmt.Sprintf("- %v\n", assoc))
		}
		md.WriteString("\n")
	}

	// Footer
	md.WriteString("---\n")
	md.WriteString(fmt.Sprintf("*Source: Apple Health (ID: %s)*\n", summary.ID))

	return md.String()
}

// createMetricSummary generates a summary view of a metric with aggregated statistics.
func createMetricSummary(metric Metric) MetricSummary {
	if len(metric.Data) == 0 {
		return MetricSummary{
			Name:       metric.Name,
			Units:      metric.Units,
			DataPoints: 0,
		}
	}

	// Calculate statistics
	var min, max, sum float64
	min = metric.Data[0].Qty
	max = metric.Data[0].Qty

	for _, data := range metric.Data {
		if data.Qty < min {
			min = data.Qty
		}
		if data.Qty > max {
			max = data.Qty
		}
		sum += data.Qty
	}

	summary := MetricSummary{
		Name:       metric.Name,
		Units:      metric.Units,
		StartDate:  metric.Data[0].Date,
		EndDate:    metric.Data[len(metric.Data)-1].Date,
		DataPoints: len(metric.Data),
		Min:        min,
		Max:        max,
		Average:    sum / float64(len(metric.Data)),
	}

	// Generate import metadata using the date range
	days := summary.EndDate.Sub(summary.StartDate).Hours() / 24
	summary.ImportMetadata = ImportMetadata{
		Date:            summary.StartDate.Format("2006-01-02"),
		Time:            summary.StartDate.Format("15:04:05"),
		DayOfWeek:       summary.StartDate.Weekday().String(),
		TimeOfDay:       getTimeOfDay(summary.StartDate),
		DurationMinutes: days * 24 * 60, // Total timespan in minutes
		HasHeartRate:    false,
		HasSteps:        false,
		HasRecovery:     false,
	}

	// Generate memory content
	summary.MemoryContent = MemoryContent{
		Title:    generateMetricTitle(summary),
		Summary:  generateMetricSummaryText(summary),
		Markdown: generateMetricMarkdown(summary),
	}

	return summary
}

// generateMetricTitle creates a title for a metric.
func generateMetricTitle(summary MetricSummary) string {
	metricName := strings.ReplaceAll(summary.Name, "_", " ")
	// Simple title case - capitalize first letter of each word
	words := strings.Fields(metricName)
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + word[1:]
		}
	}
	metricName = strings.Join(words, " ")
	return fmt.Sprintf("%s - %s to %s",
		metricName,
		summary.StartDate.Format("Jan 2"),
		summary.EndDate.Format("Jan 2, 2006"))
}

// generateMetricSummaryText creates a one-line summary of a metric.
func generateMetricSummaryText(summary MetricSummary) string {
	return fmt.Sprintf("%s: %d data points, average %.2f %s (range: %.2f-%.2f)",
		strings.ReplaceAll(summary.Name, "_", " "),
		summary.DataPoints,
		summary.Average,
		summary.Units,
		summary.Min,
		summary.Max)
}

// generateMetricMarkdown creates full markdown content for a metric.
func generateMetricMarkdown(summary MetricSummary) string {
	var md strings.Builder

	// Title
	md.WriteString(fmt.Sprintf("# %s\n\n", generateMetricTitle(summary)))

	// Time Range
	md.WriteString("## Time Range\n")
	md.WriteString(fmt.Sprintf("- **Start:** %s\n", summary.StartDate.Format("January 2, 2006")))
	md.WriteString(fmt.Sprintf("- **End:** %s\n", summary.EndDate.Format("January 2, 2006")))
	days := summary.EndDate.Sub(summary.StartDate).Hours() / 24
	md.WriteString(fmt.Sprintf("- **Duration:** %.0f days\n\n", days))

	// Statistics
	md.WriteString("## Statistics\n")
	md.WriteString(fmt.Sprintf("- **Data Points:** %d\n", summary.DataPoints))
	md.WriteString(fmt.Sprintf("- **Average:** %.2f %s\n", summary.Average, summary.Units))
	md.WriteString(fmt.Sprintf("- **Minimum:** %.2f %s\n", summary.Min, summary.Units))
	md.WriteString(fmt.Sprintf("- **Maximum:** %.2f %s\n\n", summary.Max, summary.Units))

	// Footer
	md.WriteString("---\n")
	md.WriteString("*Source: Apple Health*\n")

	return md.String()
}

// generateBatchSummary creates a JSON summary file of the batch generation process.
func generateBatchSummary(summary BatchSummary, importDir string) error {
	summaryFile := filepath.Join(importDir, "batch_summary.json")

	data, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling batch summary: %w", err)
	}

	if err := os.WriteFile(summaryFile, data, 0644); err != nil {
		return fmt.Errorf("writing batch summary: %w", err)
	}

	slog.Info("Generated batch summary", "file", summaryFile)
	return nil
}

// generateMCPImportScript creates a shell script to import all batches using the Memory MCP CLI.
func generateMCPImportScript(summary BatchSummary, importDir string) error {
	var script strings.Builder

	// Script header
	script.WriteString("#!/bin/bash\n")
	script.WriteString("# Auto-generated MCP Memory import script\n")
	script.WriteString(fmt.Sprintf("# Generated: %s\n", summary.Timestamp.Format(time.RFC3339)))
	script.WriteString(fmt.Sprintf("# Total Records: %d\n", summary.TotalRecords))
	script.WriteString(fmt.Sprintf("# Target Collections: %s\n", strings.Join(summary.TargetCollections, ", ")))
	script.WriteString("#\n")
	script.WriteString("# Usage: ./import.sh\n")
	script.WriteString("#\n")
	script.WriteString("# This script uses the Memory MCP CLI to import Apple Health data.\n")
	script.WriteString("# Ensure the 'memory' binary is in your PATH or specify with MEMORY_BIN.\n")
	script.WriteString("#\n\n")

	// Configuration
	script.WriteString("set -euo pipefail  # Exit on error, undefined vars, pipe failures\n\n")
	script.WriteString("# Configuration\n")
	script.WriteString(fmt.Sprintf("MEMORY_BIN=\"%s\"\n", memoryBinaryPath))
	script.WriteString("SCRIPT_DIR=\"$(cd \"$(dirname \"${BASH_SOURCE[0]}\")\" && pwd)\"\n")
	script.WriteString("LOG_FILE=\"${SCRIPT_DIR}/import.log\"\n")
	script.WriteString("ERROR_LOG=\"${SCRIPT_DIR}/import_errors.log\"\n\n")

	// Helper functions
	script.WriteString("# Helper functions\n")
	script.WriteString("log() { echo \"[$(date +'%Y-%m-%d %H:%M:%S')] $*\" | tee -a \"${LOG_FILE}\"; }\n")
	script.WriteString("error() { echo \"[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $*\" | tee -a \"${LOG_FILE}\" \"${ERROR_LOG}\" >&2; }\n\n")

	// Verification
	script.WriteString("# Verify memory binary exists\n")
	script.WriteString("if ! command -v \"${MEMORY_BIN}\" &> /dev/null; then\n")
	script.WriteString("    error \"Memory binary '${MEMORY_BIN}' not found in PATH\"\n")
	script.WriteString("    error \"Install memory or set MEMORY_BIN environment variable\"\n")
	script.WriteString("    exit 1\n")
	script.WriteString("fi\n\n")

	// Start import
	script.WriteString("log \"Starting MCP Memory import\"\n")
	script.WriteString(fmt.Sprintf("log \"Total batches: %d\"\n", summary.WorkoutBatches+summary.StateOfMindBatches+summary.MetricBatches))
	script.WriteString(fmt.Sprintf("log \"Total records: %d\"\n\n", summary.TotalRecords))

	// Track statistics
	script.WriteString("# Import statistics\n")
	script.WriteString("TOTAL_IMPORTED=0\n")
	script.WriteString("TOTAL_FAILED=0\n\n")

	// Import function
	script.WriteString("# Import a batch file\n")
	script.WriteString("import_batch() {\n")
	script.WriteString("    local batch_file=\"$1\"\n")
	script.WriteString("    local batch_name=$(basename \"${batch_file}\")\n")
	script.WriteString("    \n")
	script.WriteString("    log \"Importing ${batch_name}...\"\n")
	script.WriteString("    \n")
	script.WriteString("    if \"${MEMORY_BIN}\" tools run --tool memory_memory_create --input \"${batch_file}\" >> \"${LOG_FILE}\" 2>> \"${ERROR_LOG}\"; then\n")
	script.WriteString("        log \"✓ Successfully imported ${batch_name}\"\n")
	script.WriteString("        ((TOTAL_IMPORTED++))\n")
	script.WriteString("        return 0\n")
	script.WriteString("    else\n")
	script.WriteString("        error \"✗ Failed to import ${batch_name}\"\n")
	script.WriteString("        ((TOTAL_FAILED++))\n")
	script.WriteString("        return 1\n")
	script.WriteString("    fi\n")
	script.WriteString("}\n\n")

	// Import workout batches
	if summary.WorkoutBatches > 0 {
		script.WriteString(fmt.Sprintf("# Import workout batches (%d batches, %d records)\n", summary.WorkoutBatches, summary.WorkoutRecords))
		script.WriteString("log \"Importing workout batches...\"\n")
		for i := 1; i <= summary.WorkoutBatches; i++ {
			script.WriteString(fmt.Sprintf("import_batch \"${SCRIPT_DIR}/batch_%d_workouts.json\"\n", i))
		}
		script.WriteString("\n")
	}

	// Import state of mind batches
	if summary.StateOfMindBatches > 0 {
		script.WriteString(fmt.Sprintf("# Import state of mind batches (%d batches, %d records)\n", summary.StateOfMindBatches, summary.StateOfMindRecords))
		script.WriteString("log \"Importing state of mind batches...\"\n")
		for i := 1; i <= summary.StateOfMindBatches; i++ {
			script.WriteString(fmt.Sprintf("import_batch \"${SCRIPT_DIR}/batch_%d_state_of_mind.json\"\n", i))
		}
		script.WriteString("\n")
	}

	// Import metric batches
	if summary.MetricBatches > 0 {
		script.WriteString(fmt.Sprintf("# Import metric batches (%d batches, %d records)\n", summary.MetricBatches, summary.MetricRecords))
		script.WriteString("log \"Importing metric batches...\"\n")
		for i := 1; i <= summary.MetricBatches; i++ {
			script.WriteString(fmt.Sprintf("import_batch \"${SCRIPT_DIR}/batch_%d_metrics.json\"\n", i))
		}
		script.WriteString("\n")
	}

	// Summary
	script.WriteString("# Import summary\n")
	script.WriteString("log \"Import complete\"\n")
	script.WriteString("log \"Successfully imported: ${TOTAL_IMPORTED} batches\"\n")
	script.WriteString("log \"Failed imports: ${TOTAL_FAILED} batches\"\n")
	script.WriteString("log \"Total records: " + fmt.Sprintf("%d", summary.TotalRecords) + "\"\n\n")

	script.WriteString("if [ ${TOTAL_FAILED} -gt 0 ]; then\n")
	script.WriteString("    error \"Some batches failed to import. Check ${ERROR_LOG} for details.\"\n")
	script.WriteString("    exit 1\n")
	script.WriteString("fi\n\n")

	script.WriteString("log \"All batches imported successfully!\"\n")
	script.WriteString("exit 0\n")

	// Write script to file
	scriptFile := filepath.Join(importDir, "import.sh")
	if err := os.WriteFile(scriptFile, []byte(script.String()), 0755); err != nil {
		return fmt.Errorf("writing import script: %w", err)
	}

	return nil
}
