package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/http/httptrace"
	"os"
	"sort"
	"sync"
	"time"
)

// LatencyMetrics captures detailed timing information for a single request
type LatencyMetrics struct {
	// Connection timing
	DNSLookup         time.Duration `json:"dns_lookup"`
	TCPConnection     time.Duration `json:"tcp_connection"`
	TLSHandshake      time.Duration `json:"tls_handshake"`
	ServerProcessing  time.Duration `json:"server_processing"`
	ContentTransfer   time.Duration `json:"content_transfer"`

	// Total times
	TotalLatency      time.Duration `json:"total_latency"`
	TimeToFirstByte   time.Duration `json:"time_to_first_byte"`

	// Response metadata
	StatusCode        int           `json:"status_code"`
	ResponseSize      int64         `json:"response_size_bytes"`
	Timestamp         time.Time     `json:"timestamp"`

	// Error tracking
	Error             string        `json:"error,omitempty"`
}

// BenchmarkResult contains aggregated statistics from multiple requests
type BenchmarkResult struct {
	// Test configuration
	TargetURL         string        `json:"target_url"`
	TotalRequests     int           `json:"total_requests"`
	Concurrency       int           `json:"concurrency"`
	Duration          time.Duration `json:"duration"`
	StartTime         time.Time     `json:"start_time"`
	EndTime           time.Time     `json:"end_time"`

	// Success/failure tracking
	SuccessfulReqs    int           `json:"successful_requests"`
	FailedReqs        int           `json:"failed_requests"`

	// Throughput metrics
	RequestsPerSecond float64       `json:"requests_per_second"`
	BytesPerSecond    float64       `json:"bytes_per_second"`

	// Latency statistics (all in milliseconds for readability)
	Latency           LatencyStats  `json:"latency"`           // Alias for LatencyStats
	LatencyStats      LatencyStats  `json:"latency_stats"`
	TTFBStats         LatencyStats  `json:"ttfb_stats"`
	ConnectionStats   LatencyStats  `json:"connection_stats"`
	TLSStats          LatencyStats  `json:"tls_stats"`

	// Throughput alias
	Throughput        ThroughputStats `json:"throughput"`

	// Success rate
	SuccessRate       float64 `json:"success_rate"`

	// Raw data for detailed analysis
	RawMetrics        []LatencyMetrics `json:"raw_metrics,omitempty"`
}

// LatencyStats provides statistical analysis for a timing metric
// MOVED TO types.go
// type LatencyStats struct {
// 	Min               float64       `json:"min_ms"`
// 	Max               float64       `json:"max_ms"`
// 	Mean              float64       `json:"mean_ms"`
// 	Median            float64       `json:"median_ms"`
// 	P50               float64       `json:"p50_ms"`
// 	P95               float64       `json:"p95_ms"`
// 	P99               float64       `json:"p99_ms"`
// 	StdDev            float64       `json:"std_dev_ms"`
// 	Samples           int           `json:"samples"`
// }

// BenchmarkConfig defines the parameters for a benchmark run
// MOVED TO types.go
// type BenchmarkConfig struct {
// 	TargetURL         string
// 	TotalRequests     int
// 	Concurrency       int
// 	Timeout           time.Duration
// 	KeepAlive         bool
// 	IncludeRawMetrics bool
// 	CustomHeaders     map[string]string
// 	Method            string
// 	Body              []byte
// }

// Benchmarker orchestrates the benchmarking process
type Benchmarker struct {
	config     BenchmarkConfig
	client     *http.Client
	metrics    []LatencyMetrics
	metricsMux sync.Mutex
}

// NewBenchmarker creates a new benchmarker with the given configuration
func NewBenchmarker(config BenchmarkConfig) *Benchmarker {
	// Set defaults
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.Method == "" {
		config.Method = "GET"
	}
	if config.Concurrency <= 0 {
		config.Concurrency = 1
	}
	if config.TotalRequests <= 0 {
		config.TotalRequests = 100
	}

	// Create optimized HTTP client
	transport := &http.Transport{
		MaxIdleConns:        config.Concurrency,
		MaxIdleConnsPerHost: config.Concurrency,
		IdleConnTimeout:     90 * time.Second,
		DisableKeepAlives:   !config.KeepAlive,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: false,
		},
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   config.Timeout,
	}

	return &Benchmarker{
		config:  config,
		client:  client,
		metrics: make([]LatencyMetrics, 0, config.TotalRequests),
	}
}

// Run executes the benchmark and returns aggregated results
func (b *Benchmarker) Run(ctx context.Context) (*BenchmarkResult, error) {
	startTime := time.Now()

	// Create work queue
	requestQueue := make(chan int, b.config.TotalRequests)
	for i := 0; i < b.config.TotalRequests; i++ {
		requestQueue <- i
	}
	close(requestQueue)

	// Launch worker goroutines
	var wg sync.WaitGroup
	for i := 0; i < b.config.Concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			b.worker(ctx, requestQueue)
		}(i)
	}

	// Wait for all requests to complete
	wg.Wait()
	endTime := time.Now()

	// Calculate statistics
	result := b.calculateResults(startTime, endTime)

	return result, nil
}

