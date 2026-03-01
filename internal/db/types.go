package db

import (
	"encoding/json"
	"time"
)

type DBEvent struct {
	ID           int64
	RequestID    int
	EventType    string
	SourceTable  string
	SourceAction string
	Payload      json.RawMessage
	CreatedAt    time.Time
}

type RequestSnapshot struct {
	Request Request       `json:"request"`
	Items   []RequestItem `json:"items"`
	Files   []RequestFile `json:"files"`
	Logs    []RequestLog  `json:"logs"`
}

type Request struct {
	ID             int        `json:"id"`
	ObjectLevelsID string     `json:"object_levels_id"`
	Name           *string    `json:"name,omitempty"`
	CreatedBy      string     `json:"created_by"`
	Executor       *string    `json:"executor,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	StartedAt      *time.Time `json:"started_at,omitempty"`
	ApprovedAt     *time.Time `json:"approved_at,omitempty"`
	RejectedAt     *time.Time `json:"rejected_at,omitempty"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
	Deadline       *time.Time `json:"deadline,omitempty"`
	Comment        *string    `json:"comment,omitempty"`
	StatusID       string     `json:"status_id"`
}

type RequestItem struct {
	ID                  string  `json:"id"`
	RequestID           int     `json:"request_id"`
	Num                 int     `json:"num"`
	NomenclatureID      *string `json:"nomenclature_id,omitempty"`
	Name                *string `json:"name,omitempty"`
	UnitID              *string `json:"unit_id,omitempty"`
	Quantity            float64 `json:"quantity"`
	WarehouseCategoryID *string `json:"warehouse_category_id,omitempty"`
	Comment             *string `json:"comment,omitempty"`
}

type RequestFile struct {
	ID          string     `json:"id"`
	RequestID   int        `json:"request_id"`
	FileID      string     `json:"file_id"`
	LinkType    string     `json:"link_type"`
	Description *string    `json:"description,omitempty"`
	IsMain      bool       `json:"is_main"`
	SortOrder   int        `json:"sort_order"`
	CreatedAt   *time.Time `json:"created_at,omitempty"`
	CreatedBy   *string    `json:"created_by,omitempty"`
}

type RequestLog struct {
	ID           string     `json:"id"`
	UserID       string     `json:"user_id"`
	RequestID    string     `json:"request_id"`
	StatusName   string     `json:"status_name"`
	DateResponse *time.Time `json:"date_response,omitempty"`
}

type NotificationPayload struct {
	LogID         string     `json:"log_id"`
	UserID        string     `json:"user_id"`
	StatusName    string     `json:"status_name"`
	DateResponse  *time.Time `json:"date_response,omitempty"`
	RequestLogRef string     `json:"request_log_request_id"`
}
