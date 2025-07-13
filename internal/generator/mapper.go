package generator

import (
	// encoding/json 패키지 임포트 유지
	"fmt"
	"strings"
)

// MapPbTypeToGoType maps PocketBase field types to Go types and their corresponding getter methods.
func MapPbTypeToGoType(field FieldSchema, omitEmpty bool) (string, string, string) {
	var goType, comment, getterMethod string

	// MaxSelect 옵션을 기반으로 다중 선택 필드 여부를 미리 판단
	isMulti := false

	// field.Options.MaxSelect가 *int로 올바르게 파싱될 것을 기대합니다.
	if field.Options != nil && field.Options.MaxSelect != nil {
		if *field.Options.MaxSelect != 1 {
			isMulti = true
		}
	} else if field.Type == "relation" || field.Type == "file" || field.Type == "select" {
		// MaxSelect가 nil인 경우 (스키마에 maxSelect가 없거나 null인 경우),
		// relation/file/select 타입은 기본적으로 다중으로 간주합니다.
		isMulti = true
	}

	switch field.Type {
	case "text", "email", "url", "editor":
		goType = "string"
		getterMethod = "GetString"
		if omitEmpty {
			getterMethod = "GetStringPointer"
		}
	case "number":
		goType = "float64"
		getterMethod = "GetFloat"
		if omitEmpty {
			getterMethod = "GetFloatPointer"
		}
	case "bool":
		goType = "bool"
		getterMethod = "GetBool"
		if omitEmpty {
			getterMethod = "GetBoolPointer"
		}
	case "date", "autodate":
		goType = "types.DateTime"
		getterMethod = "GetDateTime"
		if omitEmpty {
			getterMethod = "GetDateTimePointer"
		}
	case "json":
		goType = "json.RawMessage"
		getterMethod = "GetRawMessage"
	case "relation", "file", "select":
		if isMulti {
			goType = "[]string"
			getterMethod = "GetStringSlice"
		} else { // 단일 선택/파일/관계
			goType = "string"
			getterMethod = "GetString"
			if omitEmpty {
				getterMethod = "GetStringPointer"
			}
		}
	default:
		goType = "interface{}"
		getterMethod = "Get"
	}

	// 디버깅 출력을 위해 MaxSelect 값을 안전하게 포맷
	var maxSelectDebugVal interface{} = "nil"
	if field.Options != nil && field.Options.MaxSelect != nil {
		maxSelectDebugVal = *field.Options.MaxSelect // 이제 *int로 올바르게 파싱되기를 기대합니다.
	}
	fmt.Printf("DEBUG: Field %s (Type: %s, MaxSelect: %v, omitEmpty: %t) -> goType: %s, getter: %s\n", field.Name, field.Type, maxSelectDebugVal, omitEmpty, goType, getterMethod)

	// 최종적으로 포인터 타입을 적용할지 결정 (이미 포인터이거나 슬라이스, json.RawMessage, interface{}인 경우는 제외)
	if omitEmpty && !strings.HasPrefix(goType, "[]") && goType != "json.RawMessage" && goType != "interface{}" && !strings.HasPrefix(goType, "*") {
		goType = "*" + goType
	}

	return goType, comment, getterMethod
}

func ToPascalCase(s string) string {
	if s == "" {
		return ""
	}

	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == '_' || r == '-' || r == ' '
	})

	for i, part := range parts {
		switch strings.ToLower(part) {
		case "id":
			parts[i] = "ID"
		case "url":
			parts[i] = "URL"
		case "html":
			parts[i] = "HTML"
		case "json":
			parts[i] = "JSON"
		default:
			parts[i] = strings.Title(part)
		}
	}

	return strings.Join(parts, "")
}
