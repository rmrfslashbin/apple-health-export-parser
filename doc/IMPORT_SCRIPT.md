# MCP Memory Import Script Generation

This document describes the import script generation feature that automates the process of importing Apple Health data batches into the MCP Memory server using the Memory CLI.

## Overview

When processing Apple Health exports with the `--generate-import-script` flag, the tool generates an executable bash script (`import.sh`) alongside the batch JSON files. This script uses the Memory MCP CLI to automate the import process.

## Benefits

- **Zero Manual Work**: No need to manually construct `memory tools run` commands
- **Error Handling**: Built-in error logging and validation
- **Progress Tracking**: Detailed logging of import operations
- **Robust Execution**: Uses bash best practices (`set -euo pipefail`)
- **Automatic Batch Discovery**: Script knows about all generated batches

## Usage

### Complete End-to-End Workflow

```bash
# Step 1: Process Apple Health export and generate import script
apple-health-export-parser process \
  --source HealthAutoExport-2025-11-17.json \
  --collections spinal_fusion_recovery \
  --generate-import-script \
  --export exports

# Step 2: Navigate to import directory
cd exports/2025-11-17/import

# Step 3: Review what will be imported
cat batch_summary.json

# Example output:
# {
#   "total_records": 31,
#   "workout_records": 1,
#   "state_of_mind_records": 4,
#   "metric_records": 26,
#   "workout_batches": 1,
#   "state_of_mind_batches": 1,
#   "metric_batches": 3,
#   "target_collections": ["spinal_fusion_recovery"],
#   "timestamp": "2025-11-19T19:54:36.617668-05:00"
# }

# Step 4: Run the import script
./import.sh

# Step 5: Check the logs
tail -20 import.log

# Step 6: Verify import success
memory tools run --tool memory_collection_stats --input '{"id": "spinal_fusion_recovery"}'
```

### Basic Usage

```bash
# Generate batches with import script
apple-health-export-parser process \
  --source HealthAutoExport-2025-11-17.json \
  --collections spinal_fusion_recovery \
  --generate-import-script

# Navigate to import directory
cd exports/2025-11-17/import

# Run the import script
./import.sh
```

### Custom Memory Binary Path

If the Memory CLI is not in your PATH:

```bash
apple-health-export-parser process \
  --source HealthAutoExport-2025-11-17.json \
  --collections spinal_fusion_recovery \
  --generate-import-script \
  --memory-binary /usr/local/bin/memory
```

Or set the `MEMORY_BIN` environment variable before running the script:

```bash
export MEMORY_BIN=/usr/local/bin/memory
./import.sh
```

## Generated Files

When the script runs, it creates:

- `import.log` - Complete log of all import operations with timestamps
- `import_errors.log` - Error-only log for troubleshooting failed imports

## Script Features

### 1. Pre-flight Validation

The script verifies the Memory CLI binary exists before attempting any imports:

```bash
if ! command -v "${MEMORY_BIN}" &> /dev/null; then
    error "Memory binary '${MEMORY_BIN}' not found in PATH"
    error "Install memory or set MEMORY_BIN environment variable"
    exit 1
fi
```

### 2. Batch Import Function

Each batch is imported using the `memory tools run` command:

```bash
import_batch() {
    local batch_file="$1"
    local batch_name=$(basename "${batch_file}")

    log "Importing ${batch_name}..."

    if "${MEMORY_BIN}" tools run --tool memory_memory_create --input "${batch_file}" >> "${LOG_FILE}" 2>> "${ERROR_LOG}"; then
        log "✓ Successfully imported ${batch_name}"
        ((TOTAL_IMPORTED++))
        return 0
    else
        error "✗ Failed to import ${batch_name}"
        ((TOTAL_FAILED++))
        return 1
    fi
}
```

### 3. Progress Tracking

The script tracks and reports:
- Total batches to import
- Total records to import
- Number of successful imports
- Number of failed imports

### 4. Error Handling

- Exits immediately on any command failure (`set -e`)
- Exits on undefined variables (`set -u`)
- Exits on pipe failures (`set -o pipefail`)
- Provides detailed error messages with timestamps
- Creates separate error log for troubleshooting

## Example Output

