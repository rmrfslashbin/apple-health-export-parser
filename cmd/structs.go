package main

import (
	"encoding/json"
	"time"
)

// HealthData represents the root structure of an Apple Health export JSON file.
// It contains all health data exported from the HealthyApps.dev service.
type HealthData struct {
	Data Data `json:"data"`
}

// Data contains the categorized health data collections.
// Each field represents a different type of health data that can be exported.
type Data struct {
	Metrics                []Metric      `json:"metrics"`                // Health metrics like heart rate, steps, etc.
	ECG                    []interface{} `json:"ecg"`                    // Electrocardiogram recordings
	HeartRateNotifications []interface{} `json:"heartRateNotifications"` // Heart rate alerts and notifications
	StateOfMind            []StateOfMind `json:"stateOfMind"`            // Mental health state recordings
	Symptoms               []interface{} `json:"symptoms"`               // Logged symptoms
	Workouts               []Workout     `json:"workouts"`               // Exercise and workout sessions
}

// Metric represents a health metric with multiple data points over time.
// Examples include heart rate, step count, blood pressure, etc.
type Metric struct {
	Name  string         `json:"name"`  // Metric name (e.g., "Heart Rate", "Steps")
	Units string         `json:"units"` // Unit of measurement (e.g., "bpm", "count")
	Data  []MetricRecord `json:"data"`  // Time-series data points
}

// MetricRecord represents a single data point for a health metric.
type MetricRecord struct {
	Date   time.Time `json:"date"`   // Timestamp of the measurement
	Qty    float64   `json:"qty"`    // Quantity/value of the measurement
	Source string    `json:"source"` // Source device or app that recorded this data
}

// StateOfMind represents a mental health or mood recording.
// It captures emotional state with valence (positive/negative) and associated labels.
type StateOfMind struct {
	Associations          []interface{} `json:"associations"`          // Related activities or contexts
	End                   time.Time     `json:"end"`                   // End time of the state recording
	ID                    string        `json:"id"`                    // Unique identifier
	Kind                  string        `json:"kind"`                  // Type of state (e.g., "momentary", "daily")
	Labels                []interface{} `json:"labels"`                // Descriptive labels for the state
	Start                 time.Time     `json:"start"`                 // Start time of the state recording
	Valence               float64       `json:"valence"`               // Numerical valence score
	ValenceClassification string        `json:"valenceClassification"` // Classification (e.g., "pleasant", "unpleasant")
}

// Workout represents a single exercise or workout session.
// It includes duration, energy burned, heart rate data, and environmental conditions.
type Workout struct {
	ActiveEnergy       []EnergyRecord  `json:"activeEnergy"`       // Energy burned over time during workout
	ActiveEnergyBurned EnergyValue     `json:"activeEnergyBurned"` // Total energy burned
	Duration           float64         `json:"duration"`           // Duration in seconds
	End                time.Time       `json:"end"`                // End time
	HeartRateData      []HeartRateData `json:"heartRateData"`      // Heart rate measurements during workout
	HeartRateRecovery  []HeartRateData `json:"heartRateRecovery"`  // Heart rate recovery measurements after workout
	Humidity           ValueWithUnits  `json:"humidity"`           // Humidity percentage
	ID                 string          `json:"id"`                 // Unique identifier
	Intensity          ValueWithUnits  `json:"intensity"`          // Workout intensity level
	Metadata           interface{}     `json:"metadata"`           // Additional metadata
	Name               string          `json:"name"`               // Workout name (e.g., "Running", "Cycling")
	Start              time.Time       `json:"start"`              // Start time
	StepCount          []StepRecord    `json:"stepCount"`          // Step count over time during workout
	Temperature        ValueWithUnits  `json:"temperature"`        // Temperature during workout
}

// EnergyRecord represents energy expenditure at a specific time.
type EnergyRecord struct {
	Date   time.Time `json:"date"`   // Timestamp of the measurement
	Qty    float64   `json:"qty"`    // Energy quantity
	Source string    `json:"source"` // Source device or app
	Units  string    `json:"units"`  // Energy units (e.g., "kcal")
}

// EnergyValue represents a total energy value with units.
type EnergyValue struct {
	Qty   float64 `json:"qty"`   // Energy quantity
	Units string  `json:"units"` // Energy units (e.g., "kcal")
}

// HeartRateData represents heart rate statistics for a time period.
type HeartRateData struct {
	Avg    float64   `json:"Avg"`    // Average heart rate
	Max    float64   `json:"Max"`    // Maximum heart rate
	Min    float64   `json:"Min"`    // Minimum heart rate
	Date   time.Time `json:"date"`   // Timestamp
	Source string    `json:"source"` // Source device or app
	Units  string    `json:"units"`  // Units (e.g., "bpm")
}

