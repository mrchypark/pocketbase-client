package pocketbase

import (
	"github.com/goccy/go-json" // gocc/go-json을 직접 사용하도록 수정
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
// ✅ 수정: Data 필드를 제거하고 deserializedData를 직접 사용합니다.
type Record struct {
	BaseModel
	CollectionID   string               `json:"collectionId"`
	CollectionName string               `json:"collectionName"`
	Expand         map[string][]*Record `json:"expand,omitempty"`

	// 데이터를 저장할 맵. 이제부터 이 필드가 유일한 데이터 소스입니다.
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

// ... (ListOptions, GetOneOptions 등 나머지 옵션 구조체는 동일)
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

type GetOneOptions struct {
	Expand string
	Fields string
}

type WriteOptions struct {
	Expand string
	Fields string
}

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

// ✅ 수정: UnmarshalJSON을 훨씬 단순하고 효율적으로 변경합니다.
func (r *Record) UnmarshalJSON(data []byte) error {
	// 임시 map에 모든 JSON 데이터를 한 번에 디코딩합니다.
	var allData map[string]interface{}
	if err := json.Unmarshal(data, &allData); err != nil {
		return err
	}

	// 공통 필드를 Record 구조체에 직접 할당합니다.
	if id, ok := allData["id"].(string); ok {
		r.ID = id
	}
	if created, ok := allData["created"].(string); ok {
		r.Created, _ = types.ParseDateTime(created)
	}
	if updated, ok := allData["updated"].(string); ok {
		r.Updated, _ = types.ParseDateTime(updated)
	}
	if colID, ok := allData["collectionId"].(string); ok {
		r.CollectionID = colID
	}
	if colName, ok := allData["collectionName"].(string); ok {
		r.CollectionName = colName
	}
	// Expand 필드도 처리합니다.
	if expandData, ok := allData["expand"]; ok {
		// expand 데이터를 다시 JSON으로 직렬화한 후 Record의 Expand 필드로 디코딩합니다.
		// 이는 expand가 복잡한 중첩 구조를 가질 수 있기 때문에 가장 안전한 방법입니다.
		expandBytes, err := json.Marshal(expandData)
		if err == nil {
			json.Unmarshal(expandBytes, &r.Expand)
		}
	}

	// 공통 필드와 expand를 map에서 제거합니다.
	delete(allData, "id")
	delete(allData, "created")
	delete(allData, "updated")
	delete(allData, "collectionId")
	delete(allData, "collectionName")
	delete(allData, "expand")

	// 나머지 데이터는 deserializedData에 저장합니다.
	r.deserializedData = allData

	return nil
}

// Get returns a raw interface{} value for a given key.
func (r *Record) Get(key string) interface{} {
	if r.deserializedData == nil {
		r.deserializedData = make(map[string]interface{})
	}
	return r.deserializedData[key]
}

// Set stores a key-value pair in the record's data.
func (r *Record) Set(key string, value interface{}) {
	if r.deserializedData == nil {
		r.deserializedData = make(map[string]interface{})
	}
	r.deserializedData[key] = value
}

// ✅ 수정: MarshalJSON도 deserializedData를 직접 사용하도록 변경합니다.
func (r *Record) MarshalJSON() ([]byte, error) {
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
	if slice, ok := val.([]string); ok {
		return slice
	}
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
// Returns nil if the key is not present or the value is not a bool.
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
// Returns nil if the key is not present or the value is not a number.
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
// Returns nil if the key is not present or the value cannot be parsed as a DateTime.
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
