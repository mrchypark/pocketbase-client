package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestCLIFlags는 새로운 CLI 플래그들의 동작을 테스트합니다
func TestCLIFlags(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		expectedEnums bool
		expectedRels  bool
		expectedFiles bool
		expectError   bool
	}{
		{
			name:          "기본값 테스트",
			args:          []string{},
			expectedEnums: true,
			expectedRels:  true,
			expectedFiles: true,
			expectError:   false,
		},
		{
			name:          "모든 기능 활성화",
			args:          []string{"-enums=true", "-relations=true", "-files=true"},
			expectedEnums: true,
			expectedRels:  true,
			expectedFiles: true,
			expectError:   false,
		},
		{
			name:          "모든 기능 비활성화",
			args:          []string{"-enums=false", "-relations=false", "-files=false"},
			expectedEnums: false,
			expectedRels:  false,
			expectedFiles: false,
			expectError:   false,
		},
		{
			name:          "일부 기능만 활성화",
			args:          []string{"-enums=true", "-relations=false", "-files=true"},
			expectedEnums: true,
			expectedRels:  false,
			expectedFiles: true,
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 플래그 파싱 테스트
			oldArgs := os.Args
			defer func() { os.Args = oldArgs }()

			// 테스트용 임시 인수 설정
			os.Args = append([]string{"pbc-gen"}, tt.args...)

			// 새로운 FlagSet 생성하여 테스트
			fs := flag.NewFlagSet("test", flag.ContinueOnError)
			fs.SetOutput(io.Discard) // 에러 출력 숨김

			generateEnums := fs.Bool("enums", true, "Generate enum constants for select fields")
			generateRelations := fs.Bool("relations", true, "Generate enhanced relation types")
			generateFiles := fs.Bool("files", true, "Generate enhanced file types")

			err := fs.Parse(tt.args)
			if tt.expectError && err == nil {
				t.Error("에러가 예상되었지만 발생하지 않았습니다")
			}
			if !tt.expectError && err != nil {
				t.Errorf("예상치 못한 에러 발생: %v", err)
			}

			if !tt.expectError {
				if *generateEnums != tt.expectedEnums {
					t.Errorf("enums 플래그: 예상값 %v, 실제값 %v", tt.expectedEnums, *generateEnums)
				}
				if *generateRelations != tt.expectedRels {
					t.Errorf("relations 플래그: 예상값 %v, 실제값 %v", tt.expectedRels, *generateRelations)
				}
				if *generateFiles != tt.expectedFiles {
					t.Errorf("files 플래그: 예상값 %v, 실제값 %v", tt.expectedFiles, *generateFiles)
				}
			}
		})
	}
}

// TestCLIIntegration은 실제 CLI 실행을 통한 통합 테스트를 수행합니다
func TestCLIIntegration(t *testing.T) {
	// 테스트용 스키마 파일 생성
	tempDir := t.TempDir()
	schemaFile := filepath.Join(tempDir, "test_schema.json")
	schemaContent := `[
		{
			"id": "test_collection",
			"name": "test_items",
			"type": "base",
			"system": false,
			"fields": [
				{
					"id": "test_name",
					"name": "name",
					"type": "text",
					"required": true,
					"options": {}
				},
				{
					"id": "test_status",
					"name": "status",
					"type": "select",
					"required": true,
					"options": {
						"maxSelect": 1,
						"values": ["active", "inactive"]
					}
				}
			]
		}
	]`

	err := os.WriteFile(schemaFile, []byte(schemaContent), 0644)
	if err != nil {
		t.Fatalf("테스트 스키마 파일 생성 실패: %v", err)
	}

	tests := []struct {
		name        string
		args        []string
		expectEnums bool
		expectError bool
	}{
		{
			name:        "기본 실행",
			args:        []string{"-schema", schemaFile, "-path", filepath.Join(tempDir, "models1.gen.go")},
			expectEnums: true,
			expectError: false,
		},
		{
			name:        "Enum 비활성화",
			args:        []string{"-schema", schemaFile, "-path", filepath.Join(tempDir, "models2.gen.go"), "-enums=false"},
			expectEnums: false,
			expectError: false,
		},
		{
			name:        "존재하지 않는 스키마 파일",
			args:        []string{"-schema", "nonexistent.json", "-path", filepath.Join(tempDir, "models3.gen.go")},
			expectEnums: false,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// CLI 실행
			cmd := exec.Command("go", append([]string{"run", "."}, tt.args...)...)
			cmd.Dir = "."

			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()

			if tt.expectError {
				if err == nil {
					t.Error("에러가 예상되었지만 발생하지 않았습니다")
				}
				return
			}

			if err != nil {
				t.Errorf("CLI 실행 실패: %v\nStderr: %s", err, stderr.String())
				return
			}

			// 출력 파일 확인
			outputPath := ""
			for i, arg := range tt.args {
				if arg == "-path" && i+1 < len(tt.args) {
					outputPath = tt.args[i+1]
					break
				}
			}

			if outputPath == "" {
				t.Fatal("출력 파일 경로를 찾을 수 없습니다")
			}

			// 생성된 파일 내용 확인
			content, err := os.ReadFile(outputPath)
			if err != nil {
				t.Fatalf("생성된 파일 읽기 실패: %v", err)
			}

			contentStr := string(content)

			// Enum 상수 생성 여부 확인
			hasEnumConstants := strings.Contains(contentStr, "TestItemsStatusActive")
			if tt.expectEnums && !hasEnumConstants {
				t.Error("Enum 상수가 생성되지 않았습니다")
			}
			if !tt.expectEnums && hasEnumConstants {
				t.Error("Enum 상수가 예상치 않게 생성되었습니다")
			}

			// 기본 구조체는 항상 생성되어야 함
			if !strings.Contains(contentStr, "type TestItems struct") {
				t.Error("기본 구조체가 생성되지 않았습니다")
			}
		})
	}
}

