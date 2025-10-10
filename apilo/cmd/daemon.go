package cmd

import (
	"apilo/internal/daemon"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	daemonPort       int
	daemonLogLevel   string
	daemonBackground bool
)

// daemonCmd represents the daemon command
var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Background daemon for automatic API optimization",
	Long: `Run apilo as a background daemon for automatic optimization of Claude Code queries.

The daemon provides:
• Persistent background process with automatic optimization
• HTTP IPC server for optimization requests
• Memory-bounded cache with intelligent eviction
• HTTP/2 connection pooling and reuse
• Circuit breaker pattern for reliability
• Real-time performance metrics

The daemon runs as a background service and can be controlled via:
  apilo daemon start   - Start the daemon
  apilo daemon stop    - Stop the daemon
  apilo daemon status  - Check daemon status
  apilo daemon restart - Restart the daemon
  apilo daemon logs    - View daemon logs`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var daemonStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the apilo daemon",
	Long:  "Start the apilo daemon as a background service for automatic API optimization",
	Run: func(cmd *cobra.Command, args []string) {
		startDaemon()
	},
}

var daemonStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the apilo daemon",
	Long:  "Stop the running apilo daemon gracefully",
	Run: func(cmd *cobra.Command, args []string) {
		stopDaemon()
	},
}

var daemonStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check daemon status",
	Long:  "Check the current status of the apilo daemon",
	Run: func(cmd *cobra.Command, args []string) {
		checkDaemonStatus()
	},
}

var daemonRestartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart the daemon",
	Long:  "Restart the apilo daemon (stop + start)",
	Run: func(cmd *cobra.Command, args []string) {
		restartDaemon()
	},
}

var daemonLogsCmd = &cobra.Command{
	Use:   "logs",
	Short: "View daemon logs",
	Long:  "View the apilo daemon logs",
	Run: func(cmd *cobra.Command, args []string) {
		viewDaemonLogs()
	},
}

func init() {
	rootCmd.AddCommand(daemonCmd)
	daemonCmd.AddCommand(daemonStartCmd)
	daemonCmd.AddCommand(daemonStopCmd)
	daemonCmd.AddCommand(daemonStatusCmd)
	daemonCmd.AddCommand(daemonRestartCmd)
	daemonCmd.AddCommand(daemonLogsCmd)

	// Flags
	daemonStartCmd.Flags().IntVarP(&daemonPort, "port", "p", 9876, "IPC server port")
	daemonStartCmd.Flags().StringVar(&daemonLogLevel, "log-level", "info", "Log level (debug, info, warn, error)")
	daemonStartCmd.Flags().BoolVarP(&daemonBackground, "background", "d", true, "Run in background")
}

func startDaemon() {
	color.Cyan("\n╔═══════════════════════════════════════════════════════════════════╗")
	color.Cyan("║                Starting Apilo Daemon                              ║")
	color.Cyan("╚═══════════════════════════════════════════════════════════════════╝\n")

	config := daemon.DefaultDaemonConfig()
	config.Port = daemonPort
	config.LogLevel = daemonLogLevel

	pidMgr := daemon.NewPIDManager(config.PIDFile)

	// Check if already running
	if running, pid, _ := pidMgr.IsRunning(); running {
		color.Yellow("⚠️  Daemon already running (PID: %d)\n", pid)
		fmt.Println(color.BlueString("💡 Use 'apilo daemon stop' to stop it first\n"))
		return
	}

	if daemonBackground {
		// Start daemon in background
		fmt.Println(color.YellowString("🚀 Starting daemon in background...\n"))

		executable, err := os.Executable()
		if err != nil {
			color.Red("❌ Failed to locate executable: %v\n", err)
			return
		}

		cmd := exec.Command(executable, "daemon", "start", "--background=false", fmt.Sprintf("--port=%d", daemonPort))
		cmd.Stdout = nil
		cmd.Stderr = nil

		if err := cmd.Start(); err != nil {
			color.Red("❌ Failed to start daemon: %v\n", err)
			return
		}

		color.Green("✅ Daemon started (PID: %d)\n", cmd.Process.Pid)
		fmt.Printf("   Port: %s\n", color.CyanString(fmt.Sprintf("%d", daemonPort)))
		fmt.Printf("   IPC:  %s\n\n", color.CyanString(fmt.Sprintf("http://localhost:%d", daemonPort)))

		fmt.Println(color.YellowString("📝 Usage:\n"))
		fmt.Println("   " + color.CyanString("apilo daemon status") + " - Check daemon status")
		fmt.Println("   " + color.CyanString("apilo daemon logs") + " - View logs")
		fmt.Println("   " + color.CyanString("apilo daemon stop") + " - Stop daemon\n")

	} else {
		// Start daemon in foreground
		fmt.Println(color.YellowString("🚀 Starting daemon (foreground mode)...\n"))

		service, err := daemon.NewService(config)
		if err != nil {
			color.Red("❌ Failed to create service: %v\n", err)
			return
		}

		if err := service.Start(); err != nil {
			color.Red("❌ Daemon error: %v\n", err)
			return
		}
	}
}

