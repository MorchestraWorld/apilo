package build

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Build-time variables injected via -ldflags
var (
	// SourceDir is the directory where apilo was built from
	SourceDir = "unknown"

	// Version is the apilo version
	Version = "2.0.0"

	// BuildTime is when the binary was built
	BuildTime = "unknown"

	// Commit is the git commit hash
	Commit = "unknown"
)

// BuildInfo contains comprehensive build and runtime information
type BuildInfo struct {
	Version        string
	BuildTime      string
	Commit         string
	SourceDir      string
	ExecutablePath string
	GoVersion      string
	GOOS           string
	GOARCH         string
}

// GetBuildInfo returns comprehensive build and runtime information
func GetBuildInfo() (*BuildInfo, error) {
	execPath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("failed to get executable path: %w", err)
	}

	// Resolve symlinks to get the actual binary path
	realExecPath, err := filepath.EvalSymlinks(execPath)
	if err != nil {
		realExecPath = execPath // fallback to original path
	}

	return &BuildInfo{
		Version:        Version,
		BuildTime:      BuildTime,
		Commit:         Commit,
		SourceDir:      SourceDir,
		ExecutablePath: realExecPath,
		GoVersion:      runtime.Version(),
		GOOS:           runtime.GOOS,
		GOARCH:         runtime.GOARCH,
	}, nil
}

// IsSourceDirValid checks if the embedded source directory exists and contains expected files
func (bi *BuildInfo) IsSourceDirValid() bool {
	if bi.SourceDir == "unknown" || bi.SourceDir == "" {
		return false
	}

	// Check for key files that should exist in the source directory
	expectedFiles := []string{
		"main.go",
		"go.mod",
		"Makefile",
		"cmd",
	}

	for _, file := range expectedFiles {
		path := filepath.Join(bi.SourceDir, file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return false
		}
	}

	return true
}

// GetSourceDirWithFallback returns the source directory with comprehensive fallback detection
func (bi *BuildInfo) GetSourceDirWithFallback() (string, error) {
	// First try the LDFLAGS-injected source directory
	if bi.IsSourceDirValid() {
		return bi.SourceDir, nil
	}

	// Fallback: try to find git repository root from executable location
	execDir := filepath.Dir(bi.ExecutablePath)

	// Walk up directory tree looking for .git directory
	currentDir := execDir
	for {
		gitDir := filepath.Join(currentDir, ".git")
		if _, err := os.Stat(gitDir); err == nil {
			// Found .git directory, check if this looks like apilo source
			if isValidSourceDirectory(currentDir) {
				return currentDir, nil
			}
		}

		parent := filepath.Dir(currentDir)
		if parent == currentDir {
			// Reached filesystem root
			break
		}
		currentDir = parent
	}

	return "", fmt.Errorf("could not locate apilo source directory (embedded: %s, executable: %s)",
		bi.SourceDir, bi.ExecutablePath)
}

// isValidSourceDirectory checks if a directory contains apilo source files
func isValidSourceDirectory(dir string) bool {
	requiredFiles := []string{
		"main.go",
		"go.mod",
		"Makefile",
	}

	for _, file := range requiredFiles {
		if _, err := os.Stat(filepath.Join(dir, file)); os.IsNotExist(err) {
			return false
		}
	}

	// Additional check: verify go.mod contains apilo module
	goModPath := filepath.Join(dir, "go.mod")
	if content, err := os.ReadFile(goModPath); err == nil {
		contentStr := string(content)
		if strings.Contains(contentStr, "apilo") {
			return true
		}
	}

	return false
}

// PrintBuildInfo prints formatted build information
func (bi *BuildInfo) PrintBuildInfo() {
	fmt.Printf("📦 API Latency Optimizer (apilo) - Build Information\n")
	fmt.Printf("====================================================\n")
	fmt.Printf("Version:        %s\n", bi.Version)
	fmt.Printf("Build Time:     %s\n", bi.BuildTime)
	fmt.Printf("Commit:         %s\n", bi.Commit)
	fmt.Printf("Source Dir:     %s", bi.SourceDir)
	if bi.IsSourceDirValid() {
		fmt.Printf(" ✅\n")
	} else {
		fmt.Printf(" ❌\n")
	}
	fmt.Printf("Executable:     %s\n", bi.ExecutablePath)
	fmt.Printf("Go Version:     %s\n", bi.GoVersion)
	fmt.Printf("Platform:       %s/%s\n", bi.GOOS, bi.GOARCH)

	// Show resolved source directory if different
	if sourceDir, err := bi.GetSourceDirWithFallback(); err == nil && sourceDir != bi.SourceDir {
		fmt.Printf("Resolved Source: %s ✅\n", sourceDir)
	}
}

// CanSelfUpdate checks if self-update is possible
func (bi *BuildInfo) CanSelfUpdate() (bool, string, error) {
	sourceDir, err := bi.GetSourceDirWithFallback()
	if err != nil {
		return false, "", err
	}

	// Check if source directory contains necessary build files
	makefilePath := filepath.Join(sourceDir, "Makefile")
	if _, err := os.Stat(makefilePath); os.IsNotExist(err) {
		return false, "", fmt.Errorf("Makefile not found in source directory")
	}

	return true, sourceDir, nil
}
