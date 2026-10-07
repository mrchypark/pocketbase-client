#!/bin/bash

# 로컬 CI 테스트 스크립트
# GitHub Actions CI 워크플로우와 동일한 단계를 로컬에서 실행합니다

set -e  # 에러 발생 시 스크립트 중단

echo "🚀 로컬 CI 테스트 시작..."

# 색상 정의
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 단계별 실행 함수
run_step() {
    local step_name="$1"
    local command="$2"
    
    echo -e "\n${BLUE}📋 단계: ${step_name}${NC}"
    echo "실행 명령: $command"
    
    if eval "$command"; then
        echo -e "${GREEN}✅ ${step_name} 성공${NC}"
    else
        echo -e "${RED}❌ ${step_name} 실패${NC}"
        exit 1
    fi
}

# Go 버전 확인
echo -e "\n${YELLOW}🔍 Go 버전 확인${NC}"
go version

# 1. 포맷 검사 (Format)
run_step "코드 포맷 검사" 'test -z "$(gofmt -l .)"'

# 2. Vet 검사
run_step "Go Vet 검사" "go vet ./..."
run_step "Installer integrity" "sh docs/tests/install_test.sh"

# 3. 일반 테스트
run_step "단위 테스트" "go test ./..."

if [ -n "${POCKETBASE_BIN:-}" ]; then
    run_step "PocketBase 0.39.10 contracts" "go test -run '^(TestPocketBaseContract|TestRuntimePocketBaseFileModifiers)$' -count=1 -timeout=3m ."
fi

# 4. Race 조건 테스트
run_step "Race 조건 테스트" "go test -race ./..."

# 5. 벤치마크 테스트
run_step "벤치마크 테스트" "go test -bench=. -benchmem ./..."

echo -e "\n${GREEN}🎉 모든 CI 테스트가 성공적으로 완료되었습니다!${NC}"
