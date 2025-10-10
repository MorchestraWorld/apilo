package daemon

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

//go:embed dashboard.html
var dashboardHTML string

// IPCServer handles inter-process communication via HTTP
type IPCServer struct {
	port    int
	service *Service
	server  *http.Server
}

// NewIPCServer creates a new IPC server
func NewIPCServer(port int, service *Service) *IPCServer {
	return &IPCServer{
		port:    port,
		service: service,
	}
}

// Start starts the IPC HTTP server
func (ipc *IPCServer) Start(ctx context.Context) error {
	mux := http.NewServeMux()

	// Register handlers
	mux.HandleFunc("/", ipc.handleRoot)
	mux.HandleFunc("/dashboard", ipc.handleDashboard)
	mux.HandleFunc("/optimize", ipc.handleOptimize)
	mux.HandleFunc("/status", ipc.handleStatus)
	mux.HandleFunc("/metrics", ipc.handleMetrics)
	mux.HandleFunc("/analytics", ipc.handleAnalytics)
	mux.HandleFunc("/requests", ipc.handleRequests)
	mux.HandleFunc("/cache/stats", ipc.handleCacheStats)
	mux.HandleFunc("/cache/invalidate", ipc.handleCacheInvalidate)
	mux.HandleFunc("/config", ipc.handleConfig)
	mux.HandleFunc("/health", ipc.handleHealth)
	mux.HandleFunc("/internal/record", ipc.handleInternalRecord)

	ipc.server = &http.Server{
		Addr:         fmt.Sprintf("localhost:%d", ipc.port),
		Handler:      ipc.loggingMiddleware(mux),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	errChan := make(chan error, 1)
	go func() {
		ipc.service.logger.Info("IPC server listening on %s", ipc.server.Addr)
		if err := ipc.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	// Wait for context cancellation or error
	select {
	case <-ctx.Done():
		ipc.service.logger.Info("Shutting down IPC server...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return ipc.server.Shutdown(shutdownCtx)
	case err := <-errChan:
		return err
	}
}

// handleRoot handles root endpoint requests
func (ipc *IPCServer) handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	status := ipc.service.GetStatus()

	apiDoc := map[string]interface{}{
		"service":     "Apilo Daemon",
		"version":     "2.0.0",
		"description": "API Latency Optimizer - Background daemon for automatic API optimization",
		"status": map[string]interface{}{
			"running": status.Running,
			"uptime":  status.Uptime.String(),
			"pid":     status.PID,
			"port":    status.Port,
		},
		"endpoints": map[string]string{
			"GET /":                          "API documentation (this page)",
			"GET /dashboard":                 "Web dashboard (GUI)",
			"GET /health":                    "Health check",
			"GET /status":                    "Daemon status and metrics",
			"GET /metrics":                   "Performance metrics (JSON)",
			"GET /analytics":                 "Advanced analytics data (JSON)",
			"GET /analytics?limit=100":       "Analytics with custom request limit",
			"GET /requests":                  "Request history (default: 100, max: 1000)",
			"GET /requests?limit=100":        "Paginated request history",
			"GET /cache/stats":               "Cache statistics (JSON)",
			"GET /cache/stats?format=visual": "Cache visualization (ASCII)",
			"POST /cache/invalidate":         "Clear cache",
			"GET /config":                    "Get daemon configuration",
			"PUT /config":                    "Update daemon configuration",
			"POST /optimize":                 "Optimize an API request",
			"POST /internal/record":          "Record proxy-intercepted request (internal use)",
		},
		"features": []string{
			"Persistent background process",
			"Memory-bounded cache with LRU eviction",
			"HTTP/2 connection pooling",
			"Circuit breaker pattern",
			"Real-time performance metrics",
			"Structured logging with rotation",
			"Cache visualization",
		},
		"documentation": "https://github.com/anthropics/api-latency-optimizer",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(apiDoc)
}

// handleDashboard serves the web dashboard
func (ipc *IPCServer) handleDashboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(dashboardHTML))
}

// handleOptimize handles optimization requests
func (ipc *IPCServer) handleOptimize(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req OptimizationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	resp, err := ipc.service.Optimize(&req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Optimization failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// handleStatus returns daemon status
func (ipc *IPCServer) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	status := ipc.service.GetStatus()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// handleMetrics returns performance metrics
func (ipc *IPCServer) handleMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	metrics := ipc.service.metrics.GetStats()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// handleAnalytics returns advanced analytics data
func (ipc *IPCServer) handleAnalytics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse limit parameter (default: 20, max: 1000)
	limit := 20
	if limitParam := r.URL.Query().Get("limit"); limitParam != "" {
		if parsedLimit, err := strconv.Atoi(limitParam); err == nil && parsedLimit > 0 {
			limit = parsedLimit
			if limit > 1000 {
				limit = 1000
			}
		}
	}

	snapshot := ipc.service.analytics.GetSnapshotWithLimit(limit)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(snapshot)
}

// handleRequests returns paginated request history
func (ipc *IPCServer) handleRequests(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse pagination parameters
	limit := 100
	if limitParam := r.URL.Query().Get("limit"); limitParam != "" {
		if parsedLimit, err := strconv.Atoi(limitParam); err == nil && parsedLimit > 0 {
			limit = parsedLimit
			if limit > 1000 {
				limit = 1000
			}
		}
	}

	// Get snapshot with specified limit
	snapshot := ipc.service.analytics.GetSnapshotWithLimit(limit)

	// Return just the requests with metadata
	response := map[string]interface{}{
		"requests":         snapshot.RecentRequests,
		"total":            len(snapshot.RecentRequests),
		"limit":            limit,
		"cache_efficiency": snapshot.CacheEfficiency,
		"token_savings":    snapshot.TokenSavings,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleConfig handles configuration get/update
func (ipc *IPCServer) handleConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		config := ipc.service.GetConfig()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(config)

	case http.MethodPut:
		var config DaemonConfig
		if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
			http.Error(w, fmt.Sprintf("Invalid config: %v", err), http.StatusBadRequest)
			return
		}

		if err := ipc.service.UpdateConfig(&config); err != nil {
			http.Error(w, fmt.Sprintf("Config update failed: %v", err), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "updated"})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleCacheStats returns detailed cache statistics
func (ipc *IPCServer) handleCacheStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats := ipc.service.optimizer.cache.GetStats()

	// Check if visual format is requested
	format := r.URL.Query().Get("format")
	if format == "visual" || format == "ascii" {
		visual := ipc.renderCacheVisual(stats)
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(visual))
		return
	}

	// Return JSON by default
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// renderCacheVisual creates an ASCII visualization of cache statistics
func (ipc *IPCServer) renderCacheVisual(stats *CacheStats) string {
	var sb strings.Builder

	sb.WriteString("\n╔═══════════════════════════════════════════════════════════════════╗\n")
	sb.WriteString("║                    Cache Statistics & Visualization               ║\n")
	sb.WriteString("╚═══════════════════════════════════════════════════════════════════╝\n\n")

	// Overall stats
	sb.WriteString(fmt.Sprintf("📊 Cache Overview:\n"))
	sb.WriteString(fmt.Sprintf("   Entries:      %d\n", stats.Entries))
	sb.WriteString(fmt.Sprintf("   Memory Used:  %.2f MB / %.2f MB (%.1f%%)\n",
		stats.MemoryUsedMB, stats.MemoryLimitMB, stats.MemoryPercent))
	sb.WriteString(fmt.Sprintf("   Default TTL:  %v\n\n", stats.DefaultTTL))

	// Memory usage bar
	sb.WriteString("💾 Memory Usage:\n")
	sb.WriteString("   [")
	barWidth := 50
	filledWidth := int(stats.MemoryPercent / 100.0 * float64(barWidth))
	for i := 0; i < barWidth; i++ {
		if i < filledWidth {
			sb.WriteString("█")
		} else {
			sb.WriteString("░")
		}
	}
	sb.WriteString(fmt.Sprintf("] %.1f%%\n\n", stats.MemoryPercent))

	// Entry details (limit to 10 for visualization)
	if len(stats.EntryDetails) > 0 {
		sb.WriteString("🗂️  Recent Cache Entries (Top 10):\n\n")
		sb.WriteString("   Key              Size        Age       TTL Remaining  Status\n")
		sb.WriteString("   ──────────────── ─────────── ───────── ───────────── ────────\n")

		maxEntries := 10
		if len(stats.EntryDetails) < maxEntries {
			maxEntries = len(stats.EntryDetails)
		}

		for i := 0; i < maxEntries; i++ {
			entry := stats.EntryDetails[i]
			status := "✓"
			if entry.Expired {
				status = "⊗"
			}

			sb.WriteString(fmt.Sprintf("   %-16s %-11s %-9s %-13s %s\n",
				entry.Key,
				formatBytes(entry.SizeBytes),
				formatDuration(entry.Age),
				formatDuration(entry.TTLRemaining),
				status))
		}

		if len(stats.EntryDetails) > maxEntries {
			sb.WriteString(fmt.Sprintf("\n   ... and %d more entries\n", len(stats.EntryDetails)-maxEntries))
		}
	} else {
		sb.WriteString("📭 Cache is empty\n")
	}

	sb.WriteString("\n")
	return sb.String()
}

// formatBytes formats bytes into human-readable format
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// formatDuration formats duration into human-readable format
func formatDuration(d time.Duration) string {
	if d < 0 {
		return "expired"
	}
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%.1fm", d.Minutes())
	}
	return fmt.Sprintf("%.1fh", d.Hours())
}

// handleCacheInvalidate handles cache invalidation requests
func (ipc *IPCServer) handleCacheInvalidate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ipc.service.optimizer.InvalidateCache()

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "cache invalidated"})
}

