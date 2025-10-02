package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"sync"
	"time"
)

// Dashboard provides a real-time web interface for monitoring
type Dashboard struct {
	port            int
	refreshInterval time.Duration
	collector       *MetricsCollector

	server  *http.Server
	mu      sync.RWMutex
	running bool
}

// NewDashboard creates a new dashboard instance
func NewDashboard(port int, refreshInterval time.Duration) *Dashboard {
	return &Dashboard{
		port:            port,
		refreshInterval: refreshInterval,
	}
}

// Start starts the dashboard HTTP server
func (d *Dashboard) Start(collector *MetricsCollector) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.running {
		return fmt.Errorf("dashboard already running")
	}

	d.collector = collector

	mux := http.NewServeMux()
	mux.HandleFunc("/", d.handleIndex)
	mux.HandleFunc("/api/current", d.handleAPICurrent)
	mux.HandleFunc("/api/snapshots", d.handleAPISnapshots)
	mux.HandleFunc("/api/summary", d.handleAPISummary)
	mux.HandleFunc("/api/trends", d.handleAPITrends)

	d.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", d.port),
		Handler: mux,
	}

	go func() {
		if err := d.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Dashboard server error: %v\n", err)
		}
	}()

	d.running = true
	return nil
}

// Stop stops the dashboard HTTP server
func (d *Dashboard) Stop() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.running {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := d.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("dashboard shutdown error: %w", err)
	}

	d.running = false
	return nil
}

