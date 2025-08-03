#!/bin/bash

# Go Cron Library - Benchmark Script
# This script runs performance benchmarks and analyzes results

set -e

echo "🏃 Go Cron - Performance Benchmark Suite"
echo "========================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if Go is installed
if ! command -v go &> /dev/null; then
    print_error "Go is not installed or not in PATH"
    exit 1
fi

GO_VERSION=$(go version | cut -d' ' -f3)
print_status "Using Go version: $GO_VERSION"

# Create benchmark output directory
mkdir -p benchmark-results

# Run benchmarks with different parameters
BENCHMARK_TIME=${BENCHMARK_TIME:-"10s"}
BENCHMARK_COUNT=${BENCHMARK_COUNT:-"5"}

print_status "Benchmark configuration:"
echo "  - Duration: $BENCHMARK_TIME"
echo "  - Iterations: $BENCHMARK_COUNT"
echo ""

# Function to run benchmarks for a package
run_package_benchmarks() {
    local package=$1
    local package_name=$(basename "$package")
    
    print_status "Running benchmarks for $package_name..."
    
    # Run benchmarks with memory allocation stats
    if go test -bench=. -benchmem -benchtime="$BENCHMARK_TIME" -count="$BENCHMARK_COUNT" \
       -run=^$ "$package" > "benchmark-results/${package_name}.txt" 2>&1; then
        print_success "Benchmarks completed for $package_name"
        
        # Show summary
        echo "Results preview:"
        tail -20 "benchmark-results/${package_name}.txt" | head -10
        echo ""
    else
        print_warning "No benchmarks found or failed for $package_name"
    fi
}

# Run CPU profile benchmark
run_cpu_profile() {
    print_status "Running CPU profile benchmark..."
    
    if go test -bench=. -benchmem -cpuprofile=benchmark-results/cpu.prof \
       -benchtime="$BENCHMARK_TIME" ./...; then
        print_success "CPU profile generated"
        
        if command -v go tool pprof &> /dev/null; then
            print_status "Analyzing CPU profile..."
            echo "Top 10 functions by CPU usage:" > benchmark-results/cpu-analysis.txt
            go tool pprof -text -top10 benchmark-results/cpu.prof >> benchmark-results/cpu-analysis.txt 2>/dev/null || true
            
            # Show preview
            if [ -f benchmark-results/cpu-analysis.txt ]; then
                echo "CPU Profile Analysis:"
                cat benchmark-results/cpu-analysis.txt
                echo ""
            fi
        fi
    else
        print_warning "CPU profiling failed or no benchmarks available"
    fi
}

# Run memory profile benchmark
run_memory_profile() {
    print_status "Running memory profile benchmark..."
    
    if go test -bench=. -benchmem -memprofile=benchmark-results/mem.prof \
       -benchtime="$BENCHMARK_TIME" ./...; then
        print_success "Memory profile generated"
        
        if command -v go tool pprof &> /dev/null; then
            print_status "Analyzing memory profile..."
            echo "Top 10 functions by memory allocation:" > benchmark-results/mem-analysis.txt
            go tool pprof -text -top10 benchmark-results/mem.prof >> benchmark-results/mem-analysis.txt 2>/dev/null || true
            
            # Show preview
            if [ -f benchmark-results/mem-analysis.txt ]; then
                echo "Memory Profile Analysis:"
                cat benchmark-results/mem-analysis.txt
                echo ""
            fi
        fi
    else
        print_warning "Memory profiling failed or no benchmarks available"
    fi
}

# Main benchmark execution
print_status "Starting benchmark suite..."

# Find all packages with benchmark tests
BENCHMARK_PACKAGES=$(find . -name "*_test.go" -exec grep -l "func Benchmark" {} \; | \
                    xargs -I {} dirname {} | sort -u | grep -v vendor || echo "")

if [ -z "$BENCHMARK_PACKAGES" ]; then
    print_warning "No benchmark tests found"
    echo ""
    echo "To add benchmarks, create functions like:"
    echo "func BenchmarkYourFunction(b *testing.B) {"
    echo "    for i := 0; i < b.N; i++ {"
    echo "        // Your code to benchmark"
    echo "    }"
    echo "}"
    exit 0
fi

echo "Found benchmark tests in packages:"
for pkg in $BENCHMARK_PACKAGES; do
    echo "  - $pkg"
done
echo ""

# Run benchmarks for each package
for package in $BENCHMARK_PACKAGES; do
    run_package_benchmarks "$package"
done

# Run profiling if benchmarks exist
if [ -n "$BENCHMARK_PACKAGES" ]; then
    run_cpu_profile
    run_memory_profile
fi

# Generate comprehensive benchmark report
print_status "Generating comprehensive benchmark report..."

cat > benchmark-results/report.md << 'EOF'
# Go Cron - Benchmark Report

## Executive Summary

This report contains performance benchmark results for the Go Cron library.

## Test Environment

EOF

echo "- **Date**: $(date)" >> benchmark-results/report.md
echo "- **Go Version**: $GO_VERSION" >> benchmark-results/report.md
echo "- **OS**: $(uname -s)" >> benchmark-results/report.md
echo "- **Architecture**: $(uname -m)" >> benchmark-results/report.md
echo "- **CPU**: $(grep 'model name' /proc/cpuinfo | head -1 | cut -d: -f2 | xargs 2>/dev/null || echo 'Unknown')" >> benchmark-results/report.md
echo "- **Memory**: $(free -h | grep '^Mem:' | awk '{print $2}' 2>/dev/null || echo 'Unknown')" >> benchmark-results/report.md
echo "" >> benchmark-results/report.md

