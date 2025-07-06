package generator

import (
	"strings"
	"unicode"
)

// MapPbTypeToGoType는 PocketBase 필드 타입을 Go 타입으로 매핑합니다.
func MapPbTypeToGoType(field FieldSchema) string {
	switch field.Type {
	case "text", "email", "url", "editor":
		return "string"
	case "number":
		return "float64"
	case "bool":
		return "bool"
	case "date", "authdate":
		// 'autodate'는 PocketBase 구버전 스키마일 수 있으나, 최신 버전은 'date'를 사용하며
		// 생성/수정 시간은 레코드의 최상위 필드에 있습니다.
		return "types.DateTime"
	case "json":
		return "json.RawMessage"

	case "relation", "file":
		if field.Options == nil || field.Options.MaxSelect == nil || *field.Options.MaxSelect != 1 {
			return "[]string"
		}
		return "string"

	case "select":
		if field.Options != nil && field.Options.MaxSelect != nil && *field.Options.MaxSelect != 1 {
			return "[]string"
		}
		return "string"

	default:
		return "interface{}"
	}
}

// ToPascalCase는 snake_case를 PascalCase로 변환합니다.
func ToPascalCase(s string) string {
	if s == "" {
		return ""
	}
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
