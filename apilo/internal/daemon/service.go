package daemon

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// Service represents the main daemon service
type Service struct {
	config       *DaemonConfig
	pidManager   *PIDManager
	ipcServer    *IPCServer
	optimizer    *Optimizer
	claudeClient *ClaudeClient
	metrics      *Metrics
	analytics    *Analytics
	logger       *Logger
	proxy        *ProxyManager
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
	startTime    time.Time
	mu           sync.RWMutex
}

// NewService creates a new daemon service
func NewService(config *DaemonConfig) (*Service, error) {
	if config == nil {
		config = DefaultDaemonConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Initialize logger
	logger, err := NewLogger(config.LogFile, ParseLogLevel(config.LogLevel))
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	service := &Service{
		config:     config,
		pidManager: NewPIDManager(config.PIDFile),
		metrics:    NewMetrics(),
		analytics:  NewAnalytics(1000), // Track last 1000 requests
		logger:     logger,
		ctx:        ctx,
		cancel:     cancel,
		startTime:  time.Now(),
	}

	// Initialize optimizer
	optimizer, err := NewOptimizer(config, service.logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create optimizer: %w", err)
	}
	service.optimizer = optimizer

	// Initialize Claude client (optional - only if API key is set)
	claudeClient, err := NewClaudeClient()
	if err != nil {
		service.logger.Warn("Claude client initialization failed: %v", err)
		service.logger.Info("Claude API features will be disabled")
	} else {
		service.claudeClient = claudeClient
		service.logger.Info("Claude API client initialized successfully")
	}

	// Initialize IPC server
	ipcServer := NewIPCServer(config.Port, service)
	service.ipcServer = ipcServer

	// Initialize proxy manager
	proxy := NewProxyManager(service.logger)
	service.proxy = proxy

	return service, nil
}

// Start starts the daemon service
func (s *Service) Start() error {
	// Check if already running
	if running, pid, _ := s.pidManager.IsRunning(); running {
		return fmt.Errorf("daemon already running (PID: %d)", pid)
	}

	// Write PID file
	if err := s.pidManager.Write(); err != nil {
		return fmt.Errorf("failed to write PID file: %w", err)
	}

	s.logger.Info("Starting apilo daemon (PID: %d)", os.Getpid())

	// Start proxy
	if err := s.proxy.Start(); err != nil {
		s.logger.Warn("Proxy failed to start: %v", err)
	}

	// Start IPC server
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		if err := s.ipcServer.Start(s.ctx); err != nil {
			s.logger.Error("IPC server error: %v", err)
		}
	}()

	s.logger.Info("Daemon started on port %d", s.config.Port)

	// Start metrics collection if enabled
	if s.config.MetricsEnabled {
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			s.collectMetrics(s.ctx)
		}()
	}

	// Setup signal handling
	s.setupSignalHandling()

	// Wait for shutdown
	s.wg.Wait()

	s.logger.Info("Daemon stopped")
	s.logger.Close()
	return nil
}

// Stop stops the daemon service
func (s *Service) Stop() error {
	s.logger.Info("Stopping daemon...")

	s.cancel()
	s.wg.Wait()

	// Stop proxy
	if s.proxy != nil {
		s.proxy.Stop()
	}

	if err := s.pidManager.Remove(); err != nil {
		s.logger.Warn("Failed to remove PID file: %v", err)
	}

	s.logger.Close()
	return nil
}

// GetStatus returns the current daemon status
func (s *Service) GetStatus() *DaemonStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	running, pid, _ := s.pidManager.IsRunning()

	stats := s.metrics.GetStats()

	return &DaemonStatus{
		Running:       running,
		PID:           pid,
		Uptime:        time.Since(s.startTime),
		Port:          s.config.Port,
		TotalRequests: stats.TotalRequests,
		CacheHitRatio: stats.CacheHitRatio,
		AvgLatency:    stats.AvgLatency,
		MemoryUsageMB: stats.MemoryUsageMB,
		CPUPercent:    stats.CPUPercent,
		ClaudeMetrics: stats.ClaudeMetrics,
	}
}

