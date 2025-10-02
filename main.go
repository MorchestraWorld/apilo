// API Latency Optimizer - Phase 1 Integration
// This is the main entry point for the integrated optimization system.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"
)

func main() {
	// Command line flags
	url := flag.String("url", "https://api.anthropic.com", "Target URL for benchmarking")
	requests := flag.Int("requests", 100, "Total number of requests")
	concurrency := flag.Int("concurrency", 10, "Number of concurrent requests")
	useOptimizations := flag.Bool("optimize", true, "Enable all optimizations (HTTP/2, caching, monitoring)")
	enableMonitoring := flag.Bool("monitor", false, "Enable monitoring dashboard")
	comparison := flag.Bool("compare", false, "Run comparison between optimized and baseline")
	validate := flag.Bool("validate", false, "Run Phase 1 validation suite")

	flag.Parse()

	if *validate {
		fmt.Println("🔍 Running Phase 1 Validation Suite...")
		if err := runPhase1ValidationSuite(); err != nil {
			log.Fatalf("❌ Validation failed: %v", err)
		}
		fmt.Println("🎉 Phase 1 validation completed successfully!")
		return
	}

	fmt.Printf(`
🚀 API Latency Optimizer - Phase 1
=====================================
Target URL: %s
Requests: %d
Concurrency: %d
Optimizations: %t
Monitoring: %t
Comparison Mode: %t

`, *url, *requests, *concurrency, *useOptimizations, *enableMonitoring, *comparison)

	// Create configuration
	config := DefaultIntegratedConfig()
	config.BenchmarkConfig.TotalRequests = *requests
	config.BenchmarkConfig.Concurrency = *concurrency
	config.MonitoringConfig.Enabled = *enableMonitoring

	// Create integrated optimizer
	optimizer, err := NewIntegratedOptimizer(config)
	if err != nil {
		log.Fatalf("Failed to create optimizer: %v", err)
	}

	// Start optimizer
	if err := optimizer.Start(); err != nil {
		log.Fatalf("Failed to start optimizer: %v", err)
	}
	defer optimizer.Stop()

	if *enableMonitoring {
		fmt.Printf("📊 Monitoring dashboard: %s\n", optimizer.GetMonitoringURL())
	}

	// Run benchmark
	runConfig := &BenchmarkRunConfig{
		URL:              *url,
		TotalRequests:    *requests,
		Concurrency:      *concurrency,
		Timeout:          30 * time.Second,
		UseOptimizations: *useOptimizations,
	}

	fmt.Println("🏁 Starting benchmark...")
	startTime := time.Now()

	if *comparison {
		result, err := optimizer.RunComparisonBenchmark(runConfig)
		if err != nil {
			log.Fatalf("Comparison benchmark failed: %v", err)
		}

		// Print comparison results
		fmt.Printf(`
📊 Comparison Results
=====================
Baseline P50:    %v
Optimized P50:   %v
Improvement:     %.1f%%

Baseline RPS:    %.2f
Optimized RPS:   %.2f
Improvement:     %.1f%%

Overall Score:   %.1f/10
`,
			result.Baseline.Latency.P50,
			result.Optimized.Latency.P50,
			result.Improvement.LatencyImprovement*100,
			result.Baseline.Throughput.RequestsPerSecond,
			result.Optimized.Throughput.RequestsPerSecond,
			result.Improvement.ThroughputImprovement*100,
			result.Improvement.EfficiencyScore*10,
		)
	} else {
		result, err := optimizer.RunBenchmark(runConfig)
		if err != nil {
			log.Fatalf("Benchmark failed: %v", err)
		}

		// Print results
		fmt.Printf(`
📊 Benchmark Results
====================
P50 Latency:     %v
P95 Latency:     %v
P99 Latency:     %v
Throughput:      %.2f req/s
Success Rate:    %.2f%%
Duration:        %v
`,
			result.Latency.P50,
			result.Latency.P95,
			result.Latency.P99,
			result.Throughput.RequestsPerSecond,
			result.SuccessRate,
			time.Since(startTime),
		)
	}

	// Print optimization stats if available
	stats := optimizer.GetCurrentStats()
	if stats.ClientStats != nil {
		fmt.Printf(`
🔧 Optimization Stats
=====================
Cache Hit Ratio:     %.1f%%
Connection Reuse:    %.1f%%
Total Requests:      %d
Average Latency:     %v
Health Status:       %s
Overall Grade:       %s
`,
			stats.ClientStats.CacheHitRatio*100,
			stats.ClientStats.ConnectionReuseRatio*100,
			stats.ClientStats.TotalRequests,
			stats.ClientStats.AverageLatency,
			stats.HealthStatus,
			stats.OverallGrade,
		)
	}

	// Check if Phase 1 targets were met
	if stats.LatencyTarget && stats.CacheTarget {
		fmt.Println("🎉 Phase 1 targets achieved! Ready for Phase 2.")
	} else {
		fmt.Println("⚠️  Phase 1 targets not fully met. Review performance.")
	}

	fmt.Println("\n✅ Benchmark completed successfully!")
}