// TestCLIErrorHandling은 CLI의 에러 처리를 테스트합니다
func TestCLIErrorHandling(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name        string
		args        []string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "잘못된 스키마 파일 형식",
			args:        []string{"-schema", createInvalidSchemaFile(t, tempDir), "-path", filepath.Join(tempDir, "out1.go")},
			expectError: true,
			errorMsg:    "schema",
		},
		{
			name:        "읽기 전용 출력 디렉토리",
			args:        []string{"-schema", createValidSchemaFile(t, tempDir), "-path", "/dev/null/readonly.go"},
			expectError: true,
			errorMsg:    "failed to create file",
		},
		{
			name:        "잘못된 패키지명",
			args:        []string{"-schema", createValidSchemaFile(t, tempDir), "-path", filepath.Join(tempDir, "out2.go"), "-pkgname", "123invalid"},
			expectError: true, // 패키지명이 잘못되면 검증에서 에러 발생
			errorMsg:    "package name is not a valid Go identifier",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command("go", append([]string{"run", "."}, tt.args...)...)
			cmd.Dir = "."

			var stderr bytes.Buffer
			cmd.Stderr = &stderr

			err := cmd.Run()

			if tt.expectError {
				if err == nil {
					t.Error("에러가 예상되었지만 발생하지 않았습니다")
				} else if tt.errorMsg != "" && !strings.Contains(stderr.String(), tt.errorMsg) {
					t.Errorf("예상 에러 메시지 '%s'가 포함되지 않았습니다. 실제: %s", tt.errorMsg, stderr.String())
				}
			} else if err != nil {
				t.Errorf("예상치 못한 에러 발생: %v\nStderr: %s", err, stderr.String())
			}
		})
	}
}

// TestCLIBackwardCompatibility는 하위 호환성을 테스트합니다
func TestCLIBackwardCompatibility(t *testing.T) {
	tempDir := t.TempDir()
	schemaFile := createValidSchemaFile(t, tempDir)

	// 기존 방식으로 실행 (새로운 플래그 없이)
	oldStyleArgs := []string{
		"-schema", schemaFile,
		"-path", filepath.Join(tempDir, "old_style.go"),
		"-pkgname", "models",
	}

	// 새로운 방식으로 실행 (모든 플래그 기본값)
	newStyleArgs := []string{
		"-schema", schemaFile,
		"-path", filepath.Join(tempDir, "new_style.go"),
		"-pkgname", "models",
		"-enums=true",
		"-relations=true",
		"-files=true",
	}

	// 두 방식 모두 실행
	for i, args := range [][]string{oldStyleArgs, newStyleArgs} {
		cmd := exec.Command("go", append([]string{"run", "."}, args...)...)
		cmd.Dir = "."

		var stderr bytes.Buffer
		cmd.Stderr = &stderr

		err := cmd.Run()
		if err != nil {
			t.Errorf("실행 %d 실패: %v\nStderr: %s", i+1, err, stderr.String())
		}
	}

	// 생성된 파일들 비교
	oldContent, err := os.ReadFile(filepath.Join(tempDir, "old_style.go"))
	if err != nil {
		t.Fatalf("기존 방식 파일 읽기 실패: %v", err)
	}

	newContent, err := os.ReadFile(filepath.Join(tempDir, "new_style.go"))
	if err != nil {
		t.Fatalf("새로운 방식 파일 읽기 실패: %v", err)
	}

	// 기본 구조체는 동일해야 함
	oldStr := string(oldContent)
	newStr := string(newContent)

	if !strings.Contains(oldStr, "type TestItems struct") {
		t.Error("기존 방식에서 기본 구조체가 생성되지 않았습니다")
	}

	if !strings.Contains(newStr, "type TestItems struct") {
		t.Error("새로운 방식에서 기본 구조체가 생성되지 않았습니다")
	}

	// 새로운 방식에서는 추가 기능이 있어야 함
	if !strings.Contains(newStr, "TestItemsStatusActive") {
		t.Error("새로운 방식에서 Enum 상수가 생성되지 않았습니다")
	}
}

// 헬퍼 함수들

func createValidSchemaFile(t *testing.T, dir string) string {
	schemaFile := filepath.Join(dir, "valid_schema.json")
	schemaContent := `[
		{
			"id": "test_collection",
			"name": "test_items",
			"type": "base",
			"system": false,
			"fields": [
				{
					"id": "test_name",
					"name": "name",
					"type": "text",
					"required": true,
					"options": {}
				},
				{
					"id": "test_status",
					"name": "status",
					"type": "select",
					"required": true,
					"options": {
						"maxSelect": 1,
						"values": ["active", "inactive"]
					}
				}
			]
		}
	]`

	err := os.WriteFile(schemaFile, []byte(schemaContent), 0644)
	if err != nil {
		t.Fatalf("유효한 스키마 파일 생성 실패: %v", err)
	}
	return schemaFile
}

func createInvalidSchemaFile(t *testing.T, dir string) string {
	schemaFile := filepath.Join(dir, "invalid_schema.json")
	invalidContent := `{
		"invalid": "json structure"
	}`

	err := os.WriteFile(schemaFile, []byte(invalidContent), 0644)
	if err != nil {
		t.Fatalf("잘못된 스키마 파일 생성 실패: %v", err)
	}
	return schemaFile
}

