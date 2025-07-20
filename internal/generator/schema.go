package generator

import (
	"encoding/json"
	"fmt" // Required for error message formatting
)

// SchemaVersionError represents errors that occur during schema version detection
type SchemaVersionError struct {
	Message string
	Cause   error
	Data    []byte
}

func (e *SchemaVersionError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("schema version detection failed: %s (cause: %v)", e.Message, e.Cause)
	}
	return fmt.Sprintf("schema version detection failed: %s", e.Message)
}

func (e *SchemaVersionError) Unwrap() error {
	return e.Cause
}

// 에러 변수들
var (
	ErrUnknownSchemaVersion = &SchemaVersionError{Message: "unknown schema version"}
	ErrInvalidSchemaFormat  = &SchemaVersionError{Message: "invalid schema format"}
	ErrMixedSchemaVersions  = &SchemaVersionError{Message: "mixed schema versions detected"}
)

// SchemaVersionDetector provides functionality to detect PocketBase schema versions
type SchemaVersionDetector struct{}

// NewSchemaVersionDetector creates a new SchemaVersionDetector instance
func NewSchemaVersionDetector() *SchemaVersionDetector {
	return &SchemaVersionDetector{}
}

// DetectVersion detects the schema version from raw schema data
func (d *SchemaVersionDetector) DetectVersion(schemaData []byte) (SchemaVersion, error) {
	if len(schemaData) == 0 {
		return SchemaVersionUnknown, &SchemaVersionError{
			Message: "empty schema data",
			Data:    schemaData,
		}
	}

	var collections []map[string]interface{}
	if err := json.Unmarshal(schemaData, &collections); err != nil {
		return SchemaVersionUnknown, &SchemaVersionError{
			Message: "failed to parse schema JSON",
			Cause:   err,
			Data:    schemaData,
		}
	}

	if len(collections) == 0 {
		return SchemaVersionUnknown, &SchemaVersionError{
			Message: "no collections found in schema",
			Data:    schemaData,
		}
	}

	return d.detectSchemaVersion(collections)
}

// detectSchemaVersion performs the actual version detection logic
func (d *SchemaVersionDetector) detectSchemaVersion(collections []map[string]interface{}) (SchemaVersion, error) {
	var detectedVersion SchemaVersion
	var hasFields, hasSchema bool

	for i, collection := range collections {
		_, fieldsExists := collection["fields"]
		_, schemaExists := collection["schema"]

		if fieldsExists && schemaExists {
			return SchemaVersionUnknown, &SchemaVersionError{
				Message: fmt.Sprintf("collection at index %d has both 'fields' and 'schema' keys", i),
			}
		}

		if !fieldsExists && !schemaExists {
			return SchemaVersionUnknown, &SchemaVersionError{
				Message: fmt.Sprintf("collection at index %d has neither 'fields' nor 'schema' key", i),
			}
		}

		var currentVersion SchemaVersion
		if fieldsExists {
			currentVersion = SchemaVersionLatest
			hasFields = true
		} else {
			currentVersion = SchemaVersionLegacy
			hasSchema = true
		}

		// 첫 번째 컬렉션에서 버전 설정
		if i == 0 {
			detectedVersion = currentVersion
		} else if detectedVersion != currentVersion {
			// 혼합된 버전이 감지된 경우
			return SchemaVersionUnknown, &SchemaVersionError{
				Message: fmt.Sprintf("mixed schema versions detected: collection at index %d uses %s format while previous collections use %s format",
					i, currentVersion.String(), detectedVersion.String()),
			}
		}
	}

	// 추가 검증: 모든 컬렉션이 동일한 버전을 사용하는지 확인
	if hasFields && hasSchema {
		return SchemaVersionUnknown, ErrMixedSchemaVersions
	}

	if detectedVersion == SchemaVersionUnknown {
		return SchemaVersionUnknown, ErrUnknownSchemaVersion
	}

	return detectedVersion, nil
}

// ValidateSchema validates the schema format for the given version
func (d *SchemaVersionDetector) ValidateSchema(schemaData []byte, expectedVersion SchemaVersion) error {
	detectedVersion, err := d.DetectVersion(schemaData)
	if err != nil {
		return err
	}

	if detectedVersion != expectedVersion {
		return &SchemaVersionError{
			Message: fmt.Sprintf("schema version mismatch: expected %s, got %s",
				expectedVersion.String(), detectedVersion.String()),
			Data: schemaData,
		}
	}

	return nil
}

// CollectionSchema represents a PocketBase collection schema with all its metadata.
// It contains information about the collection structure, rules, and field definitions.
type CollectionSchema struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Type       string   `json:"type"`
	System     bool     `json:"system"`
	Indexes    []string `json:"indexes"`
	ListRule   *string  `json:"listRule"`
	ViewRule   *string  `json:"viewRule"`
	CreateRule *string  `json:"createRule"`
	UpdateRule *string  `json:"updateRule"`
	DeleteRule *string  `json:"deleteRule"`
	Options    *struct {
		Query *string `json:"query"`
	} `json:"options"`

	Schema []FieldSchema `json:"schema"` // Keep `json:"schema"` tag
	Fields []FieldSchema `json:"fields"` // Keep `json:"fields"` tag
}

