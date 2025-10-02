package cmd

import (
	"apilo/internal/build"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update apilo CLI to the latest version from source",
	Long: `Self-update apilo CLI by rebuilding from the embedded source directory.

This command:
1. Locates the source directory (embedded at build time)
2. Pulls latest changes from git (optional)
3. Rebuilds the binary using the existing Makefile
4. Replaces the current binary with the new one
5. Installs globally if requested

The update process is safe and atomic - if the build fails,
the original binary remains unchanged.`,
	Run: func(cmd *cobra.Command, args []string) {
		buildInfo, err := build.GetBuildInfo()
		if err != nil {
			fmt.Printf("❌ Error getting build information: %v\n", err)
			return
		}

		// Check if self-update is possible
		canUpdate, sourceDir, err := buildInfo.CanSelfUpdate()
		if !canUpdate {
			fmt.Printf("❌ Self-update not available: %v\n", err)
			fmt.Printf("💡 Please manually rebuild or clone the repository\n")
			return
		}

		fmt.Printf("🔄 Updating apilo CLI from source...\n")
		fmt.Printf("📁 Source location: %s\n", sourceDir)

		// Get flags
		pullLatest, _ := cmd.Flags().GetBool("pull")
		installGlobal, _ := cmd.Flags().GetBool("install")
		force, _ := cmd.Flags().GetBool("force")

		// Step 1: Optionally pull latest changes
		if pullLatest {
			fmt.Printf("📥 Pulling latest changes from git...\n")
			if err := runGitPull(sourceDir); err != nil {
				if !force {
					fmt.Printf("❌ Git pull failed: %v\n", err)
					fmt.Printf("💡 Use --force to continue anyway, or --no-pull to skip git update\n")
					return
				}
				fmt.Printf("⚠️  Git pull failed, continuing anyway due to --force: %v\n", err)
			}
		}

		// Step 2: Build new binary
		fmt.Printf("🔨 Building new apilo binary...\n")
		newBinaryPath, err := buildNewBinary(sourceDir)
		if err != nil {
			fmt.Printf("❌ Build failed: %v\n", err)
			return
		}

		// Step 3: Replace current binary
		fmt.Printf("🔄 Replacing current binary...\n")
		if err := replaceBinary(buildInfo.ExecutablePath, newBinaryPath); err != nil {
			fmt.Printf("❌ Failed to replace binary: %v\n", err)
			// Clean up temporary binary
			os.Remove(newBinaryPath)
			return
		}

		// Step 4: Optionally install globally
		if installGlobal {
			fmt.Printf("📦 Installing globally...\n")
			if err := installGlobally(sourceDir); err != nil {
				fmt.Printf("⚠️  Global installation failed: %v\n", err)
				fmt.Printf("📦 Local update completed successfully\n")
			} else {
				fmt.Printf("✅ Global installation completed\n")
			}
		}

		// Step 5: Verify update
		fmt.Printf("🔍 Verifying update...\n")
		if err := verifyUpdate(buildInfo.ExecutablePath); err != nil {
			fmt.Printf("⚠️  Update verification failed: %v\n", err)
		} else {
			fmt.Printf("✅ Apilo CLI updated successfully!\n")
		}

		// Clean up temporary files
		if newBinaryPath != buildInfo.ExecutablePath {
			os.Remove(newBinaryPath)
		}
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)

	// Add flags
	updateCmd.Flags().Bool("pull", true, "Pull latest changes from git before building")
	updateCmd.Flags().Bool("install", false, "Install globally to Go bin directory after update")
	updateCmd.Flags().Bool("force", false, "Force update even if git pull fails")
}

// runGitPull pulls the latest changes from git
func runGitPull(sourceDir string) error {
	cmd := exec.Command("git", "pull")
	cmd.Dir = sourceDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// buildNewBinary builds a new apilo binary
func buildNewBinary(sourceDir string) (string, error) {
	// Create temporary binary name
	tempBinary := filepath.Join(sourceDir, "apilo-update-temp")

	// Use make to build the binary
	cmd := exec.Command("make", "build")
	cmd.Dir = sourceDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("make build failed: %w", err)
	}

	// The binary is built to bin/apilo, copy it to temp location
	builtBinary := filepath.Join(sourceDir, "bin", "apilo")
	if _, err := os.Stat(builtBinary); os.IsNotExist(err) {
		return "", fmt.Errorf("built binary not found at %s", builtBinary)
	}

	// Copy to temp location
	if err := copyBinaryFile(builtBinary, tempBinary); err != nil {
		return "", fmt.Errorf("failed to copy binary: %w", err)
	}

	return tempBinary, nil
}

// replaceBinary atomically replaces the current binary with the new one
func replaceBinary(currentPath, newPath string) error {
	// Make new binary executable
	if err := os.Chmod(newPath, 0755); err != nil {
		return fmt.Errorf("failed to make new binary executable: %w", err)
	}

	// On Unix systems, we can use rename for atomic replacement
	if runtime.GOOS != "windows" {
		return os.Rename(newPath, currentPath)
	}

	// On Windows, we need a different approach
	backupPath := currentPath + ".backup"

	// Create backup
	if err := copyBinaryFile(currentPath, backupPath); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	// Remove original
	if err := os.Remove(currentPath); err != nil {
		return fmt.Errorf("failed to remove original: %w", err)
	}

	// Move new binary into place
	if err := os.Rename(newPath, currentPath); err != nil {
		// Restore backup if move failed
		os.Rename(backupPath, currentPath)
		return fmt.Errorf("failed to move new binary: %w", err)
	}

	// Remove backup
	os.Remove(backupPath)
	return nil
}

// copyBinaryFile copies a file from src to dst
func copyBinaryFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	return os.WriteFile(dst, input, 0755)
}

// installGlobally installs the apilo binary globally
func installGlobally(sourceDir string) error {
	cmd := exec.Command("make", "install")
	cmd.Dir = sourceDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// verifyUpdate verifies that the update was successful
func verifyUpdate(binaryPath string) error {
	// Run version command to verify the binary works
	cmd := exec.Command(binaryPath, "version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("binary verification failed: %w", err)
	}

	fmt.Printf("📋 New version information:\n%s", output)
	return nil
}