# Add benchmark results to report
echo "## Benchmark Results" >> benchmark-results/report.md
echo "" >> benchmark-results/report.md

for package in $BENCHMARK_PACKAGES; do
    package_name=$(basename "$package")
    if [ -f "benchmark-results/${package_name}.txt" ]; then
        echo "### $package_name" >> benchmark-results/report.md
        echo "" >> benchmark-results/report.md
        echo '```' >> benchmark-results/report.md
        cat "benchmark-results/${package_name}.txt" >> benchmark-results/report.md
        echo '```' >> benchmark-results/report.md
        echo "" >> benchmark-results/report.md
    fi
done

# Add profile analysis if available
if [ -f benchmark-results/cpu-analysis.txt ]; then
    echo "## CPU Profile Analysis" >> benchmark-results/report.md
    echo "" >> benchmark-results/report.md
    echo '```' >> benchmark-results/report.md
    cat benchmark-results/cpu-analysis.txt >> benchmark-results/report.md
    echo '```' >> benchmark-results/report.md
    echo "" >> benchmark-results/report.md
fi

if [ -f benchmark-results/mem-analysis.txt ]; then
    echo "## Memory Profile Analysis" >> benchmark-results/report.md
    echo "" >> benchmark-results/report.md
    echo '```' >> benchmark-results/report.md
    cat benchmark-results/mem-analysis.txt >> benchmark-results/report.md
    echo '```' >> benchmark-results/report.md
    echo "" >> benchmark-results/report.md
fi

# Compare with previous results if available
if [ -f benchmark-results/previous.txt ] && [ -f benchmark-results/current.txt ]; then
    print_status "Comparing with previous benchmark results..."
    
    if command -v benchcmp &> /dev/null; then
        benchcmp benchmark-results/previous.txt benchmark-results/current.txt > benchmark-results/comparison.txt
        echo "## Comparison with Previous Results" >> benchmark-results/report.md
        echo "" >> benchmark-results/report.md
        echo '```' >> benchmark-results/report.md
        cat benchmark-results/comparison.txt >> benchmark-results/report.md
        echo '```' >> benchmark-results/report.md
    else
        print_warning "benchcmp not available for comparison. Install with:"
        echo "  go install golang.org/x/tools/cmd/benchcmp@latest"
    fi
fi

# Performance analysis and recommendations
cat >> benchmark-results/report.md << 'EOF'

## Performance Analysis

### Key Metrics

The benchmarks measure several important performance characteristics:

1. **Throughput**: Operations per second
2. **Latency**: Time per operation
3. **Memory Usage**: Allocations per operation
4. **Memory Efficiency**: Bytes allocated per operation

### Recommendations

Based on the benchmark results:

- Monitor memory allocations in hot paths
- Consider object pooling for frequently allocated objects
- Profile CPU usage in production workloads
- Test with realistic cron expression complexity

## Running Benchmarks

To run these benchmarks yourself:

```bash
# Run all benchmarks
go test -bench=. -benchmem ./...

# Run specific benchmark
go test -bench=BenchmarkSpecific -benchmem ./package

# Run with CPU profiling
go test -bench=. -cpuprofile=cpu.prof ./...

# Run with memory profiling  
go test -bench=. -memprofile=mem.prof ./...
```

## Benchmark Interpretation

- **ns/op**: Nanoseconds per operation (lower is better)
- **B/op**: Bytes allocated per operation (lower is better)
- **allocs/op**: Number of allocations per operation (lower is better)

EOF

print_success "Benchmark report generated: benchmark-results/report.md"

# Summary output
echo ""
echo "📊 Benchmark Summary:"
echo "===================="

# Count total benchmarks
TOTAL_BENCHMARKS=$(find benchmark-results -name "*.txt" -exec grep -c "^Benchmark" {} \; 2>/dev/null | awk '{sum+=$1} END {print sum+0}')
echo "✅ Total benchmarks run: $TOTAL_BENCHMARKS"

# Show any performance concerns
if grep -q "slow" benchmark-results/*.txt 2>/dev/null; then
    print_warning "Some benchmarks may indicate performance concerns"
else
    print_success "All benchmarks completed successfully"
fi

echo ""
echo "📁 Benchmark artifacts generated:"
echo "  - benchmark-results/report.md (comprehensive report)"
echo "  - benchmark-results/*.txt (raw benchmark data)"
if [ -f benchmark-results/cpu.prof ]; then
    echo "  - benchmark-results/cpu.prof (CPU profile)"
fi
if [ -f benchmark-results/mem.prof ]; then
    echo "  - benchmark-results/mem.prof (memory profile)"
fi

echo ""
echo "💡 Next steps:"
echo "  - Review benchmark-results/report.md for detailed analysis"
echo "  - Use 'go tool pprof' to analyze profile files interactively"
echo "  - Compare results over time to track performance changes"
echo "  - Consider setting up automated benchmark tracking"

# Save current results for future comparison
cat benchmark-results/*.txt > benchmark-results/current.txt 2>/dev/null || true

print_success "Benchmark suite completed! 🏁"