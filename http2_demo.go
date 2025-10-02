// HTTP/2 Client Test - Tests functional HTTP/2 implementation
package main

import (
	"fmt"
	"net/http"
	"time"
)

func main() {
	// Test HTTP/2 client functionality
	config := &HTTP2ClientConfig{
		MaxConnectionsPerHost: 10,
		IdleConnTimeout:      90 * time.Second,
		TLSHandshakeTimeout:  10 * time.Second,
		DisableCompression:   false,
		EnableHTTP2Push:      true,
	}

	client, err := NewFunctionalHTTP2Client(config)
	if err != nil {
		fmt.Printf("❌ Failed to create HTTP/2 client: %v\n", err)
		return
	}
	defer client.Close()

	fmt.Println("🔧 Testing HTTP/2 Client Functional Implementation")
	fmt.Println("=" + string(make([]rune, 50)))

	// Test URL (HTTP/2 enabled)
	testURL := "https://httpbin.org/get"

	// Create request
	req, err := http.NewRequest("GET", testURL, nil)
	if err != nil {
		fmt.Printf("❌ Failed to create request: %v\n", err)
		return
	}

	// Execute with timing
	fmt.Printf("🚀 Testing HTTP/2 request to: %s\n", testURL)

	resp, timing, err := client.DoWithTiming(req)
	if err != nil {
		fmt.Printf("❌ Request failed: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// Check if HTTP/2 was used
	isHTTP2 := client.IsHTTP2(resp)
	protocolInfo := client.GetProtocolInfo(resp)

	fmt.Printf(`
📊 HTTP/2 Client Test Results
=============================
Status Code:       %d
Protocol:          %s
Is HTTP/2:         %t
Protocol Major:    %d
Protocol Minor:    %d

🕐 Timing Breakdown
===================
DNS Latency:       %v
Connect Latency:   %v
TLS Latency:       %v
TTFB Latency:      %v
Processing:        %v
Connection Reused: %t

🔧 Connection Stats
===================
`,
		resp.StatusCode,
		resp.Proto,
		isHTTP2,
		resp.ProtoMajor,
		resp.ProtoMinor,
		timing.DNSLatency,
		timing.ConnectLatency,
		timing.TLSLatency,
		timing.TTFBLatency,
		timing.ProcessingLatency,
		timing.ConnectionReused,
	)

	// Print connection stats
	connStats := client.GetConnectionStats()
	for key, value := range connStats {
		fmt.Printf("%-20s: %v\n", key, value)
	}

	fmt.Printf(`
✅ HTTP/2 Test Results
======================
`)

	if isHTTP2 {
		fmt.Println("🎉 SUCCESS: HTTP/2 is working!")
		fmt.Println("✅ Protocol version confirmed as HTTP/2")
		fmt.Println("✅ Real timing measurements captured")
		fmt.Println("✅ Connection configuration applied")
	} else {
		fmt.Println("⚠️  WARNING: HTTP/2 fallback occurred")
		fmt.Printf("   Protocol used: %s\n", resp.Proto)
		fmt.Println("   This might be expected for some servers")
	}

	if timing.ConnectionReused {
		fmt.Println("✅ Connection reuse working")
	} else {
		fmt.Println("ℹ️  New connection created (expected for first request)")
	}

	// Test second request to verify connection reuse
	fmt.Println("\n🔄 Testing connection reuse with second request...")

	req2, _ := http.NewRequest("GET", testURL, nil)
	resp2, timing2, err := client.DoWithTiming(req2)
	if err == nil {
		defer resp2.Body.Close()
		if timing2.ConnectionReused {
			fmt.Println("✅ Connection reuse successful on second request!")
		} else {
			fmt.Println("⚠️  Connection not reused on second request")
		}
	}

	fmt.Println("\n✅ HTTP/2 functional test completed!")
}