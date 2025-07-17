package generator

import (
	"encoding/json"
	"fmt" // Required for error message formatting
)

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
