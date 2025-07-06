package generator

import (
	"strings"
	"unicode"
)

// MapPbTypeToGoType는 PocketBase 필드 타입을 Go 타입으로 매핑합니다.
func MapPbTypeToGoType(field FieldSchema) (string, string) {
	switch field.Type {
	case "text", "email", "url", "editor":
		return "string", ""
	case "number":
		return "float64", "" // 정수/실수 구분 정보가 없으므로 float64로 통일
	case "bool":
		return "bool", ""
	case "date":
		return "types.DateTime", ""
	case "select":
		if field.Options != nil && field.Options.MaxSelect != nil && *field.Options.MaxSelect != 1 {
			return "[]string", ""
		}
		return "string", ""
	case "file", "relation":
		if field.Options != nil && field.Options.MaxSelect != nil && *field.Options.MaxSelect != 1 {
			return "[]string", "// 다중 선택 가능"
		}
		return "string", "// 단일 선택"
	case "json":
		return "json.RawMessage", ""
	default:
		// PocketBase 시스템 기본 타입들 처리
		switch field.ID {
		case "text3208210256": // id
			return "string", ""
		case "password901924565": // password
			return "string", `json:"-"` // 비밀번호는 직렬화에서 제외
		case "autodate2990389176", "autodate3332085495": // created, updated
			return "types.DateTime", ""
		}
		return "interface{}", "// 알 수 없는 PocketBase 타입"
	}
}

func ToPascalCase(s string) string {
	if s == "" {
		return ""
	}

	// 일반적인 약어 처리
	switch strings.ToLower(s) {
	case "id":
		return "ID"
	case "url":
		return "URL"
	case "html":
		return "HTML"
	case "json":
		return "JSON"
	}

	var result strings.Builder
	capitalizeNext := true

	for _, r := range s {
		if r == '_' || r == '-' || r == ' ' {
			capitalizeNext = true
			continue
		}
		if capitalizeNext {
			result.WriteRune(unicode.ToUpper(r))
			capitalizeNext = false
		} else {
			result.WriteRune(r)
		}
	}

	return result.String()
}
