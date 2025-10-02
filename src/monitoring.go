package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// MonitoringConfig defines monitoring system configuration
type MonitoringConfig struct {
	// General settings
	Enabled            bool          // Master enable/disable switch

	// Collection settings
	MetricsInterval    time.Duration
	SnapshotInterval   time.Duration
	CleanupInterval    time.Duration

	// Dashboard settings
	DashboardEnabled   bool
	DashboardPort      int
	DashboardRefresh   time.Duration

	// Alerting settings
	AlertingEnabled    bool
	AlertsEnabled      bool          // Alias for compatibility
	AlertCheckInterval time.Duration
	AlertRules         []AlertRule

	// Prometheus settings
	PrometheusEnabled  bool
	PrometheusPort     int
	PrometheusPath     string

	// Storage settings
	RetentionPeriod    time.Duration
	MaxSnapshots       int
	OutputPath         string
}

// MonitoringSystem orchestrates all monitoring components
type MonitoringSystem struct {
	config          MonitoringConfig
	collector       *MetricsCollector
	dashboard       *Dashboard
	alertManager    *AlertManager
	promExporter    *PrometheusExporter

	// Component references for monitoring
	benchmarker     *Benchmarker
	cache           *LRUCache

	// Lifecycle management
	ctx             context.Context
	cancel          context.CancelFunc
	wg              sync.WaitGroup
	mu              sync.RWMutex

	// State tracking
	startTime       time.Time
	running         bool

	// Background tasks
	stopChannels    []chan struct{}
}

// NewMonitoringSystem creates a new monitoring system
func NewMonitoringSystem(config MonitoringConfig) *MonitoringSystem {
	ctx, cancel := context.WithCancel(context.Background())

	// Set defaults
	if config.MetricsInterval == 0 {
		config.MetricsInterval = 5 * time.Second
	}
	if config.SnapshotInterval == 0 {
		config.SnapshotInterval = 30 * time.Second
	}
	if config.CleanupInterval == 0 {
		config.CleanupInterval = 5 * time.Minute
	}
	if config.DashboardPort == 0 {
		config.DashboardPort = 8080
	}
	if config.DashboardRefresh == 0 {
		config.DashboardRefresh = 2 * time.Second
	}
	if config.PrometheusPort == 0 {
		config.PrometheusPort = 9090
	}
	if config.PrometheusPath == "" {
		config.PrometheusPath = "/metrics"
	}
	if config.AlertCheckInterval == 0 {
		config.AlertCheckInterval = 10 * time.Second
	}
	if config.RetentionPeriod == 0 {
		config.RetentionPeriod = 24 * time.Hour
	}
	if config.MaxSnapshots == 0 {
		config.MaxSnapshots = 1000
	}

	ms := &MonitoringSystem{
		config:       config,
		collector:    NewMetricsCollector(config.MaxSnapshots),
		ctx:          ctx,
		cancel:       cancel,
		startTime:    time.Now(),
		stopChannels: make([]chan struct{}, 0),
	}

	// Initialize dashboard if enabled
	if config.DashboardEnabled {
		ms.dashboard = NewDashboard(config.DashboardPort, config.DashboardRefresh)
	}

	// Initialize alert manager if enabled
	if config.AlertingEnabled {
		ms.alertManager = NewAlertManager(config.AlertRules)
	}

	// Initialize Prometheus exporter if enabled
	if config.PrometheusEnabled {
		ms.promExporter = NewPrometheusExporter(config.PrometheusPort, config.PrometheusPath)
	}

	return ms
}

// AttachBenchmarker attaches a benchmarker for monitoring
func (ms *MonitoringSystem) AttachBenchmarker(b *Benchmarker) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.benchmarker = b
	ms.collector.AttachBenchmarker(b)
}

// AttachCache attaches a cache for monitoring
func (ms *MonitoringSystem) AttachCache(c *LRUCache) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.cache = c
	ms.collector.AttachCache(c)

	// Attach alert manager to cache metrics if enabled
	if ms.alertManager != nil {
		ms.alertManager.AttachCacheMetrics(c.GetMetrics())
	}
}

