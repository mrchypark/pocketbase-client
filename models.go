package pocketbase

import (
	"bytes"
	"sync"

	"github.com/goccy/go-json"
	"github.com/pocketbase/pocketbase/tools/types"
)

// BaseModel provides common fields for all PocketBase models.
type BaseModel struct {
	ID      string         `json:"id"`
	Created types.DateTime `json:"created"`
	Updated types.DateTime `json:"updated"`
}

// Admin represents a PocketBase administrator.
type Admin struct {
	BaseModel
	Avatar int    `json:"avatar"`
	Email  string `json:"email"`
}

// Record represents a PocketBase record.
// It uses json.RawMessage for the Data field to delay parsing, improving performance.
type Record struct {
	BaseModel
	CollectionID   string               `json:"collectionId"`
	CollectionName string               `json:"collectionName"`
	Expand         map[string][]*Record `json:"expand,omitempty"`

	Data json.RawMessage `json:"-"`

	once             sync.Once
	deserializedData map[string]interface{}
}

// ListResult is a struct containing a list of records along with pagination information.
type ListResult struct {
	Page       int       `json:"page"`
	PerPage    int       `json:"perPage"`
	TotalItems int       `json:"totalItems"`
	TotalPages int       `json:"totalPages"`
	Items      []*Record `json:"items"`
}

// ListOptions defines query parameters for listing records.
type ListOptions struct {
	Page        int
	PerPage     int
	Sort        string
	Filter      string
	Expand      string
	Fields      string
	SkipTotal   bool
	QueryParams map[string]string
}

// GetOneOptions defines query parameters for retrieving a single record.
type GetOneOptions struct {
	Expand string
	Fields string
}

// WriteOptions defines query parameters for creating/updating records.
type WriteOptions struct {
	Expand string
	Fields string
}

// FileDownloadOptions defines query parameters for downloading files.
type FileDownloadOptions struct {
	Thumb    string
	Download bool
}

// AuthResponse represents the data returned after successful authentication.
type AuthResponse struct {
	Token  string  `json:"token"`
	Record *Record `json:"record,omitempty"`
	Admin  *Admin  `json:"admin,omitempty"`
}

// RealtimeEvent is an event delivered via real-time subscription.
type RealtimeEvent struct {
	Action string  `json:"action"`
	Record *Record `json:"record"`
}

// UnmarshalJSON implements a custom unmarshaler for the Record struct.
func (r *Record) UnmarshalJSON(data []byte) error {
	type RecordAlias Record
	alias := &struct {
		*RecordAlias
		RawData map[string]json.RawMessage `json:"-"`
	}{
		RecordAlias: (*RecordAlias)(r),
		RawData:     make(map[string]json.RawMessage),
	}

	decoder := json.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&alias.RawData); err != nil {
		return err
	}

	if rawID, ok := alias.RawData["id"]; ok {
		_ = json.Unmarshal(rawID, &r.ID)
		delete(alias.RawData, "id")
	}
	if rawCreated, ok := alias.RawData["created"]; ok {
		_ = json.Unmarshal(rawCreated, &r.Created)
		delete(alias.RawData, "created")
	}
	if rawUpdated, ok := alias.RawData["updated"]; ok {
		_ = json.Unmarshal(rawUpdated, &r.Updated)
		delete(alias.RawData, "updated")
	}
	if rawColID, ok := alias.RawData["collectionId"]; ok {
		_ = json.Unmarshal(rawColID, &r.CollectionID)
		delete(alias.RawData, "collectionId")
	}
	if rawColName, ok := alias.RawData["collectionName"]; ok {
		_ = json.Unmarshal(rawColName, &r.CollectionName)
		delete(alias.RawData, "collectionName")
	}
	if rawExpand, ok := alias.RawData["expand"]; ok {
		_ = json.Unmarshal(rawExpand, &r.Expand)
		delete(alias.RawData, "expand")
	}

	remainingData, err := json.Marshal(alias.RawData)
	if err != nil {
		return err
	}
	r.Data = remainingData
	return nil
}

func (r *Record) parseData() {
	r.once.Do(func() {
		if len(r.Data) == 0 {
			r.deserializedData = make(map[string]interface{})
			return
		}
		if err := json.Unmarshal(r.Data, &r.deserializedData); err != nil {
			r.deserializedData = make(map[string]interface{})
		}
	})
}

