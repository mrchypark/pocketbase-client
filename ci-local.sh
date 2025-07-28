#!/bin/bash

# Local CI test script
# Runs the same steps as GitHub Actions CI workflow locally

set -e  # Exit script on error

echo "🚀 Starting local CI test..."

# Color definitions
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Step-by-step execution function
run_step() {
    local step_name="$1"
    local command="$2"
    
    echo -e "\n${BLUE}📋 Step: ${step_name}${NC}"
    echo "Executing command: $command"
    
    if eval "$command"; then
        echo -e "${GREEN}✅ ${step_name} succeeded${NC}"
    else
        echo -e "${RED}❌ ${step_name} failed${NC}"
        exit 1
    fi
}

# Check Go version
echo -e "\n${YELLOW}🔍 Checking Go version${NC}"
go version

# 1. Format check
run_step "Code format check" "gofmt -w . && git diff --exit-code"

# 2. Vet check
run_step "Go Vet check" "go vet ./..."

# 3. Unit tests
run_step "Unit tests" "go test ./..."

# 4. Race condition tests
run_step "Race condition tests" "go test -race ./..."

# 5. Benchmark tests
run_step "Benchmark tests" "go test -bench=. -benchmem ./..."

echo -e "\n${GREEN}🎉 All CI tests completed successfully!${NC}"