// handleIndex serves the main dashboard HTML page
func (d *Dashboard) handleIndex(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.New("dashboard").Parse(dashboardHTML))

	data := map[string]interface{}{
		"RefreshInterval": d.refreshInterval.Milliseconds(),
		"Port":            d.port,
	}

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// handleAPICurrent returns the current metrics snapshot
func (d *Dashboard) handleAPICurrent(w http.ResponseWriter, r *http.Request) {
	snapshot := d.collector.GetSnapshot()
	if snapshot == nil {
		http.Error(w, "No metrics available", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(snapshot)
}

// handleAPISnapshots returns historical snapshots
func (d *Dashboard) handleAPISnapshots(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	durationParam := r.URL.Query().Get("duration")
	var snapshots []MonitoringSnapshot

	if durationParam != "" {
		dur, err := time.ParseDuration(durationParam)
		if err == nil {
			since := time.Now().Add(-dur)
			snapshots = d.collector.GetSnapshotsSince(since)
		}
	}

	if len(snapshots) == 0 {
		snapshots = d.collector.GetSnapshots()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(snapshots)
}

// handleAPISummary returns a metrics summary
func (d *Dashboard) handleAPISummary(w http.ResponseWriter, r *http.Request) {
	summary := d.collector.GetMetricsSummary()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}

// handleAPITrends returns trend analysis
func (d *Dashboard) handleAPITrends(w http.ResponseWriter, r *http.Request) {
	// Parse duration parameter
	durationParam := r.URL.Query().Get("duration")
	var analysisDuration time.Duration = 1 * time.Hour

	if durationParam != "" {
		parsed, err := time.ParseDuration(durationParam)
		if err == nil {
			analysisDuration = parsed
		}
	}

	trends := d.collector.GetTrendAnalysis(analysisDuration)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(trends)
}

// dashboardHTML is the HTML template for the dashboard
const dashboardHTML = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>API Latency Optimizer - Monitoring Dashboard</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: #333;
            padding: 20px;
        }
        .container {
            max-width: 1400px;
            margin: 0 auto;
        }
        header {
            background: white;
            padding: 20px 30px;
            border-radius: 10px;
            box-shadow: 0 4px 6px rgba(0,0,0,0.1);
            margin-bottom: 20px;
        }
        h1 {
            color: #667eea;
            font-size: 28px;
            margin-bottom: 5px;
        }
        .subtitle {
            color: #666;
            font-size: 14px;
        }
        .grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
            gap: 20px;
            margin-bottom: 20px;
        }
        .card {
            background: white;
            border-radius: 10px;
            padding: 20px;
            box-shadow: 0 4px 6px rgba(0,0,0,0.1);
        }
        .card h2 {
            font-size: 18px;
            color: #667eea;
            margin-bottom: 15px;
            border-bottom: 2px solid #667eea;
            padding-bottom: 10px;
        }
        .metric {
            display: flex;
            justify-content: space-between;
            padding: 10px 0;
            border-bottom: 1px solid #eee;
        }
        .metric:last-child {
            border-bottom: none;
        }
        .metric-label {
            color: #666;
            font-size: 14px;
        }
        .metric-value {
            font-weight: bold;
            color: #333;
            font-size: 16px;
        }
        .metric-value.good {
            color: #10b981;
        }
        .metric-value.warning {
            color: #f59e0b;
        }
        .metric-value.critical {
            color: #ef4444;
        }
        .grade {
            font-size: 48px;
            font-weight: bold;
            text-align: center;
            padding: 20px;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            border-radius: 10px;
            margin: 10px 0;
        }
        .chart-container {
            grid-column: 1 / -1;
            height: 400px;
            background: white;
            border-radius: 10px;
            padding: 20px;
            box-shadow: 0 4px 6px rgba(0,0,0,0.1);
        }
        .status-indicator {
            display: inline-block;
            width: 10px;
            height: 10px;
            border-radius: 50%;
            margin-right: 8px;
        }
        .status-indicator.active {
            background: #10b981;
            box-shadow: 0 0 10px #10b981;
        }
        .status-indicator.inactive {
            background: #ef4444;
        }
        .refresh-info {
            text-align: right;
            color: #666;
            font-size: 12px;
            margin-top: 10px;
        }
        canvas {
            max-width: 100%;
            height: 350px;
        }
        .loading {
            text-align: center;
            padding: 40px;
            color: #666;
        }
    </style>
    <script src="https://cdn.jsdelivr.net/npm/chart.js@3.9.1/dist/chart.min.js"></script>
</head>
<body>
    <div class="container">
        <header>
            <h1><span class="status-indicator active"></span>API Latency Optimizer</h1>
            <div class="subtitle">Real-time Performance Monitoring Dashboard</div>
            <div class="refresh-info">Auto-refresh: {{ .RefreshInterval }}ms | Last update: <span id="lastUpdate">--</span></div>
        </header>

        <div class="grid">
            <!-- Cache Performance Card -->
            <div class="card">
                <h2>Cache Performance</h2>
                <div class="metric">
                    <span class="metric-label">Hit Ratio</span>
                    <span class="metric-value" id="cacheHitRatio">--</span>
                </div>
                <div class="metric">
                    <span class="metric-label">Size / Capacity</span>
                    <span class="metric-value" id="cacheSize">--</span>
                </div>
                <div class="metric">
                    <span class="metric-label">Memory Usage</span>
                    <span class="metric-value" id="cacheMemory">--</span>
                </div>
                <div class="metric">
                    <span class="metric-label">Total Gets</span>
                    <span class="metric-value" id="cacheTotalGets">--</span>
                </div>
                <div class="metric">
                    <span class="metric-label">Evictions</span>
                    <span class="metric-value" id="cacheEvictions">--</span>
                </div>
            </div>

            <!-- Latency Metrics Card -->
            <div class="card">
                <h2>Latency Statistics</h2>
                <div class="metric">
                    <span class="metric-label">P50 (Median)</span>
                    <span class="metric-value" id="latencyP50">--</span>
                </div>
                <div class="metric">
                    <span class="metric-label">P95</span>
                    <span class="metric-value" id="latencyP95">--</span>
                </div>
                <div class="metric">
                    <span class="metric-label">P99</span>
                    <span class="metric-value" id="latencyP99">--</span>
                </div>
                <div class="metric">
                    <span class="metric-label">Mean</span>
                    <span class="metric-value" id="latencyMean">--</span>
                </div>
                <div class="metric">
                    <span class="metric-label">Max</span>
                    <span class="metric-value" id="latencyMax">--</span>
                </div>
            </div>

            <!-- Throughput & Reliability Card -->
            <div class="card">
                <h2>Throughput & Reliability</h2>
                <div class="metric">
                    <span class="metric-label">Requests/sec</span>
                    <span class="metric-value" id="throughputRPS">--</span>
                </div>
                <div class="metric">
                    <span class="metric-label">Bytes/sec</span>
                    <span class="metric-value" id="throughputBPS">--</span>
                </div>
                <div class="metric">
                    <span class="metric-label">Error Rate</span>
                    <span class="metric-value" id="errorRate">--</span>
                </div>
                <div class="metric">
                    <span class="metric-label">Connection Reuse</span>
                    <span class="metric-value" id="connReuse">--</span>
                </div>
                <div class="metric">
                    <span class="metric-label">Uptime</span>
                    <span class="metric-value" id="uptime">--</span>
                </div>
            </div>

            <!-- Performance Grade Card -->
            <div class="card">
                <h2>Overall Performance</h2>
                <div class="grade" id="performanceGrade">--</div>
                <div class="metric">
                    <span class="metric-label">Score</span>
                    <span class="metric-value" id="performanceScore">--</span>
                </div>
            </div>
        </div>

        <!-- Latency Chart -->
        <div class="chart-container">
            <h2>Latency Trends (Last Hour)</h2>
            <canvas id="latencyChart"></canvas>
        </div>

        <!-- Cache Performance Chart -->
        <div class="chart-container">
            <h2>Cache Hit Ratio Trends</h2>
            <canvas id="cacheChart"></canvas>
        </div>
    </div>

    <script>
        let latencyChart, cacheChart;
        const snapshotHistory = [];
        const maxDataPoints = 60;

        function initCharts() {
            const latencyCtx = document.getElementById('latencyChart').getContext('2d');
            latencyChart = new Chart(latencyCtx, {
                type: 'line',
                data: {
                    labels: [],
                    datasets: [
                        {
                            label: 'P50',
                            data: [],
                            borderColor: '#10b981',
                            backgroundColor: 'rgba(16, 185, 129, 0.1)',
                            tension: 0.4
                        },
                        {
                            label: 'P95',
                            data: [],
                            borderColor: '#f59e0b',
                            backgroundColor: 'rgba(245, 158, 11, 0.1)',
                            tension: 0.4
                        },
                        {
                            label: 'P99',
                            data: [],
                            borderColor: '#ef4444',
                            backgroundColor: 'rgba(239, 68, 68, 0.1)',
                            tension: 0.4
                        }
                    ]
                },
                options: {
                    responsive: true,
                    maintainAspectRatio: false,
                    scales: {
                        y: {
                            beginAtZero: true,
                            title: {
                                display: true,
                                text: 'Latency (ms)'
                            }
                        }
                    }
                }
            });

            const cacheCtx = document.getElementById('cacheChart').getContext('2d');
            cacheChart = new Chart(cacheCtx, {
                type: 'line',
                data: {
                    labels: [],
                    datasets: [
                        {
                            label: 'Hit Ratio',
                            data: [],
                            borderColor: '#667eea',
                            backgroundColor: 'rgba(102, 126, 234, 0.1)',
                            tension: 0.4,
                            yAxisID: 'y'
                        },
                        {
                            label: 'Memory Usage (MB)',
                            data: [],
                            borderColor: '#764ba2',
                            backgroundColor: 'rgba(118, 75, 162, 0.1)',
                            tension: 0.4,
                            yAxisID: 'y1'
                        }
                    ]
                },
                options: {
                    responsive: true,
                    maintainAspectRatio: false,
                    scales: {
                        y: {
                            type: 'linear',
                            display: true,
                            position: 'left',
                            min: 0,
                            max: 1,
                            title: {
                                display: true,
                                text: 'Hit Ratio'
                            }
                        },
                        y1: {
                            type: 'linear',
                            display: true,
                            position: 'right',
                            grid: {
                                drawOnChartArea: false,
                            },
                            title: {
                                display: true,
                                text: 'Memory (MB)'
                            }
                        }
                    }
                }
            });
        }

        function updateMetrics(data) {
            // Update cache metrics
            document.getElementById('cacheHitRatio').textContent = (data.cache_hit_ratio * 100).toFixed(2) + '%';
            document.getElementById('cacheHitRatio').className = 'metric-value ' +
                (data.cache_hit_ratio >= 0.7 ? 'good' : data.cache_hit_ratio >= 0.5 ? 'warning' : 'critical');

            document.getElementById('cacheSize').textContent = data.cache_size + ' / ' + data.cache_capacity;
            document.getElementById('cacheMemory').textContent = data.cache_memory_usage_mb.toFixed(2) + ' MB';
            document.getElementById('cacheTotalGets').textContent = data.cache_total_gets.toLocaleString();
            document.getElementById('cacheEvictions').textContent = data.cache_evictions.toLocaleString();

            // Update latency metrics
            document.getElementById('latencyP50').textContent = data.latency_p50_ms.toFixed(2) + ' ms';
            document.getElementById('latencyP95').textContent = data.latency_p95_ms.toFixed(2) + ' ms';
            document.getElementById('latencyP95').className = 'metric-value ' +
                (data.latency_p95_ms < 200 ? 'good' : data.latency_p95_ms < 500 ? 'warning' : 'critical');

            document.getElementById('latencyP99').textContent = data.latency_p99_ms.toFixed(2) + ' ms';
            document.getElementById('latencyP99').className = 'metric-value ' +
                (data.latency_p99_ms < 500 ? 'good' : data.latency_p99_ms < 1000 ? 'warning' : 'critical');

            document.getElementById('latencyMean').textContent = data.latency_mean_ms.toFixed(2) + ' ms';
            document.getElementById('latencyMax').textContent = data.latency_max_ms.toFixed(2) + ' ms';

            // Update throughput metrics
            document.getElementById('throughputRPS').textContent = data.requests_per_second.toFixed(2);
            document.getElementById('throughputBPS').textContent = (data.bytes_per_second / 1024).toFixed(2) + ' KB/s';
            document.getElementById('errorRate').textContent = (data.error_rate * 100).toFixed(2) + '%';
            document.getElementById('errorRate').className = 'metric-value ' +
                (data.error_rate === 0 ? 'good' : data.error_rate < 0.05 ? 'warning' : 'critical');

            document.getElementById('connReuse').textContent = (data.connection_reuse_rate * 100).toFixed(2) + '%';
            document.getElementById('uptime').textContent = formatUptime(data.uptime_seconds);

            // Update performance grade
            document.getElementById('performanceGrade').textContent = data.performance_grade || '--';
            document.getElementById('performanceScore').textContent = data.performance_score + ' / 100';

            // Update timestamp
            document.getElementById('lastUpdate').textContent = new Date().toLocaleTimeString();

            // Update charts
            updateCharts(data);
        }

        function updateCharts(data) {
            const timestamp = new Date(data.timestamp).toLocaleTimeString();

            // Update latency chart
            latencyChart.data.labels.push(timestamp);
            latencyChart.data.datasets[0].data.push(data.latency_p50_ms);
            latencyChart.data.datasets[1].data.push(data.latency_p95_ms);
            latencyChart.data.datasets[2].data.push(data.latency_p99_ms);

            // Update cache chart
            cacheChart.data.labels.push(timestamp);
            cacheChart.data.datasets[0].data.push(data.cache_hit_ratio);
            cacheChart.data.datasets[1].data.push(data.cache_memory_usage_mb);

            // Limit data points
            if (latencyChart.data.labels.length > maxDataPoints) {
                latencyChart.data.labels.shift();
                latencyChart.data.datasets.forEach(ds => ds.data.shift());
                cacheChart.data.labels.shift();
                cacheChart.data.datasets.forEach(ds => ds.data.shift());
            }

            latencyChart.update('none');
            cacheChart.update('none');
        }

        function formatUptime(seconds) {
            const hours = Math.floor(seconds / 3600);
            const minutes = Math.floor((seconds % 3600) / 60);
            const secs = Math.floor(seconds % 60);
            return hours + 'h ' + minutes + 'm ' + secs + 's';
        }

        async function fetchMetrics() {
            try {
                const response = await fetch('/api/current');
                if (response.ok) {
                    const data = await response.json();
                    updateMetrics(data);
                }
            } catch (error) {
                console.error('Failed to fetch metrics:', error);
            }
        }

        // Initialize
        initCharts();
        fetchMetrics();
        setInterval(fetchMetrics, {{ .RefreshInterval }});
    </script>
</body>
</html>
`
