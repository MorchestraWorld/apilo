package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"time"
)

// LoadPattern defines how requests are distributed over time
type LoadPattern string

const (
	LoadPatternConstant   LoadPattern = "constant"   // Steady request rate
	LoadPatternRampUp     LoadPattern = "ramp_up"    // Gradually increase load
	LoadPatternSpike      LoadPattern = "spike"      // Sudden load increase
	LoadPatternSinusoidal LoadPattern = "sinusoidal" // Wave pattern
)

// BenchmarkSuite manages multiple benchmark runs with different configurations
type BenchmarkSuite struct {
	Name               string         `json:"name"`
	Description        string         `json:"description"`
	Runs               []BenchmarkRun `json:"runs"`
	OutputDir          string         `json:"output_dir"`
	ComparisonBaseline string         `json:"comparison_baseline,omitempty"`
}

// BenchmarkRun represents a single benchmark configuration
type BenchmarkRun struct {
	Name             string             `json:"name"`
	Config           BenchmarkConfig    `json:"config"`
	Iterations       int                `json:"iterations"`
	WarmupIterations int                `json:"warmup_iterations"`
	LoadPattern      LoadPattern        `json:"load_pattern"`
	Results          []*BenchmarkResult `json:"results,omitempty"`
}

// BenchmarkRunner orchestrates benchmark execution with multiple iterations
type BenchmarkRunner struct {
	suite     *BenchmarkSuite
	resultDir string
}

// NewBenchmarkRunner creates a new runner for the given suite
func NewBenchmarkRunner(suite *BenchmarkSuite) *BenchmarkRunner {
	if suite.OutputDir == "" {
		suite.OutputDir = "./benchmarks/results"
	}

	resultDir := filepath.Join(suite.OutputDir, fmt.Sprintf("%s_%s",
		suite.Name, time.Now().Format("20060102_150405")))

	return &BenchmarkRunner{
		suite:     suite,
		resultDir: resultDir,
	}
}

// Run executes all benchmark runs in the suite
func (r *BenchmarkRunner) Run(ctx context.Context) error {
	// Create output directory
	if err := os.MkdirAll(r.resultDir, 0755); err != nil {
		return fmt.Errorf("failed to create result directory: %w", err)
	}

	fmt.Printf("\n=== Starting Benchmark Suite: %s ===\n", r.suite.Name)
	fmt.Printf("Description: %s\n", r.suite.Description)
	fmt.Printf("Output Directory: %s\n\n", r.resultDir)

	// Execute each benchmark run
	for i := range r.suite.Runs {
		run := &r.suite.Runs[i]

		fmt.Printf("\n--- Benchmark Run: %s ---\n", run.Name)

		if err := r.executeRun(ctx, run); err != nil {
			fmt.Printf("ERROR: Run failed: %v\n", err)
			continue
		}

		// Save individual run results
		runFile := filepath.Join(r.resultDir, fmt.Sprintf("%s.json", run.Name))
		if err := r.saveRunResults(run, runFile); err != nil {
			fmt.Printf("WARNING: Failed to save run results: %v\n", err)
		}
	}

	// Save complete suite results
	suiteFile := filepath.Join(r.resultDir, "suite_results.json")
	if err := r.saveSuiteResults(suiteFile); err != nil {
		return fmt.Errorf("failed to save suite results: %w", err)
	}

	// Generate summary report
	r.generateSummaryReport()

	fmt.Printf("\n=== Benchmark Suite Complete ===\n")
	fmt.Printf("Results saved to: %s\n", r.resultDir)

	return nil
}

