package daemon

import "time"

// OptimizationRequest represents an API call to be optimized
type OptimizationRequest struct {
	URL     string            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    []byte            `json:"body,omitempty"`
	Timeout time.Duration     `json:"timeout,omitempty"`
}

// OptimizationResponse contains the optimized response
type OptimizationResponse struct {
	StatusCode int               `json:"status_code"`
	Headers    map[string]string `json:"headers"`
	Body       []byte            `json:"body"`
	Latency    time.Duration     `json:"latency"`
	CacheHit   bool              `json:"cache_hit"`
	Optimized  bool              `json:"optimized"`
	Error      string            `json:"error,omitempty"`
	Metadata   ResponseMetadata  `json:"metadata"`
}

// ResponseMetadata provides optimization details
type ResponseMetadata struct {
	CacheStatus      string              `json:"cache_status"`
	OptimizationType string              `json:"optimization_type"`
	LatencySaved     time.Duration       `json:"latency_saved"`
	ConnectionReused bool                `json:"connection_reused"`
	HTTP2Used        bool                `json:"http2_used"`
	ClaudeTokens     *ClaudeTokenMetrics `json:"claude_tokens,omitempty"`
	TokenUsage       *TokenUsage         `json:"token_usage,omitempty"`
}

// TokenUsage represents token consumption for a request
type TokenUsage struct {
	InputTokens  int64 `json:"input_tokens"`
	OutputTokens int64 `json:"output_tokens"`
	TotalTokens  int64 `json:"total_tokens"`
	IsEstimated  bool  `json:"is_estimated"`
}

// ClaudeTokenMetrics tracks Claude API token usage and costs
type ClaudeTokenMetrics struct {
	InputTokens   int64   `json:"input_tokens"`
	OutputTokens  int64   `json:"output_tokens"`
	TotalTokens   int64   `json:"total_tokens"`
	Cost          float64 `json:"cost"`
	TotalRequests int64   `json:"total_requests"`
	Model         string  `json:"model"`
}

// DaemonStatus represents the current daemon state
type DaemonStatus struct {
	Running       bool                `json:"running"`
	PID           int                 `json:"pid"`
	Uptime        time.Duration       `json:"uptime"`
	Port          int                 `json:"port"`
	TotalRequests int64               `json:"total_requests"`
	CacheHitRatio float64             `json:"cache_hit_ratio"`
	AvgLatency    time.Duration       `json:"avg_latency"`
	MemoryUsageMB float64             `json:"memory_usage_mb"`
	CPUPercent    float64             `json:"cpu_percent"`
	ClaudeMetrics *ClaudeTokenMetrics `json:"claude_metrics,omitempty"`
}

// DaemonConfig holds daemon configuration
type DaemonConfig struct {
	Port                 int           `yaml:"port" json:"port"`
	LogLevel             string        `yaml:"log_level" json:"log_level"`
	LogFile              string        `yaml:"log_file" json:"log_file"`
	PIDFile              string        `yaml:"pid_file" json:"pid_file"`
	CacheMaxMemoryMB     int64         `yaml:"cache_max_memory_mb" json:"cache_max_memory_mb"`
	CacheDefaultTTL      time.Duration `yaml:"cache_default_ttl" json:"cache_default_ttl"`
	MaxConnections       int           `yaml:"max_connections" json:"max_connections"`
	IdleTimeout          time.Duration `yaml:"idle_timeout" json:"idle_timeout"`
	EnableHTTP2          bool          `yaml:"enable_http2" json:"enable_http2"`
	EnableCircuitBreaker bool          `yaml:"enable_circuit_breaker" json:"enable_circuit_breaker"`
	MetricsEnabled       bool          `yaml:"metrics_enabled" json:"metrics_enabled"`
}

// DefaultDaemonConfig returns default configuration
func DefaultDaemonConfig() *DaemonConfig {
	return &DaemonConfig{
		Port:                 9876,
		LogLevel:             "info",
		LogFile:              "~/.apilo/logs/daemon.log",
		PIDFile:              "~/.apilo/daemon.pid",
		CacheMaxMemoryMB:     500,
		CacheDefaultTTL:      10 * time.Minute,
		MaxConnections:       20,
		IdleTimeout:          90 * time.Second,
		EnableHTTP2:          true,
		EnableCircuitBreaker: true,
		MetricsEnabled:       true,
	}
}
