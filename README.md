# Health Data Processing Project

This project processes and exports health data from a JSON file, specifically designed for the file "HealthAutoExport-2024-08-01-2024-09-07.json".

## Project Structure

- `main.go`: Contains the main processing and export logic
- `structs.go`: Defines the data structures for parsing the JSON file
- `exports/`: Directory where exported files are saved

## Setup Instructions

1. Ensure you have Go installed on your system.
2. Place the "HealthAutoExport-2024-08-01-2024-09-07.json" file in the same directory as the Go files.
3. Run the program using the command: `go run main.go structs.go`

## Current Functionality

The program currently does the following:

1. Parses the JSON health data file
2. Exports individual records for:
   - Metrics
   - Workouts
   - State of Mind
   - ECG (if present)
   - Heart Rate Notifications (if present)
   - Symptoms (if present)
3. Saves exported data as JSON files in the `exports/` directory, using timestamps in filenames

## Exported Data Structure

```
exports/
├── metrics/
│   ├── YYYY-MM-DD_HH-MM-SS_metric_name.json
│   └── ...
├── workouts/
│   ├── YYYY-MM-DD_HH-MM-SS_workout_name.json
│   └── ...
├── state_of_mind/
│   ├── YYYY-MM-DD_HH-MM-SS_state_of_mind_type.json
│   └── ...
├── ecg/
├── heart_rate_notifications/
└── symptoms/
```

## Next Steps

To continue development in a new session:

1. Review the existing code in `main.go` and `structs.go`
2. Consider adding data analysis functions, such as:
   - Calculating average daily metrics
   - Analyzing workout patterns
   - Examining relationships between workouts and state of mind
   - Creating a timeline of mood changes
   - Identifying correlations between different health metrics
3. Implement data visualization features
4. Add error handling and logging for better debugging
5. Create unit tests for the processing and export functions

## Dependencies

- Go standard library (no external dependencies at this time)

## Notes for Future Sessions

- The current implementation assumes specific date formats. If date formats in the input file change, update the `parseDate` function in `structs.go`
- The program currently loads the entire JSON file into memory. For very large files, consider implementing streaming or chunked processing
- If additional data types are added to the health export, update the `HealthData` struct in `structs.go` and add corresponding export functions in `main.go`