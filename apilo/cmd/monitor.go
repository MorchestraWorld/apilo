package cmd

import (
	"fmt"
	"os/exec"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	monitorPort     int
	monitorInterval int
)

var monitorCmd = &cobra.Command{
	Use:   "monitor <url>",
	Short: "Start real-time monitoring",
	Long:  "Start the API Latency Optimizer with real-time monitoring dashboard",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		startMonitoring(args[0])
	},
}

func init() {
	rootCmd.AddCommand(monitorCmd)

	monitorCmd.Flags().IntVarP(&monitorPort, "port", "p", 8080, "dashboard port")
	monitorCmd.Flags().IntVarP(&monitorInterval, "interval", "i", 5, "metrics collection interval (seconds)")
}

func startMonitoring(url string) {
	// Header
	color.Cyan("\n╔═══════════════════════════════════════════════════════════════════╗")
	color.Cyan("║              API Latency Optimizer - Live Monitoring              ║")
	color.Cyan("╚═══════════════════════════════════════════════════════════════════╝\n")

	fmt.Println(color.YellowString("🔍 Monitor Configuration:"))
	fmt.Printf("   URL: %s\n", color.CyanString(url))
	fmt.Printf("   Dashboard Port: %s\n", color.CyanString(fmt.Sprintf("%d", monitorPort)))
	fmt.Printf("   Update Interval: %s seconds\n\n", color.CyanString(fmt.Sprintf("%d", monitorInterval)))

	// Dashboard URLs
	fmt.Println(color.GreenString("📊 Dashboard Access:"))
	fmt.Printf("   Dashboard: %s\n", color.CyanString(fmt.Sprintf("http://localhost:%d/dashboard", monitorPort)))
	fmt.Printf("   Metrics:   %s\n", color.CyanString(fmt.Sprintf("http://localhost:%d/metrics", monitorPort)))
	fmt.Printf("   Health:    %s\n\n", color.CyanString(fmt.Sprintf("http://localhost:%d/health", monitorPort)))

	// Available Metrics
	fmt.Println(color.YellowString("📈 Available Metrics:"))
	fmt.Println("   • Cache hit/miss ratios")
	fmt.Println("   • Memory usage and pressure")
	fmt.Println("   • Latency percentiles (P50, P95, P99)")
	fmt.Println("   • Throughput (requests/sec)")
	fmt.Println("   • Circuit breaker states")
	fmt.Println("   • Active connections")
	fmt.Println("   • GC statistics\n")

	// Try to start the actual optimizer with monitoring
	optimizerPath := "/Users/joshkornreich/Documents/Projects/api-latency-optimizer/bin/api-optimizer"

	args := []string{
		"--url", url,
		"--monitor",
		"--dashboard",
		"--port", fmt.Sprintf("%d", monitorPort),
		"--interval", fmt.Sprintf("%ds", monitorInterval),
	}

	cmd := exec.Command(optimizerPath, args...)
	cmd.Stdout = color.Output
	cmd.Stderr = color.Error

	color.Green("🚀 Starting monitoring dashboard...\n")
	fmt.Println(color.BlueString("💡 Press Ctrl+C to stop monitoring\n"))

	if err := cmd.Run(); err != nil {
		// Simulated monitoring info
		color.Yellow("⚠️  Optimizer binary not available, showing simulated monitoring setup\n")
		showSimulatedMonitoring()
		return
	}
}

func showSimulatedMonitoring() {
	fmt.Println(color.CyanString("\n🔧 To enable real-time monitoring:\n"))

	fmt.Println("1. Build the optimizer with monitoring:")
	fmt.Println("   " + color.CyanString("cd /Users/joshkornreich/Documents/Projects/api-latency-optimizer"))
	fmt.Println("   " + color.CyanString("go build -o bin/api-optimizer ./src"))

	fmt.Println("\n2. Start monitoring:")
	fmt.Println("   " + color.CyanString("apilo monitor <url> --port 8080"))

	fmt.Println("\n3. Access the dashboard:")
	fmt.Println("   " + color.CyanString("open http://localhost:8080/dashboard"))

	fmt.Println(color.YellowString("\n📊 Monitoring Features:"))
	fmt.Println("   • Real-time latency graphs")
	fmt.Println("   • Cache performance visualization")
	fmt.Println("   • Memory and CPU usage charts")
	fmt.Println("   • Request throughput monitoring")
	fmt.Println("   • Error rate tracking")
	fmt.Println("   • Circuit breaker state visualization")
	fmt.Println("   • Prometheus metrics export")
	fmt.Println("   • Alert notifications\n")

	fmt.Println(color.BlueString("💡 See also: apilo docs monitoring\n"))
}
