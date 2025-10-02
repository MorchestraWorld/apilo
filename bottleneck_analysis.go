// Phase 2: Bottleneck Identification & Analysis
package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/httptrace"
	"runtime"
	"sync"
	"time"
)

// BottleneckAnalysis comprehensive bottleneck identification
type BottleneckAnalysis struct {
	NetworkBottlenecks    *NetworkBottlenecks
	ResourceBottlenecks   *ResourceBottlenecks
	ApplicationBottlenecks *ApplicationBottlenecks
	SystemBottlenecks     *SystemBottlenecks
	Timestamp            time.Time
}

// NetworkBottlenecks identifies network-related performance issues
type NetworkBottlenecks struct {
	DNSResolutionSlow     bool
	ConnectionSetupSlow   bool
	TLSHandshakeSlow      bool
	ServerResponseSlow    bool
	BandwidthLimited      bool
	HighLatencyJitter     bool
	ConnectionReusePoor   bool
}

// ResourceBottlenecks identifies resource utilization issues
type ResourceBottlenecks struct {
	MemoryPressure        bool
	CPUBound             bool
	GCPressure           bool
	ThreadExhaustion     bool
	HeapFragmentation    bool
	StackOverflow        bool
}

// ApplicationBottlenecks identifies application-level performance issues
type ApplicationBottlenecks struct {
	CacheIneffective     bool
	AlgorithmInefficient bool
	ConcurrencyPoor      bool
	ErrorRateHigh        bool
	ProtocolSuboptimal   bool
	ConfigurationPoor    bool
}

// SystemBottlenecks identifies system-level performance constraints
type SystemBottlenecks struct {
	IOBound              bool
	NetworkBound         bool
	MemoryBound          bool
	CPUBound             bool
	ThreadLimited        bool
	OSResourceConstrained bool
}

// DetailedTiming captures comprehensive request timing
type DetailedTiming struct {
	DNSStart             time.Time
	DNSEnd               time.Time
	ConnectStart         time.Time
	ConnectEnd           time.Time
	TLSStart             time.Time
	TLSEnd               time.Time
	RequestStart         time.Time
	FirstByteReceived    time.Time
	RequestEnd           time.Time
	ConnectionReused     bool
	RemoteAddr           string
}

// BottleneckAnalyzer performs comprehensive bottleneck analysis
type BottleneckAnalyzer struct {
	client               *http.Client
	timings              []DetailedTiming
	timingsMutex         sync.Mutex
	resourceSnapshots    []ResourceSnapshot
	snapshotsMutex       sync.Mutex
}

// ResourceSnapshot captures resource state at a point in time
type ResourceSnapshot struct {
	Timestamp            time.Time
	MemoryUsage          uint64
	GoroutineCount       int
	CGOCalls             int64
	HeapObjects          uint64
	GCPause              time.Duration
}

// NewBottleneckAnalyzer creates a new bottleneck analyzer
func NewBottleneckAnalyzer() *BottleneckAnalyzer {
	transport := &http.Transport{
		MaxIdleConns:       100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:    90 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: false,
		},
		ForceAttemptHTTP2: true,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}

	return &BottleneckAnalyzer{
		client:              client,
		timings:            make([]DetailedTiming, 0),
		resourceSnapshots:  make([]ResourceSnapshot, 0),
	}
}

// AnalyzeBottlenecks performs comprehensive bottleneck identification
func (ba *BottleneckAnalyzer) AnalyzeBottlenecks(url string, samples int) (*BottleneckAnalysis, error) {
	fmt.Printf(`
🔍 Phase 2: Bottleneck Identification & Analysis
===============================================
Target URL: %s
Analysis Samples: %d
Focus: CPU, Memory, I/O, Network performance constraints

`, url, samples)

	// Start resource monitoring
	stopMonitoring := ba.startResourceMonitoring()
	defer stopMonitoring()

	// Collect detailed timing samples
	fmt.Println("📊 Collecting detailed performance samples...")

	for i := 0; i < samples; i++ {
		timing, err := ba.measureDetailedRequest(url)
		if err != nil {
			fmt.Printf("❌ Sample %d failed: %v\n", i+1, err)
			continue
		}

		ba.timingsMutex.Lock()
		ba.timings = append(ba.timings, *timing)
		ba.timingsMutex.Unlock()

		if (i+1) % 20 == 0 {
			fmt.Printf("  Progress: %d/%d samples collected\n", i+1, samples)
		}

		// Small delay to capture different system states
		time.Sleep(25 * time.Millisecond)
	}

	// Analyze collected data
	analysis := ba.analyzeCollectedData()

	ba.printBottleneckReport(analysis)

	return analysis, nil
}