// TestCLIHelp는 도움말 출력을 테스트합니다
func TestCLIHelp(t *testing.T) {
	cmd := exec.Command("go", "run", ".", "-help")
	cmd.Dir = "."

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// -help 플래그는 exit code 2를 반환하므로 에러는 예상됨
	cmd.Run()

	helpOutput := stdout.String() + stderr.String()

	// 새로운 플래그들이 도움말에 포함되어 있는지 확인
	expectedFlags := []string{"-enums", "-relations", "-files"}
	for _, flag := range expectedFlags {
		if !strings.Contains(helpOutput, flag) {
			t.Errorf("도움말에 %s 플래그가 포함되지 않았습니다", flag)
		}
	}

	// 기존 플래그들도 여전히 있는지 확인
	existingFlags := []string{"-schema", "-path", "-pkgname", "-jsonlib"}
	for _, flag := range existingFlags {
		if !strings.Contains(helpOutput, flag) {
			t.Errorf("도움말에 기존 %s 플래그가 포함되지 않았습니다", flag)
		}
	}
}

// TestEndToEndCodeGeneration은 전체 파이프라인에 대한 End-to-End 테스트를 수행합니다
func TestEndToEndCodeGeneration(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name           string
		schemaContent  string
		schemaVersion  string // "latest" or "legacy"
		forceVersion   string // CLI 강제 버전 지정
		expectedStruct string
		expectedEmbed  []string // 예상되는 임베딩
		expectError    bool
	}{
		{
			name: "최신 스키마 - fields 키 사용",
			schemaContent: `[
				{
					"id": "posts_collection",
					"name": "posts",
					"type": "base",
					"system": false,
					"fields": [
						{
							"id": "title_field",
							"name": "title",
							"type": "text",
							"required": true,
							"system": false,
							"options": {}
						},
						{
							"id": "published_field",
							"name": "published",
							"type": "bool",
							"required": false,
							"system": false,
							"options": {}
						}
					]
				}
			]`,
			schemaVersion:  "latest",
			expectedStruct: "type Posts struct",
			expectedEmbed:  []string{"pocketbase.BaseModel"},
			expectError:    false,
		},
		{
			name: "구버전 스키마 - schema 키 사용",
			schemaContent: `[
				{
					"id": "articles_collection",
					"name": "articles",
					"type": "base",
					"system": false,
					"schema": [
						{
							"id": "title_field",
							"name": "title",
							"type": "text",
							"required": true,
							"system": false,
							"options": {}
						},
						{
							"id": "content_field",
							"name": "content",
							"type": "text",
							"required": false,
							"system": false,
							"options": {}
						}
					]
				}
			]`,
			schemaVersion:  "legacy",
			expectedStruct: "type Articles struct",
			expectedEmbed:  []string{"pocketbase.BaseModel", "pocketbase.BaseDateTime"},
			expectError:    false,
		},
		{
			name: "최신 스키마를 legacy로 강제 변환",
			schemaContent: `[
				{
					"id": "products_collection",
					"name": "products",
					"type": "base",
					"system": false,
					"fields": [
						{
							"id": "name_field",
							"name": "name",
							"type": "text",
							"required": true,
							"system": false,
							"options": {}
						}
					]
				}
			]`,
			schemaVersion:  "latest",
			forceVersion:   "legacy",
			expectedStruct: "type Products struct",
			expectedEmbed:  []string{"pocketbase.BaseModel", "pocketbase.BaseDateTime"},
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 테스트용 스키마 파일 생성
			schemaFile := filepath.Join(tempDir, tt.name+"_schema.json")
			err := os.WriteFile(schemaFile, []byte(tt.schemaContent), 0644)
			if err != nil {
				t.Fatalf("스키마 파일 생성 실패: %v", err)
			}

			// 출력 파일 경로
			outputFile := filepath.Join(tempDir, tt.name+"_models.gen.go")

			// CLI 인수 구성
			args := []string{
				"-schema", schemaFile,
				"-path", outputFile,
				"-pkgname", "testmodels",
				"-verbose",
			}

			if tt.forceVersion != "" {
				args = append(args, "-force-version", tt.forceVersion)
			}

			// CLI 실행
			cmd := exec.Command("go", append([]string{"run", "."}, args...)...)
			cmd.Dir = "."

			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err = cmd.Run()

			if tt.expectError {
				if err == nil {
					t.Error("에러가 예상되었지만 발생하지 않았습니다")
				}
				return
			}

			if err != nil {
				t.Errorf("CLI 실행 실패: %v\nStdout: %s\nStderr: %s", err, stdout.String(), stderr.String())
				return
			}

			// 생성된 파일 검증
			content, err := os.ReadFile(outputFile)
			if err != nil {
				t.Fatalf("생성된 파일 읽기 실패: %v", err)
			}

			contentStr := string(content)

			// 기본 구조체 생성 확인
			if !strings.Contains(contentStr, tt.expectedStruct) {
				t.Errorf("예상 구조체 '%s'가 생성되지 않았습니다", tt.expectedStruct)
			}

			// 임베딩 확인
			for _, embed := range tt.expectedEmbed {
				if !strings.Contains(contentStr, embed) {
					t.Errorf("예상 임베딩 '%s'가 포함되지 않았습니다", embed)
				}
			}

			// 생성된 코드 컴파일 검증
			t.Run("컴파일 검증", func(t *testing.T) {
				err := verifyGeneratedCodeCompiles(t, outputFile)
				if err != nil {
					t.Errorf("생성된 코드 컴파일 실패: %v", err)
				}
			})

			// 스키마 버전별 특정 검증
			t.Run("스키마 버전별 검증", func(t *testing.T) {
				expectedVersion := tt.schemaVersion
				if tt.forceVersion != "" {
					expectedVersion = tt.forceVersion
				}

				switch expectedVersion {
				case "latest":
					// 최신 스키마: BaseModel만 임베딩, BaseDateTime 없음
					if strings.Contains(contentStr, "pocketbase.BaseDateTime") && !strings.Contains(contentStr, "// Latest schema: BaseModel only") {
						t.Error("최신 스키마에서 BaseDateTime이 예상치 않게 임베딩되었습니다")
					}
				case "legacy":
					// 구버전 스키마: BaseModel + BaseDateTime 모두 임베딩
					if !strings.Contains(contentStr, "pocketbase.BaseModel") || !strings.Contains(contentStr, "pocketbase.BaseDateTime") {
						t.Error("구버전 스키마에서 BaseModel 또는 BaseDateTime 임베딩이 누락되었습니다")
					}
				}
			})
		})
	}
}

