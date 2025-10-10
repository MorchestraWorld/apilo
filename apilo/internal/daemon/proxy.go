package daemon

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

type ProxyManager struct {
	cmd    *exec.Cmd
	logger *Logger
}

func NewProxyManager(logger *Logger) *ProxyManager {
	return &ProxyManager{logger: logger}
}

func (pm *ProxyManager) Start() error {
	script := "/tmp/anthropic_proxy.py"
	if _, err := os.Stat(script); os.IsNotExist(err) {
		return fmt.Errorf("proxy script not found: %s", script)
	}

	pm.cmd = exec.Command("mitmdump", "-p", "8765", "-s", script, "--set", "block_global=false", "-q")
	pm.cmd.Env = append(os.Environ(), "HTTPS_PROXY=http://localhost:8765")

	if err := pm.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start proxy: %w", err)
	}

	pm.logger.Info("Proxy started (PID: %d, port: 8765)", pm.cmd.Process.Pid)
	os.Setenv("HTTPS_PROXY", "http://localhost:8765")
	return nil
}

func (pm *ProxyManager) Stop() {
	if pm.cmd != nil && pm.cmd.Process != nil {
		pm.cmd.Process.Signal(syscall.SIGTERM)
		pm.logger.Info("Proxy stopped")
	}
}
