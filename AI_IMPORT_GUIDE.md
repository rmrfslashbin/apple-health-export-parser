# AI Import Guide for Apple Health Data

This guide explains how to import and analyze Apple Health Export data in an AI chat session (like Claude Code).

## Overview

The Apple Health Export Parser processes your health data into an AI-friendly structure:

1. **Summary files** - Compact, aggregated data perfect for AI context windows
2. **Detail files** - Full time-series data available when needed
3. **Manifest** - Index of all exported files for easy navigation

## Export Structure

```
export_directory/
├── manifest.json                    # Start here! Index of all files
├── metrics/                         # Health metrics (one file per metric type)
│   ├── 2025-11-10_heart_rate.json
│   ├── 2025-11-10_step_count.json
│   └── ...
├── workouts/                        # Workout summaries (AI-friendly)
│   ├── 2025-11-12_Outdoor_Walk_summary.json
│   └── ...
├── workout_details/                 # Detailed time-series data
│   └── 2025-11-12_Outdoor_Walk/
│       ├── heart_rate.json         # 166 data points
│       ├── active_energy.json      # 1,513 data points
│       └── step_count.json         # 1,399 data points
└── state_of_mind/                   # Mental health recordings
    ├── 2025-11-11_daily_mood.json
    └── ...
```

## Step-by-Step AI Import Process

### Step 1: Read the Manifest

**Prompt for AI:**
```
Please read the file `export_directory/manifest.json` and summarize what data is available.
```

The manifest contains:
- Total counts of all data types
- Relative paths to all exported files
- Generation timestamp and trace ID
- Index of workout detail files

### Step 2: Understand Data Types

**Metrics** (~30-50 lines each):
- Complete time-series for each metric type
- Examples: heart_rate, step_count, sleep_analysis, vo2_max
- Each file contains all data points for that metric

**Workout Summaries** (~50-60 lines each):
- Core metadata (name, start/end time, duration)
- Environmental conditions (temperature, humidity)
- **Aggregated statistics** for all time-series:
  - `activeEnergyStats`: min/max/avg/total energy burned
  - `heartRateStats`: min/max/avg heart rate
  - `stepCountStats`: min/max/avg/total steps
- Data point counts (tells you what detail files exist)

**Workout Details** (1,000+ lines each):
- Full time-series data arrays
- Only fetch when you need granular analysis
- Separated by data type (heart_rate, active_energy, step_count, etc.)

**State of Mind** (~15-20 lines each):
- Mental health and mood recordings
- Includes valence scores and classifications

### Step 3: Analyze Summary Data

**Prompt for AI:**
```
Read all workout summary files from the `workouts/` directory.
Analyze my workout patterns:
1. What types of workouts do I do?
2. What are my average workout durations?
3. What are my typical heart rate ranges?
4. How much energy do I typically burn?
```

The AI can answer these questions using ONLY the summary files (very efficient).

### Step 4: Deep Dive with Detail Files (When Needed)

**Prompt for AI:**
```
For the Outdoor Walk on 2025-11-12, read the detailed heart rate data from
`workout_details/2025-11-12_17-05-04_Outdoor_Walk/heart_rate.json`

Analyze:
1. How did my heart rate change over time?
2. When did I reach peak heart rate?
3. How quickly did my heart rate recover?
```

### Step 5: Cross-Reference Data Types

**Prompt for AI:**
```
Compare my step_count metric data with my workout data:
1. What percentage of my daily steps come from workouts?
2. Are there patterns between workout intensity and daily activity?
```

## Example Analysis Prompts

### Workout Analysis
```
Using the workout summaries:
1. Calculate my total exercise time this week
2. What's my average workout heart rate?
3. Which workout burned the most calories?
4. Show trends in workout intensity over time
```

### Health Metrics Analysis
```
Read my heart_rate_variability.json and resting_heart_rate.json files.
Analyze:
1. What's my average HRV?
2. Is my resting heart rate trending up or down?
3. Are there correlations between HRV and workout days?
```

### Mental Health Correlation
```
Read all state_of_mind files and compare with workout_summaries:
1. Is there a correlation between workout frequency and mood?
2. Do certain workout types correlate with better mood scores?
3. What's my mood pattern over the week?
```

### Sleep Analysis
```
Read sleep_analysis.json and analyze:
1. Average sleep duration
2. Sleep consistency (bedtime/wake time variance)
3. Correlation between sleep quality and next-day workout performance
```

## Best Practices for AI Analysis

### 1. Start Small
- Begin with the manifest to understand what's available
- Read summaries first, only fetch details when needed
- This keeps context window usage manageable

### 2. Be Specific
- Request specific files rather than "all files"
- Example: "Read the 3 most recent workout summaries" not "Read all workouts"

### 3. Use Aggregated Data
- Workout summaries contain pre-calculated statistics
- No need to fetch detail files for basic questions like "average heart rate"

### 4. Progressive Analysis
```
Step 1: Read manifest → understand what data exists
Step 2: Read summaries → get overview and trends
Step 3: Read details → deep dive specific interesting patterns
```

### 5. Combine Data Types
- Correlate workouts with daily metrics
- Compare state of mind with activity levels
- Look at environmental conditions (temperature/humidity) vs performance

## File Size Guide

| File Type | Typical Size | Lines | Best For |
|-----------|-------------|-------|----------|
| Manifest | 5-10 KB | 100-200 | Navigation, overview |
| Metric | 1-5 KB | 30-100 | Complete metric history |
| Workout Summary | 1-2 KB | 50-60 | Quick analysis, trends |
| Workout Details | 50-500 KB | 1,000-20,000 | Deep dive, time-series analysis |
| State of Mind | 0.5 KB | 15-20 | Mood tracking |

## Common Analysis Workflows

### Weekly Health Summary
```
1. Read manifest
2. Read all workout summaries from past 7 days
3. Read relevant metric files (step_count, heart_rate)
4. Generate summary report
```

### Workout Performance Deep Dive
```
1. Read specific workout summary
2. If interesting, read detailed heart_rate data
3. Compare with similar previous workouts
4. Identify performance improvements/declines
```

### Long-term Trend Analysis
```
1. Read all workout summaries (use stats, not details)
2. Plot trends over time (duration, intensity, heart rate)
3. Identify patterns and seasonality
```

## Tips for Efficient Context Usage

1. **Use the manifest** to know what files exist before asking AI to fetch them
2. **Batch related requests**: "Read these 3 files and compare them"
3. **Leverage statistics**: Summary files contain pre-calculated min/max/avg
4. **Filter by date**: Only request data for relevant time periods
5. **Progressive detail**: Start with summaries, drill down only when needed

## Data Privacy Note

All data is processed locally. The export files contain YOUR personal health data:
- Heart rate measurements
- Location data (in GPX files for outdoor workouts)
- Mental health records
- Sleep patterns

Only share with AI assistants you trust, and consider the privacy implications of cloud-based AI services.