// worker processes requests from the queue
func (b *Benchmarker) worker(ctx context.Context, queue <-chan int) {
	for requestID := range queue {
		select {
		case <-ctx.Done():
			return
		default:
			metric := b.measureRequest(ctx, requestID)
			b.metricsMux.Lock()
			b.metrics = append(b.metrics, metric)
			b.metricsMux.Unlock()
		}
	}
}

// measureRequest performs a single request and captures all timing metrics
func (b *Benchmarker) measureRequest(ctx context.Context, requestID int) LatencyMetrics {
	metric := LatencyMetrics{
		Timestamp: time.Now(),
	}

	// Timing markers
	var dnsStart, connectStart, tlsStart, reqStart, firstByteTime time.Time
	var dnsDone, connectDone, tlsDone time.Time

	// Create request with tracing
	req, err := http.NewRequestWithContext(ctx, b.config.Method, b.config.TargetURL, nil)
	if err != nil {
		metric.Error = fmt.Sprintf("request creation failed: %v", err)
		return metric
	}

	// Add custom headers
	for key, value := range b.config.CustomHeaders {
		req.Header.Set(key, value)
	}

	// Create trace to capture timing events
	trace := &httptrace.ClientTrace{
		DNSStart: func(_ httptrace.DNSStartInfo) {
			dnsStart = time.Now()
		},
		DNSDone: func(_ httptrace.DNSDoneInfo) {
			dnsDone = time.Now()
		},
		ConnectStart: func(_, _ string) {
			connectStart = time.Now()
			if dnsStart.IsZero() {
				dnsStart = connectStart
				dnsDone = connectStart
			}
		},
		ConnectDone: func(_, _ string, err error) {
			connectDone = time.Now()
			if err != nil {
				metric.Error = fmt.Sprintf("connection failed: %v", err)
			}
		},
		TLSHandshakeStart: func() {
			tlsStart = time.Now()
		},
		TLSHandshakeDone: func(_ tls.ConnectionState, err error) {
			tlsDone = time.Now()
			if err != nil {
				metric.Error = fmt.Sprintf("TLS handshake failed: %v", err)
			}
		},
		GotFirstResponseByte: func() {
			firstByteTime = time.Now()
		},
	}

	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))

	// Execute request
	reqStart = time.Now()
	resp, err := b.client.Do(req)
	if err != nil {
		metric.Error = fmt.Sprintf("request failed: %v", err)
		metric.TotalLatency = time.Since(reqStart)
		return metric
	}
	defer resp.Body.Close()

	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	responseComplete := time.Now()

	if err != nil {
		metric.Error = fmt.Sprintf("response read failed: %v", err)
	}

	// Calculate timing metrics
	metric.StatusCode = resp.StatusCode
	metric.ResponseSize = int64(len(bodyBytes))
	metric.TotalLatency = responseComplete.Sub(reqStart)

	if !dnsStart.IsZero() && !dnsDone.IsZero() {
		metric.DNSLookup = dnsDone.Sub(dnsStart)
	}

	if !connectStart.IsZero() && !connectDone.IsZero() {
		metric.TCPConnection = connectDone.Sub(connectStart)
	}

	if !tlsStart.IsZero() && !tlsDone.IsZero() {
		metric.TLSHandshake = tlsDone.Sub(tlsStart)
	}

	if !firstByteTime.IsZero() {
		metric.TimeToFirstByte = firstByteTime.Sub(reqStart)
		metric.ServerProcessing = firstByteTime.Sub(reqStart) - metric.DNSLookup - metric.TCPConnection - metric.TLSHandshake
		metric.ContentTransfer = responseComplete.Sub(firstByteTime)
	}

	return metric
}