// ValueWithUnits represents a measurement value with its associated units.
type ValueWithUnits struct {
	Qty   float64 `json:"qty"`   // Measurement value
	Units string  `json:"units"` // Unit of measurement
}

// StepRecord represents step count at a specific time.
type StepRecord struct {
	Date   time.Time `json:"date"`   // Timestamp
	Qty    float64   `json:"qty"`    // Number of steps
	Source string    `json:"source"` // Source device or app
	Units  string    `json:"units"`  // Units (typically "count")
}

// Statistics represents aggregated statistical data for time-series measurements.
type Statistics struct {
	Count  int     `json:"count"`            // Number of data points
	Min    float64 `json:"min"`              // Minimum value
	Max    float64 `json:"max"`              // Maximum value
	Avg    float64 `json:"avg"`              // Average value
	Total  float64 `json:"total,omitempty"`  // Total/sum (for cumulative metrics)
	First  float64 `json:"first,omitempty"`  // First value in series
	Last   float64 `json:"last,omitempty"`   // Last value in series
}

// WorkoutSummary represents a condensed view of workout data for AI consumption.
// It includes metadata and aggregated statistics without large time-series arrays.
type WorkoutSummary struct {
	// Core metadata
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Start       time.Time `json:"start"`
	End         time.Time `json:"end"`
	Duration    float64   `json:"duration"` // Duration in seconds

	// Environmental conditions
	Temperature ValueWithUnits `json:"temperature"`
	Humidity    ValueWithUnits `json:"humidity"`
	Intensity   ValueWithUnits `json:"intensity"`

	// Aggregated statistics
	TotalEnergyBurned   EnergyValue `json:"totalEnergyBurned"`
	ActiveEnergyStats   *Statistics `json:"activeEnergyStats,omitempty"`
	HeartRateStats      *Statistics `json:"heartRateStats,omitempty"`
	HeartRateRecoveryStats *Statistics `json:"heartRateRecoveryStats,omitempty"`
	StepCountStats      *Statistics `json:"stepCountStats,omitempty"`

	// Data point counts (so AI knows what detail files exist)
	ActiveEnergyCount       int `json:"activeEnergyCount"`
	HeartRateDataCount      int `json:"heartRateDataCount"`
	HeartRateRecoveryCount  int `json:"heartRateRecoveryCount"`
	StepCountDataCount      int `json:"stepCountDataCount"`

	// Metadata
	Metadata interface{} `json:"metadata,omitempty"`

	// Import-ready metadata for MCP import
	ImportMetadata ImportMetadata `json:"importMetadata"`
	MemoryContent  MemoryContent  `json:"memoryContent"`
}

// ImportMetadata provides domain-agnostic metadata useful for any import process.
type ImportMetadata struct {
	Date            string  `json:"date"`            // "2025-11-12"
	Time            string  `json:"time"`            // "17:05:04"
	DayOfWeek       string  `json:"dayOfWeek"`       // "Tuesday"
	TimeOfDay       string  `json:"timeOfDay"`       // "morning/afternoon/evening/night"
	DurationMinutes float64 `json:"durationMinutes"` // 25.2
	HasHeartRate    bool    `json:"hasHeartRateData"`
	HasSteps        bool    `json:"hasStepData"`
	HasRecovery     bool    `json:"hasRecoveryData"`
}

// MemoryContent provides pre-formatted content ready for import into memory systems.
type MemoryContent struct {
	Title    string `json:"title"`    // "Outdoor Walk - November 12, 2025"
	Summary  string `json:"summary"`  // One-line summary
	Markdown string `json:"markdown"` // Full formatted markdown content
}

// StateOfMindSummary represents a condensed view of mental health data for AI consumption.
type StateOfMindSummary struct {
	ID                    string    `json:"id"`
	Kind                  string    `json:"kind"` // "daily_mood" or "momentary_emotion"
	Start                 time.Time `json:"start"`
	End                   time.Time `json:"end"`
	Valence               float64   `json:"valence"`
	ValenceClassification string    `json:"valenceClassification"` // "pleasant", "unpleasant", etc.
	Labels                []interface{} `json:"labels,omitempty"`
	Associations          []interface{} `json:"associations,omitempty"`

	// Import-ready metadata for MCP import
	ImportMetadata ImportMetadata `json:"importMetadata"`
	MemoryContent  MemoryContent  `json:"memoryContent"`
}

