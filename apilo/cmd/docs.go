package cmd

import (
	"embed"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

//go:embed docs/*.md
var docsFS embed.FS

var (
	dashboardFlag bool
	portFlag      int
)

var docsCmd = &cobra.Command{
	Use:   "docs [topic]",
	Short: "View documentation",
	Long:  "Browse API Latency Optimizer documentation with beautiful terminal rendering via glow or HTML dashboard",
	Run: func(cmd *cobra.Command, args []string) {
		if dashboardFlag {
			generateAndOpenDashboard()
			return
		}

		if len(args) == 0 {
			showDocsList()
		} else {
			showDoc(args[0])
		}
	},
}

var docsDashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Generate and open HTML documentation dashboard",
	Long:  "Generate a beautiful HTML documentation dashboard and open it in your default browser",
	Run: func(cmd *cobra.Command, args []string) {
		generateAndOpenDashboard()
	},
}

func init() {
	rootCmd.AddCommand(docsCmd)
	docsCmd.AddCommand(docsDashboardCmd)
	docsCmd.Flags().BoolVarP(&dashboardFlag, "dashboard", "d", false, "Generate and open HTML documentation dashboard")
	docsCmd.Flags().IntVarP(&portFlag, "port", "p", 0, "Port for local server (if 0, just opens static file)")
}

func showDocsList() {
	color.Cyan("\n╔═══════════════════════════════════════════════════════════════════╗")
	color.Cyan("║                    Available Documentation                        ║")
	color.Cyan("╚═══════════════════════════════════════════════════════════════════╝\n")

	docs := []struct {
		topic       string
		description string
	}{
		{"quickstart", "Get started in 5 minutes"},
		{"features", "Complete feature overview"},
		{"performance", "Performance metrics and validation"},
		{"configuration", "Configuration reference"},
		{"integration", "Integration guide"},
		{"monitoring", "Monitoring and observability"},
		{"troubleshooting", "Common issues and solutions"},
		{"architecture", "System architecture overview"},
		{"deployment", "Production deployment guide"},
		{"claude-code", "Claude Code integration"},
	}

	fmt.Println(color.YellowString("📚 Documentation Topics:\n"))
	for _, doc := range docs {
		fmt.Printf("   %s - %s\n",
			color.CyanString("apilo docs "+doc.topic),
			doc.description)
	}

	fmt.Println(color.BlueString("\n💡 Tip: Use 'apilo docs <topic>' to view a specific document"))
	fmt.Println(color.BlueString("    Example: apilo docs quickstart\n"))
}

func showDoc(topic string) {
	// Map topics to embedded files
	docFiles := map[string]string{
		"quickstart":      "docs/quickstart.md",
		"features":        "docs/features.md",
		"performance":     "docs/performance.md",
		"configuration":   "docs/configuration.md",
		"integration":     "docs/integration.md",
		"monitoring":      "docs/monitoring.md",
		"troubleshooting": "docs/troubleshooting.md",
		"architecture":    "docs/architecture.md",
		"deployment":      "docs/deployment.md",
		"claude-code":     "docs/claude-code.md",
	}

	docFile, exists := docFiles[topic]
	if !exists {
		color.Red("❌ Documentation topic '%s' not found\n", topic)
		fmt.Println("\nAvailable topics:")
		showDocsList()
		return
	}

	// Read embedded documentation
	content, err := docsFS.ReadFile(docFile)
	if err != nil {
		color.Red("❌ Error reading documentation: %v\n", err)
		return
	}

	// Check if glow is available
	glowPath, err := exec.LookPath("glow")
	if err != nil {
		// Fallback to plain text display
		fmt.Println(color.YellowString("💡 Install 'glow' for beautiful markdown rendering: brew install glow\n"))
		fmt.Println(string(content))
		return
	}

	// Create temporary file for glow
	tmpFile, err := os.CreateTemp("", "apilo-docs-*.md")
	if err != nil {
		fmt.Println(string(content))
		return
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write(content); err != nil {
		fmt.Println(string(content))
		return
	}
	tmpFile.Close()

	// Display with glow
	cmd := exec.Command(glowPath, tmpFile.Name())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		// Fallback to plain text
		fmt.Println(string(content))
	}
}

// Helper function to check if glow is installed
func isGlowInstalled() bool {
	_, err := exec.LookPath("glow")
	return err == nil
}

// Helper to get all available topics
func getAvailableTopics() []string {
	return []string{
		"quickstart",
		"features",
		"performance",
		"configuration",
		"integration",
		"monitoring",
		"troubleshooting",
		"architecture",
		"deployment",
		"claude-code",
	}
}

// Helper to suggest similar topics
func suggestSimilarTopics(input string) []string {
	topics := getAvailableTopics()
	var suggestions []string

	for _, topic := range topics {
		if strings.Contains(topic, input) || strings.Contains(input, topic) {
			suggestions = append(suggestions, topic)
		}
	}

	return suggestions
}