// executeRun runs a single benchmark configuration with iterations
func (r *BenchmarkRunner) executeRun(ctx context.Context, run *BenchmarkRun) error {
	// Warmup phase
	if run.WarmupIterations > 0 {
		fmt.Printf("Warmup: Running %d iterations...\n", run.WarmupIterations)
		for i := 0; i < run.WarmupIterations; i++ {
			benchmarker := NewBenchmarker(run.Config)
			_, err := benchmarker.Run(ctx)
			if err != nil {
				fmt.Printf("Warmup iteration %d failed: %v\n", i+1, err)
			}
		}
		fmt.Printf("Warmup complete\n\n")
	}

	// Main benchmark iterations
	run.Results = make([]*BenchmarkResult, 0, run.Iterations)

	for i := 0; i < run.Iterations; i++ {
		fmt.Printf("Iteration %d/%d...\n", i+1, run.Iterations)

		benchmarker := NewBenchmarker(run.Config)
		result, err := benchmarker.Run(ctx)

		if err != nil {
			return fmt.Errorf("iteration %d failed: %w", i+1, err)
		}

		run.Results = append(run.Results, result)

		// Print iteration summary
		fmt.Printf("  Successful: %d | Failed: %d | RPS: %.2f | P95: %.2f ms\n",
			result.SuccessfulReqs, result.FailedReqs,
			result.RequestsPerSecond, result.LatencyStats.P95)

		// Small delay between iterations to avoid overwhelming the target
		if i < run.Iterations-1 {
			time.Sleep(2 * time.Second)
		}
	}

	// Calculate aggregate statistics
	r.calculateAggregateStats(run)

	return nil
}

// calculateAggregateStats computes statistics across multiple iterations
func (r *BenchmarkRunner) calculateAggregateStats(run *BenchmarkRun) {
	if len(run.Results) == 0 {
		return
	}

	fmt.Printf("\n--- Aggregate Statistics for %s ---\n", run.Name)

	var totalRPS, totalP50, totalP95, totalP99 float64
	var minRPS, maxRPS float64 = math.MaxFloat64, 0
	var minP95, maxP95 float64 = math.MaxFloat64, 0

	for _, result := range run.Results {
		totalRPS += result.RequestsPerSecond
		totalP50 += result.LatencyStats.P50
		totalP95 += result.LatencyStats.P95
		totalP99 += result.LatencyStats.P99

		if result.RequestsPerSecond < minRPS {
			minRPS = result.RequestsPerSecond
		}
		if result.RequestsPerSecond > maxRPS {
			maxRPS = result.RequestsPerSecond
		}

		if result.LatencyStats.P95 < minP95 {
			minP95 = result.LatencyStats.P95
		}
		if result.LatencyStats.P95 > maxP95 {
			maxP95 = result.LatencyStats.P95
		}
	}

	count := float64(len(run.Results))
	fmt.Printf("Iterations: %d\n", len(run.Results))
	fmt.Printf("Avg RPS: %.2f (min: %.2f, max: %.2f)\n", totalRPS/count, minRPS, maxRPS)
	fmt.Printf("Avg P50 Latency: %.2f ms\n", totalP50/count)
	fmt.Printf("Avg P95 Latency: %.2f ms (min: %.2f, max: %.2f)\n", totalP95/count, minP95, maxP95)
	fmt.Printf("Avg P99 Latency: %.2f ms\n", totalP99/count)
}

// saveRunResults saves results for a single benchmark run
func (r *BenchmarkRunner) saveRunResults(run *BenchmarkRun, filepath string) error {
	data, err := json.MarshalIndent(run, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath, data, 0644)
}

// saveSuiteResults saves the complete suite results
func (r *BenchmarkRunner) saveSuiteResults(filepath string) error {
	data, err := json.MarshalIndent(r.suite, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath, data, 0644)
}

// generateSummaryReport creates a markdown summary of all results
func (r *BenchmarkRunner) generateSummaryReport() {
	reportPath := filepath.Join(r.resultDir, "SUMMARY.md")

	report := fmt.Sprintf("# Benchmark Suite Summary: %s\n\n", r.suite.Name)
	report += fmt.Sprintf("**Description:** %s\n\n", r.suite.Description)
	report += fmt.Sprintf("**Run Date:** %s\n\n", time.Now().Format("2006-01-02 15:04:05"))
	report += "---\n\n"

	for _, run := range r.suite.Runs {
		if len(run.Results) == 0 {
			continue
		}

		report += fmt.Sprintf("## %s\n\n", run.Name)
		report += fmt.Sprintf("- **Target:** %s\n", run.Config.TargetURL)
		report += fmt.Sprintf("- **Requests:** %d\n", run.Config.TotalRequests)
		report += fmt.Sprintf("- **Concurrency:** %d\n", run.Config.Concurrency)
		report += fmt.Sprintf("- **Iterations:** %d\n\n", run.Iterations)

		// Calculate averages
		var avgRPS, avgP50, avgP95, avgP99, avgTTFB float64
		for _, result := range run.Results {
			avgRPS += result.RequestsPerSecond
			avgP50 += result.LatencyStats.P50
			avgP95 += result.LatencyStats.P95
			avgP99 += result.LatencyStats.P99
			avgTTFB += result.TTFBStats.P95
		}
		count := float64(len(run.Results))

		report += "### Performance Metrics\n\n"
		report += "| Metric | Value |\n"
		report += "|--------|-------|\n"
		report += fmt.Sprintf("| Avg Requests/sec | %.2f |\n", avgRPS/count)
		report += fmt.Sprintf("| Avg P50 Latency | %.2f ms |\n", avgP50/count)
		report += fmt.Sprintf("| Avg P95 Latency | %.2f ms |\n", avgP95/count)
		report += fmt.Sprintf("| Avg P99 Latency | %.2f ms |\n", avgP99/count)
		report += fmt.Sprintf("| Avg P95 TTFB | %.2f ms |\n\n", avgTTFB/count)
	}

	os.WriteFile(reportPath, []byte(report), 0644)
	fmt.Printf("\nSummary report generated: %s\n", reportPath)
}

