package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	// Define CLI flags
	sourceFile := flag.String("source", "HealthAutoExport-2024-08-01-2024-09-07.json", "Source JSON file to process")
	exportDir := flag.String("export", "exports", "Directory to export processed data")

	// Parse the flags
	flag.Parse()

	// Validate the source file
	if _, err := os.Stat(*sourceFile); os.IsNotExist(err) {
		fmt.Printf("Error: Source file '%s' does not exist.\n", *sourceFile)
		flag.Usage()
		os.Exit(1)
	}

	// Create the export directory if it doesn't exist
	err := os.MkdirAll(*exportDir, os.ModePerm)
	if err != nil {
		fmt.Printf("Error creating export directory '%s': %v\n", *exportDir, err)
		os.Exit(1)
	}

	// Open and parse the source file
	file, err := os.Open(*sourceFile)
	if err != nil {
		fmt.Printf("Error opening file %s: %v\n", *sourceFile, err)
		os.Exit(1)
	}
	defer file.Close()

	var healthData HealthData
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&healthData)
	if err != nil {
		fmt.Printf("Error decoding JSON: %v\n", err)
		os.Exit(1)
	}

	// Process and export the data
	exportData(healthData, *exportDir)
}

func exportData(healthData HealthData, exportDir string) {
	exportMetrics(healthData.Data.Metrics, exportDir)
	exportWorkouts(healthData.Data.Workouts, exportDir)
	exportStateOfMind(healthData.Data.StateOfMind, exportDir)
	exportOtherData(healthData.Data.ECG, "ecg", exportDir)
	exportOtherData(healthData.Data.HeartRateNotifications, "heart_rate_notifications", exportDir)
	exportOtherData(healthData.Data.Symptoms, "symptoms", exportDir)

	fmt.Println("Data export completed.")
}

func exportMetrics(metrics []Metric, exportDir string) {
	metricsDir := filepath.Join(exportDir, "metrics")
	os.MkdirAll(metricsDir, os.ModePerm)

	for _, metric := range metrics {
		if len(metric.Data) > 0 {
			timestamp := metric.Data[0].Date.Format("2006-01-02_15-04-05")
			filename := filepath.Join(metricsDir, fmt.Sprintf("%s_%s.json", timestamp, sanitizeFilename(metric.Name)))
			exportToJSON(metric, filename)
		}
	}

	fmt.Printf("Exported %d metrics\n", len(metrics))
}

func exportWorkouts(workouts []Workout, exportDir string) {
	workoutsDir := filepath.Join(exportDir, "workouts")
	os.MkdirAll(workoutsDir, os.ModePerm)

	for _, workout := range workouts {
		timestamp := workout.Start.Format("2006-01-02_15-04-05")
		filename := filepath.Join(workoutsDir, fmt.Sprintf("%s_%s.json", timestamp, sanitizeFilename(workout.Name)))
		exportToJSON(workout, filename)
	}

	fmt.Printf("Exported %d workouts\n", len(workouts))
}

func exportStateOfMind(stateOfMind []StateOfMind, exportDir string) {
	somDir := filepath.Join(exportDir, "state_of_mind")
	os.MkdirAll(somDir, os.ModePerm)

	for _, som := range stateOfMind {
		timestamp := som.Start.Format("2006-01-02_15-04-05")
		filename := filepath.Join(somDir, fmt.Sprintf("%s_%s.json", timestamp, sanitizeFilename(som.Kind)))
		exportToJSON(som, filename)
	}

	fmt.Printf("Exported %d state of mind records\n", len(stateOfMind))
}

func exportOtherData(data []interface{}, name string, exportDir string) {
	if len(data) > 0 {
		dataDir := filepath.Join(exportDir, name)
		os.MkdirAll(dataDir, os.ModePerm)

		for i, item := range data {
			timestamp := time.Now().Format("2006-01-02_15-04-05")
			filename := filepath.Join(dataDir, fmt.Sprintf("%s_%s_%03d.json", timestamp, name, i+1))
			exportToJSON(item, filename)
		}
	}

	fmt.Printf("Exported %d %s records\n", len(data), name)
}

func exportToJSON(data interface{}, filename string) {
	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("Error creating file %s: %v\n", filename, err)
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(data)
	if err != nil {
		fmt.Printf("Error encoding JSON to file %s: %v\n", filename, err)
	}
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
