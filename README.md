# Apple Health Export Parser

A command-line tool to process and organize Apple Health JSON export files from [HealthyApps.dev](https://www.healthyapps.dev/) into structured, categorized data files.

## Features

- Parse Apple Health JSON exports
- Organize data by type (metrics, workouts, state of mind, ECG, heart rate notifications, symptoms)
- Export individual records as separate JSON files with timestamps
- Configurable logging with multiple output formats
- Built-in validation and error handling
- Comprehensive trace ID support for debugging

## Author

Robert Sigler (code@sigler.io)

## License

MIT License - see [LICENSE](LICENSE) file for details

## Installation

### Prerequisites

- Go 1.25.4 or later

### Building from Source

```bash
# Clone the repository
git clone git@github.com:rmrfslashbin/apple-health-export-parser.git
cd apple-health-export-parser

# Build the binary
make build

# Or build for multiple platforms
make build-all
```

The binary will be created in the `bin/` directory.

## Usage

The tool uses subcommands for different operations, powered by Cobra and Viper for flexible configuration.

### Available Commands

```bash
# Display help
apple-health-export-parser --help

# Process a health export file
apple-health-export-parser process --source health-export.json

# Display version information
apple-health-export-parser version
```

### Global Flags

These flags apply to all commands:

```
--config string       Config file (default is $HOME/.apple-health-export-parser.yaml)
--log-level string    Log level: debug, info, warn, error (default "info")
--log-format string   Log format: json or text (default "text")
--log-output string   Log output: stderr, /path/to/file, or /path/to/dir/ (default "stderr")
```

### Process Command

Process an Apple Health export file and organize the data.

```bash
apple-health-export-parser process [flags]
```

**Flags:**
```
-s, --source string              Source JSON file to process (required)
-e, --export string              Directory to export processed data (default "exports")
-c, --collections strings        Target collections for MCP import (comma-separated)
    --batch-size-workouts int    Batch size for workout records (default 20)
    --batch-size-som int         Batch size for state of mind records (default 20)
    --batch-size-metrics int     Batch size for metric records (default 10)
```

**Examples:**

Process a health export file:
```bash
apple-health-export-parser process --source HealthAutoExport-2024-08-01.json
```

Process with MCP Memory import preparation:
```bash
apple-health-export-parser process \
  --source HealthAutoExport-2024-08-01.json \
  --collections spinal_fusion_recovery
```

Process with multiple collections and custom batch sizes:
```bash
apple-health-export-parser process \
  --source HealthAutoExport-2024-08-01.json \
  --collections health_tracking,recovery_log \
  --batch-size-workouts 30 \
  --batch-size-som 25 \
  --batch-size-metrics 15
```

Process with custom export directory:
```bash
apple-health-export-parser process \
  --source HealthAutoExport-2024-08-01.json \
  --export ./my-health-data
```

Process with debug logging to file:
```bash
apple-health-export-parser process \
  --source HealthAutoExport-2024-08-01.json \
  --log-level debug \
  --log-output ./logs/
```

Process with JSON logging:
```bash
apple-health-export-parser process \
  --source HealthAutoExport-2024-08-01.json \
  --log-format json \
  --log-output ./logs/
```

### Version Command

Display detailed version information:

```bash
apple-health-export-parser version
```

### Configuration File

You can create a configuration file to set default values. The tool looks for:
- `$HOME/.apple-health-export-parser.yaml`
- Or specify with `--config /path/to/config.yaml`

**Example config file:**

```yaml
log-level: debug
log-format: text
log-output: ./logs/
```

### Environment Variables

Configuration can also be set via environment variables with the `AHEP_` prefix:

```bash
export AHEP_LOG_LEVEL=debug
export AHEP_LOG_FORMAT=json
apple-health-export-parser process --source health-export.json
```

## Output Structure

The tool creates the following directory structure:

```
exports/
├── import/
│   ├── batch_1_workouts.json
│   ├── batch_2_workouts.json
│   ├── batch_1_state_of_mind.json
│   ├── batch_2_state_of_mind.json
│   ├── batch_1_metrics.json
│   ├── batch_summary.json
│   └── ...
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
│   └── ...
├── heart_rate_notifications/
│   └── ...
└── symptoms/
    └── ...
```

Each exported file contains the complete data for a single record, making it easy to analyze individual metrics, workouts, or health events.

### MCP Memory Import Batches

The `import/` directory contains batch files ready for import into the MCP Memory server:

- **Batch Files**: JSON arrays of memory objects formatted for direct MCP import
- **Collections Field**: Each memory includes the target collection(s) specified via `--collections` flag
- **Batch Summary**: `batch_summary.json` provides an overview of all generated batches

**Batch Summary Format:**
```json
{
  "total_records": 160,
  "workout_records": 49,
  "state_of_mind_records": 81,
  "metric_records": 30,
  "workout_batches": 3,
  "state_of_mind_batches": 5,
  "metric_batches": 3,
  "target_collections": ["spinal_fusion_recovery"],
  "timestamp": "2025-11-17T18:35:54Z"
}
```

**Memory Object Format:**
```json
{
  "type": "workout_log",
  "content": "# Outdoor Walk - November 17, 2025...",
  "metadata": {
    "workout_type": "Outdoor Walk",
    "date": "2025-11-17",
    "time": "14:30:00",
    "day_of_week": "Sunday",
    "time_of_day": "afternoon",
    "duration_minutes": 45,
    "data_source": "apple_health",
    "apple_health_id": "ABC123...",
    "review_status": "unreviewed",
    "privacy_level": "private"
  },
  "collections": ["spinal_fusion_recovery"]
}
```

## Development

### Running Tests

```bash
# Run tests with coverage
make test

# View coverage report
open coverage.html
```

### Running Linters

```bash
# Run all linters
make lint

# Run individual checks
make vet
make staticcheck

# Check for vulnerabilities
make vulncheck
```

### Building

```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Clean build artifacts
make clean
```

## Contributing

Contributions are welcome! Please follow these guidelines:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass (`make test`)
6. Run linters (`make lint`)
7. Commit your changes with descriptive messages
8. Push to your branch
9. Open a Pull Request

### Code Style

- Follow standard Go formatting (enforced by `gofmt` and `goimports`)
- Write godoc comments for all exported symbols
- Use table-driven tests
- Aim for 80% test coverage
- Handle errors explicitly

## Troubleshooting

### Common Issues

**"source file is required" error**
- Make sure to specify the `--source` flag with a valid JSON file path
- Use the `process` subcommand: `apple-health-export-parser process --source file.json`

**Permission denied when creating export directory**
- Ensure you have write permissions in the export directory
- Try specifying a different directory with `--export`

**JSON parsing errors**
- Verify your source file is valid JSON
- Ensure the file follows the HealthyApps.dev export format

### Debug Logging

Enable debug logging for detailed execution information:
```bash
apple-health-export-parser process \
  --source your-file.json \
  --log-level debug \
  --log-output ./logs/
```

## Roadmap

Future enhancements under consideration:

- [ ] Data analysis functions (average daily metrics, workout patterns)
- [ ] Correlation analysis between metrics
- [ ] Data visualization output
- [ ] Support for additional export formats (CSV, SQLite)
- [ ] Streaming processing for large files
- [ ] Web interface for interactive exploration

## Project History

- **v2025.11.17** - Project restructuring with enhanced logging and testing
- **Initial Release** - Basic parsing and export functionality

## Support

For bugs, feature requests, or questions:
- Open an issue on [GitHub](https://github.com/rmrfslashbin/apple-health-export-parser/issues)
- Contact: code@sigler.io