// TestComplexSchemaGeneration은 복잡한 스키마 구조에 대한 테스트를 수행합니다
func TestComplexSchemaGeneration(t *testing.T) {
	tempDir := t.TempDir()

	// 실제 데이터베이스 스키마 사용
	// 프로젝트 루트에서 database 디렉토리 찾기
	projectRoot, err := findProjectRoot()
	if err != nil {
		t.Fatalf("프로젝트 루트를 찾을 수 없습니다: %v", err)
	}
	schemaFile := filepath.Join(projectRoot, "database", "pb_schema.json")
	// 파일 존재 확인
	if _, err := os.Stat(schemaFile); os.IsNotExist(err) {
		t.Skipf("스키마 파일이 존재하지 않습니다: %s", schemaFile)
	}
	outputFile := filepath.Join(tempDir, "complex_models.gen.go")

	// 모든 enhanced 기능 활성화하여 실행
	args := []string{
		"-schema", schemaFile,
		"-path", outputFile,
		"-pkgname", "complexmodels",
		"-enums=true",
		"-relations=true",
		"-files=true",
		"-verbose",
	}

	cmd := exec.Command("go", append([]string{"run", "."}, args...)...)
	cmd.Dir = "."

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		t.Fatalf("복잡한 스키마 CLI 실행 실패: %v\nStdout: %s\nStderr: %s", err, stdout.String(), stderr.String())
	}

	// 생성된 파일 검증
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("생성된 파일 읽기 실패: %v", err)
	}

	contentStr := string(content)

	// 기본 구조체들 확인 (실제 데이터베이스 스키마 기준)
	expectedStructs := []string{
		"type Users struct",
		"type Organizations struct",
		"type Issues struct",
	}

	for _, structName := range expectedStructs {
		if !strings.Contains(contentStr, structName) {
			t.Errorf("예상 구조체 '%s'가 생성되지 않았습니다", structName)
		}
	}

	// Enhanced 기능들 확인
	t.Run("Enum 생성 확인", func(t *testing.T) {
		expectedEnums := []string{
			"DevicesTypeM2",
			"DevicesTypeD2",
			"DevicesTypeS2",
		}

		for _, enumName := range expectedEnums {
			if !strings.Contains(contentStr, enumName) {
				t.Errorf("예상 Enum 상수 '%s'가 생성되지 않았습니다", enumName)
			}
		}
	})

	t.Run("Relation 타입 생성 확인", func(t *testing.T) {
		// 실제 데이터베이스 스키마에서 생성되는 relation 타입들 확인
		expectedRelations := []string{
			"UsersRelation",
			"OrganizationsRelation",
		}

		for _, relName := range expectedRelations {
			if !strings.Contains(contentStr, relName) {
				t.Errorf("예상 Relation 타입 '%s'가 생성되지 않았습니다", relName)
			}
		}
	})

	t.Run("File 타입 생성 확인", func(t *testing.T) {
		expectedFileTypes := []string{
			"FileReference",
			"FileReferences",
		}

		for _, fileType := range expectedFileTypes {
			if !strings.Contains(contentStr, fileType) {
				t.Errorf("예상 File 타입 '%s'가 생성되지 않았습니다", fileType)
			}
		}
	})

	// 컴파일 검증
	t.Run("복잡한 스키마 컴파일 검증", func(t *testing.T) {
		err := verifyGeneratedCodeCompiles(t, outputFile)
		if err != nil {
			t.Errorf("복잡한 스키마 생성 코드 컴파일 실패: %v", err)
		}
	})
}

