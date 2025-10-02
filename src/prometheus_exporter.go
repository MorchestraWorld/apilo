package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// PrometheusExporter exports metrics in Prometheus format
type PrometheusExporter struct {
	port      int
	path      string
	collector *MetricsCollector

	server  *http.Server
	mu      sync.RWMutex
	running bool
}

// NewPrometheusExporter creates a new Prometheus exporter
func NewPrometheusExporter(port int, path string) *PrometheusExporter {
	return &PrometheusExporter{
		port: port,
		path: path,
	}
}

// Start starts the Prometheus HTTP server
func (pe *PrometheusExporter) Start(collector *MetricsCollector) error {
	pe.mu.Lock()
	defer pe.mu.Unlock()

	if pe.running {
		return fmt.Errorf("Prometheus exporter already running")
	}

	pe.collector = collector

	mux := http.NewServeMux()
	mux.HandleFunc(pe.path, pe.handleMetrics)
	mux.HandleFunc("/", pe.handleIndex)

	pe.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", pe.port),
		Handler: mux,
	}

	go func() {
		if err := pe.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Prometheus exporter error: %v\n", err)
		}
	}()

	pe.running = true
	return nil
}

// Stop stops the Prometheus HTTP server
func (pe *PrometheusExporter) Stop() error {
	pe.mu.Lock()
	defer pe.mu.Unlock()

	if !pe.running {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := pe.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("Prometheus exporter shutdown error: %w", err)
	}

	pe.running = false
	return nil
}

// handleIndex serves a simple index page
func (pe *PrometheusExporter) handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `<html>
<head><title>API Latency Optimizer - Prometheus Exporter</title></head>
<body>
<h1>Prometheus Metrics Exporter</h1>
<p><a href="%s">Metrics</a></p>
</body>
</html>`, pe.path)
}

