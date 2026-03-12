package tools

import (
	"encoding/json"
	"time"
)

// APIError is the MCP-facing normalized error payload.
type APIError struct {
	StatusCode int            `json:"statusCode"`
	Message    string         `json:"message"`
	Status     string         `json:"status,omitempty"`
	Detail     string         `json:"detail,omitempty"`
	Upstream   map[string]any `json:"upstream,omitempty"`
}

func (e *APIError) Error() string {
	if e.Detail != "" {
		return e.Message + ": " + e.Detail
	}
	return e.Message
}

// SearchHit maps a /api/search hit object.
type SearchHit struct {
	ID          int64    `json:"id,omitempty"`
	UID         string   `json:"uid,omitempty"`
	Title       string   `json:"title,omitempty"`
	Type        string   `json:"type,omitempty"`
	URL         string   `json:"url,omitempty"`
	URI         string   `json:"uri,omitempty"`
	Slug        string   `json:"slug,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	FolderID    int64    `json:"folderId,omitempty"`
	FolderUID   string   `json:"folderUid,omitempty"`
	FolderTitle string   `json:"folderTitle,omitempty"`
	FolderURL   string   `json:"folderUrl,omitempty"`
	IsStarred   bool     `json:"isStarred,omitempty"`
}

// FolderItem is the contract type for folder list results.
type FolderItem struct {
	ID    int64  `json:"id,omitempty"`
	UID   string `json:"uid,omitempty"`
	Title string `json:"title,omitempty"`
	URL   string `json:"url,omitempty"`
}

// DatasourceModel is the contract type for datasource details.
type DatasourceModel struct {
	ID               int64           `json:"id,omitempty"`
	UID              string          `json:"uid,omitempty"`
	Name             string          `json:"name,omitempty"`
	Type             string          `json:"type,omitempty"`
	TypeName         string          `json:"typeName,omitempty"`
	TypeLogoURL      string          `json:"typeLogoUrl,omitempty"`
	URL              string          `json:"url,omitempty"`
	Access           string          `json:"access,omitempty"`
	Database         string          `json:"database,omitempty"`
	User             string          `json:"user,omitempty"`
	OrgID            int64           `json:"orgId,omitempty"`
	IsDefault        bool            `json:"isDefault,omitempty"`
	BasicAuth        bool            `json:"basicAuth,omitempty"`
	WithCredentials  bool            `json:"withCredentials,omitempty"`
	ReadOnly         bool            `json:"readOnly,omitempty"`
	Version          int64           `json:"version,omitempty"`
	JSONData         map[string]any  `json:"jsonData,omitempty"`
	SecureJSONFields map[string]bool `json:"secureJsonFields,omitempty"`
}

// AnnotationItem is the contract type for annotation results.
type AnnotationItem struct {
	ID           int64          `json:"id,omitempty"`
	AlertID      int64          `json:"alertId,omitempty"`
	AlertUID     string         `json:"alertUID,omitempty"`
	AlertName    string         `json:"alertName,omitempty"`
	DashboardID  int64          `json:"dashboardId,omitempty"`
	DashboardUID string         `json:"dashboardUID,omitempty"`
	PanelID      int64          `json:"panelId,omitempty"`
	UserID       int64          `json:"userId,omitempty"`
	Login        string         `json:"login,omitempty"`
	Email        string         `json:"email,omitempty"`
	AvatarURL    string         `json:"avatarUrl,omitempty"`
	NewState     string         `json:"newState,omitempty"`
	PrevState    string         `json:"prevState,omitempty"`
	Text         string         `json:"text,omitempty"`
	Tags         []string       `json:"tags,omitempty"`
	Time         int64          `json:"time,omitempty"`
	TimeEnd      int64          `json:"timeEnd,omitempty"`
	Created      int64          `json:"created,omitempty"`
	Updated      int64          `json:"updated,omitempty"`
	Data         map[string]any `json:"data,omitempty"`
}

// LegacyAlertItem is the contract type for legacy alert results.
type LegacyAlertItem struct {
	ID             int64          `json:"id,omitempty"`
	DashboardID    int64          `json:"dashboardId,omitempty"`
	DashboardUID   string         `json:"dashboardUid,omitempty"`
	DashboardSlug  string         `json:"dashboardSlug,omitempty"`
	PanelID        int64          `json:"panelId,omitempty"`
	Name           string         `json:"name,omitempty"`
	State          string         `json:"state,omitempty"`
	NewStateDate   time.Time      `json:"newStateDate,omitempty"`
	EvalDate       time.Time      `json:"evalDate,omitempty"`
	ExecutionError string         `json:"executionError,omitempty"`
	URL            string         `json:"url,omitempty"`
	EvalData       map[string]any `json:"evalData,omitempty"`
}

// LegacyNotificationChannel is the contract type for notification channel results.
type LegacyNotificationChannel struct {
	ID                    int64           `json:"id,omitempty"`
	UID                   string          `json:"uid,omitempty"`
	Name                  string          `json:"name,omitempty"`
	Type                  string          `json:"type,omitempty"`
	IsDefault             bool            `json:"isDefault,omitempty"`
	SendReminder          bool            `json:"sendReminder,omitempty"`
	DisableResolveMessage bool            `json:"disableResolveMessage,omitempty"`
	Frequency             string          `json:"frequency,omitempty"`
	Created               time.Time       `json:"created,omitempty"`
	Updated               time.Time       `json:"updated,omitempty"`
	SecureFields          map[string]bool `json:"secureFields,omitempty"`
	Settings              map[string]any  `json:"settings,omitempty"`
}

// OrgUserItem is the contract type for org user results.
type OrgUserItem struct {
	UserID        int64           `json:"userId,omitempty"`
	OrgID         int64           `json:"orgId,omitempty"`
	Login         string          `json:"login,omitempty"`
	Name          string          `json:"name,omitempty"`
	Email         string          `json:"email,omitempty"`
	AvatarURL     string          `json:"avatarUrl,omitempty"`
	Role          string          `json:"role,omitempty"`
	LastSeenAt    time.Time       `json:"lastSeenAt,omitempty"`
	LastSeenAtAge string          `json:"lastSeenAtAge,omitempty"`
	AccessControl map[string]bool `json:"accessControl,omitempty"`
}

// TeamItem is the contract type for team results.
type TeamItem struct {
	ID            int64           `json:"id,omitempty"`
	OrgID         int64           `json:"orgId,omitempty"`
	Name          string          `json:"name,omitempty"`
	Email         string          `json:"email,omitempty"`
	AvatarURL     string          `json:"avatarUrl,omitempty"`
	MemberCount   int64           `json:"memberCount,omitempty"`
	Permission    int64           `json:"permission,omitempty"`
	AccessControl map[string]bool `json:"accessControl,omitempty"`
}

// FlexibleID handles IDs that may be either int or string in upsert responses.
type FlexibleID = json.RawMessage
