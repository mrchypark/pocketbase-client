package generator

import (
	"strings"
	"unicode"
)

// MapPbTypeToGoType maps PocketBase field types to Go types and their corresponding getter methods.
func MapPbTypeToGoType(field FieldSchema, omitEmpty bool) (string, string, string) {
	var goType, comment, getterMethod string

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
		getterMethod = "GetRawMessage" // 범용 Get 사용 후 사용자가 직접 Unmarshal
	case "relation", "file":
		if field.Options != nil && field.Options.MaxSelect != nil && *field.Options.MaxSelect == 1 {
			goType = "string"
			getterMethod = "GetString"
			if omitEmpty {
				getterMethod = "GetStringPointer"
			}
		} else {
			goType = "[]string"
			getterMethod = "GetStringSlice"
		}
	case "select":
		if field.Options != nil && field.Options.MaxSelect != nil && *field.Options.MaxSelect == 1 {
			goType = "string"
			getterMethod = "GetString"
			if omitEmpty {
				getterMethod = "GetStringPointer"
			}
		} else {
			goType = "[]string"
			getterMethod = "GetStringSlice"
		}
	default:
		goType = "interface{}"
		getterMethod = "Get"
	}

	if omitEmpty && !strings.HasPrefix(goType, "[]") && goType != "json.RawMessage" && goType != "interface{}" {
		goType = "*" + goType
	}

	return goType, comment, getterMethod
}

// ... ToPascalCase 함수는 동일 ...
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
