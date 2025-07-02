package pocketbase

import (
	"time"

	"github.com/goccy/go-json"
)

// BaseModel provides common fields for all PocketBase models.
type BaseModel struct {
	ID      string    `json:"id"`
	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`
}

// Admin represents a PocketBase administrator.
type Admin struct {
	BaseModel
	Avatar int    `json:"avatar"`
	Email  string `json:"email"`
}

// Record represents a PocketBase record.
type Record struct {
	BaseModel
	CollectionID   string               `json:"collectionId"`
	CollectionName string               `json:"collectionName"`
	Expand         map[string][]*Record `json:"expand,omitempty"`
	Data           map[string]interface{}
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
	Page      int
	PerPage   int
	Sort      string
	Filter    string
	Expand    string
	Fields    string
	SkipTotal bool
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

// UnmarshalJSON is a custom unmarshaler for the Record struct.
func (r *Record) UnmarshalJSON(data []byte) error {
	// 1. 맵으로 한 번만 언마샬링합니다.
	var allData map[string]interface{}
	if err := json.Unmarshal(data, &allData); err != nil {
		return err
	}

	// 2. 맵에서 알려진 키를 찾아 구조체 필드에 할당합니다.
	if v, ok := allData["id"].(string); ok {
		r.ID = v
	}
	if v, ok := allData["collectionId"].(string); ok {
		r.CollectionID = v
	}
	if v, ok := allData["collectionName"].(string); ok {
		r.CollectionName = v
	}

	// 2-1. 시간 필드는 파싱이 필요합니다.
	if v, ok := allData["created"].(string); ok {
		if t, err := time.Parse(time.RFC3339Nano, v); err == nil {
			r.Created = t
		}
	}
	if v, ok := allData["updated"].(string); ok {
		if t, err := time.Parse(time.RFC3339Nano, v); err == nil {
			r.Updated = t
		}
	}
	// expand 필드 처리
	if exp, ok := allData["expand"]; ok {
		b, err := json.Marshal(exp)
		if err == nil {
			_ = json.Unmarshal(b, &r.Expand)
		}
	}

	// 3. 사용된 키를 맵에서 제거합니다.
	delete(allData, "id")
	delete(allData, "created")
	delete(allData, "updated")
	delete(allData, "collectionId")
	delete(allData, "collectionName")
	delete(allData, "expand")

	// 4. 남은 맵을 Data 필드에 할당합니다.
	r.Data = allData

	return nil
}
