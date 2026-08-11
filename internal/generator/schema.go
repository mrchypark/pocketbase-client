package generator

import (
	"encoding/json"
)

// CollectionSchema represents a PocketBase collection schema with all its metadata.
// It contains information about the collection structure, rules, and field definitions.
//
// The schema follows the PocketBase v0.23+ export format: the collection's fields
// are exposed under the "fields" key and type-specific collection options (e.g.
// "viewQuery" for views, auth options for auth collections) are flattened onto
// the collection itself rather than nested under an "options" key.
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

	// ViewQuery is the SQL query of "view" type collections (flattened "viewQuery" key).
	ViewQuery string `json:"viewQuery"`

	Fields []FieldSchema `json:"fields"`
}

// FieldSchema represents a single field definition within a PocketBase collection.
// It contains all metadata about the field including type, validation rules, and options.
//
// In the latest schema format (v0.23+) field options are flattened onto the field
// itself rather than nested under an "options" object. Custom unmarshaling captures
// those option keys into the Options field.
type FieldSchema struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	System      bool   `json:"system"`
	Required    bool   `json:"required"`
	Presentable bool   `json:"presentable"`
	Hidden      bool   `json:"hidden"`

	// Options holds the flattened field options. It is excluded from direct JSON
	// decoding (json:"-") and populated by UnmarshalJSON from the field-level
	// option keys.
	Options *FieldOptions `json:"-"`
}

// UnmarshalJSON decodes a field's metadata and its flattened option keys.
//
// FieldOptions only declares struct tags for option keys (maxSelect, values,
// collectionId, thumbs, ...), so decoding the raw field JSON into it captures
// the options while naturally ignoring the field metadata keys (id, name, type,
// system, ...).
func (fs *FieldSchema) UnmarshalJSON(data []byte) error {
	type Alias FieldSchema
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(fs),
	}

	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	opts := FieldOptions{}
	if err := json.Unmarshal(data, &opts); err != nil {
		return err
	}
	fs.Options = &opts

	return nil
}

// FieldOptions represents the options/configuration for a PocketBase field.
// It contains validation rules, constraints, and field-specific settings.
//
// The JSON tags mirror the flattened field-level keys of the latest schema format.
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

	// Help is the field description shown in the dashboard (added in v0.39.0).
	Help string `json:"help"`

	// Email/URL options
	ExceptDomains []string `json:"exceptDomains"`
	OnlyDomains   []string `json:"onlyDomains"`

	// Editor options
	ConvertURLs *bool `json:"convertUrls"`

	// Password options
	Cost *int `json:"cost"`
}
