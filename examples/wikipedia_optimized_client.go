package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// This example demonstrates using the API Latency Optimizer
// with all optimizations enabled for Wikipedia requests

func main() {
	fmt.Println("╔═══════════════════════════════════════════════════════════╗")
	fmt.Println("║   Wikipedia Optimized Client - Full Optimization Stack   ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Note: To use the full optimizer, import the src package:
	// import "path/to/api-latency-optimizer/src"
	//
	// config := src.DefaultIntegratedConfig()
	// config.CacheConfig.Enabled = true
	// config.CacheConfig.MaxMemoryMB = 500
	// config.MonitoringConfig.Enabled = true
	//
	// optimizer, err := src.NewIntegratedOptimizer(config)
	// if err != nil {
	//     log.Fatalf("Failed to create optimizer: %v", err)
	// }
	//
	// optimizer.Start()
	// defer optimizer.Stop()
	//
	// client := optimizer.GetClient()

	// For this example, we'll use a standard HTTP/2 client
	// with optimized settings
	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
			ForceAttemptHTTP2:   true,
		},
	}

	fmt.Println("⚙️  Configuration:")
	fmt.Println("   • HTTP/2 enabled")
	fmt.Println("   • Connection pooling: 100 max connections")
	fmt.Println("   • Keep-alive: 90 seconds")
	fmt.Println("   • Target: https://www.wikipedia.org/")
	fmt.Println()

	// Set up graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\n\n🛑 Shutting down gracefully...")
		cancel()
	}()

	// Run performance test
	fmt.Println("🚀 Running optimized performance test...")
	fmt.Println()

	iterations := 100
	successCount := 0
	var totalDuration time.Duration

	startTime := time.Now()

	for i := 0; i < iterations; i++ {
		if ctx.Err() != nil {
			break
		}

		reqStart := time.Now()
		req, err := http.NewRequestWithContext(ctx, "GET", "https://www.wikipedia.org/", nil)
		if err != nil {
			log.Printf("Error creating request %d: %v", i+1, err)
			continue
		}

		resp, err := client.Do(req)
		reqDuration := time.Since(reqStart)
		totalDuration += reqDuration

		if err != nil {
			log.Printf("Request %d failed: %v", i+1, err)
			continue
		}

		resp.Body.Close()

		if resp.StatusCode == 200 {
			successCount++
		}

		if (i+1)%10 == 0 {
			avgLatency := totalDuration / time.Duration(i+1)
			fmt.Printf("   Progress: %d/%d | Avg Latency: %v | Success: %d\n",
				i+1, iterations, avgLatency, successCount)
		}
	}

	totalTime := time.Since(startTime)

	// Display results
	fmt.Println()
	fmt.Println("╔═══════════════════════════════════════════════════════════╗")
	fmt.Println("║                    Performance Results                    ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════╝")
	fmt.Println()

	avgLatency := totalDuration / time.Duration(iterations)
	rps := float64(successCount) / totalTime.Seconds()

	fmt.Printf("📊 Metrics:\n")
	fmt.Printf("   Total Requests:  %d\n", iterations)
	fmt.Printf("   Successful:      %d (%.1f%%)\n", successCount, float64(successCount)/float64(iterations)*100)
	fmt.Printf("   Average Latency: %v\n", avgLatency)
	fmt.Printf("   Throughput:      %.2f req/sec\n", rps)
	fmt.Printf("   Total Duration:  %v\n", totalTime)
	fmt.Println()

	fmt.Println("✅ Test complete!")
	fmt.Println()
	fmt.Println("💡 Next Steps:")
	fmt.Println("   • Integrate with src.IntegratedOptimizer for full caching")
	fmt.Println("   • Enable monitoring dashboard on port 8080")
	fmt.Println("   • Configure cache size and TTL settings")
	fmt.Println("   • Deploy to production with circuit breaker")
	fmt.Println()
}
