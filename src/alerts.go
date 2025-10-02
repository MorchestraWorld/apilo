package main

import (
	"fmt"
	"sync"
	"time"
)

// AlertSeverity defines the severity level of an alert
type AlertSeverity string

const (
	AlertSeverityInfo     AlertSeverity = "INFO"
	AlertSeverityWarning  AlertSeverity = "WARNING"
	AlertSeverityCritical AlertSeverity = "CRITICAL"
)

// AlertType defines the type of metric being monitored
type AlertType string

const (
	AlertTypeLatency        AlertType = "latency"
	AlertTypeTTFB           AlertType = "ttfb"
	AlertTypeCacheHitRatio  AlertType = "cache_hit_ratio"
	AlertTypeCacheMemory    AlertType = "cache_memory"
	AlertTypeErrorRate      AlertType = "error_rate"
	AlertTypeThroughput     AlertType = "throughput"
	AlertTypeCustom         AlertType = "custom"
)

// AlertRule defines a monitoring rule
type AlertRule struct {
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Type        AlertType     `json:"type"`
	Threshold   float64       `json:"threshold"`
	Comparator  string        `json:"comparator"` // "gt", "lt", "gte", "lte", "eq"
	Severity    AlertSeverity `json:"severity"`
	Cooldown    time.Duration `json:"cooldown"`
	Enabled     bool          `json:"enabled"`
}

// Alert represents an active or historical alert
type Alert struct {
	Rule        AlertRule     `json:"rule"`
	Timestamp   time.Time     `json:"timestamp"`
	Value       float64       `json:"value"`
	Message     string        `json:"message"`
	Severity    AlertSeverity `json:"severity"`
	Active      bool          `json:"active"`
	AckedAt     *time.Time    `json:"acknowledged_at,omitempty"`
	ResolvedAt  *time.Time    `json:"resolved_at,omitempty"`
}

// AlertManager manages alert rules and active alerts
type AlertManager struct {
	rules          []AlertRule
	activeAlerts   map[string]*Alert
	alertHistory   []Alert
	cacheMetrics   *CacheMetrics

	// Callbacks
	onAlert        func(alert *Alert)
	onResolve      func(alert *Alert)

	// Cooldown tracking
	lastTriggered  map[string]time.Time

	mu             sync.RWMutex
	maxHistory     int
}

// NewAlertManager creates a new alert manager
func NewAlertManager(rules []AlertRule) *AlertManager {
	// Set defaults for rules
	for i := range rules {
		if rules[i].Comparator == "" {
			rules[i].Comparator = "gt"
		}
		if rules[i].Cooldown == 0 {
			rules[i].Cooldown = 5 * time.Minute
		}
		if !rules[i].Enabled {
			rules[i].Enabled = true
		}
	}

	return &AlertManager{
		rules:         rules,
		activeAlerts:  make(map[string]*Alert),
		alertHistory:  make([]Alert, 0, 1000),
		lastTriggered: make(map[string]time.Time),
		maxHistory:    1000,
	}
}

// AttachCacheMetrics attaches cache metrics for monitoring
func (am *AlertManager) AttachCacheMetrics(metrics *CacheMetrics) {
	am.mu.Lock()
	defer am.mu.Unlock()
	am.cacheMetrics = metrics
}

// SetOnAlert sets a callback for when alerts are triggered
func (am *AlertManager) SetOnAlert(callback func(alert *Alert)) {
	am.mu.Lock()
	defer am.mu.Unlock()
	am.onAlert = callback
}

// SetOnResolve sets a callback for when alerts are resolved
func (am *AlertManager) SetOnResolve(callback func(alert *Alert)) {
	am.mu.Lock()
	defer am.mu.Unlock()
	am.onResolve = callback
}

// Start starts the alert checking loop
func (am *AlertManager) Start(checkInterval time.Duration) chan struct{} {
	stopChan := make(chan struct{})

	go func() {
		ticker := time.NewTicker(checkInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				am.CheckAlerts()
			case <-stopChan:
				return
			}
		}
	}()

	return stopChan
}