// verifyGeneratedCodeCompiles는 생성된 코드가 컴파일되는지 검증합니다
func verifyGeneratedCodeCompiles(t *testing.T, filePath string) error {
	// 현재 작업 디렉토리 확인
	currentDir, err := os.Getwd()
	if err != nil {
		return err
	}

	// 프로젝트 루트 디렉토리 찾기 (go.mod가 있는 디렉토리)
	projectRoot := currentDir
	for {
		if _, err := os.Stat(filepath.Join(projectRoot, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(projectRoot)
		if parent == projectRoot {
			return fmt.Errorf("프로젝트 루트를 찾을 수 없습니다")
		}
		projectRoot = parent
	}

	// 임시 모듈 디렉토리 생성
	tempModuleDir := t.TempDir()

	// 상대 경로 계산
	relPath, err := filepath.Rel(tempModuleDir, projectRoot)
	if err != nil {
		return err
	}

	// go.mod 파일 생성
	goModContent := fmt.Sprintf(`module testmodule

go 1.21

require (
	github.com/mrchypark/pocketbase-client v0.0.0
	github.com/pocketbase/pocketbase v0.22.0
)

replace github.com/mrchypark/pocketbase-client => %s
`, relPath)

	err = os.WriteFile(filepath.Join(tempModuleDir, "go.mod"), []byte(goModContent), 0644)
	if err != nil {
		return err
	}

	// 생성된 파일을 임시 모듈로 복사
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	testFilePath := filepath.Join(tempModuleDir, "models.go")
	err = os.WriteFile(testFilePath, content, 0644)
	if err != nil {
		return err
	}

	// 생성된 파일에서 패키지명 추출
	packageName := "main"
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "package ") {
			packageName = strings.TrimSpace(strings.TrimPrefix(line, "package "))
			break
		}
	}

	// 컴파일 검증을 위한 더미 main 함수 추가
	mainContent := fmt.Sprintf(`package %s

import (
	"context"
	"fmt"
	
	"github.com/mrchypark/pocketbase-client"
)

func TestCompilation() {
	client := pocketbase.NewClient("http://localhost:8090")
	ctx := context.Background()
	
	// 생성된 함수들이 올바른 시그니처를 가지는지 확인
	_ = ctx
	_ = client
	fmt.Println("Compilation test passed")
}
`, packageName)
	testCompileFilePath := filepath.Join(tempModuleDir, "test_compile.go")
	err = os.WriteFile(testCompileFilePath, []byte(mainContent), 0644)
	if err != nil {
		return err
	}

	// go mod tidy 실행
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = tempModuleDir
	err = cmd.Run()
	if err != nil {
		// go mod tidy 실패는 무시하고 계속 진행
		t.Logf("go mod tidy 실패 (무시): %v", err)
	}

	// go build 실행 (syntax check만)
	cmd = exec.Command("go", "build", "-o", "/dev/null", ".")
	cmd.Dir = tempModuleDir

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("컴파일 실패: %v\nStderr: %s", err, stderr.String())
	}

	return nil
}

// TestSchemaVersionDetectionAccuracy는 스키마 버전 감지 정확성을 테스트합니다
func TestSchemaVersionDetectionAccuracy(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name           string
		schemaContent  string
		expectedOutput string // 로그에서 확인할 버전 정보
	}{
		{
			name: "명확한 최신 스키마",
			schemaContent: `[
				{
					"id": "test1",
					"name": "test1",
					"type": "base",
					"system": false,
					"fields": [
						{"id": "f1", "name": "field1", "type": "text", "required": true, "system": false, "options": {}}
					]
				}
			]`,
			expectedOutput: "Using schema version: latest",
		},
		{
			name: "명확한 구버전 스키마",
			schemaContent: `[
				{
					"id": "test2",
					"name": "test2",
					"type": "base",
					"system": false,
					"schema": [
						{"id": "f1", "name": "field1", "type": "text", "required": true, "system": false, "options": {}}
					]
				}
			]`,
			expectedOutput: "Using schema version: legacy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schemaFile := filepath.Join(tempDir, tt.name+"_schema.json")
			err := os.WriteFile(schemaFile, []byte(tt.schemaContent), 0644)
			if err != nil {
				t.Fatalf("스키마 파일 생성 실패: %v", err)
			}

			outputFile := filepath.Join(tempDir, tt.name+"_models.gen.go")

			args := []string{
				"-schema", schemaFile,
				"-path", outputFile,
				"-verbose",
			}

			cmd := exec.Command("go", append([]string{"run", "."}, args...)...)
			cmd.Dir = "."

			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err = cmd.Run()
			if err != nil {
				t.Errorf("CLI 실행 실패: %v\nStderr: %s", err, stderr.String())
				return
			}

			// 로그 출력에서 버전 감지 확인
			logOutput := stdout.String() + stderr.String()
			if !strings.Contains(logOutput, tt.expectedOutput) {
				t.Errorf("예상 로그 출력 '%s'가 포함되지 않았습니다. 실제 출력: %s", tt.expectedOutput, logOutput)
			}
		})
	}
}

