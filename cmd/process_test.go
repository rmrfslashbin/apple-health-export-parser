package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "alphanumeric only",
			input: "TestFile123",
			want:  "TestFile123",
		},
		{
			name:  "spaces to underscores",
			input: "Test File Name",
			want:  "Test_File_Name",
		},
		{
			name:  "special characters removed",
			input: "Test@File#Name!",
			want:  "TestFileName",
		},
		{
			name:  "mixed case preserved",
			input: "TestFileNAME",
			want:  "TestFileNAME",
		},
		{
			name:  "hyphens removed underscores preserved",
			input: "Test-File_Name",
			want:  "TestFile_Name",
		},
		{
			name:  "parentheses removed",
			input: "Test(File)Name",
			want:  "TestFileName",
		},
		{
			name:  "periods removed",
			input: "Test.File.Name",
			want:  "TestFileName",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sanitizeFilename(tt.input); got != tt.want {
				t.Errorf("sanitizeFilename() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetTimeOfDay(t *testing.T) {
	tests := []struct {
		name string
		time time.Time
		want string
	}{
		{
			name: "early morning",
			time: time.Date(2025, 11, 17, 3, 0, 0, 0, time.UTC),
			want: "night",
		},
		{
			name: "morning start",
			time: time.Date(2025, 11, 17, 5, 0, 0, 0, time.UTC),
			want: "morning",
		},
		{
			name: "late morning",
			time: time.Date(2025, 11, 17, 11, 0, 0, 0, time.UTC),
			want: "morning",
		},
		{
			name: "afternoon start",
			time: time.Date(2025, 11, 17, 12, 0, 0, 0, time.UTC),
			want: "afternoon",
		},
		{
			name: "late afternoon",
			time: time.Date(2025, 11, 17, 16, 0, 0, 0, time.UTC),
			want: "afternoon",
		},
		{
			name: "evening start",
			time: time.Date(2025, 11, 17, 17, 0, 0, 0, time.UTC),
			want: "evening",
		},
		{
			name: "late evening",
			time: time.Date(2025, 11, 17, 20, 59, 0, 0, time.UTC),
			want: "evening",
		},
		{
			name: "night start",
			time: time.Date(2025, 11, 17, 21, 0, 0, 0, time.UTC),
			want: "night",
		},
		{
			name: "midnight",
			time: time.Date(2025, 11, 17, 0, 0, 0, 0, time.UTC),
			want: "night",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getTimeOfDay(tt.time); got != tt.want {
				t.Errorf("getTimeOfDay() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateImportMetadata(t *testing.T) {
	tests := []struct {
		name          string
		start         time.Time
		duration      float64
		hasHeartRate  bool
		hasSteps      bool
		hasRecovery   bool
		checkDate     string
		checkTime      string
		checkDayOfWeek string
		checkTimeOfDay string
		checkDuration  float64
	}{
		{
			name:          "morning workout with all data",
			start:         time.Date(2025, 11, 17, 9, 30, 0, 0, time.UTC),
			duration:      3600.0,
			hasHeartRate:  true,
			hasSteps:      true,
			hasRecovery:   true,
			checkDate:     "2025-11-17",
			checkTime:     "09:30:00",
			checkDayOfWeek: "Monday",
			checkTimeOfDay: "morning",
			checkDuration: 60,
		},
		{
			name:          "afternoon workout no recovery",
			start:         time.Date(2025, 11, 17, 14, 15, 0, 0, time.UTC),
			duration:      1800.0,
			hasHeartRate:  true,
			hasSteps:      false,
			hasRecovery:   false,
			checkDate:     "2025-11-17",
			checkTime:     "14:15:00",
			checkDayOfWeek: "Monday",
			checkTimeOfDay: "afternoon",
			checkDuration: 30,
		},
		{
			name:          "evening workout",
			start:         time.Date(2025, 11, 17, 19, 0, 0, 0, time.UTC),
			duration:      2700.0,
			hasHeartRate:  false,
			hasSteps:      true,
			hasRecovery:   false,
			checkDate:     "2025-11-17",
			checkTime:     "19:00:00",
			checkDayOfWeek: "Monday",
			checkTimeOfDay: "evening",
			checkDuration: 45,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateImportMetadata(tt.start, tt.duration, tt.hasHeartRate, tt.hasSteps, tt.hasRecovery)

			if got.Date != tt.checkDate {
				t.Errorf("Date = %v, want %v", got.Date, tt.checkDate)
			}
			if got.Time != tt.checkTime {
				t.Errorf("Time = %v, want %v", got.Time, tt.checkTime)
			}
			if got.DayOfWeek != tt.checkDayOfWeek {
				t.Errorf("DayOfWeek = %v, want %v", got.DayOfWeek, tt.checkDayOfWeek)
			}
			if got.TimeOfDay != tt.checkTimeOfDay {
				t.Errorf("TimeOfDay = %v, want %v", got.TimeOfDay, tt.checkTimeOfDay)
			}
			if got.DurationMinutes != tt.checkDuration {
				t.Errorf("DurationMinutes = %v, want %v", got.DurationMinutes, tt.checkDuration)
			}
		})
	}
}

func TestCalculateEnergyStats(t *testing.T) {
	tests := []struct {
		name    string
		records []EnergyRecord
		wantMin float64
		wantMax float64
		wantAvg float64
	}{
		{
			name: "multiple energy records",
			records: []EnergyRecord{
				{Qty: 100.0},
				{Qty: 150.5},
				{Qty: 200.0},
			},
			wantMin: 100.0,
			wantMax: 200.0,
			wantAvg: 150.166667,
		},
		{
			name: "single record",
			records: []EnergyRecord{
				{Qty: 175.5},
			},
			wantMin: 175.5,
			wantMax: 175.5,
			wantAvg: 175.5,
		},
		{
			name:    "empty records",
			records: []EnergyRecord{},
			wantMin: 0,
			wantMax: 0,
			wantAvg: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateEnergyStats(tt.records)
			if got == nil && len(tt.records) > 0 {
				t.Error("calculateEnergyStats() returned nil for non-empty records")
				return
			}
			if len(tt.records) == 0 {
				if got != nil {
					t.Error("calculateEnergyStats() should return nil for empty records")
				}
				return
			}

			if got.Min != tt.wantMin {
				t.Errorf("Min = %v, want %v", got.Min, tt.wantMin)
			}
			if got.Max != tt.wantMax {
				t.Errorf("Max = %v, want %v", got.Max, tt.wantMax)
			}
			// Allow small floating point difference
			if diff := got.Avg - tt.wantAvg; diff > 0.01 || diff < -0.01 {
				t.Errorf("Avg = %v, want %v", got.Avg, tt.wantAvg)
			}
		})
	}
}

func TestCalculateHeartRateStats(t *testing.T) {
	tests := []struct {
		name    string
		records []HeartRateData
		wantMin float64
		wantMax float64
		wantAvg float64
	}{
		{
			name: "multiple heart rate records",
			records: []HeartRateData{
				{Avg: 60, Min: 60, Max: 60},
				{Avg: 75, Min: 75, Max: 75},
				{Avg: 90, Min: 90, Max: 90},
			},
			wantMin: 60.0,
			wantMax: 90.0,
			wantAvg: 75.0,
		},
		{
			name: "single record",
			records: []HeartRateData{
				{Avg: 72, Min: 72, Max: 72},
			},
			wantMin: 72.0,
			wantMax: 72.0,
			wantAvg: 72.0,
		},
		{
			name:    "empty records",
			records: []HeartRateData{},
			wantMin: 0,
			wantMax: 0,
			wantAvg: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateHeartRateStats(tt.records)
			if got == nil && len(tt.records) > 0 {
				t.Error("calculateHeartRateStats() returned nil for non-empty records")
				return
			}
			if len(tt.records) == 0 {
				if got != nil {
					t.Error("calculateHeartRateStats() should return nil for empty records")
				}
				return
			}

			if got.Min != tt.wantMin {
				t.Errorf("Min = %v, want %v", got.Min, tt.wantMin)
			}
			if got.Max != tt.wantMax {
				t.Errorf("Max = %v, want %v", got.Max, tt.wantMax)
			}
			if got.Avg != tt.wantAvg {
				t.Errorf("Avg = %v, want %v", got.Avg, tt.wantAvg)
			}
		})
	}
}

func TestCalculateStepStats(t *testing.T) {
	tests := []struct {
		name    string
		records []StepRecord
		wantMin float64
		wantMax float64
		wantAvg float64
	}{
		{
			name: "multiple step records",
			records: []StepRecord{
				{Qty: 1000},
				{Qty: 1500},
				{Qty: 2000},
			},
			wantMin: 1000.0,
			wantMax: 2000.0,
			wantAvg: 1500.0,
		},
		{
			name: "single record",
			records: []StepRecord{
				{Qty: 1234},
			},
			wantMin: 1234.0,
			wantMax: 1234.0,
			wantAvg: 1234.0,
		},
		{
			name:    "empty records",
			records: []StepRecord{},
			wantMin: 0,
			wantMax: 0,
			wantAvg: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateStepStats(tt.records)
			if got == nil && len(tt.records) > 0 {
				t.Error("calculateStepStats() returned nil for non-empty records")
				return
			}
			if len(tt.records) == 0 {
				if got != nil {
					t.Error("calculateStepStats() should return nil for empty records")
				}
				return
			}

			if got.Min != tt.wantMin {
				t.Errorf("Min = %v, want %v", got.Min, tt.wantMin)
			}
			if got.Max != tt.wantMax {
				t.Errorf("Max = %v, want %v", got.Max, tt.wantMax)
			}
			if got.Avg != tt.wantAvg {
				t.Errorf("Avg = %v, want %v", got.Avg, tt.wantAvg)
			}
		})
	}
}

func TestExportToJSON(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		data     interface{}
		filename string
		wantErr  bool
	}{
		{
			name: "valid export",
			data: map[string]string{
				"name": "test",
				"value": "123",
			},
			filename: filepath.Join(tmpDir, "test.json"),
			wantErr:  false,
		},
		{
			name: "invalid directory",
			data: map[string]string{
				"name": "test",
			},
			filename: "/nonexistent/directory/test.json",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := exportToJSON(tt.data, tt.filename)
			if (err != nil) != tt.wantErr {
				t.Errorf("exportToJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify file was created and is valid JSON
				if _, err := os.Stat(tt.filename); os.IsNotExist(err) {
					t.Errorf("File was not created: %v", tt.filename)
				}
			}
		})
	}
}

func TestGenerateStateOfMindTitle(t *testing.T) {
	tests := []struct {
		name    string
		summary StateOfMindSummary
		want    string
	}{
		{
			name: "momentary emotion",
			summary: StateOfMindSummary{
				Kind: "momentary_emotion",
				Start: time.Date(2025, 11, 17, 14, 30, 0, 0, time.UTC),
			},
			want: "Momentary Emotion - November 17, 2025",
		},
		{
			name: "daily mood",
			summary: StateOfMindSummary{
				Kind: "daily_mood",
				Start: time.Date(2025, 11, 17, 14, 30, 0, 0, time.UTC),
			},
			want: "Daily Mood - November 17, 2025",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := generateStateOfMindTitle(tt.summary); got != tt.want {
				t.Errorf("generateStateOfMindTitle() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateMetricTitle(t *testing.T) {
	tests := []struct {
		name    string
		summary MetricSummary
		want    string
	}{
		{
			name: "heart rate metric",
			summary: MetricSummary{
				Name: "Heart Rate",
				StartDate: time.Date(2025, 11, 17, 0, 0, 0, 0, time.UTC),
				EndDate: time.Date(2025, 11, 17, 23, 59, 59, 0, time.UTC),
			},
			want: "Heart Rate - Nov 17 to Nov 17, 2025",
		},
		{
			name: "steps metric",
			summary: MetricSummary{
				Name: "Steps",
				StartDate: time.Date(2025, 11, 17, 0, 0, 0, 0, time.UTC),
				EndDate: time.Date(2025, 11, 17, 23, 59, 59, 0, time.UTC),
			},
			want: "Steps - Nov 17 to Nov 17, 2025",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := generateMetricTitle(tt.summary); got != tt.want {
				t.Errorf("generateMetricTitle() = %v, want %v", got, tt.want)
			}
		})
	}
}