// CheckAlerts evaluates all rules against current metrics
func (am *AlertManager) CheckAlerts() {
	am.mu.Lock()
	defer am.mu.Unlock()

	if am.cacheMetrics == nil {
		return
	}

	for _, rule := range am.rules {
		if !rule.Enabled {
			continue
		}

		// Check cooldown
		if lastTrigger, exists := am.lastTriggered[rule.Name]; exists {
			if time.Since(lastTrigger) < rule.Cooldown {
				continue
			}
		}

		// Evaluate rule
		value, shouldAlert := am.evaluateRule(rule)
		if shouldAlert {
			am.triggerAlert(rule, value)
		} else {
			am.resolveAlert(rule.Name)
		}
	}
}

// evaluateRule evaluates a single rule against current metrics
func (am *AlertManager) evaluateRule(rule AlertRule) (float64, bool) {
	var value float64

	switch rule.Type {
	case AlertTypeLatency:
		// Use average access latency as proxy
		value = float64(am.cacheMetrics.AvgAccessLatency().Milliseconds())

	case AlertTypeTTFB:
		// Use average access latency as proxy for TTFB
		value = float64(am.cacheMetrics.AvgAccessLatency().Milliseconds())

	case AlertTypeCacheHitRatio:
		value = am.cacheMetrics.HitRatio()

	case AlertTypeCacheMemory:
		value = float64(am.cacheMetrics.CurrentMemoryUsage()) / (1024 * 1024) // Convert to MB

	case AlertTypeErrorRate:
		// Error rate would need to come from benchmark results
		// For now, we'll use 0 as cache operations don't typically error
		value = 0.0

	case AlertTypeThroughput:
		value = am.cacheMetrics.RequestsPerSecond()

	default:
		return 0, false
	}

	// Evaluate comparator
	return value, am.compare(value, rule.Threshold, rule.Comparator)
}

// compare compares a value against a threshold using the specified comparator
func (am *AlertManager) compare(value, threshold float64, comparator string) bool {
	switch comparator {
	case "gt":
		return value > threshold
	case "gte":
		return value >= threshold
	case "lt":
		return value < threshold
	case "lte":
		return value <= threshold
	case "eq":
		return value == threshold
	default:
		return false
	}
}

// triggerAlert creates and activates an alert
func (am *AlertManager) triggerAlert(rule AlertRule, value float64) {
	// Check if alert already active
	if _, exists := am.activeAlerts[rule.Name]; exists {
		return
	}

	alert := &Alert{
		Rule:      rule,
		Timestamp: time.Now(),
		Value:     value,
		Message:   am.formatAlertMessage(rule, value),
		Severity:  rule.Severity,
		Active:    true,
	}

	am.activeAlerts[rule.Name] = alert
	am.alertHistory = append(am.alertHistory, *alert)

	// Limit history size
	if len(am.alertHistory) > am.maxHistory {
		am.alertHistory = am.alertHistory[1:]
	}

	// Update cooldown tracker
	am.lastTriggered[rule.Name] = time.Now()

	// Trigger callback
	if am.onAlert != nil {
		go am.onAlert(alert)
	}

	// Log alert
	fmt.Printf("\n[%s ALERT] %s\n", alert.Severity, alert.Message)
}

// resolveAlert resolves an active alert
func (am *AlertManager) resolveAlert(ruleName string) {
	alert, exists := am.activeAlerts[ruleName]
	if !exists {
		return
	}

	now := time.Now()
	alert.Active = false
	alert.ResolvedAt = &now

	delete(am.activeAlerts, ruleName)

	// Trigger callback
	if am.onResolve != nil {
		go am.onResolve(alert)
	}

	// Log resolution
	fmt.Printf("\n[RESOLVED] %s alert: %s\n", alert.Severity, alert.Rule.Name)
}

// AcknowledgeAlert acknowledges an active alert
func (am *AlertManager) AcknowledgeAlert(ruleName string) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	alert, exists := am.activeAlerts[ruleName]
	if !exists {
		return fmt.Errorf("alert not found: %s", ruleName)
	}

	now := time.Now()
	alert.AckedAt = &now

	return nil
}

