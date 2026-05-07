package models

import "time"

// Health Response - GET /health
type HealthResponse struct {
	Status    string `json:"status"`
	IsRunning bool   `json:"is_running"`
}

// Command Response - POST /restart
type CommandResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// Time Diff Response - GET /time-diff
type TimeDiffResponse struct {
	Success      bool      `json:"success"`
	TimeDiffNs   int64     `json:"time_diff_ns"`
	TimeDiffStr  string    `json:"time_diff_str"`
	SystemTime   time.Time `json:"system_time"`
	NtpTime      time.Time `json:"ntp_time"`
	ErrorMessage string    `json:"error_message,omitempty"`
}

// Privileges Response - GET /debug/privileges
type PrivilegesResponse struct {
	User             string `json:"user"`
	IsAdministrator  bool   `json:"isAdministrator"`
	ProcessId        int    `json:"processId"`
	CurrentDirectory string `json:"currentDirectory"`
}

// Database Health Response - GET /db-health
type DatabaseHealthResponse struct {
	Success       bool      `json:"success"`
	SqlConnection string    `json:"sql_connection"`
	XmlStatus     string    `json:"xml_status"`
	LastCheck     time.Time `json:"last_check"`
}

// Data Delays Response - GET /data-delays
type DelayedPoint struct {
	Name       string    `json:"name"`
	LastDate   time.Time `json:"last_date"`
	DelayHours float64   `json:"delay_hours"`
}

type DataDelaysResponse struct {
	Success            bool           `json:"success"`
	HasDelays          bool           `json:"has_delays"`
	DelayedPointsCount int            `json:"delayed_points_count"`
	DelayedPoints      []DelayedPoint `json:"delayed_points"`
	LastCheck          time.Time      `json:"last_check"`
	Error              string         `json:"error,omitempty"`
}