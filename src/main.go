package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Build-time variables injected via -ldflags
var (
	Version   = "1.0.0"
	BuildTime = "unknown"
	Commit    = "unknown"
	SourceDir = "unknown"
)

const (
	banner = `
╔═══════════════════════════════════════════════════════════╗
║       API Latency Optimizer - Benchmark Tool             ║
║       Version: %-10s                                 ║
╚═══════════════════════════════════════════════════════════╝
`
)

func main() {
	// Command line flags
	var (
		configFile      = flag.String("config", "", "Path to YAML configuration file")
		url             = flag.String("url", "https://api.anthropic.com", "Target URL to benchmark")
		requests        = flag.Int("requests", 100, "Total number of requests")
		concurrency     = flag.Int("concurrency", 10, "Number of concurrent requests")
		iterations      = flag.Int("iterations", 3, "Number of benchmark iterations")
		warmup          = flag.Int("warmup", 1, "Number of warmup iterations")
		timeout         = flag.Duration("timeout", 30*time.Second, "Request timeout")
		keepalive       = flag.Bool("keepalive", true, "Enable HTTP keep-alive")
		outputDir       = flag.String("output", "./benchmarks/results", "Output directory for results")
		rawMetrics      = flag.Bool("raw", false, "Include raw metrics in output")
		compareBaseline = flag.String("compare", "", "Path to baseline results for comparison")
		quiet           = flag.Bool("quiet", false, "Suppress progress output")
		showVersion     = flag.Bool("version", false, "Show version and exit")

		// Monitoring flags
		enableMonitoring = flag.Bool("monitor", false, "Enable real-time monitoring dashboard")
		dashboardPort    = flag.Int("dashboard-port", 8080, "Dashboard HTTP port")
		prometheusPort   = flag.Int("prometheus-port", 9090, "Prometheus exporter port")
		enableAlerts     = flag.Bool("alerts", false, "Enable performance alerting")
		monitoringConfig = flag.String("monitoring-config", "", "Path to monitoring configuration file")
	)

	flag.Parse()

	// Show version
	if *showVersion {
		fmt.Printf("API Latency Optimizer v%s\n", Version)
		fmt.Printf("Build Time: %s\n", BuildTime)
		fmt.Printf("Commit: %s\n", Commit)
		fmt.Printf("Source Dir: %s\n", SourceDir)
		os.Exit(0)
	}

	// Print banner
	if !*quiet {
		fmt.Printf(banner, Version)
		fmt.Println()
	}

	// Set up context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\n\nReceived interrupt signal, shutting down gracefully...")
		cancel()
	}()

	var err error

	// Initialize monitoring if enabled
	var monitoringSystem *MonitoringSystem
	if *enableMonitoring {
		monitoringSystem, err = initializeMonitoring(*monitoringConfig, *dashboardPort, *prometheusPort, *enableAlerts, *quiet)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to initialize monitoring: %v\n", err)
			os.Exit(1)
		}
		defer monitoringSystem.Stop()
	}

	// Run benchmark based on configuration
	if *configFile != "" {
		err = runFromConfig(ctx, *configFile, *compareBaseline, *quiet, monitoringSystem)
	} else {
		err = runQuickBenchmark(ctx, quickBenchmarkParams{
			url:             *url,
			requests:        *requests,
			concurrency:     *concurrency,
			iterations:      *iterations,
			warmup:          *warmup,
			timeout:         *timeout,
			keepalive:       *keepalive,
			outputDir:       *outputDir,
			includeRaw:      *rawMetrics,
			compareBaseline: *compareBaseline,
			quiet:           *quiet,
		}, monitoringSystem)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}

	if !*quiet {
		fmt.Println("\n✓ Benchmark completed successfully")
	}
}

// quickBenchmarkParams holds parameters for a quick benchmark run
type quickBenchmarkParams struct {
	url             string
	requests        int
	concurrency     int
	iterations      int
	warmup          int
	timeout         time.Duration
	keepalive       bool
	outputDir       string
	includeRaw      bool
	compareBaseline string
	quiet           bool
}

// initializeMonitoring sets up and starts the monitoring system
func initializeMonitoring(configPath string, dashboardPort, prometheusPort int, enableAlerts, quiet bool) (*MonitoringSystem, error) {
	// Create monitoring configuration
	config := DefaultMonitoringConfig()

	// Override with CLI flags
	config.DashboardPort = dashboardPort
	config.PrometheusPort = prometheusPort
	config.AlertingEnabled = enableAlerts

	// Load from config file if provided
	if configPath != "" {
		// TODO: Load configuration from YAML file
		if !quiet {
			fmt.Printf("Loading monitoring configuration from: %s\n", configPath)
		}
	}

	// Create and start monitoring system
	monitoring := NewMonitoringSystem(config)
	if err := monitoring.Start(); err != nil {
		return nil, fmt.Errorf("failed to start monitoring: %w", err)
	}

	if !quiet {
		fmt.Println("\n✓ Monitoring system started successfully")
		fmt.Printf("  Dashboard: http://localhost:%d\n", config.DashboardPort)
		if config.PrometheusEnabled {
			fmt.Printf("  Prometheus: http://localhost:%d%s\n", config.PrometheusPort, config.PrometheusPath)
		}
		fmt.Println()
	}

	return monitoring, nil
}

