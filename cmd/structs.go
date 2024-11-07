package main

import (
	"encoding/json"
	"time"
)

type HealthData struct {
	Data Data `json:"data"`
}

type Data struct {
	Metrics                []Metric      `json:"metrics"`
	ECG                    []interface{} `json:"ecg"`
	HeartRateNotifications []interface{} `json:"heartRateNotifications"`
	StateOfMind            []StateOfMind `json:"stateOfMind"`
	Symptoms               []interface{} `json:"symptoms"`
	Workouts               []Workout     `json:"workouts"`
}

type Metric struct {
	Name  string         `json:"name"`
	Units string         `json:"units"`
	Data  []MetricRecord `json:"data"`
}

type MetricRecord struct {
	Date   time.Time `json:"date"`
	Qty    float64   `json:"qty"`
	Source string    `json:"source"`
}

type StateOfMind struct {
	Associations          []interface{} `json:"associations"`
	End                   time.Time     `json:"end"`
	ID                    string        `json:"id"`
	Kind                  string        `json:"kind"`
	Labels                []interface{} `json:"labels"`
	Start                 time.Time     `json:"start"`
	Valence               float64       `json:"valence"`
	ValenceClassification string        `json:"valenceClassification"`
}

type Workout struct {
	ActiveEnergy       []EnergyRecord  `json:"activeEnergy"`
	ActiveEnergyBurned EnergyValue     `json:"activeEnergyBurned"`
	Duration           float64         `json:"duration"`
	End                time.Time       `json:"end"`
	HeartRateData      []HeartRateData `json:"heartRateData"`
	Humidity           ValueWithUnits  `json:"humidity"`
	ID                 string          `json:"id"`
	Intensity          ValueWithUnits  `json:"intensity"`
	Metadata           interface{}     `json:"metadata"`
	Name               string          `json:"name"`
	Start              time.Time       `json:"start"`
	StepCount          []StepRecord    `json:"stepCount"`
	Temperature        ValueWithUnits  `json:"temperature"`
}

type EnergyRecord struct {
	Date   time.Time `json:"date"`
	Qty    float64   `json:"qty"`
	Source string    `json:"source"`
	Units  string    `json:"units"`
}

type EnergyValue struct {
	Qty   float64 `json:"qty"`
	Units string  `json:"units"`
}

type HeartRateData struct {
	Avg    float64   `json:"Avg"`
	Max    float64   `json:"Max"`
	Min    float64   `json:"Min"`
	Date   time.Time `json:"date"`
	Source string    `json:"source"`
	Units  string    `json:"units"`
}

type ValueWithUnits struct {
	Qty   float64 `json:"qty"`
	Units string  `json:"units"`
}

type StepRecord struct {
	Date   time.Time `json:"date"`
	Qty    float64   `json:"qty"`
	Source string    `json:"source"`
	Units  string    `json:"units"`
}

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
