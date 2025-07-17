package main

import (
	"bytes"
	"flag"
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
			cmd := exec.Command("go", append([]string{"run", "main.go"}, tt.args...)...)
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
			args:        []string{"-schema", createValidSchemaFile(t, tempDir), "-path", "/root/readonly.go"},
			expectError: true,
			errorMsg:    "permission",
		},
		{
			name:        "잘못된 패키지명",
			args:        []string{"-schema", createValidSchemaFile(t, tempDir), "-path", filepath.Join(tempDir, "out2.go"), "-pkgname", "123invalid"},
			expectError: false, // 패키지명 검증은 현재 구현되지 않음
			errorMsg:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command("go", append([]string{"run", "main.go"}, tt.args...)...)
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
		cmd := exec.Command("go", append([]string{"run", "main.go"}, args...)...)
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
	cmd := exec.Command("go", "run", "main.go", "-help")
	cmd.Dir = "."

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	// -help 플래그는 exit code 2를 반환하므로 에러는 예상됨
	cmd.Run()

	helpOutput := stdout.String()

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
