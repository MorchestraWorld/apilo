package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install apilo globally",
	Long: `Install apilo CLI globally for system-wide access.

This command installs apilo to your system's binary directory,
making it available from any location in your terminal.

Installation locations by platform:
  • macOS/Linux: /usr/local/bin/apilo (or $GOPATH/bin if set)
  • Windows:     C:\Program Files\apilo\apilo.exe

The installer will:
1. Detect your system architecture and OS
2. Build or locate the apilo binary
3. Install to the appropriate system location
4. Verify the installation
5. Optionally add to PATH if needed`,
	Run: func(cmd *cobra.Command, args []string) {
		global, _ := cmd.Flags().GetBool("global")
		symlink, _ := cmd.Flags().GetBool("symlink")
		force, _ := cmd.Flags().GetBool("force")

		performInstallation(global, symlink, force)
	},
}

func init() {
	rootCmd.AddCommand(installCmd)

	// Installation flags
	installCmd.Flags().Bool("global", true, "Install globally for all users")
	installCmd.Flags().Bool("symlink", false, "Create symlink instead of copying binary")
	installCmd.Flags().Bool("force", false, "Force installation even if binary exists")
}

func performInstallation(global, symlink, force bool) {
	color.Cyan("\n╔═══════════════════════════════════════════════════════════════════╗")
	color.Cyan("║                  Apilo Global Installation                        ║")
	color.Cyan("╚═══════════════════════════════════════════════════════════════════╝\n")

	// Step 1: Detect system information
	fmt.Println(color.YellowString("🔍 Detecting system configuration...\n"))

	sysInfo := detectSystemInfo()
	displaySystemInfo(sysInfo)

	// Step 2: Locate current binary
	fmt.Println(color.YellowString("\n📦 Locating apilo binary...\n"))

	currentBinary, err := os.Executable()
	if err != nil {
		color.Red("❌ Failed to locate current binary: %v\n", err)
		return
	}

	fmt.Printf("   Current binary: %s\n", color.CyanString(currentBinary))

	// Step 3: Determine installation target
	installPath := determineInstallPath(global, sysInfo.OS)
	fmt.Printf("   Target location: %s\n", color.CyanString(installPath))

	// Step 4: Check if already installed
	if _, err := os.Stat(installPath); err == nil && !force {
		color.Yellow("\n⚠️  Apilo is already installed at: %s\n", installPath)
		fmt.Println(color.BlueString("💡 Use --force to reinstall\n"))
		return
	}

	// Step 5: Ensure target directory exists
	installDir := filepath.Dir(installPath)
	if err := os.MkdirAll(installDir, 0755); err != nil {
		color.Red("❌ Failed to create installation directory: %v\n", err)
		return
	}

	// Step 6: Perform installation
	fmt.Println(color.YellowString("\n🚀 Installing apilo...\n"))

	if symlink {
		if err := createSymlink(currentBinary, installPath); err != nil {
			color.Red("❌ Failed to create symlink: %v\n", err)
			return
		}
		color.Green("✅ Symlink created successfully\n")
	} else {
		if err := copyBinary(currentBinary, installPath); err != nil {
			color.Red("❌ Failed to copy binary: %v\n", err)
			return
		}
		color.Green("✅ Binary installed successfully\n")
	}

	// Step 7: Verify installation
	fmt.Println(color.YellowString("\n🔍 Verifying installation...\n"))

	if err := verifyInstallation(installPath); err != nil {
		color.Red("❌ Installation verification failed: %v\n", err)
		return
	}

	// Step 8: Check PATH
	fmt.Println(color.YellowString("\n🛣️  Checking PATH configuration...\n"))

	checkPathConfiguration(installDir)

	// Installation complete
	color.Green("\n✅ Apilo installed successfully!\n")

	// Display usage instructions
	displayUsageInstructions(installPath)
}

type SystemInfo struct {
	OS   string
	Arch string
	Home string
	User string
}

func detectSystemInfo() SystemInfo {
	homeDir, _ := os.UserHomeDir()
	currentUser := os.Getenv("USER")
	if currentUser == "" {
		currentUser = os.Getenv("USERNAME")
	}

	return SystemInfo{
		OS:   runtime.GOOS,
		Arch: runtime.GOARCH,
		Home: homeDir,
		User: currentUser,
	}
}

func displaySystemInfo(info SystemInfo) {
	fmt.Printf("   OS:           %s\n", color.CyanString(info.OS))
	fmt.Printf("   Architecture: %s\n", color.CyanString(info.Arch))
	fmt.Printf("   User:         %s\n", color.CyanString(info.User))
}