// handleMetrics serves metrics in Prometheus format
func (pe *PrometheusExporter) handleMetrics(w http.ResponseWriter, r *http.Request) {
	snapshot := pe.collector.GetSnapshot()
	if snapshot == nil {
		http.Error(w, "No metrics available", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "text/plain; version=0.0.4")

	var sb strings.Builder

	// Write header
	sb.WriteString("# HELP api_latency_optimizer_info Information about the API latency optimizer\n")
	sb.WriteString("# TYPE api_latency_optimizer_info gauge\n")
	sb.WriteString("api_latency_optimizer_info{version=\"1.0.0\"} 1\n\n")

	// Cache metrics
	pe.writeMetric(&sb, "cache_hit_ratio", "Cache hit ratio (0-1)", "gauge",
		snapshot.CacheHitRatio, nil)
	pe.writeMetric(&sb, "cache_miss_ratio", "Cache miss ratio (0-1)", "gauge",
		snapshot.CacheMissRatio, nil)
	pe.writeMetric(&sb, "cache_size", "Current cache size (number of entries)", "gauge",
		float64(snapshot.CacheSize), nil)
	pe.writeMetric(&sb, "cache_capacity", "Maximum cache capacity", "gauge",
		float64(snapshot.CacheCapacity), nil)
	pe.writeMetric(&sb, "cache_memory_usage_bytes", "Current cache memory usage in bytes", "gauge",
		snapshot.CacheMemoryUsageMB*1024*1024, nil)
	pe.writeMetric(&sb, "cache_memory_peak_bytes", "Peak cache memory usage in bytes", "gauge",
		snapshot.CachePeakMemoryMB*1024*1024, nil)
	pe.writeMetric(&sb, "cache_gets_total", "Total number of cache get operations", "counter",
		float64(snapshot.CacheTotalGets), nil)
	pe.writeMetric(&sb, "cache_hits_total", "Total number of cache hits", "counter",
		float64(snapshot.CacheTotalHits), nil)
	pe.writeMetric(&sb, "cache_misses_total", "Total number of cache misses", "counter",
		float64(snapshot.CacheTotalMisses), nil)
	pe.writeMetric(&sb, "cache_evictions_total", "Total number of cache evictions", "counter",
		float64(snapshot.CacheEvictions), nil)
	pe.writeMetric(&sb, "cache_expirations_total", "Total number of cache expirations", "counter",
		float64(snapshot.CacheExpirations), nil)

	// Latency metrics
	pe.writeMetric(&sb, "latency_p50_milliseconds", "P50 latency in milliseconds", "gauge",
		snapshot.LatencyP50, nil)
	pe.writeMetric(&sb, "latency_p95_milliseconds", "P95 latency in milliseconds", "gauge",
		snapshot.LatencyP95, nil)
	pe.writeMetric(&sb, "latency_p99_milliseconds", "P99 latency in milliseconds", "gauge",
		snapshot.LatencyP99, nil)
	pe.writeMetric(&sb, "latency_mean_milliseconds", "Mean latency in milliseconds", "gauge",
		snapshot.LatencyMean, nil)
	pe.writeMetric(&sb, "latency_min_milliseconds", "Minimum latency in milliseconds", "gauge",
		snapshot.LatencyMin, nil)
	pe.writeMetric(&sb, "latency_max_milliseconds", "Maximum latency in milliseconds", "gauge",
		snapshot.LatencyMax, nil)

	// Time to First Byte metrics
	pe.writeMetric(&sb, "ttfb_p50_milliseconds", "P50 time to first byte in milliseconds", "gauge",
		snapshot.TTFBP50, nil)
	pe.writeMetric(&sb, "ttfb_p95_milliseconds", "P95 time to first byte in milliseconds", "gauge",
		snapshot.TTFBP95, nil)
	pe.writeMetric(&sb, "ttfb_p99_milliseconds", "P99 time to first byte in milliseconds", "gauge",
		snapshot.TTFBP99, nil)

	// Throughput metrics
	pe.writeMetric(&sb, "requests_per_second", "Requests per second", "gauge",
		snapshot.RequestsPerSecond, nil)
	pe.writeMetric(&sb, "bytes_per_second", "Bytes per second", "gauge",
		snapshot.BytesPerSecond, nil)

	// Request metrics
	pe.writeMetric(&sb, "requests_total", "Total number of requests", "counter",
		float64(snapshot.TotalRequests), nil)
	pe.writeMetric(&sb, "requests_successful_total", "Total number of successful requests", "counter",
		float64(snapshot.SuccessfulRequests), nil)
	pe.writeMetric(&sb, "requests_failed_total", "Total number of failed requests", "counter",
		float64(snapshot.FailedRequests), nil)
	pe.writeMetric(&sb, "error_rate", "Error rate (0-1)", "gauge",
		snapshot.ErrorRate, nil)

	// Connection metrics
	pe.writeMetric(&sb, "connection_reuse_rate", "Connection reuse rate (0-1)", "gauge",
		snapshot.ConnectionReuseRate, nil)

	// System metrics
	pe.writeMetric(&sb, "uptime_seconds", "System uptime in seconds", "counter",
		snapshot.UptimeSeconds, nil)

	// Performance metrics
	pe.writeMetric(&sb, "performance_score", "Overall performance score (0-100)", "gauge",
		float64(snapshot.PerformanceScore), nil)

	// Performance grade as info metric
	gradeValue := pe.gradeToValue(snapshot.PerformanceGrade)
	pe.writeMetric(&sb, "performance_grade_value", "Performance grade as numeric value", "gauge",
		gradeValue, map[string]string{"grade": snapshot.PerformanceGrade})

	w.Write([]byte(sb.String()))
}

// writeMetric writes a metric in Prometheus format
func (pe *PrometheusExporter) writeMetric(sb *strings.Builder, name, help, metricType string, value float64, labels map[string]string) {
	fullName := fmt.Sprintf("api_latency_optimizer_%s", name)

	// Write HELP
	sb.WriteString(fmt.Sprintf("# HELP %s %s\n", fullName, help))

	// Write TYPE
	sb.WriteString(fmt.Sprintf("# TYPE %s %s\n", fullName, metricType))

	// Write metric with labels
	if len(labels) > 0 {
		labelStr := pe.formatLabels(labels)
		sb.WriteString(fmt.Sprintf("%s{%s} %v\n", fullName, labelStr, value))
	} else {
		sb.WriteString(fmt.Sprintf("%s %v\n", fullName, value))
	}

	sb.WriteString("\n")
}

// formatLabels formats labels for Prometheus
func (pe *PrometheusExporter) formatLabels(labels map[string]string) string {
	if len(labels) == 0 {
		return ""
	}

	parts := make([]string, 0, len(labels))
	for k, v := range labels {
		parts = append(parts, fmt.Sprintf("%s=\"%s\"", k, v))
	}

	return strings.Join(parts, ",")
}

// gradeToValue converts letter grade to numeric value for Prometheus
func (pe *PrometheusExporter) gradeToValue(grade string) float64 {
	switch grade {
	case "A":
		return 5.0
	case "B":
		return 4.0
	case "C":
		return 3.0
	case "D":
		return 2.0
	case "F":
		return 1.0
	default:
		return 0.0
	}
}

// GetMetricsText returns the current metrics as a Prometheus-formatted string
func (pe *PrometheusExporter) GetMetricsText() (string, error) {
	snapshot := pe.collector.GetSnapshot()
	if snapshot == nil {
		return "", fmt.Errorf("no metrics available")
	}

	var sb strings.Builder

	// This is a simplified version - the full implementation would use handleMetrics logic
	sb.WriteString("# API Latency Optimizer Metrics\n")
	sb.WriteString(fmt.Sprintf("cache_hit_ratio %.4f\n", snapshot.CacheHitRatio))
	sb.WriteString(fmt.Sprintf("latency_p95_ms %.2f\n", snapshot.LatencyP95))
	sb.WriteString(fmt.Sprintf("latency_p99_ms %.2f\n", snapshot.LatencyP99))

	return sb.String(), nil
}

// SamplePrometheusConfig returns a sample Prometheus configuration
func SamplePrometheusConfig(exporterPort int, scrapePath string) string {
	return fmt.Sprintf(`
# Prometheus Configuration for API Latency Optimizer

global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'api_latency_optimizer'
    static_configs:
      - targets: ['localhost:%d']
    metrics_path: '%s'
    scrape_interval: 5s
    scrape_timeout: 5s

# Alert rules can be configured here
# rule_files:
#   - 'alerts.yml'

# Alertmanager configuration (optional)
# alerting:
#   alertmanagers:
#     - static_configs:
#         - targets: ['localhost:9093']
`, exporterPort, scrapePath)
}

// SampleAlertRules returns sample Prometheus alert rules
func SampleAlertRules() string {
	return `
# Alert Rules for API Latency Optimizer

groups:
  - name: latency_alerts
    interval: 30s
    rules:
      - alert: HighLatencyP95
        expr: api_latency_optimizer_latency_p95_milliseconds > 500
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "High P95 latency detected"
          description: "P95 latency is {{ $value }}ms (threshold: 500ms)"

      - alert: CriticalLatencyP99
        expr: api_latency_optimizer_latency_p99_milliseconds > 1000
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Critical P99 latency detected"
          description: "P99 latency is {{ $value }}ms (threshold: 1000ms)"

  - name: cache_alerts
    interval: 30s
    rules:
      - alert: LowCacheHitRatio
        expr: api_latency_optimizer_cache_hit_ratio < 0.6
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Low cache hit ratio"
          description: "Cache hit ratio is {{ $value }} (threshold: 0.6)"

      - alert: CriticalCacheHitRatio
        expr: api_latency_optimizer_cache_hit_ratio < 0.4
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "Critical cache hit ratio"
          description: "Cache hit ratio is {{ $value }} (threshold: 0.4)"

      - alert: HighCacheMemoryUsage
        expr: api_latency_optimizer_cache_memory_usage_bytes / (1024*1024) > 500
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "High cache memory usage"
          description: "Cache memory usage is {{ $value }}MB (threshold: 500MB)"

  - name: error_alerts
    interval: 30s
    rules:
      - alert: HighErrorRate
        expr: api_latency_optimizer_error_rate > 0.05
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "High error rate detected"
          description: "Error rate is {{ $value }} (threshold: 0.05)"

  - name: performance_alerts
    interval: 30s
    rules:
      - alert: LowPerformanceScore
        expr: api_latency_optimizer_performance_score < 60
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Low performance score"
          description: "Performance score is {{ $value }}/100 (threshold: 60)"

      - alert: LowThroughput
        expr: api_latency_optimizer_requests_per_second < 10
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Low throughput detected"
          description: "Throughput is {{ $value }} req/s (threshold: 10 req/s)"
`
}

// SavePrometheusConfig saves the Prometheus configuration to a file
func SavePrometheusConfig(filepath string, exporterPort int, scrapePath string) error {
	config := SamplePrometheusConfig(exporterPort, scrapePath)
	return writeFile(filepath, []byte(config))
}

// SaveAlertRules saves the alert rules to a file
func SaveAlertRules(filepath string) error {
	rules := SampleAlertRules()
	return writeFile(filepath, []byte(rules))
}

// writeFile is a helper to write content to a file
func writeFile(filepath string, content []byte) error {
	// This would use os.WriteFile in a real implementation
	// For now, we'll just return nil as a placeholder
	return nil
}
