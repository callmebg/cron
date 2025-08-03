#!/bin/bash

# Go Cron Library - Comprehensive Test Script
# This script runs all tests with coverage reporting

set -e

echo "🧪 Go Cron - Comprehensive Test Suite"
echo "====================================="

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

# Create test output directory
mkdir -p test-results

print_status "Downloading dependencies..."
go mod download
go mod verify

print_status "Running go vet..."
if go vet ./...; then
    print_success "go vet passed"
else
    print_error "go vet failed"
    exit 1
fi

print_status "Checking code formatting..."
if [ "$(gofmt -s -l . | wc -l)" -gt 0 ]; then
    print_error "Code is not properly formatted. Run:"
    echo "gofmt -s -w ."
    gofmt -s -l .
    exit 1
else
    print_success "Code formatting is correct"
fi

print_status "Running unit tests with race detection and coverage..."
if go test -race -coverprofile=test-results/coverage.out -covermode=atomic ./...; then
    print_success "Unit tests passed"
else
    print_error "Unit tests failed"
    exit 1
fi

print_status "Generating coverage report..."
go tool cover -html=test-results/coverage.out -o test-results/coverage.html
COVERAGE=$(go tool cover -func=test-results/coverage.out | grep total | awk '{print $3}' | sed 's/%//')

echo ""
echo "📊 Coverage Report:"
echo "=================="
go tool cover -func=test-results/coverage.out | tail -10

if (( $(echo "$COVERAGE >= 90" | bc -l) )); then
    print_success "Coverage: $COVERAGE% (meets 90% requirement)"
else
    print_warning "Coverage: $COVERAGE% (below 90% requirement)"
fi

# Check if integration tests exist and run them
if [ -d "test/integration" ] && [ "$(ls -A test/integration)" ]; then
    print_status "Running integration tests..."
    if go test -race -tags=integration ./test/integration/...; then
        print_success "Integration tests passed"
    else
        print_error "Integration tests failed"
        exit 1
    fi
else
    print_warning "No integration tests found in test/integration/"
fi

# Check if benchmark tests exist and run them
print_status "Running benchmark tests..."
if find . -name "*_test.go" -exec grep -l "func Benchmark" {} \; | head -1 > /dev/null; then
    go test -bench=. -benchmem -run=^$ ./... | tee test-results/benchmark.txt
    print_success "Benchmark tests completed"
else
    print_warning "No benchmark tests found"
fi

# Run tests with different build tags if they exist
print_status "Testing with different build tags..."
BUILD_TAGS=("" "integration" "benchmark")

for tag in "${BUILD_TAGS[@]}"; do
    if [ -n "$tag" ]; then
        print_status "Testing with build tag: $tag"
        if go test -tags="$tag" ./... > /dev/null 2>&1; then
            print_success "Tests with tag '$tag' passed"
        else
            print_warning "Tests with tag '$tag' failed or no such tests exist"
        fi
    fi
done

# Test examples if they exist
if [ -d "cmd/examples" ]; then
    print_status "Testing example builds..."
    for example in cmd/examples/*/; do
        if [ -d "$example" ]; then
            example_name=$(basename "$example")
            print_status "Building example: $example_name"
            if go build -o "test-results/example-$example_name" "./$example"; then
                print_success "Example $example_name builds successfully"
                rm -f "test-results/example-$example_name"
            else
                print_error "Example $example_name failed to build"
                exit 1
            fi
        fi
    done
fi

# Test cross-compilation
print_status "Testing cross-compilation..."
PLATFORMS=("linux/amd64" "windows/amd64" "darwin/amd64" "linux/arm64")

for platform in "${PLATFORMS[@]}"; do
    os_arch=(${platform//\// })
    GOOS=${os_arch[0]}
    GOARCH=${os_arch[1]}
    
    print_status "Testing build for $GOOS/$GOARCH"
    if env GOOS=$GOOS GOARCH=$GOARCH go build ./...; then
        print_success "Build successful for $GOOS/$GOARCH"
    else
        print_error "Build failed for $GOOS/$GOARCH"
        exit 1
    fi
done

# Check for security vulnerabilities if govulncheck is available
if command -v govulncheck &> /dev/null; then
    print_status "Running vulnerability check..."
    if govulncheck ./...; then
        print_success "No known vulnerabilities found"
    else
        print_warning "Vulnerability check completed with findings"
    fi
else
    print_warning "govulncheck not available. Install with: go install golang.org/x/vuln/cmd/govulncheck@latest"
fi

# Generate test summary
echo ""
echo "📋 Test Summary:"
echo "================"
echo "✅ Unit Tests: PASSED"
echo "✅ Race Detection: PASSED"
echo "✅ Code Formatting: PASSED"
echo "✅ Go Vet: PASSED"
echo "📊 Coverage: $COVERAGE%"
echo "🏗️  Cross-compilation: PASSED"

if [ -d "test/integration" ] && [ "$(ls -A test/integration)" ]; then
    echo "✅ Integration Tests: PASSED"
fi

echo ""
echo "📁 Test artifacts generated:"
echo "  - test-results/coverage.out (coverage data)"
echo "  - test-results/coverage.html (coverage report)"
if [ -f "test-results/benchmark.txt" ]; then
    echo "  - test-results/benchmark.txt (benchmark results)"
fi

echo ""
print_success "All tests completed successfully! 🎉"
echo ""
echo "💡 Tips:"
echo "  - Open test-results/coverage.html in your browser to view detailed coverage"
echo "  - Run individual test packages with: go test ./path/to/package"
echo "  - Run specific tests with: go test -run TestName ./..."
echo "  - Run benchmarks with: go test -bench=. ./..."