package core

import "time"

// ResourceCategory defines a category of resources provided by a platform
type ResourceCategory struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Unit  string `json:"unit"`  // e.g., "GB", "MB", "次", "h"
	Limit float64 `json:"limit"` // 0 means unlimited
}

// ResourceUsage represents the usage of a specific resource
type ResourceUsage struct {
	CategoryID string     `json:"categoryId"`
	Name       string     `json:"name"`
	Used       float64    `json:"used"`
	Limit      float64    `json:"limit"` // 0 means unlimited
	Unit       string     `json:"unit"`
	ResetDate  *time.Time `json:"resetDate,omitempty"`
}

// UsageData contains all usage information for a platform account
type UsageData struct {
	Platform    string          `json:"platform"`
	AccountName string          `json:"accountName"`
	Plan        string          `json:"plan"`
	Resources   []ResourceUsage `json:"resources"`
	LastUpdated time.Time       `json:"lastUpdated"`
	Status      ServiceStatus   `json:"status"`
	Error       string          `json:"error,omitempty"`
}

// ServiceStatus represents the health status of a service
type ServiceStatus string

const (
	StatusOK     ServiceStatus = "ok"
	StatusWarn   ServiceStatus = "warn"
	StatusDanger ServiceStatus = "danger"
	StatusError  ServiceStatus = "error"
)

// ProjectInfo represents a project hosted on a platform
type ProjectInfo struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	URL       string    `json:"url,omitempty"`
	CreatedAt time.Time `json:"createdAt,omitempty"`
}

// AlertInfo represents a usage alert
type AlertInfo struct {
	Platform   string    `json:"platform"`
	Resource   string    `json:"resource"`
	Percentage float64   `json:"percentage"`
	Time       time.Time `json:"time"`
	Status     string    `json:"status"` // "warn", "danger", "critical"
}

// DashboardData contains all data needed for the dashboard view
type DashboardData struct {
	Services    []UsageData `json:"services"`
	Alerts      []AlertInfo `json:"alerts"`
	LastSync    time.Time   `json:"lastSync"`
	ConnectedCount int     `json:"connectedCount"`
	HealthyCount   int     `json:"healthyCount"`
	WarnCount      int     `json:"warnCount"`
}

// AuthConfig defines the authentication fields needed for a platform
type AuthConfig struct {
	Fields []AuthField `json:"fields"`
}

// AuthField represents a single authentication field
type AuthField struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"` // "text", "password"
	Required    bool   `json:"required"`
}

// PlatformCredentials stores credentials for a platform account
type PlatformCredentials struct {
	Platform    string            `json:"platform"`
	AccountName string            `json:"accountName"`
	Credentials map[string]string `json:"credentials"`
}