// GetActiveAlerts returns all active alerts
func (am *AlertManager) GetActiveAlerts() []*Alert {
	am.mu.RLock()
	defer am.mu.RUnlock()

	alerts := make([]*Alert, 0, len(am.activeAlerts))
	for _, alert := range am.activeAlerts {
		alerts = append(alerts, alert)
	}

	return alerts
}

// GetAlertHistory returns the alert history
func (am *AlertManager) GetAlertHistory() []Alert {
	am.mu.RLock()
	defer am.mu.RUnlock()

	history := make([]Alert, len(am.alertHistory))
	copy(history, am.alertHistory)
	return history
}

// GetAlertHistorySince returns alerts since the given time
func (am *AlertManager) GetAlertHistorySince(since time.Time) []Alert {
	am.mu.RLock()
	defer am.mu.RUnlock()

	filtered := make([]Alert, 0)
	for _, alert := range am.alertHistory {
		if alert.Timestamp.After(since) {
			filtered = append(filtered, alert)
		}
	}

	return filtered
}

// AddRule adds a new alert rule
func (am *AlertManager) AddRule(rule AlertRule) {
	am.mu.Lock()
	defer am.mu.Unlock()

	// Set defaults
	if rule.Comparator == "" {
		rule.Comparator = "gt"
	}
	if rule.Cooldown == 0 {
		rule.Cooldown = 5 * time.Minute
	}
	if !rule.Enabled {
		rule.Enabled = true
	}

	am.rules = append(am.rules, rule)
}

