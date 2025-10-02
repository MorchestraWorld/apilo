package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the complete benchmark configuration
type Config struct {
	Name              string          `yaml:"name"`
	Description       string          `yaml:"description"`
	OutputDir         string          `yaml:"output_dir"`
	ComparisonBaseline string         `yaml:"comparison_baseline,omitempty"`
	Runs              []RunConfig     `yaml:"runs"`
}

// RunConfig represents configuration for a single benchmark run
type RunConfig struct {
	Name             string               `yaml:"name"`
	Config           BenchmarkSettings    `yaml:"config"`
	Iterations       int                  `yaml:"iterations"`
	WarmupIterations int                  `yaml:"warmup_iterations"`
	LoadPattern      string               `yaml:"load_pattern"`
}

// BenchmarkSettings contains the actual benchmark parameters
type BenchmarkSettings struct {
	TargetURL     string            `yaml:"target_url"`
	TotalRequests int               `yaml:"total_requests"`
	Concurrency   int               `yaml:"concurrency"`
	Timeout       Duration          `yaml:"timeout"`
	KeepAlive     bool              `yaml:"keep_alive"`
	Method        string            `yaml:"method"`
	CustomHeaders map[string]string `yaml:"custom_headers,omitempty"`
	Body          string            `yaml:"body,omitempty"`
	Cache         *CacheConfig      `yaml:"cache,omitempty"`
}

// CacheConfig represents cache configuration
type CacheConfig struct {
	Enabled       bool          `yaml:"enabled"`
	Capacity      int           `yaml:"capacity"`
	MaxMemoryMB   int64         `yaml:"max_memory_mb"`
	Policy        PolicyConfig  `yaml:"policy"`
	Warmup        WarmupConfig  `yaml:"warmup"`
	CleanupInterval string      `yaml:"cleanup_interval"`
	MetricsInterval string      `yaml:"metrics_interval"`
}

// PolicyConfig represents cache policy configuration
type PolicyConfig struct {
	Type           string `yaml:"type"`
	BaseTTL        string `yaml:"base_ttl"`
	MinTTL         string `yaml:"min_ttl"`
	MaxTTL         string `yaml:"max_ttl"`
	MaxCacheSizeMB int64  `yaml:"max_cache_size_mb"`
	MinAccessCount int64  `yaml:"min_access_count"`
}

// WarmupConfig represents cache warmup configuration
type WarmupConfig struct {
	Enabled          bool     `yaml:"enabled"`
	Strategy         string   `yaml:"strategy"`
	Interval         string   `yaml:"interval"`
	StaticURLs       []string `yaml:"static_urls"`
	PredictionWindow string   `yaml:"prediction_window"`
	TopN             int      `yaml:"top_n"`
}

// Duration is a custom type to handle YAML duration parsing
type Duration struct {
	time.Duration
}

// UnmarshalYAML implements custom unmarshaling for duration
func (d *Duration) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}

	duration, err := time.ParseDuration(s)
	if err != nil {
		return fmt.Errorf("invalid duration format: %w", err)
	}

	d.Duration = duration
	return nil
}

// MarshalYAML implements custom marshaling for duration
func (d Duration) MarshalYAML() (interface{}, error) {
	return d.Duration.String(), nil
}

// LoadConfig loads configuration from a YAML file
func LoadConfig(filepath string) (*Config, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Set defaults
	if config.OutputDir == "" {
		config.OutputDir = "./benchmarks/results"
	}

	for i := range config.Runs {
		run := &config.Runs[i]
		if run.Iterations <= 0 {
			run.Iterations = 1
		}
		if run.Config.Method == "" {
			run.Config.Method = "GET"
		}
		if run.Config.Timeout.Duration == 0 {
			run.Config.Timeout.Duration = 30 * time.Second
		}
		if run.Config.Concurrency <= 0 {
			run.Config.Concurrency = 1
		}
		if run.Config.TotalRequests <= 0 {
			run.Config.TotalRequests = 100
		}
	}

	return &config, nil
}

// SaveConfig saves configuration to a YAML file
func SaveConfig(config *Config, filepath string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(filepath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("configuration name is required")
	}

	if len(c.Runs) == 0 {
		return fmt.Errorf("at least one benchmark run is required")
	}

	for i, run := range c.Runs {
		if err := run.Validate(); err != nil {
			return fmt.Errorf("run %d (%s) validation failed: %w", i, run.Name, err)
		}
	}

	return nil
}

// Validate checks if a run configuration is valid
func (r *RunConfig) Validate() error {
	if r.Name == "" {
		return fmt.Errorf("run name is required")
	}

	if r.Config.TargetURL == "" {
		return fmt.Errorf("target URL is required")
	}

	if r.Config.TotalRequests <= 0 {
		return fmt.Errorf("total requests must be positive")
	}

	if r.Config.Concurrency <= 0 {
		return fmt.Errorf("concurrency must be positive")
	}

	if r.Config.Concurrency > r.Config.TotalRequests {
		return fmt.Errorf("concurrency cannot exceed total requests")
	}

	return nil
}

// DefaultConfig returns a sensible default configuration
func DefaultConfig() *Config {
	return &Config{
		Name:        "default_benchmark",
		Description: "Default benchmark configuration",
		OutputDir:   "./benchmarks/results",
		Runs: []RunConfig{
			{
				Name: "baseline",
				Config: BenchmarkSettings{
					TargetURL:     "https://api.anthropic.com",
					TotalRequests: 100,
					Concurrency:   10,
					Timeout:       Duration{Duration: 30 * time.Second},
					KeepAlive:     true,
					Method:        "GET",
				},
				Iterations:       3,
				WarmupIterations: 1,
				LoadPattern:      "constant",
			},
		},
	}
}