// MetricSummary represents a condensed view of metric time-series data for AI consumption.
type MetricSummary struct {
	Name  string `json:"name"`  // Metric name (e.g., "Heart Rate", "Steps")
	Units string `json:"units"` // Unit of measurement (e.g., "bpm", "count")

	// Time range
	StartDate     time.Time `json:"startDate"`
	EndDate       time.Time `json:"endDate"`
	DataPoints    int       `json:"dataPoints"`

	// Statistics
	Min           float64   `json:"min,omitempty"`
	Max           float64   `json:"max,omitempty"`
	Average       float64   `json:"average,omitempty"`

	// Import-ready metadata for MCP import
	ImportMetadata ImportMetadata `json:"importMetadata"`
	MemoryContent  MemoryContent  `json:"memoryContent"`
}

// ExportManifest provides an index of all exported files for AI navigation.
type ExportManifest struct {
	GeneratedAt time.Time `json:"generatedAt"`
	TraceID     string    `json:"traceId"`
	SourceFile  string    `json:"sourceFile"`
	Version     string    `json:"version"`

	// Date range analysis
	DateRange struct {
		Earliest  time.Time `json:"earliest"`
		Latest    time.Time `json:"latest"`
		TotalDays int       `json:"totalDays"`
	} `json:"dateRange"`

	Summary struct {
		TotalMetrics      int `json:"totalMetrics"`
		TotalWorkouts     int `json:"totalWorkouts"`
		TotalStateOfMind  int `json:"totalStateOfMind"`
	} `json:"summary"`

	Metrics      []string `json:"metrics"`      // List of metric files
	Workouts     []string `json:"workouts"`     // List of workout summary files
	StateOfMind  []string `json:"stateOfMind"`  // List of state of mind files

	// Detail file directories
	WorkoutDetails struct {
		HeartRate       []string `json:"heartRate,omitempty"`
		HeartRateRecovery []string `json:"heartRateRecovery,omitempty"`
		Energy          []string `json:"energy,omitempty"`
		Steps           []string `json:"steps,omitempty"`
	} `json:"workoutDetails"`

	// Import hints for MCP clients
	ImportHints struct {
		RecommendedMemoryTypes struct {
			Workouts     string `json:"workouts"`     // "workout_log"
			Metrics      string `json:"metrics"`      // "health_metric"
			StateOfMind  string `json:"stateOfMind"`  // "mental_health_log"
		} `json:"recommendedMemoryTypes"`

		BatchRecommendations struct {
			Workouts struct {
				TotalItems         int      `json:"totalItems"`
				SuggestedBatchSize int      `json:"suggestedBatchSize"`
				EstimatedBatches   int      `json:"estimatedBatches"`
				GroupingOptions    []string `json:"groupingOptions"`
			} `json:"workouts"`
			Metrics struct {
				TotalTypes         int      `json:"totalTypes"`
				SuggestedBatchSize int      `json:"suggestedBatchSize"`
				GroupingOptions    []string `json:"groupingOptions"`
			} `json:"metrics"`
			StateOfMind struct {
				TotalItems         int      `json:"totalItems"`
				SuggestedBatchSize int      `json:"suggestedBatchSize"`
				EstimatedBatches   int      `json:"estimatedBatches"`
			} `json:"stateOfMind"`
		} `json:"batchRecommendations"`

		DataQuality struct {
			WorkoutsWithHeartRate int `json:"workoutsWithHeartRate"`
			WorkoutsWithSteps     int `json:"workoutsWithSteps"`
			WorkoutsWithRecovery  int `json:"workoutsWithRecovery"`
		} `json:"dataQuality"`

		ContextWindowEstimates struct {
			WorkoutSummaryAvgChars int `json:"workoutSummaryAvgChars"`
			MetricFileAvgChars     int `json:"metricFileAvgChars"`
			StateOfMindAvgChars    int `json:"stateOfMindAvgChars"`
			SafeBatchSizeChars     int `json:"safeBatchSizeChars"` // ~75k recommended for Claude
		} `json:"contextWindowEstimates"`
	} `json:"importHints"`
}

// ImportBatch represents a pre-grouped set of memories ready for MCP import.
type ImportBatch struct {
	Batch           int      `json:"batch"`
	Description     string   `json:"description"`
	Count           int      `json:"count"`
	EstimatedChars  int      `json:"estimatedChars"`
	TargetCollection string  `json:"targetCollection,omitempty"` // Optional, for specific collections
	Memories        []Memory `json:"memories"`
}

