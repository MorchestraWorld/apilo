package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	main "api-latency-optimizer/src"
)

// Type aliases for main package types
type BenchmarkConfig = main.BenchmarkConfig

var (
	NewBenchmarker   = main.NewBenchmarker
	calculateStats   = main.CalculateStats
)

// MockServer creates a test HTTP server with configurable latency
func MockServer(responseDelay time.Duration, statusCode int, body string) *httptest.Server {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if responseDelay > 0 {
			time.Sleep(responseDelay)
		}
		w.WriteHeader(statusCode)
		w.Write([]byte(body))
	})
	return httptest.NewServer(handler)
}

// TestBenchmarkMetrics validates that metrics are captured correctly
func TestBenchmarkMetrics(t *testing.T) {
	// Create a test server with 100ms delay
	server := MockServer(100*time.Millisecond, http.StatusOK, "test response")
	defer server.Close()

	config := BenchmarkConfig{
		TargetURL:     server.URL,
		TotalRequests: 10,
		Concurrency:   2,
		Timeout:       5 * time.Second,
		KeepAlive:     true,
		Method:        "GET",
	}

	benchmarker := NewBenchmarker(config)
	result, err := benchmarker.Run(context.Background())

	if err != nil {
		t.Fatalf("Benchmark failed: %v", err)
	}

	// Validate results
	if result.SuccessfulReqs != 10 {
		t.Errorf("Expected 10 successful requests, got %d", result.SuccessfulReqs)
	}

	if result.FailedReqs != 0 {
		t.Errorf("Expected 0 failed requests, got %d", result.FailedReqs)
	}

	// Check that latency is reasonable (should be > 100ms due to server delay)
	if result.LatencyStats.Min < 100 {
		t.Errorf("Minimum latency too low: %.2f ms", result.LatencyStats.Min)
	}

	// Check that percentiles are ordered correctly
	if result.LatencyStats.P50 > result.LatencyStats.P95 {
		t.Errorf("P50 (%.2f) should be less than P95 (%.2f)",
			result.LatencyStats.P50, result.LatencyStats.P95)
	}

	if result.LatencyStats.P95 > result.LatencyStats.P99 {
		t.Errorf("P95 (%.2f) should be less than P99 (%.2f)",
			result.LatencyStats.P95, result.LatencyStats.P99)
	}
}

// TestConcurrentRequests validates that concurrent requests work correctly
func TestConcurrentRequests(t *testing.T) {
	server := MockServer(50*time.Millisecond, http.StatusOK, "concurrent test")
	defer server.Close()

	config := BenchmarkConfig{
		TargetURL:     server.URL,
		TotalRequests: 100,
		Concurrency:   10,
		Timeout:       10 * time.Second,
		KeepAlive:     true,
		Method:        "GET",
	}

	benchmarker := NewBenchmarker(config)
	start := time.Now()
	result, err := benchmarker.Run(context.Background())
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("Benchmark failed: %v", err)
	}

	// With 10 concurrent requests and 100 total requests (10 batches),
	// each taking 50ms, total time should be roughly 500ms, not 5000ms
	expectedMax := 2 * time.Second // Allow some overhead
	if elapsed > expectedMax {
		t.Errorf("Concurrent execution took too long: %v (expected < %v)", elapsed, expectedMax)
	}

	if result.SuccessfulReqs != 100 {
		t.Errorf("Expected 100 successful requests, got %d", result.SuccessfulReqs)
	}
}

// TestErrorHandling validates that errors are captured correctly
func TestErrorHandling(t *testing.T) {
	// Server that returns errors
	server := MockServer(0, http.StatusInternalServerError, "error")
	defer server.Close()

	config := BenchmarkConfig{
		TargetURL:     server.URL,
		TotalRequests: 10,
		Concurrency:   2,
		Timeout:       5 * time.Second,
		KeepAlive:     true,
		Method:        "GET",
	}

	benchmarker := NewBenchmarker(config)
	result, err := benchmarker.Run(context.Background())

	if err != nil {
		t.Fatalf("Benchmark failed: %v", err)
	}

	// All requests should complete (even with 500 status)
	totalReqs := result.SuccessfulReqs + result.FailedReqs
	if totalReqs != 10 {
		t.Errorf("Expected 10 total requests, got %d", totalReqs)
	}
}

// TestTimeoutHandling validates timeout behavior
func TestTimeoutHandling(t *testing.T) {
	// Server with long delay
	server := MockServer(5*time.Second, http.StatusOK, "slow response")
	defer server.Close()

	config := BenchmarkConfig{
		TargetURL:     server.URL,
		TotalRequests: 5,
		Concurrency:   1,
		Timeout:       100 * time.Millisecond, // Short timeout
		KeepAlive:     true,
		Method:        "GET",
	}

	benchmarker := NewBenchmarker(config)
	result, err := benchmarker.Run(context.Background())

	if err != nil {
		t.Fatalf("Benchmark failed: %v", err)
	}

	// Most/all requests should fail due to timeout
	if result.FailedReqs == 0 {
		t.Errorf("Expected some failed requests due to timeout, got 0")
	}
}

