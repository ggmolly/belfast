package types

type ServerStatusResponse struct {
	Running     bool   `json:"running"`
	Accepting   bool   `json:"accepting"`
	UptimeSec   int64  `json:"uptime_sec"`
	UptimeHuman string `json:"uptime_human"`
	ClientCount int    `json:"client_count"`
}

type ServerConfigResponse struct {
	BindAddress string `json:"bind_address"`
	Port        int    `json:"port"`
	Region      string `json:"region"`
}

type ServerConfigUpdate struct {
	BindAddress string `json:"bind_address"`
	Port        int    `json:"port"`
	Region      string `json:"region"`
}

type ServerMaintenanceUpdate struct {
	Enabled bool `json:"enabled"`
}

type ServerMaintenanceResponse struct {
	Enabled bool `json:"enabled"`
}

type ServerStatsResponse struct {
	ClientCount int `json:"client_count"`
}

type ServerMetricsResponse struct {
	ClientCount   int     `json:"client_count"`
	QueueMax      int     `json:"queue_max"`
	QueueBlocks   uint64  `json:"queue_blocks"`
	HandlerErrors uint64  `json:"handler_errors"`
	WriteErrors   uint64  `json:"write_errors"`
	PacketsPerSec float64 `json:"pps"`
}

type ServerUptimeResponse struct {
	UptimeSec   int64  `json:"uptime_sec"`
	UptimeHuman string `json:"uptime_human"`
}

type ConnectionSummary struct {
	Hash        uint32 `json:"hash"`
	RemoteAddr  string `json:"remote_address"`
	ConnectedAt string `json:"connected_at"`
	CommanderID uint32 `json:"commander_id,omitempty"`
}

type ConnectionDetail struct {
	Hash          uint32 `json:"hash"`
	RemoteAddr    string `json:"remote_address"`
	ConnectedAt   string `json:"connected_at"`
	CommanderID   uint32 `json:"commander_id,omitempty"`
	QueueMax      int    `json:"queue_max"`
	QueueBlocks   uint64 `json:"queue_blocks"`
	HandlerErrors uint64 `json:"handler_errors"`
	WriteErrors   uint64 `json:"write_errors"`
}
