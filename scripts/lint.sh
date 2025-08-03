#!/bin/bash

# Go Cron Library - Code Quality and Linting Script
# This script runs various linting tools and code quality checks

set -e

echo "🔍 Go Cron - Code Quality Check"
echo "==============================="

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

# Create output directory
mkdir -p lint-results

# Initialize error flag
HAS_ERRORS=0

print_status "Running go fmt check..."
if [ "$(gofmt -s -l . | wc -l)" -gt 0 ]; then
    print_error "Code is not properly formatted:"
    gofmt -s -l .
    echo ""
    echo "Run the following command to fix formatting:"
    echo "gofmt -s -w ."
    HAS_ERRORS=1
else
    print_success "Code formatting is correct"
fi

print_status "Running go vet..."
if go vet ./... 2> lint-results/vet.log; then
    print_success "go vet passed"
else
    print_error "go vet found issues:"
    cat lint-results/vet.log
    HAS_ERRORS=1
fi

# Check for golangci-lint
if command -v golangci-lint &> /dev/null; then
    print_status "Running golangci-lint..."
    if golangci-lint run --out-format=checkstyle > lint-results/golangci-lint.xml 2>&1; then
        print_success "golangci-lint passed"
    else
        print_error "golangci-lint found issues:"
        golangci-lint run
        HAS_ERRORS=1
    fi
else
    print_warning "golangci-lint not found. Install with:"
    echo "  curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b \$(go env GOPATH)/bin"
fi

# Check for gosec (security linter)
if command -v gosec &> /dev/null; then
    print_status "Running gosec security scan..."
    if gosec -fmt json -out lint-results/gosec.json ./... 2>/dev/null; then
        # Check if any issues were found
        ISSUES=$(jq '.Issues | length' lint-results/gosec.json 2>/dev/null || echo "0")
        if [ "$ISSUES" -gt 0 ]; then
            print_warning "gosec found $ISSUES security issues:"
            gosec ./...
        else
            print_success "gosec found no security issues"
        fi
    else
        print_warning "gosec scan completed with warnings"
    fi
else
    print_warning "gosec not found. Install with:"
    echo "  go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"
fi

# Check for staticcheck
if command -v staticcheck &> /dev/null; then
    print_status "Running staticcheck..."
    if staticcheck ./... > lint-results/staticcheck.log 2>&1; then
        print_success "staticcheck passed"
    else
        print_error "staticcheck found issues:"
        cat lint-results/staticcheck.log
        HAS_ERRORS=1
    fi
else
    print_warning "staticcheck not found. Install with:"
    echo "  go install honnef.co/go/tools/cmd/staticcheck@latest"
fi

# Check for ineffassign
if command -v ineffassign &> /dev/null; then
    print_status "Running ineffassign..."
    if ineffassign ./... > lint-results/ineffassign.log 2>&1; then
        if [ -s lint-results/ineffassign.log ]; then
            print_warning "ineffassign found issues:"
            cat lint-results/ineffassign.log
        else
            print_success "ineffassign passed"
        fi
    else
        print_warning "ineffassign scan completed"
    fi
else
    print_warning "ineffassign not found. Install with:"
    echo "  go install github.com/gordonklaus/ineffassign@latest"
fi

# Check for misspell
if command -v misspell &> /dev/null; then
    print_status "Running misspell check..."
    if misspell -error . > lint-results/misspell.log 2>&1; then
        print_success "misspell check passed"
    else
        print_warning "misspell found issues:"
        cat lint-results/misspell.log
    fi
else
    print_warning "misspell not found. Install with:"
    echo "  go install github.com/client9/misspell/cmd/misspell@latest"
fi

# Check for gocyclo (cyclomatic complexity)
if command -v gocyclo &> /dev/null; then
    print_status "Running gocyclo (complexity check)..."
    if gocyclo -over 15 . > lint-results/gocyclo.log 2>&1; then
        if [ -s lint-results/gocyclo.log ]; then
            print_warning "gocyclo found complex functions:"
            cat lint-results/gocyclo.log
        else
            print_success "gocyclo check passed (complexity < 15)"
        fi
    else
        print_warning "gocyclo scan completed"
    fi
else
    print_warning "gocyclo not found. Install with:"
    echo "  go install github.com/fzipp/gocyclo/cmd/gocyclo@latest"
fi

