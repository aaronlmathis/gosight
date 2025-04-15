package model

import "time"

type LogEntry struct {
	Timestamp time.Time         `json:"timestamp"`      // When the log was emitted
	Level     string            `json:"level"`          // info, warn, error, debug, etc.
	Message   string            `json:"message"`        // The actual log content
	Source    string            `json:"source"`         // Collector name or service name (e.g., journald, nginx)
	Category  string            `json:"category"`       // Optional: auth, network, system, app, etc.
	Host      string            `json:"host"`           // Hostname or agent ID
	PID       int               `json:"pid,omitempty"`  // Process ID if available
	Fields    map[string]string `json:"fields"`         // Structured fields (JSON logs, key/values)
	Tags      map[string]string `json:"tags"`           // Custom labels/tags (user-defined or enriched)
	Meta      *LogMeta          `json:"meta,omitempty"` // Optional platform/service-specific metadata
}

type LogMeta struct {
	EndPointID    string            `json:"endpoint_id,omitempty"`    // Unique gosight ID for the endpoint
	OS            string            `json:"os,omitempty"`             // linux, windows, darwin, etc.
	Platform      string            `json:"platform,omitempty"`       // journald, eventlog, syslog, etc.
	AppName       string            `json:"app_name,omitempty"`       // e.g., nginx, sshd
	AppVersion    string            `json:"app_version,omitempty"`    // if known
	ContainerID   string            `json:"container_id,omitempty"`   // if inside a container
	ContainerName string            `json:"container_name,omitempty"` // optional
	Unit          string            `json:"unit,omitempty"`           // For journald: systemd unit name
	Service       string            `json:"service,omitempty"`        // For syslog/Windows Event Log
	EventID       string            `json:"event_id,omitempty"`       // Windows event ID, etc.
	User          string            `json:"user,omitempty"`           // User associated with log entry
	Executable    string            `json:"exe,omitempty"`            // Path to binary if available
	Path          string            `json:"path,omitempty"`           // Original source log path
	Extra         map[string]string `json:"extra,omitempty"`          // For collector-specific fields
	AgentID       string            `json:"agent_id,omitempty"`       // Unique ID for the agent
}

type LogPayload struct {
	EndpointID string
	Timestamp  time.Time
	Logs       []LogEntry
	Meta       *Meta
}