```
[2025-11-19 20:52:51] Starting MCP Memory import
[2025-11-19 20:52:51] Total batches: 5
[2025-11-19 20:52:51] Total records: 31
[2025-11-19 20:52:51] Importing workout batches...
[2025-11-19 20:52:51] Importing batch_1_workouts.json...
[2025-11-19 20:52:52] ✓ Successfully imported batch_1_workouts.json
[2025-11-19 20:52:52] Importing state of mind batches...
[2025-11-19 20:52:52] Importing batch_1_state_of_mind.json...
[2025-11-19 20:52:53] ✓ Successfully imported batch_1_state_of_mind.json
[2025-11-19 20:52:53] Importing metric batches...
[2025-11-19 20:52:53] Importing batch_1_metrics.json...
[2025-11-19 20:52:54] ✓ Successfully imported batch_1_metrics.json
[2025-11-19 20:52:54] Importing batch_2_metrics.json...
[2025-11-19 20:52:55] ✓ Successfully imported batch_2_metrics.json
[2025-11-19 20:52:55] Importing batch_3_metrics.json...
[2025-11-19 20:52:56] ✓ Successfully imported batch_3_metrics.json
[2025-11-19 20:52:56] Import complete
[2025-11-19 20:52:56] Successfully imported: 5 batches
[2025-11-19 20:52:56] Failed imports: 0 batches
[2025-11-19 20:52:56] Total records: 31
[2025-11-19 20:52:56] All batches imported successfully!
```

## Memory CLI Integration

The script uses the Memory MCP CLI tools interface:

```bash
memory tools run --tool memory_memory_create --input batch_file.json
```

This command:
1. Invokes the `memory_memory_create` MCP tool
2. Reads the batch JSON file as input
3. Creates memories in the MCP Memory server
4. Returns success/failure status

## Troubleshooting

### Script Fails with "Memory binary not found"

**Problem**: The Memory CLI binary is not in your PATH.

**Solution**:
- Install the Memory CLI and ensure it's in your PATH, or
- Set `MEMORY_BIN` environment variable: `export MEMORY_BIN=/path/to/memory`, or
- Re-generate the script with `--memory-binary /path/to/memory`

### Import Fails for Specific Batches

**Problem**: Some batches fail to import while others succeed.

**Solution**:
1. Check `import_errors.log` for detailed error messages
2. Verify the batch JSON files are valid
3. Ensure the target collections exist in MCP Memory
4. Check MCP Memory server connectivity and credentials

### Script Exits Early

**Problem**: Script stops after first error due to `set -e`.

**Solution**: This is intentional behavior to prevent partial imports. Fix the error and re-run the script. The script is idempotent - successfully imported batches can be imported again without issues.

## Advanced Usage

### Running in Background

```bash
nohup ./import.sh > console.log 2>&1 &
```

### Monitoring Progress

```bash
tail -f import.log
```

### Checking for Errors

```bash
tail -f import_errors.log
```

### Re-running Failed Imports

If some batches fail, fix the underlying issue and re-run the script. The Memory MCP server handles duplicate imports gracefully.

### Manual Import Commands

If you prefer not to use the generated script, you can import batches manually:

```bash
# Navigate to import directory
cd exports/2025-11-17/import

# Import each batch type manually
memory tools run --tool memory_memory_create --input batch_1_workouts.json
memory tools run --tool memory_memory_create --input batch_1_state_of_mind.json
memory tools run --tool memory_memory_create --input batch_1_metrics.json
memory tools run --tool memory_memory_create --input batch_2_metrics.json
memory tools run --tool memory_memory_create --input batch_3_metrics.json
```

### Verifying Imports

After importing, verify the data was successfully added to MCP Memory:

```bash
# Check collection statistics
memory tools run --tool memory_collection_stats --input '{"id": "spinal_fusion_recovery"}'

# Search for workouts in the collection
memory tools run --tool memory_memory_search --input '{
  "collection": "spinal_fusion_recovery",
  "query": "workout",
  "limit": 10
}'

# Search for specific metrics
memory tools run --tool memory_memory_search --input '{
  "collection": "spinal_fusion_recovery",
  "query": "heart rate",
  "limit": 5
}'

# Search with metadata filters
memory tools run --tool memory_memory_search --input '{
  "collection": "spinal_fusion_recovery",
  "query": "outdoor walk",
  "filter": {"metadata.distance": {"$gt": 2.0}},
  "limit": 5
}'
```

### Batch Import with Verification Loop

For production workflows, you might want to verify each batch import:

```bash
#!/bin/bash
IMPORT_DIR="exports/2025-11-17/import"

for batch_file in "${IMPORT_DIR}"/batch_*.json; do
    echo "Importing $(basename "${batch_file}")..."

    if memory tools run --tool memory_memory_create --input "${batch_file}"; then
        echo "✓ Success: $(basename "${batch_file}")"
    else
        echo "✗ Failed: $(basename "${batch_file}")"
        exit 1
    fi
done

echo "All batches imported successfully!"
```

## Implementation Details

The import script is generated by the `generateMCPImportScript()` function in `cmd/process.go`. The function:

1. Reads the `BatchSummary` struct containing batch statistics
2. Generates a bash script with:
   - Script header with metadata
   - Configuration variables
   - Helper functions for logging
   - Pre-flight validation
   - Import function
   - Batch import calls (organized by type)
   - Summary reporting
3. Writes the script to `import.sh` with executable permissions (0755)

The generated script filename is `import.sh` and is always created in the same directory as the batch files.
