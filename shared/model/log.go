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
	EndPointID    string            `json:"endpoint_id"`     // Unique gosight ID for the endpoint
	OS            string            `json:"os"`              // linux, windows, darwin, etc.
	Platform      string            `json:"platform"`        // journald, eventlog, syslog, etc.
	AppName       string            `json:"app_name"`        // e.g., nginx, sshd
	AppVersion    string            `json:"app_version"`     // if known
	ContainerID   string            `json:"container_id"`    // if inside a container
	ContainerName string            `json:"container_name"`  // optional
	Unit          string            `json:"unit"`            // For journald: systemd unit name
	Service       string            `json:"service"`         // For syslog/Windows Event Log
	EventID       string            `json:"event_id"`        // Windows event ID, etc.
	User          string            `json:"user,omitempty"`  // User associated with log entry
	Executable    string            `json:"exe,omitempty"`   // Path to binary if available
	Path          string            `json:"path,omitempty"`  // Original source log path
	Extra         map[string]string `json:"extra,omitempty"` // For collector-specific fields
	EndpointID    string            `json:"endpoint_id"`     // Unique ID for the endpoint
}

type LogPayload struct {
	EndpointID string
	Timestamp  time.Time
	Logs       []LogEntry
	Meta       *LogMeta
}