func stopDaemon() {
	color.Cyan("\n╔═══════════════════════════════════════════════════════════════════╗")
	color.Cyan("║                Stopping Apilo Daemon                              ║")
	color.Cyan("╚═══════════════════════════════════════════════════════════════════╝\n")

	config := daemon.DefaultDaemonConfig()
	pidMgr := daemon.NewPIDManager(config.PIDFile)

	running, pid, err := pidMgr.IsRunning()
	if err != nil {
		color.Red("❌ Error: %v\n", err)
		return
	}

	if !running {
		color.Yellow("⚠️  Daemon is not running\n\n")
		return
	}

	fmt.Printf("🛑 Stopping daemon (PID: %d)...\n\n", pid)

	if err := pidMgr.Stop(); err != nil {
		color.Red("❌ Failed to stop daemon: %v\n", err)
		return
	}

	color.Green("✅ Daemon stopped successfully\n\n")
}

func checkDaemonStatus() {
	color.Cyan("\n╔═══════════════════════════════════════════════════════════════════╗")
	color.Cyan("║                  Apilo Daemon Status                              ║")
	color.Cyan("╚═══════════════════════════════════════════════════════════════════╝\n")

	config := daemon.DefaultDaemonConfig()
	pidMgr := daemon.NewPIDManager(config.PIDFile)

	running, pid, err := pidMgr.IsRunning()

	if err != nil {
		color.Yellow("⚠️  Status: %s\n", color.RedString("Not Running"))
		fmt.Printf("   Error: %v\n\n", err)
		return
	}

	if !running {
		color.Yellow("⚠️  Status: %s\n\n", color.RedString("Not Running"))
		fmt.Println(color.BlueString("💡 Start with: apilo daemon start\n"))
		return
	}

	fmt.Println(color.YellowString("📊 Daemon Status:\n"))
	fmt.Printf("   Status:  %s\n", color.GreenString("Running"))
	fmt.Printf("   PID:     %s\n", color.CyanString(fmt.Sprintf("%d", pid)))
	fmt.Printf("   Port:    %s\n", color.CyanString(fmt.Sprintf("%d", config.Port)))
	fmt.Printf("   IPC:     %s\n\n", color.CyanString(fmt.Sprintf("http://localhost:%d", config.Port)))

	// Get detailed status from IPC
	ipcURL := fmt.Sprintf("http://localhost:%d", config.Port)
	metricsURL := fmt.Sprintf("%s/metrics", ipcURL)
	cacheURL := fmt.Sprintf("%s/cache/stats?format=visual", ipcURL)

	// Fetch and display metrics
	if metrics := fetchMetrics(metricsURL); metrics != nil {
		fmt.Println(color.YellowString("📈 Performance Metrics:\n"))
		fmt.Printf("   Total Requests:  %s\n", color.CyanString(fmt.Sprintf("%d", metrics.TotalRequests)))
		fmt.Printf("   Cache Hit Ratio: %s\n", color.GreenString(fmt.Sprintf("%.2f%%", metrics.CacheHitRatio*100)))
		fmt.Printf("   Avg Latency:     %s\n", color.CyanString(fmt.Sprintf("%v", metrics.AvgLatency)))
		fmt.Printf("   Memory Usage:    %s\n", color.CyanString(fmt.Sprintf("%.2f MB", metrics.MemoryUsageMB)))
		fmt.Println()
	}

	// Fetch and display cache visualization
	if cacheVisual := fetchCacheVisual(cacheURL); cacheVisual != "" {
		fmt.Print(cacheVisual)
	}

	fmt.Println(color.YellowString("🎯 Available Commands:\n"))
	fmt.Println("   " + color.CyanString("apilo daemon logs") + " - View logs")
	fmt.Println("   " + color.CyanString("apilo daemon stop") + " - Stop daemon")
	fmt.Println("   " + color.CyanString("apilo daemon restart") + " - Restart daemon\n")
}