// Get returns a raw interface{} value for a given key.
func (r *Record) Get(key string) interface{} {
	r.parseData()
	return r.deserializedData[key]
}

// GetString returns a string value for a given key.
func (r *Record) GetString(key string) string {
	val := r.Get(key)
	if str, ok := val.(string); ok {
		return str
	}
	return ""
}

// GetBool returns a boolean value for a given key.
func (r *Record) GetBool(key string) bool {
	val := r.Get(key)
	if b, ok := val.(bool); ok {
		return b
	}
	return false
}

// GetFloat returns a float64 value for a given key.
func (r *Record) GetFloat(key string) float64 {
	val := r.Get(key)
	if f, ok := val.(float64); ok {
		return f
	}
	if i, ok := val.(int); ok {
		return float64(i)
	}
	// For JSON numbers
	if num, ok := val.(json.Number); ok {
		f, _ := num.Float64()
		return f
	}
	return 0
}

// GetDateTime returns a types.DateTime value for a given key.
func (r *Record) GetDateTime(key string) types.DateTime {
	val := r.Get(key)
	if dt, ok := val.(types.DateTime); ok {
		return dt
	}
	if str, ok := val.(string); ok {
		dt, err := types.ParseDateTime(str)
		if err == nil {
			return dt
		}
	}
	return types.DateTime{}
}

// GetStringSlice returns a slice of strings for a given key.
func (r *Record) GetStringSlice(key string) []string {
	val := r.Get(key)
	if slice, ok := val.([]interface{}); ok {
		result := make([]string, len(slice))
		for i, v := range slice {
			if str, ok := v.(string); ok {
				result[i] = str
			}
		}
		return result
	}
	return []string{}
}

// GetRawMessage returns a json.RawMessage value for a given key.
func (r *Record) GetRawMessage(key string) json.RawMessage {
	val := r.Get(key)
	if raw, ok := val.(json.RawMessage); ok {
		return raw
	}
	// If it was parsed into a map/slice, re-marshal it.
	// This can happen if Set() was called before.
	if val != nil {
		bytes, err := json.Marshal(val)
		if err == nil {
			return bytes
		}
	}
	return nil
}

// GetStringPointer returns a pointer to a string value for a given key.
// Returns nil if the key is not present or the value is not a string.
func (r *Record) GetStringPointer(key string) *string {
	val := r.Get(key)
	if str, ok := val.(string); ok {
		return &str
	}
	return nil
}

// GetBoolPointer returns a pointer to a boolean value for a given key.
// Returns nil if the key is not present or the value is not a bool.
func (r *Record) GetBoolPointer(key string) *bool {
	val := r.Get(key)
	if b, ok := val.(bool); ok {
		return &b
	}
	return nil
}

// GetFloatPointer returns a pointer to a float64 value for a given key.
// Returns nil if the key is not present or the value is not a number.
func (r *Record) GetFloatPointer(key string) *float64 {
	val := r.Get(key)
	var f float64
	var ok bool

	if num, isNum := val.(json.Number); isNum {
		f, err := num.Float64()
		if err == nil {
			return &f
		}
	} else if f, ok = val.(float64); ok {
		return &f
	}
	return nil
}

// GetDateTimePointer returns a pointer to a types.DateTime value for a given key.
// Returns nil if the key is not present or the value cannot be parsed as a DateTime.
func (r *Record) GetDateTimePointer(key string) *types.DateTime {
	val := r.Get(key)
	if str, ok := val.(string); ok {
		dt, err := types.ParseDateTime(str)
		if err == nil {
			return &dt
		}
	}
	return nil
}

// Set stores a key-value pair in the record's data.
func (r *Record) Set(key string, value interface{}) {
	r.parseData()
	r.deserializedData[key] = value

	newData, err := json.Marshal(r.deserializedData)
	if err == nil {
		r.Data = newData
	}
}

// MarshalJSON implements a custom marshaler for the Record struct.
func (r *Record) MarshalJSON() ([]byte, error) {
	r.parseData()
	combinedData := make(map[string]interface{}, len(r.deserializedData)+6)
	for k, v := range r.deserializedData {
		combinedData[k] = v
	}

	combinedData["id"] = r.ID
	combinedData["collectionId"] = r.CollectionID
	combinedData["collectionName"] = r.CollectionName
	combinedData["created"] = r.Created
	combinedData["updated"] = r.Updated
	if r.Expand != nil {
		combinedData["expand"] = r.Expand
	}

	return json.Marshal(combinedData)
}
