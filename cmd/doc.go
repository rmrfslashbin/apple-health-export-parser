// apple-health-export-parser processes Apple Health JSON export files from
// HealthyApps.dev and organizes the data into structured, categorized JSON files.
//
// The parser handles multiple types of health data including:
//   - Metrics (heart rate, steps, blood pressure, etc.)
//   - Workouts (exercise sessions with duration, energy, heart rate data)
//   - State of Mind (mental health and mood recordings)
//   - ECG (electrocardiogram data)
//   - Heart Rate Notifications (alerts and abnormal readings)
//   - Symptoms (logged health symptoms)
//
// Each data type is exported to separate JSON files organized by timestamp,
// making it easy to analyze individual records or track trends over time.
//
// Usage:
//
//	apple-health-export-parser -source health-export.json -export ./output/
//
// The tool supports configurable logging with multiple output formats and
// destinations, and includes comprehensive trace ID support for debugging.
package main