# Check imports
print_status "Checking import organization..."
if command -v goimports &> /dev/null; then
    if [ "$(goimports -l . | wc -l)" -gt 0 ]; then
        print_warning "Import organization issues found:"
        goimports -l .
        echo ""
        echo "Run the following command to fix imports:"
        echo "goimports -w ."
    else
        print_success "Import organization is correct"
    fi
else
    print_warning "goimports not found. Install with:"
    echo "  go install golang.org/x/tools/cmd/goimports@latest"
fi

# Check for go mod issues
print_status "Checking go.mod and go.sum..."
if go mod verify; then
    print_success "go.mod and go.sum are valid"
else
    print_error "go.mod or go.sum validation failed"
    HAS_ERRORS=1
fi

if go mod tidy -diff > lint-results/mod-tidy.diff 2>&1; then
    if [ -s lint-results/mod-tidy.diff ]; then
        print_warning "go.mod needs tidying:"
        cat lint-results/mod-tidy.diff
        echo ""
        echo "Run: go mod tidy"
    else
        print_success "go.mod is properly tidied"
    fi
else
    print_warning "Could not check go mod tidy status"
fi

# Check for TODO/FIXME/HACK comments
print_status "Checking for TODO/FIXME/HACK comments..."
TODO_COUNT=$(grep -r -n "TODO\|FIXME\|HACK" --include="*.go" . | wc -l || echo "0")
if [ "$TODO_COUNT" -gt 0 ]; then
    print_warning "Found $TODO_COUNT TODO/FIXME/HACK comments:"
    grep -r -n "TODO\|FIXME\|HACK" --include="*.go" . | head -10
    if [ "$TODO_COUNT" -gt 10 ]; then
        echo "... and $((TODO_COUNT - 10)) more"
    fi
else
    print_success "No TODO/FIXME/HACK comments found"
fi

# Check for panic/println statements
print_status "Checking for panic/println statements..."
PANIC_COUNT=$(grep -r -n "panic\|println" --include="*.go" . | grep -v "_test.go" | wc -l || echo "0")
if [ "$PANIC_COUNT" -gt 0 ]; then
    print_warning "Found $PANIC_COUNT panic/println statements in non-test code:"
    grep -r -n "panic\|println" --include="*.go" . | grep -v "_test.go"
else
    print_success "No panic/println statements found in non-test code"
fi

# Check for hardcoded credentials or sensitive data
print_status "Checking for potential secrets..."
SECRETS_PATTERNS=(
    "password.*=.*['\"][^'\"]*['\"]"
    "token.*=.*['\"][^'\"]*['\"]"
    "secret.*=.*['\"][^'\"]*['\"]"
    "key.*=.*['\"][^'\"]*['\"]"
    "api_key.*=.*['\"][^'\"]*['\"]"
)

SECRET_FOUND=0
for pattern in "${SECRETS_PATTERNS[@]}"; do
    if grep -r -i -n "$pattern" --include="*.go" . > /dev/null 2>&1; then
        if [ $SECRET_FOUND -eq 0 ]; then
            print_warning "Potential secrets found:"
            SECRET_FOUND=1
        fi
        grep -r -i -n "$pattern" --include="*.go" .
    fi
done

if [ $SECRET_FOUND -eq 0 ]; then
    print_success "No obvious secrets detected"
fi

# Generate lint summary
echo ""
echo "📋 Lint Summary:"
echo "================"

if [ $HAS_ERRORS -eq 0 ]; then
    echo "✅ Code Formatting: PASSED"
    echo "✅ Go Vet: PASSED"
    echo "✅ Module Verification: PASSED"
    
    if command -v golangci-lint &> /dev/null; then
        echo "✅ GolangCI-Lint: PASSED"
    fi
    
    if command -v staticcheck &> /dev/null; then
        echo "✅ Staticcheck: PASSED"
    fi
    
    echo ""
    print_success "All linting checks passed! 🎉"
else
    echo "❌ Some linting checks failed"
    echo ""
    print_error "Please fix the issues above before proceeding"
fi

echo ""
echo "📁 Lint artifacts generated:"
echo "  - lint-results/ (detailed reports)"
echo ""
echo "💡 Recommended tools to install:"
echo "  - golangci-lint: comprehensive linter suite"
echo "  - gosec: security-focused linter"
echo "  - staticcheck: advanced static analysis"
echo "  - goimports: import organization"

exit $HAS_ERRORS