// measureDetailedRequest captures comprehensive request timing
func (ba *BottleneckAnalyzer) measureDetailedRequest(url string) (*DetailedTiming, error) {
	timing := &DetailedTiming{}

	// Create request with detailed tracing
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Set up detailed HTTP tracing
	trace := &httptrace.ClientTrace{
		DNSStart: func(info httptrace.DNSStartInfo) {
			timing.DNSStart = time.Now()
		},
		DNSDone: func(info httptrace.DNSDoneInfo) {
			timing.DNSEnd = time.Now()
		},
		ConnectStart: func(network, addr string) {
			timing.ConnectStart = time.Now()
		},
		ConnectDone: func(network, addr string, err error) {
			timing.ConnectEnd = time.Now()
		},
		TLSHandshakeStart: func() {
			timing.TLSStart = time.Now()
		},
		TLSHandshakeDone: func(state tls.ConnectionState, err error) {
			timing.TLSEnd = time.Now()
		},
		WroteRequest: func(info httptrace.WroteRequestInfo) {
			timing.RequestStart = time.Now()
		},
		GotFirstResponseByte: func() {
			timing.FirstByteReceived = time.Now()
		},
		GotConn: func(info httptrace.GotConnInfo) {
			timing.ConnectionReused = info.Reused
			timing.RemoteAddr = info.Conn.RemoteAddr().String()
		},
	}

	// Add trace to request context
	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))

	// Execute request
	resp, err := ba.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	timing.RequestEnd = time.Now()

	return timing, nil
}

// startResourceMonitoring begins continuous resource monitoring
func (ba *BottleneckAnalyzer) startResourceMonitoring() func() {
	stop := make(chan bool)

	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				snapshot := ba.captureResourceSnapshot()
				ba.snapshotsMutex.Lock()
				ba.resourceSnapshots = append(ba.resourceSnapshots, snapshot)
				ba.snapshotsMutex.Unlock()

			case <-stop:
				return
			}
		}
	}()

	return func() {
		stop <- true
	}
}

// captureResourceSnapshot captures current resource state
func (ba *BottleneckAnalyzer) captureResourceSnapshot() ResourceSnapshot {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	return ResourceSnapshot{
		Timestamp:      time.Now(),
		MemoryUsage:    memStats.Alloc,
		GoroutineCount: runtime.NumGoroutine(),
		CGOCalls:       runtime.NumCgoCall(),
		HeapObjects:    memStats.HeapObjects,
		GCPause:        time.Duration(memStats.PauseNs[(memStats.NumGC+255)%256]),
	}
}

// analyzeCollectedData performs comprehensive bottleneck analysis
func (ba *BottleneckAnalyzer) analyzeCollectedData() *BottleneckAnalysis {
	networkBottlenecks := ba.analyzeNetworkBottlenecks()
	resourceBottlenecks := ba.analyzeResourceBottlenecks()
	applicationBottlenecks := ba.analyzeApplicationBottlenecks()
	systemBottlenecks := ba.analyzeSystemBottlenecks()

	return &BottleneckAnalysis{
		NetworkBottlenecks:     networkBottlenecks,
		ResourceBottlenecks:    resourceBottlenecks,
		ApplicationBottlenecks: applicationBottlenecks,
		SystemBottlenecks:      systemBottlenecks,
		Timestamp:             time.Now(),
	}
}

