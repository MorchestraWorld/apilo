package daemon

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

// PIDManager handles process ID file management
type PIDManager struct {
	pidFile string
}

// NewPIDManager creates a new PID manager
func NewPIDManager(pidFile string) *PIDManager {
	// Expand ~ to home directory
	if strings.HasPrefix(pidFile, "~/") {
		home, _ := os.UserHomeDir()
		pidFile = filepath.Join(home, pidFile[2:])
	}

	// Ensure directory exists
	dir := filepath.Dir(pidFile)
	os.MkdirAll(dir, 0755)

	return &PIDManager{
		pidFile: pidFile,
	}
}

// Write writes the current process ID to the PID file
func (pm *PIDManager) Write() error {
	pid := os.Getpid()
	return os.WriteFile(pm.pidFile, []byte(fmt.Sprintf("%d", pid)), 0644)
}

// Read reads the PID from the PID file
func (pm *PIDManager) Read() (int, error) {
	data, err := os.ReadFile(pm.pidFile)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, fmt.Errorf("daemon not running (PID file not found)")
		}
		return 0, fmt.Errorf("failed to read PID file: %w", err)
	}

	pidStr := strings.TrimSpace(string(data))
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return 0, fmt.Errorf("invalid PID in file: %w", err)
	}

	return pid, nil
}

// Remove removes the PID file
func (pm *PIDManager) Remove() error {
	return os.Remove(pm.pidFile)
}

// IsRunning checks if the process with the stored PID is running
func (pm *PIDManager) IsRunning() (bool, int, error) {
	pid, err := pm.Read()
	if err != nil {
		return false, 0, err
	}

	// Check if process exists by sending signal 0
	process, err := os.FindProcess(pid)
	if err != nil {
		return false, pid, nil
	}

	err = process.Signal(syscall.Signal(0))
	if err != nil {
		return false, pid, nil
	}

	return true, pid, nil
}

// Stop stops the daemon by sending SIGTERM to the process
func (pm *PIDManager) Stop() error {
	pid, err := pm.Read()
	if err != nil {
		return err
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("process not found: %w", err)
	}

	// Send SIGTERM for graceful shutdown
	if err := process.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("failed to stop daemon: %w", err)
	}

	return nil
}

// Restart restarts the daemon
func (pm *PIDManager) Restart() error {
	// Stop existing process
	if err := pm.Stop(); err != nil {
		return fmt.Errorf("failed to stop daemon: %w", err)
	}

	// Note: Actual restart requires external process to start new daemon
	// This function only handles stopping the old one
	return nil
}