// Generate and open HTML documentation dashboard
func generateAndOpenDashboard() {
	color.Cyan("\n📊 Generating documentation dashboard...\n")

	// Read all documentation files
	docs := []struct {
		id          string
		title       string
		file        string
		description string
	}{
		{"quickstart", "Quick Start", "docs/quickstart.md", "Get started in 5 minutes"},
		{"features", "Features", "docs/features.md", "Complete feature overview"},
		{"performance", "Performance", "docs/performance.md", "Performance metrics and validation"},
		{"configuration", "Configuration", "docs/configuration.md", "Configuration reference"},
		{"integration", "Integration", "docs/integration.md", "Integration guide"},
		{"monitoring", "Monitoring", "docs/monitoring.md", "Monitoring and observability"},
		{"troubleshooting", "Troubleshooting", "docs/troubleshooting.md", "Common issues and solutions"},
		{"architecture", "Architecture", "docs/architecture.md", "System architecture overview"},
		{"deployment", "Deployment", "docs/deployment.md", "Production deployment guide"},
		{"claude-code", "Claude Code", "docs/claude-code.md", "Claude Code integration"},
	}

	html := generateDashboardHTML(docs)

	// Save to temporary file
	tmpFile, err := os.CreateTemp("", "apilo-dashboard-*.html")
	if err != nil {
		color.Red("❌ Error creating temporary file: %v\n", err)
		return
	}

	if _, err := tmpFile.WriteString(html); err != nil {
		color.Red("❌ Error writing HTML file: %v\n", err)
		return
	}
	tmpFile.Close()

	color.Green("✅ Documentation dashboard generated: %s\n", tmpFile.Name())
	color.Cyan("🌐 Opening in browser...\n")

	// Open in browser
	var cmd *exec.Cmd
	switch {
	case fileExists("/usr/bin/open"): // macOS
		cmd = exec.Command("open", tmpFile.Name())
	case fileExists("/usr/bin/xdg-open"): // Linux
		cmd = exec.Command("xdg-open", tmpFile.Name())
	default: // Windows or fallback
		cmd = exec.Command("cmd", "/c", "start", tmpFile.Name())
	}

	if err := cmd.Start(); err != nil {
		color.Yellow("⚠️  Could not open browser automatically\n")
		color.Cyan("📄 Please open this file manually: %s\n", tmpFile.Name())
	} else {
		color.Green("✅ Dashboard opened in your default browser\n")
		color.Blue("\n💡 Tip: The HTML file will remain accessible until you restart your system\n")
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func generateDashboardHTML(docs []struct {
	id          string
	title       string
	file        string
	description string
}) string {
	// Read all markdown contents
	docsContent := make(map[string]string)
	for _, doc := range docs {
		content, err := docsFS.ReadFile(doc.file)
		if err != nil {
			docsContent[doc.id] = fmt.Sprintf("Error loading documentation: %v", err)
		} else {
			docsContent[doc.id] = string(content)
		}
	}

	return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>API Latency Optimizer Documentation</title>
    <script src="https://cdn.jsdelivr.net/npm/marked/marked.min.js"></script>
    <style>
        *, *::before, *::after { margin: 0; padding: 0; box-sizing: border-box; }

        body {
            font: 16px/1.6 -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
            color: #111827;
            background: #fff;
            -webkit-font-smoothing: antialiased;
        }

        .container {
            display: flex;
            min-height: 100vh;
            max-width: 1400px;
            margin: 0 auto;
        }

        .sidebar {
            width: 240px;
            border-right: 1px solid #E5E7EB;
            padding: 32px 20px;
            position: sticky;
            top: 0;
            height: 100vh;
            overflow-y: auto;
        }

        .sidebar h1 {
            font-size: 14px;
            font-weight: 700;
            color: #111827;
            margin-bottom: 4px;
            letter-spacing: -0.01em;
        }

        .sidebar .subtitle {
            font-size: 12px;
            color: #6B7280;
            margin-bottom: 24px;
        }

        .performance-badge {
            display: inline-block;
            background: #10B981;
            color: #fff;
            padding: 2px 8px;
            border-radius: 3px;
            font-size: 11px;
            font-weight: 600;
            margin-top: 4px;
        }

        nav { margin-top: 32px; }

        .nav-item {
            display: block;
            padding: 8px 12px;
            margin: 2px 0;
            border-radius: 6px;
            cursor: pointer;
            transition: all 120ms ease;
            text-decoration: none;
            color: #6B7280;
        }

        .nav-item:hover {
            background: #F9FAFB;
            color: #111827;
        }

        .nav-item.active {
            background: #111827;
            color: #fff;
        }

        .nav-item .title {
            font-size: 13px;
            font-weight: 500;
            margin-bottom: 2px;
        }

        .nav-item .description {
            font-size: 11px;
            opacity: 0.7;
        }

        .content {
            flex: 1;
            padding: 64px 48px;
            overflow-y: auto;
        }

        article {
            max-width: 680px;
            margin: 0 auto;
        }

        article h1 {
            font-size: 32px;
            font-weight: 700;
            line-height: 1.25;
            color: #111827;
            margin-bottom: 8px;
            letter-spacing: -0.02em;
        }

        article h2 {
            font-size: 24px;
            font-weight: 600;
            line-height: 1.3;
            color: #111827;
            margin: 48px 0 16px;
            letter-spacing: -0.01em;
        }

        article h3 {
            font-size: 18px;
            font-weight: 600;
            line-height: 1.4;
            color: #374151;
            margin: 32px 0 12px;
        }

        article p {
            margin-bottom: 16px;
            color: #374151;
        }

        article a {
            color: #2563EB;
            text-decoration: none;
        }

        article a:hover {
            text-decoration: underline;
        }

        article ul, article ol {
            margin: 16px 0;
            padding-left: 24px;
        }

        article li {
            margin-bottom: 8px;
            color: #374151;
        }

        article code {
            background: #F3F4F6;
            color: #111827;
            padding: 2px 6px;
            border-radius: 3px;
            font-size: 14px;
            font-family: 'SF Mono', Consolas, monospace;
        }

        article pre {
            background: #1F2937;
            color: #F3F4F6;
            padding: 16px 20px;
            border-radius: 6px;
            overflow-x: auto;
            margin: 24px 0;
            line-height: 1.6;
        }

        article pre code {
            background: transparent;
            color: inherit;
            padding: 0;
            font-size: 13px;
        }

        article table {
            width: 100%;
            border-collapse: collapse;
            margin: 24px 0;
            font-size: 14px;
        }

        article th, article td {
            padding: 10px 12px;
            text-align: left;
            border: 1px solid #E5E7EB;
        }

        article th {
            background: #F9FAFB;
            font-weight: 600;
            color: #111827;
        }

        article blockquote {
            border-left: 3px solid #E5E7EB;
            padding-left: 16px;
            margin: 24px 0;
            color: #6B7280;
            font-style: italic;
        }

        article hr {
            border: none;
            border-top: 1px solid #E5E7EB;
            margin: 32px 0;
        }

        .loading {
            text-align: center;
            padding: 64px;
            color: #9CA3AF;
        }

        .sidebar::-webkit-scrollbar, .content::-webkit-scrollbar {
            width: 6px;
        }

        .sidebar::-webkit-scrollbar-thumb, .content::-webkit-scrollbar-thumb {
            background: #D1D5DB;
            border-radius: 3px;
        }

        @media (max-width: 768px) {
            .container { flex-direction: column; }
            .sidebar { width: 100%; height: auto; position: relative; border-right: none; border-bottom: 1px solid #E5E7EB; }
            .content { padding: 32px 20px; }
            article h1 { font-size: 28px; }
            article h2 { font-size: 20px; }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="sidebar">
            <h1>API Latency Optimizer</h1>
            <div class="subtitle">Documentation</div>
            <span class="performance-badge">93.69% Faster</span>
            <nav id="navigation"></nav>
        </div>
        <div class="content">
            <article id="content-area">
                <div class="loading">Select a topic</div>
            </article>
        </div>
    </div>

    <script>
        const docs = ` + generateDocsJSON(docs, docsContent) + `;

        let currentDoc = 'quickstart';

        function renderNavigation() {
            const nav = document.getElementById('navigation');
            nav.innerHTML = docs.map(doc => ` + "`" + `
                <div class="nav-item ${doc.id === currentDoc ? 'active' : ''}"
                     onclick="loadDoc('${doc.id}')">
                    <div class="title">${doc.title}</div>
                    <div class="description">${doc.description}</div>
                </div>
            ` + "`" + `).join('');
        }

        function loadDoc(id) {
            currentDoc = id;
            const doc = docs.find(d => d.id === id);
            if (!doc) return;

            const contentArea = document.getElementById('content-area');
            contentArea.innerHTML = marked.parse(doc.content);
            renderNavigation();

            // Scroll content to top
            document.querySelector('.content').scrollTop = 0;
        }

        // Initial load
        renderNavigation();
        loadDoc('quickstart');
    </script>
</body>
</html>`
}

func generateDocsJSON(docs []struct {
	id          string
	title       string
	file        string
	description string
}, docsContent map[string]string) string {
	var items []string
	for _, doc := range docs {
		content := docsContent[doc.id]
		// Escape content for JSON
		content = strings.ReplaceAll(content, "\\", "\\\\")
		content = strings.ReplaceAll(content, "`", "\\`")
		content = strings.ReplaceAll(content, "$", "\\$")

		items = append(items, fmt.Sprintf(`{
			"id": "%s",
			"title": "%s",
			"description": "%s",
			"content": `+"`%s`"+`
		}`, doc.id, doc.title, doc.description, content))
	}

	return "[" + strings.Join(items, ",") + "]"
}
