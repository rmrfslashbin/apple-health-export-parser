package main

import (
	"encoding/json"
	"testing"
	"time"
)

func TestParseDate(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    time.Time
		wantErr bool
	}{
		{
			name:    "RFC3339 format",
			input:   "2024-01-15T10:30:00Z",
			want:    time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			wantErr: false,
		},
		{
			name:    "Custom format with timezone",
			input:   "2024-01-15 10:30:00 +0000",
			want:    time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			wantErr: false,
		},
		{
			name:    "Invalid format",
			input:   "not-a-date",
			want:    time.Time{},
			wantErr: true,
		},
		{
			name:    "Empty string",
			input:   "",
			want:    time.Time{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseDate(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseDate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !got.Equal(tt.want) {
				t.Errorf("parseDate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMetricRecord_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		want    MetricRecord
		wantErr bool
	}{
		{
			name: "Valid metric record",
			json: `{"date":"2024-01-15T10:30:00Z","qty":72.5,"source":"Apple Watch"}`,
			want: MetricRecord{
				Date:   time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				Qty:    72.5,
				Source: "Apple Watch",
			},
			wantErr: false,
		},
		{
			name:    "Invalid JSON",
			json:    `{invalid}`,
			want:    MetricRecord{},
			wantErr: true,
		},
		{
			name:    "Invalid date format",
			json:    `{"date":"invalid-date","qty":72.5,"source":"Apple Watch"}`,
			want:    MetricRecord{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got MetricRecord
			err := json.Unmarshal([]byte(tt.json), &got)
			if (err != nil) != tt.wantErr {
				t.Errorf("MetricRecord.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if !got.Date.Equal(tt.want.Date) {
					t.Errorf("MetricRecord.Date = %v, want %v", got.Date, tt.want.Date)
				}
				if got.Qty != tt.want.Qty {
					t.Errorf("MetricRecord.Qty = %v, want %v", got.Qty, tt.want.Qty)
				}
				if got.Source != tt.want.Source {
					t.Errorf("MetricRecord.Source = %v, want %v", got.Source, tt.want.Source)
				}
			}
		})
	}
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "Simple name",
			input: "Heart Rate",
			want:  "Heart_Rate",
		},
		{
			name:  "Special characters",
			input: "Test/File:Name*With?Chars",
			want:  "TestFileNameWithChars",
		},
		{
			name:  "Multiple spaces",
			input: "Multiple   Spaces   Here",
			want:  "Multiple___Spaces___Here",
		},
		{
			name:  "Alphanumeric with underscores",
			input: "Valid_Name_123",
			want:  "Valid_Name_123",
		},
		{
			name:  "Empty string",
			input: "",
			want:  "",
		},
		{
			name:  "Only special characters",
			input: "!@#$%^&*()",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeFilename(tt.input)
			if got != tt.want {
				t.Errorf("sanitizeFilename() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  config
		wantErr bool
	}{
		{
			name: "Valid config",
			config: config{
				sourceFile: "/path/to/file.json",
				exportDir:  "exports",
			},
			wantErr: false,
		},
		{
			name: "Missing source file",
			config: config{
				sourceFile: "",
				exportDir:  "exports",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWorkout_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		wantErr bool
	}{
		{
			name: "Valid workout",
			json: `{
				"start": "2024-01-15T10:00:00Z",
				"end": "2024-01-15T11:00:00Z",
				"name": "Running",
				"id": "test-id",
				"duration": 3600,
				"activeEnergy": [],
				"activeEnergyBurned": {"qty": 300, "units": "kcal"},
				"heartRateData": [],
				"humidity": {"qty": 60, "units": "%"},
				"intensity": {"qty": 5, "units": "intensity"},
				"stepCount": [],
				"temperature": {"qty": 20, "units": "C"}
			}`,
			wantErr: false,
		},
		{
			name:    "Invalid JSON",
			json:    `{invalid}`,
			wantErr: true,
		},
		{
			name: "Invalid start date",
			json: `{
				"start": "invalid-date",
				"end": "2024-01-15T11:00:00Z",
				"name": "Running"
			}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got Workout
			err := json.Unmarshal([]byte(tt.json), &got)
			if (err != nil) != tt.wantErr {
				t.Errorf("Workout.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStateOfMind_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		wantErr bool
	}{
		{
			name: "Valid state of mind",
			json: `{
				"start": "2024-01-15T10:00:00Z",
				"end": "2024-01-15T10:05:00Z",
				"id": "test-id",
				"kind": "momentary",
				"valence": 0.5,
				"valenceClassification": "pleasant",
				"associations": [],
				"labels": []
			}`,
			wantErr: false,
		},
		{
			name:    "Invalid JSON",
			json:    `{invalid}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got StateOfMind
			err := json.Unmarshal([]byte(tt.json), &got)
			if (err != nil) != tt.wantErr {
				t.Errorf("StateOfMind.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEnergyRecord_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		wantErr bool
	}{
		{
			name:    "Valid energy record",
			json:    `{"date":"2024-01-15T10:30:00Z","qty":50.5,"source":"Apple Watch","units":"kcal"}`,
			wantErr: false,
		},
		{
			name:    "Invalid JSON",
			json:    `{invalid}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got EnergyRecord
			err := json.Unmarshal([]byte(tt.json), &got)
			if (err != nil) != tt.wantErr {
				t.Errorf("EnergyRecord.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHeartRateData_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		wantErr bool
	}{
		{
			name:    "Valid heart rate data",
			json:    `{"date":"2024-01-15T10:30:00Z","Avg":75,"Max":85,"Min":65,"source":"Apple Watch","units":"bpm"}`,
			wantErr: false,
		},
		{
			name:    "Invalid JSON",
			json:    `{invalid}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got HeartRateData
			err := json.Unmarshal([]byte(tt.json), &got)
			if (err != nil) != tt.wantErr {
				t.Errorf("HeartRateData.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStepRecord_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		wantErr bool
	}{
		{
			name:    "Valid step record",
			json:    `{"date":"2024-01-15T10:30:00Z","qty":100,"source":"iPhone","units":"count"}`,
			wantErr: false,
		},
		{
			name:    "Invalid JSON",
			json:    `{invalid}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got StepRecord
			err := json.Unmarshal([]byte(tt.json), &got)
			if (err != nil) != tt.wantErr {
				t.Errorf("StepRecord.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