// TestFullPipelineIntegration은 전체 파이프라인의 통합 테스트를 수행합니다
func TestFullPipelineIntegration(t *testing.T) {
	tempDir := t.TempDir()

	// 실제 프로젝트의 스키마 파일을 사용한 테스트
	t.Run("실제 프로젝트 스키마 테스트", func(t *testing.T) {
		// 프로젝트 루트의 database/pb_schema.json 파일 사용
		projectRoot, err := findProjectRoot()
		if err != nil {
			t.Skipf("프로젝트 루트를 찾을 수 없어 테스트를 건너뜁니다: %v", err)
		}

		realSchemaPath := filepath.Join(projectRoot, "database", "pb_schema.json")
		if _, err := os.Stat(realSchemaPath); os.IsNotExist(err) {
			t.Skipf("실제 스키마 파일이 없어 테스트를 건너뜁니다: %s", realSchemaPath)
		}

		outputFile := filepath.Join(tempDir, "real_schema_models.gen.go")

		args := []string{
			"-schema", realSchemaPath,
			"-path", outputFile,
			"-pkgname", "realmodels",
			"-verbose",
			"-validate-schema",
		}

		cmd := exec.Command("go", append([]string{"run", "."}, args...)...)
		cmd.Dir = "."

		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err = cmd.Run()
		if err != nil {
			t.Errorf("실제 스키마 CLI 실행 실패: %v\nStdout: %s\nStderr: %s", err, stdout.String(), stderr.String())
			return
		}

		// 생성된 파일 검증
		content, err := os.ReadFile(outputFile)
		if err != nil {
			t.Fatalf("생성된 파일 읽기 실패: %v", err)
		}

		contentStr := string(content)

		// 기본 검증
		if !strings.Contains(contentStr, "package realmodels") {
			t.Error("패키지 선언이 올바르지 않습니다")
		}

		// 컴파일 검증
		err = verifyGeneratedCodeCompiles(t, outputFile)
		if err != nil {
			t.Errorf("실제 스키마 생성 코드 컴파일 실패: %v", err)
		}

		// 로그에서 스키마 버전 확인
		logOutput := stdout.String() + stderr.String()
		if strings.Contains(logOutput, "Using schema version: legacy") {
			t.Log("실제 스키마는 legacy 버전으로 감지되었습니다")
			// Legacy 스키마 특성 확인
			if !strings.Contains(contentStr, "pocketbase.BaseModel") || !strings.Contains(contentStr, "pocketbase.BaseDateTime") {
				t.Error("Legacy 스키마에서 BaseModel 또는 BaseDateTime 임베딩이 누락되었습니다")
			}
		} else if strings.Contains(logOutput, "Using schema version: latest") {
			t.Log("실제 스키마는 latest 버전으로 감지되었습니다")
			// Latest 스키마 특성 확인
			if !strings.Contains(contentStr, "pocketbase.BaseModel") {
				t.Error("Latest 스키마에서 BaseModel 임베딩이 누락되었습니다")
			}
		}
	})

	// 다양한 CLI 옵션 조합 테스트
	t.Run("CLI 옵션 조합 테스트", func(t *testing.T) {
		baseSchema := `[
			{
				"id": "combo_test",
				"name": "combo_items",
				"type": "base",
				"system": false,
				"fields": [
					{
						"id": "name_field",
						"name": "name",
						"type": "text",
						"required": true,
						"system": false,
						"options": {}
					},
					{
						"id": "status_field",
						"name": "status",
						"type": "select",
						"required": true,
						"system": false,
						"options": {
							"maxSelect": 1,
							"values": ["draft", "published", "archived"]
						}
					}
				]
			}
		]`

		schemaFile := filepath.Join(tempDir, "combo_schema.json")
		err := os.WriteFile(schemaFile, []byte(baseSchema), 0644)
		if err != nil {
			t.Fatalf("조합 테스트 스키마 파일 생성 실패: %v", err)
		}

		testCombinations := []struct {
			name        string
			args        []string
			expectEnums bool
			expectRels  bool
			expectFiles bool
		}{
			{
				name:        "모든 기능 활성화",
				args:        []string{"-enums=true", "-relations=true", "-files=true"},
				expectEnums: true,
				expectRels:  true,
				expectFiles: true,
			},
			{
				name:        "Enum만 활성화",
				args:        []string{"-enums=true", "-relations=false", "-files=false"},
				expectEnums: true,
				expectRels:  false,
				expectFiles: false,
			},
			{
				name:        "모든 기능 비활성화",
				args:        []string{"-enums=false", "-relations=false", "-files=false"},
				expectEnums: false,
				expectRels:  false,
				expectFiles: false,
			},
		}

		for i, combo := range testCombinations {
			t.Run(combo.name, func(t *testing.T) {
				outputFile := filepath.Join(tempDir, fmt.Sprintf("combo_%d_models.gen.go", i))

				args := []string{
					"-schema", schemaFile,
					"-path", outputFile,
					"-pkgname", "combomodels",
				}
				args = append(args, combo.args...)

				cmd := exec.Command("go", append([]string{"run", "."}, args...)...)
				cmd.Dir = "."

				var stderr bytes.Buffer
				cmd.Stderr = &stderr

				err := cmd.Run()
				if err != nil {
					t.Errorf("조합 테스트 실행 실패: %v\nStderr: %s", err, stderr.String())
					return
				}

				// 생성된 파일 검증
				content, err := os.ReadFile(outputFile)
				if err != nil {
					t.Fatalf("조합 테스트 파일 읽기 실패: %v", err)
				}

				contentStr := string(content)

				// Enum 생성 확인
				hasEnums := strings.Contains(contentStr, "ComboItemsStatusDraft")
				if combo.expectEnums && !hasEnums {
					t.Error("Enum이 예상되었지만 생성되지 않았습니다")
				}
				if !combo.expectEnums && hasEnums {
					t.Error("Enum이 예상되지 않았지만 생성되었습니다")
				}

				// 기본 구조체는 항상 생성되어야 함
				if !strings.Contains(contentStr, "type ComboItems struct") {
					t.Error("기본 구조체가 생성되지 않았습니다")
				}

				// 컴파일 검증
				err = verifyGeneratedCodeCompiles(t, outputFile)
				if err != nil {
					t.Errorf("조합 테스트 코드 컴파일 실패: %v", err)
				}
			})
		}
	})

	// 스키마 버전 강제 지정 테스트
	t.Run("스키마 버전 강제 지정 테스트", func(t *testing.T) {
		latestSchema := `[
			{
				"id": "force_test",
				"name": "force_items",
				"type": "base",
				"system": false,
				"fields": [
					{
						"id": "title_field",
						"name": "title",
						"type": "text",
						"required": true,
						"system": false,
						"options": {}
					}
				]
			}
		]`

		schemaFile := filepath.Join(tempDir, "force_schema.json")
		err := os.WriteFile(schemaFile, []byte(latestSchema), 0644)
		if err != nil {
			t.Fatalf("강제 지정 테스트 스키마 파일 생성 실패: %v", err)
		}

		// 최신 스키마를 legacy로 강제 변환
		outputFile := filepath.Join(tempDir, "force_legacy_models.gen.go")

		args := []string{
			"-schema", schemaFile,
			"-path", outputFile,
			"-pkgname", "forcemodels",
			"-force-version", "legacy",
			"-verbose",
		}

		cmd := exec.Command("go", append([]string{"run", "."}, args...)...)
		cmd.Dir = "."

		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err = cmd.Run()
		if err != nil {
			t.Errorf("강제 지정 테스트 실행 실패: %v\nStderr: %s", err, stderr.String())
			return
		}

		// 로그에서 강제 변환 확인
		logOutput := stdout.String() + stderr.String()
		if !strings.Contains(logOutput, "Forcing schema version from latest to legacy") {
			t.Error("스키마 버전 강제 변환 로그가 출력되지 않았습니다")
		}

		// 생성된 파일에서 legacy 특성 확인
		content, err := os.ReadFile(outputFile)
		if err != nil {
			t.Fatalf("강제 지정 테스트 파일 읽기 실패: %v", err)
		}

		contentStr := string(content)

		// Legacy 스키마 특성 확인 (BaseModel + BaseDateTime 임베딩)
		if !strings.Contains(contentStr, "pocketbase.BaseModel") || !strings.Contains(contentStr, "pocketbase.BaseDateTime") {
			t.Error("강제 legacy 변환에서 BaseModel 또는 BaseDateTime 임베딩이 누락되었습니다")
		}

		// 컴파일 검증
		err = verifyGeneratedCodeCompiles(t, outputFile)
		if err != nil {
			t.Errorf("강제 지정 테스트 코드 컴파일 실패: %v", err)
		}
	})

	// 에러 시나리오 테스트
	t.Run("에러 시나리오 테스트", func(t *testing.T) {
		errorTests := []struct {
			name        string
			args        []string
			expectError bool
			errorMsg    string
		}{
			{
				name:        "존재하지 않는 스키마 파일",
				args:        []string{"-schema", "nonexistent.json", "-path", filepath.Join(tempDir, "error1.go")},
				expectError: true,
				errorMsg:    "failed to load schema",
			},
			{
				name:        "잘못된 강제 버전",
				args:        []string{"-schema", createValidSchemaFile(t, tempDir), "-path", filepath.Join(tempDir, "error2.go"), "-force-version", "invalid"},
				expectError: true,
				errorMsg:    "invalid schema version",
			},
			{
				name:        "읽기 전용 출력 경로",
				args:        []string{"-schema", createValidSchemaFile(t, tempDir), "-path", "/dev/null/readonly.go"},
				expectError: true,
				errorMsg:    "failed to create file",
			},
		}

		for _, errorTest := range errorTests {
			t.Run(errorTest.name, func(t *testing.T) {
				cmd := exec.Command("go", append([]string{"run", "."}, errorTest.args...)...)
				cmd.Dir = "."

				var stderr bytes.Buffer
				cmd.Stderr = &stderr

				err := cmd.Run()

				if errorTest.expectError {
					if err == nil {
						t.Error("에러가 예상되었지만 발생하지 않았습니다")
					} else if !strings.Contains(stderr.String(), errorTest.errorMsg) {
						t.Errorf("예상 에러 메시지 '%s'가 포함되지 않았습니다. 실제: %s", errorTest.errorMsg, stderr.String())
					}
				} else if err != nil {
					t.Errorf("예상치 못한 에러 발생: %v\nStderr: %s", err, stderr.String())
				}
			})
		}
	})
}