// calculateResults aggregates metrics into statistical summary
func (b *Benchmarker) calculateResults(startTime, endTime time.Time) *BenchmarkResult {
	result := &BenchmarkResult{
		TargetURL:     b.config.TargetURL,
		TotalRequests: b.config.TotalRequests,
		Concurrency:   b.config.Concurrency,
		StartTime:     startTime,
		EndTime:       endTime,
		Duration:      endTime.Sub(startTime),
	}

	// Separate successful and failed requests
	var totalLatencies []float64
	var ttfbLatencies []float64
	var connectionLatencies []float64
	var tlsLatencies []float64
	var totalBytes int64

	for _, m := range b.metrics {
		if m.Error != "" {
			result.FailedReqs++
			continue
		}

		result.SuccessfulReqs++
		totalBytes += m.ResponseSize

		totalLatencies = append(totalLatencies, float64(m.TotalLatency.Microseconds())/1000.0)

		if m.TimeToFirstByte > 0 {
			ttfbLatencies = append(ttfbLatencies, float64(m.TimeToFirstByte.Microseconds())/1000.0)
		}

		if m.TCPConnection > 0 {
			connectionLatencies = append(connectionLatencies, float64(m.TCPConnection.Microseconds())/1000.0)
		}

		if m.TLSHandshake > 0 {
			tlsLatencies = append(tlsLatencies, float64(m.TLSHandshake.Microseconds())/1000.0)
		}
	}

	// Calculate throughput
	durationSecs := result.Duration.Seconds()
	if durationSecs > 0 {
		result.RequestsPerSecond = float64(result.SuccessfulReqs) / durationSecs
		result.BytesPerSecond = float64(totalBytes) / durationSecs
	}

	// Calculate statistics for each metric
	result.LatencyStats = CalculateStats(totalLatencies)
	result.TTFBStats = CalculateStats(ttfbLatencies)
	result.ConnectionStats = CalculateStats(connectionLatencies)
	result.TLSStats = CalculateStats(tlsLatencies)

	// Include raw metrics if requested
	if b.config.IncludeRawMetrics {
		result.RawMetrics = b.metrics
	}

	return result
}

// CalculateStats computes statistical measures for a set of values
func CalculateStats(values []float64) LatencyStats {
	stats := LatencyStats{
		Samples: len(values),
	}

	if len(values) == 0 {
		return stats
	}

	// Sort for percentile calculations
	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)

	// Min and Max
	stats.Min = sorted[0]
	stats.Max = sorted[len(sorted)-1]

	// Mean
	var sum float64
	for _, v := range values {
		sum += v
	}
	stats.Mean = sum / float64(len(values))

	// Percentiles
	stats.P50 = percentile(sorted, 50)
	stats.Median = stats.P50
	stats.P95 = percentile(sorted, 95)
	stats.P99 = percentile(sorted, 99)

	// Standard deviation
	var variance float64
	for _, v := range values {
		diff := v - stats.Mean
		variance += diff * diff
	}
	variance /= float64(len(values))
	stats.StdDev = math.Sqrt(variance)

	return stats
}

// percentile calculates the nth percentile from sorted values
func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}

	index := (p / 100.0) * float64(len(sorted)-1)
	lower := int(math.Floor(index))
	upper := int(math.Ceil(index))

	if lower == upper {
		return sorted[lower]
	}

	// Linear interpolation
	weight := index - float64(lower)
	return sorted[lower]*(1-weight) + sorted[upper]*weight
}

// SaveJSON writes the benchmark results to a JSON file
func (r *BenchmarkResult) SaveJSON(path string) error {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal results: %w", err)
	}

	return os.WriteFile(path, data, 0644)
}

// PrintSummary displays a human-readable summary of the results
func (r *BenchmarkResult) PrintSummary() {
	fmt.Printf("\n=== Benchmark Results ===\n")
	fmt.Printf("Target URL: %s\n", r.TargetURL)
	fmt.Printf("Total Duration: %v\n", r.Duration)
	fmt.Printf("Total Requests: %d\n", r.TotalRequests)
	fmt.Printf("Successful: %d | Failed: %d\n", r.SuccessfulReqs, r.FailedReqs)
	fmt.Printf("Concurrency: %d\n", r.Concurrency)
	fmt.Printf("\n--- Throughput ---\n")
	fmt.Printf("Requests/sec: %.2f\n", r.RequestsPerSecond)
	fmt.Printf("Bytes/sec: %.2f (%.2f KB/s)\n", r.BytesPerSecond, r.BytesPerSecond/1024)

	fmt.Printf("\n--- Total Latency Statistics ---\n")
	printLatencyStats(r.LatencyStats)

	fmt.Printf("\n--- Time to First Byte (TTFB) ---\n")
	printLatencyStats(r.TTFBStats)

	fmt.Printf("\n--- TCP Connection Time ---\n")
	printLatencyStats(r.ConnectionStats)

	fmt.Printf("\n--- TLS Handshake Time ---\n")
	printLatencyStats(r.TLSStats)
}

func printLatencyStats(stats LatencyStats) {
	if stats.Samples == 0 {
		fmt.Printf("No data available\n")
		return
	}

	fmt.Printf("Min: %.2f ms\n", stats.Min)
	fmt.Printf("P50: %.2f ms\n", stats.P50)
	fmt.Printf("P95: %.2f ms\n", stats.P95)
	fmt.Printf("P99: %.2f ms\n", stats.P99)
	fmt.Printf("Max: %.2f ms\n", stats.Max)
	fmt.Printf("Mean: %.2f ms\n", stats.Mean)
	fmt.Printf("Std Dev: %.2f ms\n", stats.StdDev)
	fmt.Printf("Samples: %d\n", stats.Samples)
}