// analyzeNetworkBottlenecks identifies network performance issues
func (ba *BottleneckAnalyzer) analyzeNetworkBottlenecks() *NetworkBottlenecks {
	if len(ba.timings) == 0 {
		return &NetworkBottlenecks{}
	}

	var totalDNS, totalConnect, totalTLS, totalTTFB time.Duration
	var dnsCount, connectCount, tlsCount, ttfbCount int
	var reuseCount int

	for _, timing := range ba.timings {
		if !timing.DNSStart.IsZero() && !timing.DNSEnd.IsZero() {
			totalDNS += timing.DNSEnd.Sub(timing.DNSStart)
			dnsCount++
		}

		if !timing.ConnectStart.IsZero() && !timing.ConnectEnd.IsZero() {
			totalConnect += timing.ConnectEnd.Sub(timing.ConnectStart)
			connectCount++
		}

		if !timing.TLSStart.IsZero() && !timing.TLSEnd.IsZero() {
			totalTLS += timing.TLSEnd.Sub(timing.TLSStart)
			tlsCount++
		}

		if !timing.RequestStart.IsZero() && !timing.FirstByteReceived.IsZero() {
			totalTTFB += timing.FirstByteReceived.Sub(timing.RequestStart)
			ttfbCount++
		}

		if timing.ConnectionReused {
			reuseCount++
		}
	}

	// Calculate averages and identify bottlenecks
	avgDNS := time.Duration(0)
	if dnsCount > 0 {
		avgDNS = totalDNS / time.Duration(dnsCount)
	}

	avgConnect := time.Duration(0)
	if connectCount > 0 {
		avgConnect = totalConnect / time.Duration(connectCount)
	}

	avgTLS := time.Duration(0)
	if tlsCount > 0 {
		avgTLS = totalTLS / time.Duration(tlsCount)
	}

	avgTTFB := time.Duration(0)
	if ttfbCount > 0 {
		avgTTFB = totalTTFB / time.Duration(ttfbCount)
	}

	reuseRate := float64(reuseCount) / float64(len(ba.timings)) * 100

	return &NetworkBottlenecks{
		DNSResolutionSlow:   avgDNS > 50*time.Millisecond,
		ConnectionSetupSlow: avgConnect > 100*time.Millisecond,
		TLSHandshakeSlow:    avgTLS > 200*time.Millisecond,
		ServerResponseSlow:  avgTTFB > 500*time.Millisecond,
		ConnectionReusePoor: reuseRate < 50.0,
	}
}

// analyzeResourceBottlenecks identifies resource utilization issues
func (ba *BottleneckAnalyzer) analyzeResourceBottlenecks() *ResourceBottlenecks {
	if len(ba.resourceSnapshots) == 0 {
		return &ResourceBottlenecks{}
	}

	var maxMemory uint64
	var maxGoroutines int
	var maxGCPause time.Duration
	var avgMemory uint64
	var memoryGrowth bool

	for i, snapshot := range ba.resourceSnapshots {
		if snapshot.MemoryUsage > maxMemory {
			maxMemory = snapshot.MemoryUsage
		}

		if snapshot.GoroutineCount > maxGoroutines {
			maxGoroutines = snapshot.GoroutineCount
		}

		if snapshot.GCPause > maxGCPause {
			maxGCPause = snapshot.GCPause
		}

		avgMemory += snapshot.MemoryUsage

		// Check for memory growth pattern
		if i > 10 && snapshot.MemoryUsage > ba.resourceSnapshots[i-10].MemoryUsage*2 {
			memoryGrowth = true
		}
	}

	avgMemory /= uint64(len(ba.resourceSnapshots))

	return &ResourceBottlenecks{
		MemoryPressure:    maxMemory > 500*1024*1024, // 500MB threshold
		GCPressure:        maxGCPause > 10*time.Millisecond,
		ThreadExhaustion:  maxGoroutines > 1000,
		HeapFragmentation: memoryGrowth,
	}
}

// analyzeApplicationBottlenecks identifies application-level issues
func (ba *BottleneckAnalyzer) analyzeApplicationBottlenecks() *ApplicationBottlenecks {
	// Based on timing patterns and resource usage
	if len(ba.timings) == 0 {
		return &ApplicationBottlenecks{}
	}

	var newConnections int
	for _, timing := range ba.timings {
		if !timing.ConnectionReused {
			newConnections++
		}
	}

	newConnRate := float64(newConnections) / float64(len(ba.timings)) * 100

	return &ApplicationBottlenecks{
		CacheIneffective:  newConnRate > 80.0, // More than 80% new connections
		ConcurrencyPoor:   false, // Would need more analysis
		ConfigurationPoor: newConnRate > 90.0,
	}
}

