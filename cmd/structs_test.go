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
			name:    "valid date with timezone",
			input:   "2025-11-17T14:30:00-05:00",
			want:    time.Date(2025, 11, 17, 14, 30, 0, 0, time.FixedZone("", -5*3600)),
			wantErr: false,
		},
		{
			name:    "valid UTC date",
			input:   "2025-11-17T14:30:00Z",
			want:    time.Date(2025, 11, 17, 14, 30, 0, 0, time.UTC),
			wantErr: false,
		},
		{
			name:    "invalid date format",
			input:   "invalid-date",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
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

func TestMetricRecordUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		check   func(*testing.T, *MetricRecord)
	}{
		{
			name: "valid metric with all fields",
			input: `{
				"qty": 75.0,
				"source": "Apple Watch",
				"date": "2025-11-17T14:30:00Z"
			}`,
			wantErr: false,
			check: func(t *testing.T, m *MetricRecord) {
				if m.Qty != 75.0 {
					t.Errorf("Qty = %v, want 75.0", m.Qty)
				}
				if m.Source != "Apple Watch" {
					t.Errorf("Source = %v, want Apple Watch", m.Source)
				}
			},
		},
		{
			name:    "invalid JSON",
			input:   `{invalid}`,
			wantErr: true,
		},
		{
			name: "missing date field",
			input: `{
				"qty": 75.0,
				"source": "Apple Watch"
			}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var m MetricRecord
			err := json.Unmarshal([]byte(tt.input), &m)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.check != nil {
				tt.check(t, &m)
			}
		})
	}
}

func TestStateOfMindUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		check   func(*testing.T, *StateOfMind)
	}{
		{
			name: "valid state of mind",
			input: `{
				"start": "2025-11-17T14:30:00Z",
				"end": "2025-11-17T14:30:00Z",
				"kind": "momentaryEmotion",
				"valence": 0.75,
				"valenceClassification": "pleasant",
				"labels": ["Happy", "Content"],
				"id": "test-123"
			}`,
			wantErr: false,
			check: func(t *testing.T, s *StateOfMind) {
				if s.Kind != "momentaryEmotion" {
					t.Errorf("Kind = %v, want momentaryEmotion", s.Kind)
				}
				if s.Valence != 0.75 {
					t.Errorf("Valence = %v, want 0.75", s.Valence)
				}
			},
		},
		{
			name:    "invalid JSON",
			input:   `{invalid}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var s StateOfMind
			err := json.Unmarshal([]byte(tt.input), &s)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.check != nil {
				tt.check(t, &s)
			}
		})
	}
}

func TestWorkoutUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		check   func(*testing.T, *Workout)
	}{
		{
			name: "valid workout",
			input: `{
				"name": "Outdoor Walk",
				"start": "2025-11-17T14:00:00Z",
				"end": "2025-11-17T15:00:00Z",
				"duration": 3600.0,
				"totalDistance": {"qty": "5.0", "units": "km"}
			}`,
			wantErr: false,
			check: func(t *testing.T, w *Workout) {
				if w.Name != "Outdoor Walk" {
					t.Errorf("Name = %v, want Outdoor Walk", w.Name)
				}
				if w.Duration != 3600.0 {
					t.Errorf("Duration = %v, want 3600.0", w.Duration)
				}
			},
		},
		{
			name:    "invalid JSON",
			input:   `{invalid}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var w Workout
			err := json.Unmarshal([]byte(tt.input), &w)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.check != nil {
				tt.check(t, &w)
			}
		})
	}
}

func TestEnergyRecordUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		check   func(*testing.T, *EnergyRecord)
	}{
		{
			name: "valid energy record",
			input: `{
				"date": "2025-11-17T14:30:00Z",
				"qty": 150.5,
				"units": "kcal"
			}`,
			wantErr: false,
			check: func(t *testing.T, e *EnergyRecord) {
				if e.Qty != 150.5 {
					t.Errorf("Qty = %v, want 150.5", e.Qty)
				}
				if e.Units != "kcal" {
					t.Errorf("Units = %v, want kcal", e.Units)
				}
			},
		},
		{
			name:    "invalid JSON",
			input:   `{invalid}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var e EnergyRecord
			err := json.Unmarshal([]byte(tt.input), &e)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.check != nil {
				tt.check(t, &e)
			}
		})
	}
}

func TestHeartRateDataUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		check   func(*testing.T, *HeartRateData)
	}{
		{
			name: "valid heart rate data",
			input: `{
				"date": "2025-11-17T14:30:00Z",
				"Avg": 75,
				"Min": 70,
				"Max": 80,
				"units": "count/min"
			}`,
			wantErr: false,
			check: func(t *testing.T, h *HeartRateData) {
				if h.Avg != 75 {
					t.Errorf("Avg = %v, want 75", h.Avg)
				}
				if h.Units != "count/min" {
					t.Errorf("Units = %v, want count/min", h.Units)
				}
			},
		},
		{
			name:    "invalid JSON",
			input:   `{invalid}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var h HeartRateData
			err := json.Unmarshal([]byte(tt.input), &h)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.check != nil {
				tt.check(t, &h)
			}
		})
	}
}

func TestStepRecordUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		check   func(*testing.T, *StepRecord)
	}{
		{
			name: "valid step record",
			input: `{
				"date": "2025-11-17T14:30:00Z",
				"qty": 1000,
				"units": "count"
			}`,
			wantErr: false,
			check: func(t *testing.T, s *StepRecord) {
				if s.Qty != 1000 {
					t.Errorf("Qty = %v, want 1000", s.Qty)
				}
				if s.Units != "count" {
					t.Errorf("Units = %v, want count", s.Units)
				}
			},
		},
		{
			name:    "invalid JSON",
			input:   `{invalid}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var s StepRecord
			err := json.Unmarshal([]byte(tt.input), &s)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.check != nil {
				tt.check(t, &s)
			}
		})
	}
}