// Memory represents a single memory ready for MCP import.
type Memory struct {
	Type        string                 `json:"type"`
	Content     string                 `json:"content"`
	Metadata    map[string]interface{} `json:"metadata"`
	Collections []string               `json:"collections"` // Target collections for MCP Memory server
}

// BatchSummary provides an overview of generated import batches.
type BatchSummary struct {
	TotalRecords        int       `json:"total_records"`
	WorkoutRecords      int       `json:"workout_records"`
	StateOfMindRecords  int       `json:"state_of_mind_records"`
	MetricRecords       int       `json:"metric_records"`
	WorkoutBatches      int       `json:"workout_batches"`
	StateOfMindBatches  int       `json:"state_of_mind_batches"`
	MetricBatches       int       `json:"metric_batches"`
	TargetCollections   []string  `json:"target_collections"`
	Timestamp           time.Time `json:"timestamp"`
}

// parseDate attempts to parse a date string using multiple common formats.
// It tries each format in order until one succeeds or all fail.
// Supported formats:
//   - "2006-01-02 15:04:05 -0700" (custom format with timezone)
//   - RFC3339 (standard ISO 8601 format)
func parseDate(dateStr string) (time.Time, error) {
	formats := []string{
		"2006-01-02 15:04:05 -0700",
		time.RFC3339,
	}

	var t time.Time
	var err error
	for _, format := range formats {
		t, err = time.Parse(format, dateStr)
		if err == nil {
			return t, nil
		}
	}
	return t, err
}

// UnmarshalJSON implements custom JSON unmarshaling for MetricRecord.
// It handles date parsing from string format to time.Time.
func (m *MetricRecord) UnmarshalJSON(data []byte) error {
	type Alias MetricRecord
	aux := &struct {
		Date string `json:"date"`
		*Alias
	}{
		Alias: (*Alias)(m),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	var err error
	m.Date, err = parseDate(aux.Date)
	if err != nil {
		return err
	}
	return nil
}

// UnmarshalJSON implements custom JSON unmarshaling for StateOfMind.
// It handles date parsing for both Start and End fields from string format to time.Time.
func (s *StateOfMind) UnmarshalJSON(data []byte) error {
	type Alias StateOfMind
	aux := &struct {
		Start string `json:"start"`
		End   string `json:"end"`
		*Alias
	}{
		Alias: (*Alias)(s),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	var err error
	s.Start, err = parseDate(aux.Start)
	if err != nil {
		return err
	}
	s.End, err = parseDate(aux.End)
	if err != nil {
		return err
	}
	return nil
}

// UnmarshalJSON implements custom JSON unmarshaling for Workout.
// It handles date parsing for both Start and End fields from string format to time.Time.
func (w *Workout) UnmarshalJSON(data []byte) error {
	type Alias Workout
	aux := &struct {
		Start string `json:"start"`
		End   string `json:"end"`
		*Alias
	}{
		Alias: (*Alias)(w),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	var err error
	w.Start, err = parseDate(aux.Start)
	if err != nil {
		return err
	}
	w.End, err = parseDate(aux.End)
	if err != nil {
		return err
	}
	return nil
}

// UnmarshalJSON implements custom JSON unmarshaling for EnergyRecord.
// It handles date parsing from string format to time.Time.
func (e *EnergyRecord) UnmarshalJSON(data []byte) error {
	type Alias EnergyRecord
	aux := &struct {
		Date string `json:"date"`
		*Alias
	}{
		Alias: (*Alias)(e),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	var err error
	e.Date, err = parseDate(aux.Date)
	if err != nil {
		return err
	}
	return nil
}

// UnmarshalJSON implements custom JSON unmarshaling for HeartRateData.
// It handles date parsing from string format to time.Time.
func (h *HeartRateData) UnmarshalJSON(data []byte) error {
	type Alias HeartRateData
	aux := &struct {
		Date string `json:"date"`
		*Alias
	}{
		Alias: (*Alias)(h),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	var err error
	h.Date, err = parseDate(aux.Date)
	if err != nil {
		return err
	}
	return nil
}

// UnmarshalJSON implements custom JSON unmarshaling for StepRecord.
// It handles date parsing from string format to time.Time.
func (s *StepRecord) UnmarshalJSON(data []byte) error {
	type Alias StepRecord
	aux := &struct {
		Date string `json:"date"`
		*Alias
	}{
		Alias: (*Alias)(s),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	var err error
	s.Date, err = parseDate(aux.Date)
	if err != nil {
		return err
	}
	return nil
}
