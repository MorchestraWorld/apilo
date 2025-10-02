// Package src provides a functional HTTP/2 client implementation
// This replaces the stub HTTP/2 client to enable actual HTTP/2 optimizations
package main

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"sync"
	"time"

	"golang.org/x/net/http2"
)

// FunctionalHTTP2Client provides a working HTTP/2 client with real optimizations
type FunctionalHTTP2Client struct {
	config     *HTTP2ClientConfig
	client     *http.Client
	lastTiming *HTTP2RequestTiming
	timingMu   sync.RWMutex
}

// NewFunctionalHTTP2Client creates a new functional HTTP/2 client
func NewFunctionalHTTP2Client(config *HTTP2ClientConfig) (*FunctionalHTTP2Client, error) {
	if config == nil {
		config = &HTTP2ClientConfig{
			MaxConnectionsPerHost: 10,
			IdleConnTimeout:      90 * time.Second,
			TLSHandshakeTimeout:  10 * time.Second,
			DisableCompression:   false,
			EnableHTTP2Push:      true,
		}
	}

	// Create transport with HTTP/2 configuration
	transport := &http.Transport{
		// Connection pooling configuration
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: config.MaxConnectionsPerHost,
		IdleConnTimeout:     config.IdleConnTimeout,

		// TLS configuration
		TLSHandshakeTimeout: config.TLSHandshakeTimeout,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: false,
			NextProtos:        []string{"h2", "http/1.1"}, // Prefer HTTP/2
		},

		// Connection settings
		DisableKeepAlives:     false,
		DisableCompression:    config.DisableCompression,
		MaxIdleConnsPerHost:   config.MaxConnectionsPerHost,
		ResponseHeaderTimeout: 30 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,

		// Dial configuration
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,

		// Force HTTP/2
		ForceAttemptHTTP2: true,
	}

	// Configure HTTP/2 transport
	if err := http2.ConfigureTransport(transport); err != nil {
		return nil, err
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}

	return &FunctionalHTTP2Client{
		config: config,
		client: client,
	}, nil
}

// DoWithTiming executes an HTTP request with detailed timing measurement
func (c *FunctionalHTTP2Client) DoWithTiming(req *http.Request) (*http.Response, *HTTP2RequestTiming, error) {
	timing := &HTTP2RequestTiming{}

	// Create a context for tracing
	ctx := req.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	var dnsStart, connectStart, tlsStart, requestStart time.Time
	var dnsEnd, connectEnd, tlsEnd time.Time
	var connectionReused bool

	// Set up HTTP tracing
	trace := &httptrace.ClientTrace{
		DNSStart: func(info httptrace.DNSStartInfo) {
			dnsStart = time.Now()
		},
		DNSDone: func(info httptrace.DNSDoneInfo) {
			dnsEnd = time.Now()
			timing.DNSLatency = dnsEnd.Sub(dnsStart)
		},
		ConnectStart: func(network, addr string) {
			connectStart = time.Now()
		},
		ConnectDone: func(network, addr string, err error) {
			connectEnd = time.Now()
			timing.ConnectLatency = connectEnd.Sub(connectStart)
		},
		TLSHandshakeStart: func() {
			tlsStart = time.Now()
		},
		TLSHandshakeDone: func(state tls.ConnectionState, err error) {
			tlsEnd = time.Now()
			timing.TLSLatency = tlsEnd.Sub(tlsStart)
		},
		GotConn: func(info httptrace.GotConnInfo) {
			connectionReused = info.Reused
			timing.ConnectionReused = connectionReused
		},
		WroteRequest: func(info httptrace.WroteRequestInfo) {
			requestStart = time.Now()
		},
		GotFirstResponseByte: func() {
			if !requestStart.IsZero() {
				timing.TTFBLatency = time.Since(requestStart)
			}
		},
	}

	// Add trace to context
	ctx = httptrace.WithClientTrace(ctx, trace)
	req = req.WithContext(ctx)

	// Execute request
	startTime := time.Now()
	resp, err := c.client.Do(req)
	endTime := time.Now()

	if err != nil {
		return nil, timing, err
	}

	// Calculate processing latency (server processing time)
	if !timing.TTFBLatency.IsZero() {
		timing.ProcessingLatency = timing.TTFBLatency - timing.ConnectLatency - timing.TLSLatency
		if timing.ProcessingLatency < 0 {
			timing.ProcessingLatency = 0
		}
	}

	// Store timing for later retrieval
	c.timingMu.Lock()
	c.lastTiming = timing
	c.timingMu.Unlock()

	return resp, timing, nil
}

// Do executes an HTTP request (compatibility with existing interface)
func (c *FunctionalHTTP2Client) Do(req *http.Request) (*http.Response, error) {
	resp, _, err := c.DoWithTiming(req)
	return resp, err
}

// GetLastRequestTiming returns timing for the last request (real data)
func (c *FunctionalHTTP2Client) GetLastRequestTiming() *HTTP2RequestTiming {
	c.timingMu.RLock()
	defer c.timingMu.RUnlock()

	if c.lastTiming == nil {
		// Return zero values if no request has been made
		return &HTTP2RequestTiming{}
	}

	// Return a copy to avoid race conditions
	return &HTTP2RequestTiming{
		DNSLatency:        c.lastTiming.DNSLatency,
		ConnectLatency:    c.lastTiming.ConnectLatency,
		TLSLatency:        c.lastTiming.TLSLatency,
		TTFBLatency:       c.lastTiming.TTFBLatency,
		ProcessingLatency: c.lastTiming.ProcessingLatency,
		ConnectionReused:  c.lastTiming.ConnectionReused,
	}
}

// Close closes the HTTP/2 client
func (c *FunctionalHTTP2Client) Close() error {
	// Close idle connections
	if transport, ok := c.client.Transport.(*http.Transport); ok {
		transport.CloseIdleConnections()
	}
	return nil
}

// GetConnectionStats returns connection pool statistics
func (c *FunctionalHTTP2Client) GetConnectionStats() map[string]interface{} {
	return map[string]interface{}{
		"max_idle_conns_per_host": c.config.MaxConnectionsPerHost,
		"idle_conn_timeout":       c.config.IdleConnTimeout.String(),
		"tls_handshake_timeout":   c.config.TLSHandshakeTimeout.String(),
		"compression_disabled":    c.config.DisableCompression,
		"http2_push_enabled":      c.config.EnableHTTP2Push,
	}
}

// IsHTTP2 checks if the last response used HTTP/2
func (c *FunctionalHTTP2Client) IsHTTP2(resp *http.Response) bool {
	return resp.ProtoMajor == 2
}

// GetProtocolInfo returns protocol information for the last response
func (c *FunctionalHTTP2Client) GetProtocolInfo(resp *http.Response) map[string]interface{} {
	return map[string]interface{}{
		"protocol":       resp.Proto,
		"protocol_major": resp.ProtoMajor,
		"protocol_minor": resp.ProtoMinor,
		"is_http2":       c.IsHTTP2(resp),
	}
}