// Optimize processes an optimization request
func (s *Service) Optimize(req *OptimizationRequest) (*OptimizationResponse, error) {
	s.metrics.IncrementRequests()

	start := time.Now()
	resp, err := s.optimizer.Optimize(req)
	latency := time.Since(start)

	// Record analytics
	record := RequestRecord{
		Timestamp:  start,
		URL:        req.URL,
		Method:     req.Method,
		StatusCode: 0,
		Latency:    int64(latency),
		CacheHit:   false,
	}

	if err != nil {
		s.metrics.IncrementErrors()
		s.logger.Error("Optimization failed for %s: %v", req.URL, err)
		record.Error = err.Error()
		s.analytics.RecordRequest(record)
		return nil, err
	}

	record.StatusCode = resp.StatusCode
	record.CacheHit = resp.CacheHit

	// Extract token usage from response metadata
	if resp.Metadata.TokenUsage != nil {
		record.InputTokens = resp.Metadata.TokenUsage.InputTokens
		record.OutputTokens = resp.Metadata.TokenUsage.OutputTokens
		record.TotalTokens = resp.Metadata.TokenUsage.TotalTokens
		record.IsEstimated = resp.Metadata.TokenUsage.IsEstimated
	}

	s.metrics.RecordLatency(latency)
	if resp.CacheHit {
		s.metrics.IncrementCacheHits()
	} else {
		s.metrics.IncrementCacheMisses()
	}

	s.logger.LogOptimization(req.URL, resp.CacheHit, latency)
	s.analytics.RecordRequest(record)

	resp.Latency = latency
	return resp, nil
}

// OptimizeWithClaude processes an optimization request with Claude API analysis
func (s *Service) OptimizeWithClaude(req *OptimizationRequest, prompt string, maxTokens int) (*OptimizationResponse, error) {
	if s.claudeClient == nil {
		return nil, fmt.Errorf("Claude API client not initialized (set ANTHROPIC_API_KEY)")
	}

	s.metrics.IncrementRequests()

	start := time.Now()

	// Make Claude API request
	claudeResp, err := s.claudeClient.MakeRequest(prompt, maxTokens)
	if err != nil {
		s.metrics.IncrementErrors()
		return nil, fmt.Errorf("Claude API error: %w", err)
	}

	latency := time.Since(start)
	s.metrics.RecordLatency(latency)

	// Track Claude tokens in service metrics
	claudeMetrics := s.claudeClient.GetMetrics()
	s.metrics.IncrementClaudeTokens(claudeMetrics.InputTokens, claudeMetrics.OutputTokens)
	s.metrics.RecordClaudeCost(claudeMetrics.Cost)

	// Extract response text
	var responseText string
	if len(claudeResp.Content) > 0 {
		responseText = claudeResp.Content[0].Text
	}

	return &OptimizationResponse{
		StatusCode: 200,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       []byte(responseText),
		Latency:    latency,
		CacheHit:   false,
		Optimized:  true,
		Metadata: ResponseMetadata{
			CacheStatus:      "claude-api",
			OptimizationType: "ai-powered",
			ConnectionReused: false,
			HTTP2Used:        true,
			ClaudeTokens:     &claudeMetrics,
		},
	}, nil
}

// setupSignalHandling configures signal handlers for graceful shutdown
func (s *Service) setupSignalHandling() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		sig := <-sigChan
		s.logger.Info("Received signal: %v", sig)

		switch sig {
		case syscall.SIGHUP:
			// Reload configuration
			s.logger.Info("Reloading configuration...")
			// TODO: Implement config reload
		case syscall.SIGINT, syscall.SIGTERM:
			// Graceful shutdown
			s.Stop()
		}
	}()
}

// collectMetrics periodically collects system metrics
func (s *Service) collectMetrics(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.metrics.UpdateSystemMetrics()
		}
	}
}

// GetConfig returns the daemon configuration
func (s *Service) GetConfig() *DaemonConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config
}

// UpdateConfig updates the daemon configuration
func (s *Service) UpdateConfig(config *DaemonConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Validate config
	if config.Port < 1024 || config.Port > 65535 {
		return fmt.Errorf("invalid port: must be between 1024 and 65535")
	}

	s.config = config
	s.logger.Info("Configuration updated")
	return nil
}
