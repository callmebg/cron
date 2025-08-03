# Test Data Files

This directory contains test data files used by various test suites.

## Files

- `cron_expressions.json` - Valid and invalid cron expressions for testing
- `job_configs.json` - Various job configurations for testing
- `timezone_data.json` - Timezone test data
- `benchmark_data.json` - Data for performance benchmarks

## Usage

These files are used by:
- Integration tests (`../integration/`)
- Benchmark tests (`../benchmark/`)
- Unit tests in the main packages

## Format

All data files use JSON format for easy parsing and modification.