// Start initializes and starts all monitoring components
func (ms *MonitoringSystem) Start() error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if ms.running {
		return fmt.Errorf("monitoring system already running")
	}

	fmt.Printf("\n=== Starting Monitoring System ===\n")
	fmt.Printf("Metrics Interval: %v\n", ms.config.MetricsInterval)
	fmt.Printf("Snapshot Interval: %v\n", ms.config.SnapshotInterval)

	// Start metrics collection
	stopMetrics := ms.startMetricsCollection()
	ms.stopChannels = append(ms.stopChannels, stopMetrics)

	// Start snapshot capture
	stopSnapshots := ms.startSnapshotCapture()
	ms.stopChannels = append(ms.stopChannels, stopSnapshots)

	// Start dashboard if enabled
	if ms.config.DashboardEnabled && ms.dashboard != nil {
		fmt.Printf("Dashboard: http://localhost:%d\n", ms.config.DashboardPort)
		if err := ms.dashboard.Start(ms.collector); err != nil {
			return fmt.Errorf("failed to start dashboard: %w", err)
		}
	}

	// Start alert manager if enabled
	if ms.config.AlertingEnabled && ms.alertManager != nil {
		fmt.Printf("Alert Manager: Enabled (%d rules)\n", len(ms.config.AlertRules))
		stopAlerts := ms.alertManager.Start(ms.config.AlertCheckInterval)
		ms.stopChannels = append(ms.stopChannels, stopAlerts)
	}

	// Start Prometheus exporter if enabled
	if ms.config.PrometheusEnabled && ms.promExporter != nil {
		fmt.Printf("Prometheus Exporter: http://localhost:%d%s\n",
			ms.config.PrometheusPort, ms.config.PrometheusPath)
		if err := ms.promExporter.Start(ms.collector); err != nil {
			return fmt.Errorf("failed to start Prometheus exporter: %w", err)
		}
	}

	// Start cleanup routine
	stopCleanup := ms.startCleanupRoutine()
	ms.stopChannels = append(ms.stopChannels, stopCleanup)

	ms.running = true
	fmt.Printf("Monitoring System: Active\n\n")

	return nil
}

// Stop gracefully shuts down all monitoring components
func (ms *MonitoringSystem) Stop() error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if !ms.running {
		return nil
	}

	fmt.Printf("\n=== Stopping Monitoring System ===\n")

	// Stop all background tasks
	for _, stopChan := range ms.stopChannels {
		close(stopChan)
	}

	// Stop dashboard
	if ms.dashboard != nil {
		ms.dashboard.Stop()
	}

	// Stop Prometheus exporter
	if ms.promExporter != nil {
		ms.promExporter.Stop()
	}

	// Cancel context
	ms.cancel()

	// Wait for all goroutines
	ms.wg.Wait()

	ms.running = false
	fmt.Printf("Monitoring System: Stopped\n")

	return nil
}

// GetCollector returns the metrics collector
func (ms *MonitoringSystem) GetCollector() *MetricsCollector {
	return ms.collector
}

// GetSnapshot returns current monitoring snapshot
func (ms *MonitoringSystem) GetSnapshot() *MonitoringSnapshot {
	return ms.collector.GetSnapshot()
}

// GetDashboard returns the dashboard instance
func (ms *MonitoringSystem) GetDashboard() *Dashboard {
	return ms.dashboard
}

// GetAlertManager returns the alert manager instance
func (ms *MonitoringSystem) GetAlertManager() *AlertManager {
	return ms.alertManager
}

// GetPrometheusExporter returns the Prometheus exporter instance
func (ms *MonitoringSystem) GetPrometheusExporter() *PrometheusExporter {
	return ms.promExporter
}

// IsRunning returns whether the monitoring system is running
func (ms *MonitoringSystem) IsRunning() bool {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	return ms.running
}

// Uptime returns how long the monitoring system has been running
func (ms *MonitoringSystem) Uptime() time.Duration {
	return time.Since(ms.startTime)
}

// startMetricsCollection starts periodic metrics collection
func (ms *MonitoringSystem) startMetricsCollection() chan struct{} {
	stopChan := make(chan struct{})

	ms.wg.Add(1)
	go func() {
		defer ms.wg.Done()
		ticker := time.NewTicker(ms.config.MetricsInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				ms.collector.Collect()
			case <-stopChan:
				return
			case <-ms.ctx.Done():
				return
			}
		}
	}()

	return stopChan
}