// RemoveRule removes an alert rule
func (am *AlertManager) RemoveRule(ruleName string) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	for i, rule := range am.rules {
		if rule.Name == ruleName {
			am.rules = append(am.rules[:i], am.rules[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("rule not found: %s", ruleName)
}

// EnableRule enables an alert rule
func (am *AlertManager) EnableRule(ruleName string) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	for i, rule := range am.rules {
		if rule.Name == ruleName {
			am.rules[i].Enabled = true
			return nil
		}
	}

	return fmt.Errorf("rule not found: %s", ruleName)
}

// DisableRule disables an alert rule
func (am *AlertManager) DisableRule(ruleName string) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	for i, rule := range am.rules {
		if rule.Name == ruleName {
			am.rules[i].Enabled = false
			// Also resolve any active alerts for this rule
			am.resolveAlert(ruleName)
			return nil
		}
	}

	return fmt.Errorf("rule not found: %s", ruleName)
}

// GetRules returns all configured rules
func (am *AlertManager) GetRules() []AlertRule {
	am.mu.RLock()
	defer am.mu.RUnlock()

	rules := make([]AlertRule, len(am.rules))
	copy(rules, am.rules)
	return rules
}

// GetSummary returns a summary of alert status
func (am *AlertManager) GetSummary() map[string]interface{} {
	am.mu.RLock()
	defer am.mu.RUnlock()

	activeCount := len(am.activeAlerts)
	criticalCount := 0
	warningCount := 0
	infoCount := 0

	for _, alert := range am.activeAlerts {
		switch alert.Severity {
		case AlertSeverityCritical:
			criticalCount++
		case AlertSeverityWarning:
			warningCount++
		case AlertSeverityInfo:
			infoCount++
		}
	}

	return map[string]interface{}{
		"total_rules":     len(am.rules),
		"active_alerts":   activeCount,
		"critical_alerts": criticalCount,
		"warning_alerts":  warningCount,
		"info_alerts":     infoCount,
		"total_history":   len(am.alertHistory),
	}
}

// PrintSummary displays a summary of alert status
func (am *AlertManager) PrintSummary() {
	summary := am.GetSummary()

	fmt.Printf("\n=== Alert Manager Summary ===\n")
	fmt.Printf("Total Rules: %d\n", summary["total_rules"])
	fmt.Printf("Active Alerts: %d\n", summary["active_alerts"])
	fmt.Printf("  Critical: %d\n", summary["critical_alerts"])
	fmt.Printf("  Warning: %d\n", summary["warning_alerts"])
	fmt.Printf("  Info: %d\n", summary["info_alerts"])
	fmt.Printf("Alert History: %d events\n", summary["total_history"])

	activeAlerts := am.GetActiveAlerts()
	if len(activeAlerts) > 0 {
		fmt.Printf("\nActive Alerts:\n")
		for _, alert := range activeAlerts {
			fmt.Printf("  [%s] %s: %s\n", alert.Severity, alert.Rule.Name, alert.Message)
		}
	}
}

// formatAlertMessage creates a human-readable alert message
func (am *AlertManager) formatAlertMessage(rule AlertRule, value float64) string {
	var unit string
	var formattedValue string

	switch rule.Type {
	case AlertTypeLatency, AlertTypeTTFB:
		unit = "ms"
		formattedValue = fmt.Sprintf("%.2f", value)
	case AlertTypeCacheHitRatio:
		unit = "%"
		formattedValue = fmt.Sprintf("%.2f", value*100)
	case AlertTypeCacheMemory:
		unit = "MB"
		formattedValue = fmt.Sprintf("%.2f", value)
	case AlertTypeErrorRate:
		unit = "%"
		formattedValue = fmt.Sprintf("%.2f", value*100)
	case AlertTypeThroughput:
		unit = "req/s"
		formattedValue = fmt.Sprintf("%.2f", value)
	default:
		unit = ""
		formattedValue = fmt.Sprintf("%.2f", value)
	}

	threshold := rule.Threshold
	if rule.Type == AlertTypeCacheHitRatio || rule.Type == AlertTypeErrorRate {
		threshold *= 100
	}

	return fmt.Sprintf("%s: %s %s %s (threshold: %.2f %s)",
		rule.Description,
		formattedValue,
		unit,
		rule.Comparator,
		threshold,
		unit)
}

// DefaultAlertRules returns a set of sensible default alert rules
func DefaultAlertRules() []AlertRule {
	return []AlertRule{
		{
			Name:        "high_latency",
			Description: "High P95 latency detected",
			Type:        AlertTypeLatency,
			Threshold:   500.0, // 500ms
			Comparator:  "gt",
			Severity:    AlertSeverityWarning,
			Cooldown:    5 * time.Minute,
			Enabled:     true,
		},
		{
			Name:        "critical_latency",
			Description: "Critical P99 latency detected",
			Type:        AlertTypeLatency,
			Threshold:   1000.0, // 1s
			Comparator:  "gt",
			Severity:    AlertSeverityCritical,
			Cooldown:    5 * time.Minute,
			Enabled:     true,
		},
		{
			Name:        "low_cache_hit_ratio",
			Description: "Cache hit ratio below threshold",
			Type:        AlertTypeCacheHitRatio,
			Threshold:   0.60, // 60%
			Comparator:  "lt",
			Severity:    AlertSeverityWarning,
			Cooldown:    10 * time.Minute,
			Enabled:     true,
		},
		{
			Name:        "critical_cache_hit_ratio",
			Description: "Cache hit ratio critically low",
			Type:        AlertTypeCacheHitRatio,
			Threshold:   0.40, // 40%
			Comparator:  "lt",
			Severity:    AlertSeverityCritical,
			Cooldown:    10 * time.Minute,
			Enabled:     true,
		},
		{
			Name:        "high_cache_memory",
			Description: "Cache memory usage high",
			Type:        AlertTypeCacheMemory,
			Threshold:   500.0, // 500MB
			Comparator:  "gt",
			Severity:    AlertSeverityWarning,
			Cooldown:    15 * time.Minute,
			Enabled:     true,
		},
		{
			Name:        "high_error_rate",
			Description: "Error rate exceeds threshold",
			Type:        AlertTypeErrorRate,
			Threshold:   0.05, // 5%
			Comparator:  "gt",
			Severity:    AlertSeverityCritical,
			Cooldown:    5 * time.Minute,
			Enabled:     true,
		},
		{
			Name:        "low_throughput",
			Description: "Throughput below expected rate",
			Type:        AlertTypeThroughput,
			Threshold:   10.0, // 10 req/s
			Comparator:  "lt",
			Severity:    AlertSeverityWarning,
			Cooldown:    10 * time.Minute,
			Enabled:     false, // Disabled by default
		},
	}
}
