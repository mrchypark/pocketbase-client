package pocketbase

import (
	"maps"
	"net/url"

	"github.com/goccy/go-json"
	"github.com/pocketbase/pocketbase/tools/types"
)

// BaseModel is the interface that all PocketBase models must implement.
// It provides common accessors for shared fields.
type BaseModel interface {
	GetID() string
	GetCollectionName() string
}

// BaseDatetime provides common datetime fields for PocketBase models.
type BaseDatetime struct {
	Created types.DateTime `json:"created"`
	Updated types.DateTime `json:"updated"`
}

// Admin represents a PocketBase administrator.
type Admin struct {
	ID             string `json:"id"`
	CollectionID   string `json:"collectionId"`
	CollectionName string `json:"collectionName"`
	Avatar         int    `json:"avatar"`
	Email          string `json:"email"`
}

func (a *Admin) GetID() string             { return a.ID }
func (a *Admin) GetCollectionName() string { return "_superusers" }

// Record represents a PocketBase record with type-safe accessors.
type Record struct {
	ID             string               `json:"id"`
	CollectionID   string               `json:"collectionId"`
	CollectionName string               `json:"collectionName"`
	Expand         map[string][]*Record `json:"expand,omitempty"`

	// Map to store data. From now on, this field is the only data source.
	deserializedData map[string]any
}

func (r *Record) GetID() string             { return r.ID }
func (r *Record) GetCollectionName() string { return r.CollectionName }

// ListResult is a struct containing a list of records along with pagination information.
type ListResult struct {
	Page       int       `json:"page"`
	PerPage    int       `json:"perPage"`
	TotalItems int       `json:"totalItems"`
	TotalPages int       `json:"totalPages"`
	Items      []*Record `json:"items"`
}

// ListOptions contains options for listing records.
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

// GetOneOptions contains options for retrieving a single record.
type GetOneOptions struct {
	Expand string
	Fields string
}

// WriteOptions contains options for create and update operations.
type WriteOptions struct {
	Expand string
	Fields string
}

func (o *WriteOptions) apply(q url.Values) {
	if o == nil {
		return
	}
	if o.Expand != "" {
		q.Set("expand", o.Expand)
	}
	if o.Fields != "" {
		q.Set("fields", o.Fields)
	}
}

// FileDownloadOptions contains options for file download operations.
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

// UnmarshalJSON deserializes JSON data into the Record struct efficiently.
func (r *Record) UnmarshalJSON(data []byte) error {
	// Decode all JSON data into raw fields to avoid re-marshaling.
	var allData map[string]json.RawMessage
	if err := json.Unmarshal(data, &allData); err != nil {
		return err
	}

	// Assign common fields directly to Record struct.
	if raw, ok := allData["id"]; ok {
		_ = json.Unmarshal(raw, &r.ID)
	}
	if raw, ok := allData["collectionId"]; ok {
		_ = json.Unmarshal(raw, &r.CollectionID)
	}
	if raw, ok := allData["collectionName"]; ok {
		_ = json.Unmarshal(raw, &r.CollectionName)
	}
	// Also handle Expand field.
	if raw, ok := allData["expand"]; ok {
		_ = json.Unmarshal(raw, &r.Expand)
	}

	delete(allData, "id")
	delete(allData, "collectionId")
	delete(allData, "collectionName")
	delete(allData, "expand")

	// Store remaining data in deserializedData.
	r.deserializedData = make(map[string]any, len(allData))
	for key, raw := range allData {
		var value any
		if err := json.Unmarshal(raw, &value); err != nil {
			continue
		}
		r.deserializedData[key] = value
	}

	return nil
}

// Get returns a raw any value for a given key.
func (r *Record) Get(key string) any {
	if r.deserializedData == nil {
		r.deserializedData = make(map[string]any)
	}
	return r.deserializedData[key]
}

// Set stores a key-value pair in the record's data.
func (r *Record) Set(key string, value any) {
	if r.deserializedData == nil {
		r.deserializedData = make(map[string]any)
	}
	r.deserializedData[key] = value
}

// MarshalJSON serializes the Record to JSON using deserializedData directly.
func (r *Record) MarshalJSON() ([]byte, error) {
	combinedData := make(map[string]any, len(r.deserializedData)+6)
	maps.Copy(combinedData, r.deserializedData)

	combinedData["id"] = r.ID
	combinedData["collectionId"] = r.CollectionID
	combinedData["collectionName"] = r.CollectionName
	if r.Expand != nil {
		combinedData["expand"] = r.Expand
	}

	return json.Marshal(combinedData)
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
	if slice, ok := val.([]string); ok {
		return slice
	}
	if slice, ok := val.([]any); ok {
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
	if val != nil {
		bytes, err := json.Marshal(val)
		if err == nil {
			return bytes
		}
	}
	return nil
}

// GetStringPointer returns a pointer to a string value for a given key.
func (r *Record) GetStringPointer(key string) *string {
	val := r.Get(key)
	if val == nil {
		return nil
	}
	if ptr, ok := val.(*string); ok {
		return ptr
	}
	if str, ok := val.(string); ok {
		return &str
	}
	return nil
}

// GetBoolPointer returns a pointer to a boolean value for a given key.
func (r *Record) GetBoolPointer(key string) *bool {
	val := r.Get(key)
	if val == nil {
		return nil
	}
	if ptr, ok := val.(*bool); ok {
		return ptr
	}
	if b, ok := val.(bool); ok {
		return &b
	}
	return nil
}

// GetFloatPointer returns a pointer to a float64 value for a given key.
func (r *Record) GetFloatPointer(key string) *float64 {
	val := r.Get(key)
	if val == nil {
		return nil
	}
	if ptr, ok := val.(*float64); ok {
		return ptr
	}

	var f float64
	var ok bool

	if num, isNum := val.(json.Number); isNum {
		f, err := num.Float64()
		if err == nil {
			return &f
		}
	} else if f, ok = val.(float64); ok {
		return &f
	} else if i, ok := val.(int); ok {
		f = float64(i)
		return &f
	} else if i64, ok := val.(int64); ok {
		f = float64(i64)
		return &f
	}

	return nil
}

// GetDateTimePointer returns a pointer to a types.DateTime value for a given key.
func (r *Record) GetDateTimePointer(key string) *types.DateTime {
	val := r.Get(key)
	if val == nil {
		return nil
	}
	if ptr, ok := val.(*types.DateTime); ok {
		return ptr
	}
	if dt, ok := val.(types.DateTime); ok {
		return &dt
	}
	if str, ok := val.(string); ok {
		dt, err := types.ParseDateTime(str)
		if err == nil {
			return &dt
		}
	}
	return nil
}