// TestStatisticalAccuracy validates statistical calculations
func TestStatisticalAccuracy(t *testing.T) {
	// Test with known values
	values := []float64{10, 20, 30, 40, 50, 60, 70, 80, 90, 100}

	stats := calculateStats(values)

	if stats.Min != 10 {
		t.Errorf("Expected min=10, got %.2f", stats.Min)
	}

	if stats.Max != 100 {
		t.Errorf("Expected max=100, got %.2f", stats.Max)
	}

	if stats.Mean != 55 {
		t.Errorf("Expected mean=55, got %.2f", stats.Mean)
	}

	if stats.Median != 55 {
		t.Errorf("Expected median=55, got %.2f", stats.Median)
	}

	// P95 should be between 90 and 100
	if stats.P95 < 90 || stats.P95 > 100 {
		t.Errorf("Expected P95 between 90-100, got %.2f", stats.P95)
	}
}

// TestKeepAliveImpact validates that keep-alive affects performance
func TestKeepAliveImpact(t *testing.T) {
	server := MockServer(10*time.Millisecond, http.StatusOK, "keepalive test")
	defer server.Close()

	// Test with keep-alive enabled
	configWithKeepAlive := BenchmarkConfig{
		TargetURL:     server.URL,
		TotalRequests: 50,
		Concurrency:   5,
		Timeout:       10 * time.Second,
		KeepAlive:     true,
		Method:        "GET",
	}

	benchmarkerWithKA := NewBenchmarker(configWithKeepAlive)
	resultWithKA, err := benchmarkerWithKA.Run(context.Background())
	if err != nil {
		t.Fatalf("Benchmark with keep-alive failed: %v", err)
	}

	// Test without keep-alive
	configWithoutKeepAlive := BenchmarkConfig{
		TargetURL:     server.URL,
		TotalRequests: 50,
		Concurrency:   5,
		Timeout:       10 * time.Second,
		KeepAlive:     false,
		Method:        "GET",
	}

	benchmarkerWithoutKA := NewBenchmarker(configWithoutKeepAlive)
	resultWithoutKA, err := benchmarkerWithoutKA.Run(context.Background())
	if err != nil {
		t.Fatalf("Benchmark without keep-alive failed: %v", err)
	}

	// Keep-alive should generally result in better performance
	// (though this is not guaranteed in all environments)
	t.Logf("With keep-alive - P95: %.2f ms, RPS: %.2f",
		resultWithKA.LatencyStats.P95, resultWithKA.RequestsPerSecond)
	t.Logf("Without keep-alive - P95: %.2f ms, RPS: %.2f",
		resultWithoutKA.LatencyStats.P95, resultWithoutKA.RequestsPerSecond)
}

// TestContextCancellation validates that context cancellation works
func TestContextCancellation(t *testing.T) {
	server := MockServer(100*time.Millisecond, http.StatusOK, "test")
	defer server.Close()

	config := BenchmarkConfig{
		TargetURL:     server.URL,
		TotalRequests: 100,
		Concurrency:   5,
		Timeout:       10 * time.Second,
		KeepAlive:     true,
		Method:        "GET",
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Cancel context after a short delay
	go func() {
		time.Sleep(200 * time.Millisecond)
		cancel()
	}()

	benchmarker := NewBenchmarker(config)
	result, err := benchmarker.Run(ctx)

	// Should complete without error (may have fewer results)
	if err != nil {
		t.Fatalf("Benchmark failed: %v", err)
	}

	// Should have processed fewer than all requests
	totalProcessed := result.SuccessfulReqs + result.FailedReqs
	if totalProcessed >= 100 {
		t.Logf("Warning: Expected fewer than 100 requests processed due to cancellation, got %d", totalProcessed)
	}
}

// BenchmarkPerformance measures the overhead of the benchmarking tool itself
func BenchmarkPerformance(b *testing.B) {
	server := MockServer(1*time.Millisecond, http.StatusOK, "benchmark")
	defer server.Close()

	config := BenchmarkConfig{
		TargetURL:     server.URL,
		TotalRequests: 100,
		Concurrency:   10,
		Timeout:       5 * time.Second,
		KeepAlive:     true,
		Method:        "GET",
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		benchmarker := NewBenchmarker(config)
		_, err := benchmarker.Run(context.Background())
		if err != nil {
			b.Fatalf("Benchmark failed: %v", err)
		}
	}
}