// handleHealth returns health check status
func (ipc *IPCServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	health := map[string]interface{}{
		"status":  "healthy",
		"uptime":  time.Since(ipc.service.startTime).String(),
		"version": "2.0.0",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

// handleInternalRecord receives request records from the proxy
func (ipc *IPCServer) handleInternalRecord(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var record RequestRecord
	if err := json.NewDecoder(r.Body).Decode(&record); err != nil {
		http.Error(w, fmt.Sprintf("Invalid record: %v", err), http.StatusBadRequest)
		return
	}

	// Convert timestamp from Unix seconds to time.Time
	if record.Timestamp.IsZero() {
		record.Timestamp = time.Now()
	}

	// Record in analytics
	ipc.service.analytics.RecordRequest(record)

	// Update metrics
	ipc.service.metrics.IncrementRequests()
	if record.CacheHit {
		ipc.service.metrics.IncrementCacheHits()
	} else {
		ipc.service.metrics.IncrementCacheMisses()
	}
	ipc.service.metrics.RecordLatency(time.Duration(record.Latency))

	// Track token usage if available
	if record.TotalTokens > 0 {
		ipc.service.metrics.IncrementClaudeTokens(record.InputTokens, record.OutputTokens)
	}

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"status": "recorded"})
}

// loggingMiddleware logs all HTTP requests
func (ipc *IPCServer) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response writer wrapper to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		latency := time.Since(start)
		ipc.service.logger.LogRequest(r.Method, r.URL.Path, wrapped.statusCode, latency)
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