func restartDaemon() {
	color.Cyan("\n╔═══════════════════════════════════════════════════════════════════╗")
	color.Cyan("║                Restarting Apilo Daemon                            ║")
	color.Cyan("╚═══════════════════════════════════════════════════════════════════╝\n")

	// Stop daemon
	config := daemon.DefaultDaemonConfig()
	pidMgr := daemon.NewPIDManager(config.PIDFile)

	if running, pid, _ := pidMgr.IsRunning(); running {
		fmt.Printf("🛑 Stopping daemon (PID: %d)...\n", pid)
		if err := pidMgr.Stop(); err != nil {
			color.Red("❌ Failed to stop daemon: %v\n", err)
			return
		}
		color.Green("✅ Daemon stopped\n\n")
	}

	// Start daemon
	fmt.Println("🚀 Starting daemon...\n")
	startDaemon()
}

func viewDaemonLogs() {
	color.Cyan("\n╔═══════════════════════════════════════════════════════════════════╗")
	color.Cyan("║                    Apilo Daemon Logs                              ║")
	color.Cyan("╚═══════════════════════════════════════════════════════════════════╝\n")

	config := daemon.DefaultDaemonConfig()
	logFile := config.LogFile

	// Expand ~ to home directory
	if strings.HasPrefix(logFile, "~/") {
		home, _ := os.UserHomeDir()
		logFile = filepath.Join(home, logFile[2:])
	}

	// Check if log file exists
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		color.Yellow("⚠️  Log file not found: %s\n", logFile)
		fmt.Println(color.BlueString("💡 Logs will be created when daemon starts\n"))
		return
	}

	// Read and display last 50 lines
	content, err := os.ReadFile(logFile)
	if err != nil {
		color.Red("❌ Failed to read log file: %v\n", err)
		return
	}

	lines := strings.Split(string(content), "\n")
	start := 0
	if len(lines) > 50 {
		start = len(lines) - 50
	}

	fmt.Printf("📄 Log file: %s\n", color.CyanString(logFile))
	fmt.Println(color.YellowString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"))

	for _, line := range lines[start:] {
		if line != "" {
			fmt.Println(line)
		}
	}

	fmt.Println(color.YellowString("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(color.BlueString("💡 Use 'tail -f %s' for live logs\n", logFile))
}

// fetchMetrics fetches metrics from the daemon IPC server
func fetchMetrics(url string) *daemon.MetricsStats {
	client := &http.Client{Timeout: 2 * time.Second}

	resp, err := client.Get(url)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	var metrics daemon.MetricsStats
	if err := json.NewDecoder(resp.Body).Decode(&metrics); err != nil {
		return nil
	}

	return &metrics
}

// fetchCacheVisual fetches cache visualization from the daemon IPC server
func fetchCacheVisual(url string) string {
	client := &http.Client{Timeout: 2 * time.Second}

	resp, err := client.Get(url)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ""
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}

	return string(body)
}