func determineInstallPath(global bool, osType string) string {
	// Check if GOPATH/bin exists
	gopath := os.Getenv("GOPATH")
	if gopath != "" {
		gopathBin := filepath.Join(gopath, "bin", "apilo")
		if global {
			return gopathBin
		}
	}

	// Platform-specific defaults
	switch osType {
	case "darwin", "linux":
		if global {
			return "/usr/local/bin/apilo"
		}
		homeDir, _ := os.UserHomeDir()
		return filepath.Join(homeDir, ".local", "bin", "apilo")
	case "windows":
		if global {
			return filepath.Join("C:", "Program Files", "apilo", "apilo.exe")
		}
		homeDir, _ := os.UserHomeDir()
		return filepath.Join(homeDir, "AppData", "Local", "apilo", "apilo.exe")
	default:
		return "/usr/local/bin/apilo"
	}
}

func createSymlink(source, target string) error {
	// Remove existing symlink if present
	if _, err := os.Lstat(target); err == nil {
		os.Remove(target)
	}

	return os.Symlink(source, target)
}

func copyBinary(source, target string) error {
	// Read source file
	input, err := os.ReadFile(source)
	if err != nil {
		return fmt.Errorf("failed to read source binary: %w", err)
	}

	// Write to target with executable permissions
	if err := os.WriteFile(target, input, 0755); err != nil {
		return fmt.Errorf("failed to write target binary: %w", err)
	}

	return nil
}

func verifyInstallation(installPath string) error {
	// Check if file exists
	if _, err := os.Stat(installPath); os.IsNotExist(err) {
		return fmt.Errorf("installation file not found: %s", installPath)
	}

	// Try to execute version command
	cmd := exec.Command(installPath, "version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to execute installed binary: %w", err)
	}

	fmt.Printf("   %s Binary is executable\n", color.GreenString("✅"))
	fmt.Printf("   %s Version check passed\n", color.GreenString("✅"))
	fmt.Printf("\n   Version info:\n   %s", string(output))

	return nil
}

func checkPathConfiguration(installDir string) {
	pathEnv := os.Getenv("PATH")

	// Check if install directory is in PATH
	if contains(filepath.SplitList(pathEnv), installDir) {
		fmt.Printf("   %s Installation directory is in PATH\n", color.GreenString("✅"))
		fmt.Printf("   %s You can run 'apilo' from anywhere\n", color.GreenString("✅"))
	} else {
		color.Yellow("   ⚠️  Installation directory is NOT in PATH\n")
		fmt.Println(color.BlueString("\n   To add to PATH, run one of these commands:"))

		shellConfig := detectShellConfig()
		fmt.Printf("\n   %s\n", color.CyanString(fmt.Sprintf("echo 'export PATH=\"%s:$PATH\"' >> %s", installDir, shellConfig)))
		fmt.Printf("   %s\n", color.CyanString("source "+shellConfig))
	}
}

func detectShellConfig() string {
	shell := os.Getenv("SHELL")
	homeDir, _ := os.UserHomeDir()

	switch {
	case contains([]string{shell}, "zsh"):
		return filepath.Join(homeDir, ".zshrc")
	case contains([]string{shell}, "bash"):
		return filepath.Join(homeDir, ".bashrc")
	case contains([]string{shell}, "fish"):
		return filepath.Join(homeDir, ".config", "fish", "config.fish")
	default:
		return filepath.Join(homeDir, ".profile")
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func displayUsageInstructions(installPath string) {
	fmt.Println(color.YellowString("\n📝 Next Steps:\n"))

	fmt.Println("   1. " + color.CyanString("Verify installation:"))
	fmt.Println("      " + color.BlueString("apilo version"))

	fmt.Println("\n   2. " + color.CyanString("Run your first benchmark:"))
	fmt.Println("      " + color.BlueString("apilo benchmark https://api.example.com"))

	fmt.Println("\n   3. " + color.CyanString("Start monitoring:"))
	fmt.Println("      " + color.BlueString("apilo monitor https://api.example.com"))

	fmt.Println("\n   4. " + color.CyanString("View documentation:"))
	fmt.Println("      " + color.BlueString("apilo docs"))

	fmt.Println(color.YellowString("\n🎯 Advanced Features:\n"))
	fmt.Println("   • " + color.CyanString("apilo performance") + " - View validated metrics")
	fmt.Println("   • " + color.CyanString("apilo config init") + " - Create custom configuration")
	fmt.Println("   • " + color.CyanString("apilo claude setup") + " - Enable AI integration")

	fmt.Println(color.GreenString("\n🚀 Happy optimizing!\n"))
}