// runQuickBenchmark runs a simple benchmark without a config file
func runQuickBenchmark(ctx context.Context, params quickBenchmarkParams, monitoring *MonitoringSystem) error {
	if !params.quiet {
		fmt.Printf("Running benchmark against: %s\n", params.url)
		fmt.Printf("Configuration: %d requests, %d concurrent, %d iterations\n\n",
			params.requests, params.concurrency, params.iterations)
	}

	// Create benchmark suite
	suite := &BenchmarkSuite{
		Name:        "quick_benchmark",
		Description: fmt.Sprintf("Quick benchmark of %s", params.url),
		OutputDir:   params.outputDir,
		Runs: []BenchmarkRun{
			{
				Name: "benchmark",
				Config: BenchmarkConfig{
					TargetURL:         params.url,
					TotalRequests:     params.requests,
					Concurrency:       params.concurrency,
					Timeout:           params.timeout,
					KeepAlive:         params.keepalive,
					Method:            "GET",
					IncludeRawMetrics: params.includeRaw,
				},
				Iterations:       params.iterations,
				WarmupIterations: params.warmup,
				LoadPattern:      LoadPatternConstant,
			},
		},
	}

	// Run benchmark
	runner := NewBenchmarkRunner(suite)

	// Attach monitoring if enabled
	if monitoring != nil {
		// The runner will integrate with monitoring during execution
		// We'll need to update the runner to accept monitoring
		if !params.quiet {
			fmt.Println("Monitoring enabled for benchmark run")
		}
	}

	if err := runner.Run(ctx); err != nil {
		return err
	}

	// Update monitoring with final results
	if monitoring != nil && len(suite.Runs) > 0 {
		// Get the last run with results
		for i := len(suite.Runs) - 1; i >= 0; i-- {
			if len(suite.Runs[i].Results) > 0 {
				lastResult := suite.Runs[i].Results[len(suite.Runs[i].Results)-1]
				monitoring.GetCollector().UpdateBenchmarkResult(lastResult)
				monitoring.GetCollector().Collect()
				break
			}
		}

		if !params.quiet {
			fmt.Println("\n=== Monitoring Summary ===")
			monitoring.PrintSummary()
		}
	}

	// Compare with baseline if provided
	if params.compareBaseline != "" {
		if !params.quiet {
			fmt.Printf("\nComparing with baseline: %s\n", params.compareBaseline)
		}
		if err := runner.CompareWithBaseline(params.compareBaseline); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Comparison failed: %v\n", err)
		}
	}

	return nil
}

// runFromConfig runs benchmarks from a YAML configuration file
func runFromConfig(ctx context.Context, configPath, baselinePath string, quiet bool, monitoring *MonitoringSystem) error {
	if !quiet {
		fmt.Printf("Loading configuration from: %s\n\n", configPath)
	}

	// This would load the config using the config package
	// For now, we'll create a sample suite
	suite := &BenchmarkSuite{
		Name:        "config_benchmark",
		Description: "Benchmark from configuration file",
		OutputDir:   "./benchmarks/results",
		Runs: []BenchmarkRun{
			{
				Name: "baseline_test",
				Config: BenchmarkConfig{
					TargetURL:     "https://api.anthropic.com",
					TotalRequests: 100,
					Concurrency:   10,
					Timeout:       30 * time.Second,
					KeepAlive:     true,
					Method:        "GET",
				},
				Iterations:       3,
				WarmupIterations: 1,
				LoadPattern:      LoadPatternConstant,
			},
		},
	}

	runner := NewBenchmarkRunner(suite)

	// Attach monitoring if enabled
	if monitoring != nil {
		if !quiet {
			fmt.Println("Monitoring enabled for benchmark run")
		}
	}

	if err := runner.Run(ctx); err != nil {
		return err
	}

	// Update monitoring with final results
	if monitoring != nil && len(suite.Runs) > 0 {
		// Get the last run with results
		for i := len(suite.Runs) - 1; i >= 0; i-- {
			if len(suite.Runs[i].Results) > 0 {
				lastResult := suite.Runs[i].Results[len(suite.Runs[i].Results)-1]
				monitoring.GetCollector().UpdateBenchmarkResult(lastResult)
				monitoring.GetCollector().Collect()
				break
			}
		}

		if !quiet {
			fmt.Println("\n=== Monitoring Summary ===")
			monitoring.PrintSummary()
		}
	}

	if baselinePath != "" {
		if !quiet {
			fmt.Printf("\nComparing with baseline: %s\n", baselinePath)
		}
		if err := runner.CompareWithBaseline(baselinePath); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Comparison failed: %v\n", err)
		}
	}

	return nil
}