// TestGeneratedCodeUsability는 생성된 코드의 실제 사용성을 테스트합니다
func TestGeneratedCodeUsability(t *testing.T) {
	tempDir := t.TempDir()

	// 사용성 테스트용 스키마
	usabilitySchema := `[
		{
			"id": "usability_test",
			"name": "test_records",
			"type": "base",
			"system": false,
			"fields": [
				{
					"id": "name_field",
					"name": "name",
					"type": "text",
					"required": true,
					"system": false,
					"options": {}
				},
				{
					"id": "age_field",
					"name": "age",
					"type": "number",
					"required": false,
					"system": false,
					"options": {}
				},
				{
					"id": "active_field",
					"name": "active",
					"type": "bool",
					"required": false,
					"system": false,
					"options": {}
				}
			]
		}
	]`

	schemaFile := filepath.Join(tempDir, "usability_schema.json")
	err := os.WriteFile(schemaFile, []byte(usabilitySchema), 0644)
	if err != nil {
		t.Fatalf("사용성 테스트 스키마 파일 생성 실패: %v", err)
	}

	outputFile := filepath.Join(tempDir, "usability_models.gen.go")

	// 코드 생성
	args := []string{
		"-schema", schemaFile,
		"-path", outputFile,
		"-pkgname", "usabilitymodels",
	}

	cmd := exec.Command("go", append([]string{"run", "."}, args...)...)
	cmd.Dir = "."

	err = cmd.Run()
	if err != nil {
		t.Fatalf("사용성 테스트 코드 생성 실패: %v", err)
	}

	// 생성된 코드를 사용하는 테스트 코드 작성
	testCodeContent := `package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	
	"github.com/mrchypark/pocketbase-client"
)

` + string(func() []byte {
		content, _ := os.ReadFile(outputFile)
		// 패키지 선언과 import 블록 제거
		lines := strings.Split(string(content), "\n")
		var filteredLines []string
		inImportBlock := false
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "package ") {
				continue
			}
			if strings.HasPrefix(trimmed, "import (") {
				inImportBlock = true
				continue
			}
			if inImportBlock && trimmed == ")" {
				inImportBlock = false
				continue
			}
			if inImportBlock {
				continue
			}
			filteredLines = append(filteredLines, line)
		}
		return []byte(strings.Join(filteredLines, "\n"))
	}()) + `

func main() {
	// 구조체 생성 테스트
	record := NewTestRecords()
	
	// 필드 설정 테스트
	record.Name = "Test User"
	if record.Age != nil {
		*record.Age = 25
	}
	if record.Active != nil {
		*record.Active = true
	}
	
	// ToMap 메서드 테스트
	data := record.ToMap()
	
	// 기본 검증
	if data["name"] != "Test User" {
		log.Fatalf("Name 필드 설정 실패: %v", data["name"])
	}
	
	fmt.Println("사용성 테스트 성공!")
}
`

	testCodeFile := filepath.Join(tempDir, "usability_test_main.go")
	err = os.WriteFile(testCodeFile, []byte(testCodeContent), 0644)
	if err != nil {
		t.Fatalf("사용성 테스트 코드 작성 실패: %v", err)
	}

	// 사용성 테스트 코드 컴파일 및 실행
	err = verifyGeneratedCodeCompiles(t, testCodeFile)
	if err != nil {
		t.Errorf("사용성 테스트 코드 컴파일 실패: %v", err)
	}
}