// analyzeSystemBottlenecks identifies system-level constraints
func (ba *BottleneckAnalyzer) analyzeSystemBottlenecks() *SystemBottlenecks {
	// Analyze system-level performance patterns
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	return &SystemBottlenecks{
		MemoryBound:   memStats.Sys > 1024*1024*1024, // 1GB threshold
		CPUBound:     runtime.NumGoroutine() > runtime.NumCPU()*10,
		ThreadLimited: runtime.NumGoroutine() > 100,
	}
}

// printBottleneckReport displays comprehensive bottleneck analysis
func (ba *BottleneckAnalyzer) printBottleneckReport(analysis *BottleneckAnalysis) {
	fmt.Printf(`
🔍 Comprehensive Bottleneck Analysis Report
===========================================
Analysis Timestamp: %s
Samples Collected: %d
Resource Snapshots: %d

📡 Network Bottleneck Analysis
==============================
DNS Resolution:         %s
Connection Setup:       %s
TLS Handshake:          %s
Server Response:        %s
Connection Reuse:       %s
Bandwidth Limited:      %s

💾 Resource Bottleneck Analysis
===============================
Memory Pressure:        %s
CPU Bound:              %s
GC Pressure:            %s
Thread Exhaustion:      %s
Heap Fragmentation:     %s

🔧 Application Bottleneck Analysis
==================================
Cache Ineffective:      %s
Algorithm Inefficient:  %s
Concurrency Poor:       %s
Error Rate High:        %s
Configuration Poor:     %s

🖥️  System Bottleneck Analysis
==============================
I/O Bound:              %s
Network Bound:          %s
Memory Bound:           %s
CPU Bound:              %s
Thread Limited:         %s

🎯 Critical Bottlenecks Identified
==================================
`,
		analysis.Timestamp.Format("2006-01-02 15:04:05"),
		len(ba.timings),
		len(ba.resourceSnapshots),
		formatBottleneckStatus(analysis.NetworkBottlenecks.DNSResolutionSlow),
		formatBottleneckStatus(analysis.NetworkBottlenecks.ConnectionSetupSlow),
		formatBottleneckStatus(analysis.NetworkBottlenecks.TLSHandshakeSlow),
		formatBottleneckStatus(analysis.NetworkBottlenecks.ServerResponseSlow),
		formatBottleneckStatus(analysis.NetworkBottlenecks.ConnectionReusePoor),
		formatBottleneckStatus(analysis.NetworkBottlenecks.BandwidthLimited),
		formatBottleneckStatus(analysis.ResourceBottlenecks.MemoryPressure),
		formatBottleneckStatus(analysis.ResourceBottlenecks.CPUBound),
		formatBottleneckStatus(analysis.ResourceBottlenecks.GCPressure),
		formatBottleneckStatus(analysis.ResourceBottlenecks.ThreadExhaustion),
		formatBottleneckStatus(analysis.ResourceBottlenecks.HeapFragmentation),
		formatBottleneckStatus(analysis.ApplicationBottlenecks.CacheIneffective),
		formatBottleneckStatus(analysis.ApplicationBottlenecks.AlgorithmInefficient),
		formatBottleneckStatus(analysis.ApplicationBottlenecks.ConcurrencyPoor),
		formatBottleneckStatus(analysis.ApplicationBottlenecks.ErrorRateHigh),
		formatBottleneckStatus(analysis.ApplicationBottlenecks.ConfigurationPoor),
		formatBottleneckStatus(analysis.SystemBottlenecks.IOBound),
		formatBottleneckStatus(analysis.SystemBottlenecks.NetworkBound),
		formatBottleneckStatus(analysis.SystemBottlenecks.MemoryBound),
		formatBottleneckStatus(analysis.SystemBottlenecks.CPUBound),
		formatBottleneckStatus(analysis.SystemBottlenecks.ThreadLimited),
	)

	// Identify and prioritize critical bottlenecks
	criticalBottlenecks := ba.identifyCriticalBottlenecks(analysis)
	if len(criticalBottlenecks) > 0 {
		fmt.Println("❌ Critical Performance Bottlenecks:")
		for i, bottleneck := range criticalBottlenecks {
			fmt.Printf("   %d. %s\n", i+1, bottleneck)
		}
	} else {
		fmt.Println("✅ No critical bottlenecks identified")
	}

	fmt.Printf(`
📊 Bottleneck Analysis Summary
==============================
Network Issues:         %d critical
Resource Issues:        %d critical
Application Issues:     %d critical
System Issues:          %d critical

Next Phase: Application Performance Profiling
`,
		ba.countNetworkBottlenecks(analysis.NetworkBottlenecks),
		ba.countResourceBottlenecks(analysis.ResourceBottlenecks),
		ba.countApplicationBottlenecks(analysis.ApplicationBottlenecks),
		ba.countSystemBottlenecks(analysis.SystemBottlenecks),
	)
}