// UnmarshalJSON unmarshals the 'schema' or 'fields' array of CollectionSchema to cs.Fields.
// This method handles both legacy 'schema' field and newer 'fields' field formats.
func (cs *CollectionSchema) UnmarshalJSON(data []byte) error {
	type Alias CollectionSchema
	aux := &struct {
		SchemaRaw json.RawMessage `json:"schema"`
		FieldsRaw json.RawMessage `json:"fields"`
		*Alias
	}{
		Alias: (*Alias)(cs),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if len(aux.SchemaRaw) > 0 && string(aux.SchemaRaw) != "null" {
		if err := json.Unmarshal(aux.SchemaRaw, &cs.Fields); err != nil {
			return err
		}
	} else if len(aux.FieldsRaw) > 0 && string(aux.FieldsRaw) != "null" {
		if err := json.Unmarshal(aux.FieldsRaw, &cs.Fields); err != nil {
			return err
		}
	}

	cs.Schema = nil // Clear Schema field since Fields field is populated with data.

	return nil
}

// FieldSchema represents a single field definition within a PocketBase collection.
// It contains all metadata about the field including type, validation rules, and options.
type FieldSchema struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	System      bool   `json:"system"`
	Required    bool   `json:"required"`
	Presentable bool   `json:"presentable"`
	Unique      bool   `json:"unique"`
	Hidden      bool   `json:"hidden"`

	// options field is kept as FieldOptions type (pointer)
	Options *FieldOptions `json:"options"`

	// RawMessage fields for maxSelect, minSelect directly at field level (for temporary parsing)
	// These fields are only used inside FieldSchema.UnmarshalJSON.
	MinSelectRaw json.RawMessage `json:"minSelect"` // May be at field level depending on schema
	MaxSelectRaw json.RawMessage `json:"maxSelect"` // May be at field level depending on schema
}

// UnmarshalJSON is custom unmarshaling logic for FieldSchema.
// This method handles both minSelect/maxSelect inside 'options' object or directly at field level.
func (fs *FieldSchema) UnmarshalJSON(data []byte) error {
	// Use temporary struct to unmarshal basic fields and raw 'options' data to avoid infinite recursion.
	type Alias FieldSchema // Alias that includes all other fields of FieldSchema
	aux := &struct {
		OptionsRaw json.RawMessage `json:"options"` // Capture raw 'options' object as json.RawMessage
		*Alias
	}{
		Alias: (*Alias)(fs), // Automatically bind other fields to fs instance
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Initialize fs.Options
	fs.Options = &FieldOptions{}

	// If aux.OptionsRaw (raw 'options' object) exists, unmarshal it first.
	if len(aux.OptionsRaw) > 0 && string(aux.OptionsRaw) != "null" {
		if err := json.Unmarshal(aux.OptionsRaw, fs.Options); err != nil {
			return fmt.Errorf("failed to unmarshal FieldOptions from raw options: %w", err)
		}
	}

	// Now check MaxSelectRaw/MinSelectRaw directly at field level,
	// and overwrite only if values are not already set in fs.Options (or give priority).
	// In PocketBase schema, values inside `options` may be more explicit,
	// so prioritize values inside `options` and use field-level values as auxiliary.
	// Here we will unconditionally overwrite with field-level values.
	if len(aux.MinSelectRaw) > 0 && string(aux.MinSelectRaw) != "null" {
		var val int
		if err := json.Unmarshal(aux.MinSelectRaw, &val); err == nil {
			fs.Options.MinSelect = &val // Prioritize field-level value
		}
	}
	if len(aux.MaxSelectRaw) > 0 && string(aux.MaxSelectRaw) != "null" {
		var val int
		if err := json.Unmarshal(aux.MaxSelectRaw, &val); err == nil {
			fs.Options.MaxSelect = &val // Prioritize field-level value
		}
	}

	return nil
}

// FieldOptions represents the options/configuration for a PocketBase field.
// It contains validation rules, constraints, and field-specific settings.
type FieldOptions struct {
	CollectionID  string `json:"collectionId"`
	CascadeDelete bool   `json:"cascadeDelete"`

	Min json.RawMessage `json:"min"`
	Max json.RawMessage `json:"max"`

	MinSelect *int `json:"minSelect"` // Keep as *int, handled in FieldSchema.UnmarshalJSON
	MaxSelect *int `json:"maxSelect"` // Keep as *int, handled in FieldSchema.UnmarshalJSON

	Pattern string `json:"pattern"`

	MimeTypes []string `json:"mimeTypes"`
	Thumbs    []string `json:"thumbs"`
	MaxSize   int      `json:"maxSize"`
	Protected bool     `json:"protected"`
	Values    []string `json:"values"`
}

// UnmarshalJSON method for FieldOptions is no longer needed (deleted).
// FieldSchema.UnmarshalJSON handles all parsing.
