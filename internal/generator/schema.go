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

	// RawMessage fields for flattened options (v0.23+ format)
	// These fields can appear at field level or inside options
	MinSelectRaw           json.RawMessage `json:"minSelect"`
	MaxSelectRaw           json.RawMessage `json:"maxSelect"`
	AutogeneratePatternRaw json.RawMessage `json:"autogeneratePattern"`
	OnlyIntRaw             json.RawMessage `json:"onlyInt"`
	OnCreateRaw            json.RawMessage `json:"onCreate"`
	OnUpdateRaw            json.RawMessage `json:"onUpdate"`
	ExceptDomainsRaw       json.RawMessage `json:"exceptDomains"`
	OnlyDomainsRaw         json.RawMessage `json:"onlyDomains"`
	ConvertURLsRaw         json.RawMessage `json:"convertUrls"`
	CostRaw                json.RawMessage `json:"cost"`
	PrimaryKeyRaw          json.RawMessage `json:"primaryKey"`
}

// UnmarshalJSON is custom unmarshaling logic for FieldSchema.
// This method handles both legacy (options nested) and latest (flattened) formats.
func (fs *FieldSchema) UnmarshalJSON(data []byte) error {
	type Alias FieldSchema
	aux := &struct {
		OptionsRaw json.RawMessage `json:"options"`
		*Alias
	}{
		Alias: (*Alias)(fs),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Initialize fs.Options if nil
	if fs.Options == nil {
		fs.Options = &FieldOptions{}
	}

	// Unmarshal options object if present (legacy format)
	if len(aux.OptionsRaw) > 0 && string(aux.OptionsRaw) != "null" {
		if err := json.Unmarshal(aux.OptionsRaw, fs.Options); err != nil {
			return fmt.Errorf("failed to unmarshal FieldOptions: %w", err)
		}
	}

	// Handle flattened options (v0.23+ format) - field-level values override options
	// These fields can appear either at field level or inside options

	// minSelect / maxSelect
	if len(aux.MinSelectRaw) > 0 && string(aux.MinSelectRaw) != "null" {
		var val int
		if err := json.Unmarshal(aux.MinSelectRaw, &val); err == nil {
			fs.Options.MinSelect = &val
		}
	}
	if len(aux.MaxSelectRaw) > 0 && string(aux.MaxSelectRaw) != "null" {
		var val int
		if err := json.Unmarshal(aux.MaxSelectRaw, &val); err == nil {
			fs.Options.MaxSelect = &val
		}
	}

	// autogeneratePattern (text)
	if len(aux.AutogeneratePatternRaw) > 0 && string(aux.AutogeneratePatternRaw) != "null" {
		var val string
		if err := json.Unmarshal(aux.AutogeneratePatternRaw, &val); err == nil {
			fs.Options.AutogeneratePattern = &val
		}
	}

	// onlyInt (number)
	if len(aux.OnlyIntRaw) > 0 && string(aux.OnlyIntRaw) != "null" {
		var val bool
		if err := json.Unmarshal(aux.OnlyIntRaw, &val); err == nil {
			fs.Options.OnlyInt = &val
		}
	}

	// onCreate / onUpdate (autodate)
	if len(aux.OnCreateRaw) > 0 && string(aux.OnCreateRaw) != "null" {
		var val bool
		if err := json.Unmarshal(aux.OnCreateRaw, &val); err == nil {
			fs.Options.OnCreate = &val
		}
	}
	if len(aux.OnUpdateRaw) > 0 && string(aux.OnUpdateRaw) != "null" {
		var val bool
		if err := json.Unmarshal(aux.OnUpdateRaw, &val); err == nil {
			fs.Options.OnUpdate = &val
		}
	}

	// exceptDomains / onlyDomains (email, url)
	if len(aux.ExceptDomainsRaw) > 0 && string(aux.ExceptDomainsRaw) != "null" {
		var val []string
		if err := json.Unmarshal(aux.ExceptDomainsRaw, &val); err == nil {
			fs.Options.ExceptDomains = val
		}
	}
	if len(aux.OnlyDomainsRaw) > 0 && string(aux.OnlyDomainsRaw) != "null" {
		var val []string
		if err := json.Unmarshal(aux.OnlyDomainsRaw, &val); err == nil {
			fs.Options.OnlyDomains = val
		}
	}

	// convertURLs (editor)
	if len(aux.ConvertURLsRaw) > 0 && string(aux.ConvertURLsRaw) != "null" {
		var val bool
		if err := json.Unmarshal(aux.ConvertURLsRaw, &val); err == nil {
			fs.Options.ConvertURLs = &val
		}
	}

	// cost (password)
	if len(aux.CostRaw) > 0 && string(aux.CostRaw) != "null" {
		var val int
		if err := json.Unmarshal(aux.CostRaw, &val); err == nil {
			fs.Options.Cost = &val
		}
	}

	// primaryKey (text)
	if len(aux.PrimaryKeyRaw) > 0 && string(aux.PrimaryKeyRaw) != "null" {
		var val bool
		if err := json.Unmarshal(aux.PrimaryKeyRaw, &val); err == nil {
			fs.Options.PrimaryKey = &val
		}
	}

	return nil
}

// FieldOptions represents the options/configuration for a PocketBase field.
// It contains validation rules, constraints, and field-specific settings.
type FieldOptions struct {
	// Relation options
	CollectionID  string `json:"collectionId"`
	CascadeDelete bool   `json:"cascadeDelete"`

	// Number options
	OnlyInt *bool `json:"onlyInt"`

	// Text options
	AutogeneratePattern *string `json:"autogeneratePattern"`
	PrimaryKey          *bool   `json:"primaryKey"`

	// Date options
	OnCreate *bool `json:"onCreate"`
	OnUpdate *bool `json:"onUpdate"`

	// Select/Relation/File options
	MinSelect *int     `json:"minSelect"`
	MaxSelect *int     `json:"maxSelect"`
	Values    []string `json:"values"`

	// File options
	MimeTypes []string `json:"mimeTypes"`
	Thumbs    []string `json:"thumbs"`
	MaxSize   int      `json:"maxSize"`
	Protected bool     `json:"protected"`

	// String validation options
	Pattern string `json:"pattern"`

	// Email/URL options
	ExceptDomains []string `json:"exceptDomains"`
	OnlyDomains   []string `json:"onlyDomains"`

	// Editor options
	ConvertURLs *bool `json:"convertUrls"`

	// Password options
	Cost *int `json:"cost"`
}