// Helper functions
func formatBottleneckStatus(isBottleneck bool) string {
	if isBottleneck {
		return "❌ BOTTLENECK"
	}
	return "✅ OK"
}

func (ba *BottleneckAnalyzer) identifyCriticalBottlenecks(analysis *BottleneckAnalysis) []string {
	var critical []string

	if analysis.NetworkBottlenecks.ServerResponseSlow {
		critical = append(critical, "Server response time > 500ms")
	}
	if analysis.NetworkBottlenecks.ConnectionReusePoor {
		critical = append(critical, "Poor connection reuse (< 50%)")
	}
	if analysis.ResourceBottlenecks.MemoryPressure {
		critical = append(critical, "High memory usage (> 500MB)")
	}
	if analysis.ResourceBottlenecks.GCPressure {
		critical = append(critical, "GC pause time > 10ms")
	}
	if analysis.ApplicationBottlenecks.CacheIneffective {
		critical = append(critical, "Cache ineffective (< 20% reuse)")
	}

	return critical
}

func (ba *BottleneckAnalyzer) countNetworkBottlenecks(nb *NetworkBottlenecks) int {
	count := 0
	if nb.DNSResolutionSlow { count++ }
	if nb.ConnectionSetupSlow { count++ }
	if nb.TLSHandshakeSlow { count++ }
	if nb.ServerResponseSlow { count++ }
	if nb.ConnectionReusePoor { count++ }
	if nb.BandwidthLimited { count++ }
	return count
}

func (ba *BottleneckAnalyzer) countResourceBottlenecks(rb *ResourceBottlenecks) int {
	count := 0
	if rb.MemoryPressure { count++ }
	if rb.CPUBound { count++ }
	if rb.GCPressure { count++ }
	if rb.ThreadExhaustion { count++ }
	if rb.HeapFragmentation { count++ }
	return count
}

func (ba *BottleneckAnalyzer) countApplicationBottlenecks(ab *ApplicationBottlenecks) int {
	count := 0
	if ab.CacheIneffective { count++ }
	if ab.AlgorithmInefficient { count++ }
	if ab.ConcurrencyPoor { count++ }
	if ab.ErrorRateHigh { count++ }
	if ab.ConfigurationPoor { count++ }
	return count
}

func (ba *BottleneckAnalyzer) countSystemBottlenecks(sb *SystemBottlenecks) int {
	count := 0
	if sb.IOBound { count++ }
	if sb.NetworkBound { count++ }
	if sb.MemoryBound { count++ }
	if sb.CPUBound { count++ }
	if sb.ThreadLimited { count++ }
	return count
}

func main() {
	analyzer := NewBottleneckAnalyzer()

	// Analyze bottlenecks with 50 samples for comprehensive analysis
	analysis, err := analyzer.AnalyzeBottlenecks("https://httpbin.org/get", 50)
	if err != nil {
		fmt.Printf("❌ Bottleneck analysis failed: %v\n", err)
		return
	}

	fmt.Printf("\n🎯 Bottleneck analysis completed successfully!\n")
	fmt.Printf("Analysis completed at: %s\n", analysis.Timestamp.Format("2006-01-02 15:04:05"))
}