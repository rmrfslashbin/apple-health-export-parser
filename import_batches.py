#!/usr/bin/env python3
"""
Import Apple Health batch files to MCP Memory server.
"""

import json
import sys
from pathlib import Path

def read_batch(filepath):
    """Read a batch JSON file."""
    with open(filepath, 'r') as f:
        return json.load(f)

def main():
    import_dir = Path('exports/2025-11-17-18/import')

    # Find all metric batch files
    metric_batches = sorted(import_dir.glob('batch_*_metrics.json'))

    print(f"Found {len(metric_batches)} metric batch files to import")

    # Combine all metric batches
    all_metrics = []
    for batch_file in metric_batches:
        metrics = read_batch(batch_file)
        all_metrics.extend(metrics)
        print(f"Loaded {len(metrics)} metrics from {batch_file.name}")

    print(f"\nTotal metrics to import: {len(all_metrics)}")

    # Output as JSON for MCP tool
    print("\nJSON output for MCP import:")
    print(json.dumps(all_metrics))

if __name__ == '__main__':
    main()