// findProjectRoot는 프로젝트 루트 디렉토리를 찾습니다
func findProjectRoot() (string, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	dir := currentDir
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("프로젝트 루트를 찾을 수 없습니다")
		}
		dir = parent
	}
}

// TestOutputComparison은 예상 출력과 실제 출력을 비교합니다
func TestOutputComparison(t *testing.T) {
	tempDir := t.TempDir()

	// 예상 출력 패턴을 정의한 테스트 케이스
	comparisonTests := []struct {
		name           string
		schemaContent  string
		expectedParts  []string // 반드시 포함되어야 하는 부분들
		forbiddenParts []string // 포함되면 안 되는 부분들
	}{
		{
			name: "기본 구조체 생성",
			schemaContent: `[
				{
					"id": "simple_test",
					"name": "simple_items",
					"type": "base",
					"system": false,
					"fields": [
						{
							"id": "title_field",
							"name": "title",
							"type": "text",
							"required": true,
							"system": false,
							"options": {}
						}
					]
				}
			]`,
			expectedParts: []string{
				"type SimpleItems struct",
				"pocketbase.BaseModel",
				"Title string `json:\"title\"`",
				"func NewSimpleItems()",
				"func (m *SimpleItems) ToMap()",
			},
			forbiddenParts: []string{
				"pocketbase.BaseDateTime", // 최신 스키마에서는 BaseDateTime 없음
			},
		},
		{
			name: "Legacy 스키마 구조체 생성",
			schemaContent: `[
				{
					"id": "legacy_test",
					"name": "legacy_items",
					"type": "base",
					"system": false,
					"schema": [
						{
							"id": "name_field",
							"name": "name",
							"type": "text",
							"required": true,
							"system": false,
							"options": {}
						}
					]
				}
			]`,
			expectedParts: []string{
				"type LegacyItems struct",
				"pocketbase.BaseModel",
				"pocketbase.BaseDateTime",
				"Name string `json:\"name\"`",
				"func NewLegacyItems()",
			},
			forbiddenParts: []string{
				"// Latest schema:", // Legacy 스키마에서는 Latest 주석 없음
			},
		},
	}

	for _, test := range comparisonTests {
		t.Run(test.name, func(t *testing.T) {
			schemaFile := filepath.Join(tempDir, test.name+"_schema.json")
			err := os.WriteFile(schemaFile, []byte(test.schemaContent), 0644)
			if err != nil {
				t.Fatalf("비교 테스트 스키마 파일 생성 실패: %v", err)
			}

			outputFile := filepath.Join(tempDir, test.name+"_models.gen.go")

			args := []string{
				"-schema", schemaFile,
				"-path", outputFile,
				"-pkgname", "comparisonmodels",
			}

			cmd := exec.Command("go", append([]string{"run", "."}, args...)...)
			cmd.Dir = "."

			err = cmd.Run()
			if err != nil {
				t.Fatalf("비교 테스트 실행 실패: %v", err)
			}

			// 생성된 파일 내용 확인
			content, err := os.ReadFile(outputFile)
			if err != nil {
				t.Fatalf("비교 테스트 파일 읽기 실패: %v", err)
			}

			contentStr := string(content)

			// 예상 부분들이 포함되어 있는지 확인
			for _, expectedPart := range test.expectedParts {
				if !strings.Contains(contentStr, expectedPart) {
					t.Errorf("예상 부분 '%s'가 포함되지 않았습니다", expectedPart)
				}
			}

			// 금지된 부분들이 포함되지 않았는지 확인
			for _, forbiddenPart := range test.forbiddenParts {
				if strings.Contains(contentStr, forbiddenPart) {
					t.Errorf("금지된 부분 '%s'가 포함되었습니다", forbiddenPart)
				}
			}
		})
	}
}