// CompareWithBaseline compares current results with a baseline benchmark
func (r *BenchmarkRunner) CompareWithBaseline(baselinePath string) error {
	// Load baseline data
	baselineData, err := os.ReadFile(baselinePath)
	if err != nil {
		return fmt.Errorf("failed to read baseline: %w", err)
	}

	var baseline BenchmarkSuite
	if err := json.Unmarshal(baselineData, &baseline); err != nil {
		return fmt.Errorf("failed to parse baseline: %w", err)
	}

	// Generate comparison report
	reportPath := filepath.Join(r.resultDir, "COMPARISON.md")
	report := "# Benchmark Comparison Report\n\n"
	report += fmt.Sprintf("**Current:** %s\n", r.suite.Name)
	report += fmt.Sprintf("**Baseline:** %s\n\n", baseline.Name)
	report += "---\n\n"

	// Compare matching runs
	for _, currentRun := range r.suite.Runs {
		for _, baselineRun := range baseline.Runs {
			if currentRun.Name != baselineRun.Name {
				continue
			}

			report += r.generateComparisonSection(&currentRun, &baselineRun)
		}
	}

	os.WriteFile(reportPath, []byte(report), 0644)
	fmt.Printf("Comparison report generated: %s\n", reportPath)

	return nil
}

// generateComparisonSection creates a comparison for two benchmark runs
func (r *BenchmarkRunner) generateComparisonSection(current, baseline *BenchmarkRun) string {
	if len(current.Results) == 0 || len(baseline.Results) == 0 {
		return ""
	}

	section := fmt.Sprintf("## %s\n\n", current.Name)

	// Calculate current averages
	var currRPS, currP95 float64
	for _, result := range current.Results {
		currRPS += result.RequestsPerSecond
		currP95 += result.LatencyStats.P95
	}
	currRPS /= float64(len(current.Results))
	currP95 /= float64(len(current.Results))

	// Calculate baseline averages
	var baseRPS, baseP95 float64
	for _, result := range baseline.Results {
		baseRPS += result.RequestsPerSecond
		baseP95 += result.LatencyStats.P95
	}
	baseRPS /= float64(len(baseline.Results))
	baseP95 /= float64(len(baseline.Results))

	// Calculate improvements
	rpsChange := ((currRPS - baseRPS) / baseRPS) * 100
	p95Change := ((currP95 - baseP95) / baseP95) * 100

	section += "| Metric | Baseline | Current | Change |\n"
	section += "|--------|----------|---------|--------|\n"
	section += fmt.Sprintf("| Requests/sec | %.2f | %.2f | %.1f%% |\n", baseRPS, currRPS, rpsChange)
	section += fmt.Sprintf("| P95 Latency | %.2f ms | %.2f ms | %.1f%% |\n\n", baseP95, currP95, p95Change)

	if rpsChange > 5 {
		section += "✅ **Improvement:** Throughput increased significantly\n\n"
	} else if rpsChange < -5 {
		section += "❌ **Regression:** Throughput decreased\n\n"
	}

	if p95Change < -5 {
		section += "✅ **Improvement:** Latency reduced significantly\n\n"
	} else if p95Change > 5 {
		section += "❌ **Regression:** Latency increased\n\n"
	}

	return section
}