// startSnapshotCapture starts periodic snapshot capture
func (ms *MonitoringSystem) startSnapshotCapture() chan struct{} {
	stopChan := make(chan struct{})

	ms.wg.Add(1)
	go func() {
		defer ms.wg.Done()
		ticker := time.NewTicker(ms.config.SnapshotInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				ms.collector.CaptureSnapshot()
			case <-stopChan:
				return
			case <-ms.ctx.Done():
				return
			}
		}
	}()

	return stopChan
}

// startCleanupRoutine starts periodic cleanup of old data
func (ms *MonitoringSystem) startCleanupRoutine() chan struct{} {
	stopChan := make(chan struct{})

	ms.wg.Add(1)
	go func() {
		defer ms.wg.Done()
		ticker := time.NewTicker(ms.config.CleanupInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				ms.collector.CleanupOldSnapshots(ms.config.RetentionPeriod)
			case <-stopChan:
				return
			case <-ms.ctx.Done():
				return
			}
		}
	}()

	return stopChan
}

// SaveReport saves a comprehensive monitoring report
func (ms *MonitoringSystem) SaveReport(filepath string) error {
	return ms.collector.SaveReport(filepath)
}

// PrintSummary displays a monitoring summary
func (ms *MonitoringSystem) PrintSummary() {
	snapshot := ms.collector.GetSnapshot()

	fmt.Printf("\n=== Monitoring System Summary ===\n")
	fmt.Printf("Uptime: %v\n", ms.Uptime())
	fmt.Printf("Status: %s\n", map[bool]string{true: "Running", false: "Stopped"}[ms.running])
	fmt.Printf("\n")

	if snapshot != nil {
		snapshot.PrintSummary()
	}

	// Print component status
	fmt.Printf("\n--- Component Status ---\n")
	fmt.Printf("Dashboard: %s\n", map[bool]string{true: "Enabled", false: "Disabled"}[ms.config.DashboardEnabled])
	fmt.Printf("Alerting: %s\n", map[bool]string{true: "Enabled", false: "Disabled"}[ms.config.AlertingEnabled])
	fmt.Printf("Prometheus: %s\n", map[bool]string{true: "Enabled", false: "Disabled"}[ms.config.PrometheusEnabled])

	if ms.alertManager != nil {
		activeAlerts := ms.alertManager.GetActiveAlerts()
		fmt.Printf("Active Alerts: %d\n", len(activeAlerts))
		if len(activeAlerts) > 0 {
			fmt.Printf("\nActive Alerts:\n")
			for _, alert := range activeAlerts {
				fmt.Printf("  - %s: %s\n", alert.Severity, alert.Message)
			}
		}
	}
}

// DefaultMonitoringConfig returns a default monitoring configuration
func DefaultMonitoringConfig() MonitoringConfig {
	return MonitoringConfig{
		Enabled:            true,
		MetricsInterval:    5 * time.Second,
		SnapshotInterval:   30 * time.Second,
		CleanupInterval:    5 * time.Minute,
		DashboardEnabled:   true,
		DashboardPort:      8080,
		DashboardRefresh:   2 * time.Second,
		AlertingEnabled:    true,
		AlertsEnabled:      true,
		AlertCheckInterval: 10 * time.Second,
		PrometheusEnabled:  true,
		PrometheusPort:     9090,
		PrometheusPath:     "/metrics",
		RetentionPeriod:    24 * time.Hour,
		MaxSnapshots:       1000,
		OutputPath:         "./monitoring",
		AlertRules: []AlertRule{
			{
				Name:        "high_latency",
				Description: "Alert when P95 latency exceeds threshold",
				Type:        AlertTypeLatency,
				Threshold:   500.0, // 500ms
				Severity:    AlertSeverityWarning,
			},
			{
				Name:        "critical_latency",
				Description: "Alert when P99 latency exceeds threshold",
				Type:        AlertTypeLatency,
				Threshold:   1000.0, // 1s
				Severity:    AlertSeverityCritical,
			},
			{
				Name:        "low_cache_hit_ratio",
				Description: "Alert when cache hit ratio drops below 60%",
				Type:        AlertTypeCacheHitRatio,
				Threshold:   0.60,
				Severity:    AlertSeverityWarning,
			},
			{
				Name:        "high_error_rate",
				Description: "Alert when error rate exceeds 5%",
				Type:        AlertTypeErrorRate,
				Threshold:   0.05,
				Severity:    AlertSeverityCritical,
			},
		},
	}
}
