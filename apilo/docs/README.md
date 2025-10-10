# Apilo Documentation System

## Overview

This directory contains comprehensive documentation for the Apilo API Latency Optimizer, including both markdown sources and generated static HTML documentation.

## Structure

```
docs/
├── README.md                    # This file
├── index.md                     # Main documentation index
├── html/                        # Generated static HTML site
│   ├── index.html              # Main documentation page
│   ├── css/
│   │   └── documentation.css   # Professional styling
│   ├── js/
│   │   └── documentation.js    # Interactive features
│   └── images/                 # Documentation images
├── embedded/                    # Embedded docs (in binary)
│   ├── architecture.md
│   ├── quickstart.md
│   ├── performance.md
│   └── ...
└── ...                         # Additional markdown files
```

## Available Documentation

### Root Level (apilo/)
- **README.md** - Project overview
- **INSTALLATION.md** - Complete installation guide
- **DAEMON.md** - Daemon service documentation
- **HOOK_GUIDE.md** - Claude Code hook guide
- **BENCHMARK_RESULTS.md** - Performance benchmarks
- **CLAUDE_INSTRUMENTATION.md** - Token tracking and metrics

### Embedded Documentation (cmd/docs/)
- **quickstart.md** - 5-minute getting started guide
- **architecture.md** - System architecture and design
- **configuration.md** - Configuration options
- **performance.md** - Performance metrics and tuning
- **monitoring.md** - Observability and metrics
- **deployment.md** - Production deployment guide
- **troubleshooting.md** - Common issues and solutions
- **integration.md** - Integration patterns
- **features.md** - Feature documentation
- **claude-code.md** - Claude Code integration

## Viewing Documentation

### Option 1: Static HTML Site (Recommended)
```bash
# Open in browser
open docs/html/index.html

# Or serve with Python
cd docs/html
python3 -m http.server 8000
# Visit: http://localhost:8000
```

### Option 2: Markdown Files
```bash
# View with any markdown reader
cat docs/index.md
cat DAEMON.md
cat HOOK_GUIDE.md

# Or use a markdown viewer
glow docs/index.md
```

### Option 3: CLI Embedded Docs
```bash
# Access via apilo CLI
apilo docs
apilo docs quickstart
apilo docs architecture
```

## Features of HTML Documentation

### Professional Design
- **Dark theme** optimized for readability
- **Responsive layout** works on all devices
- **Professional styling** consistent with apilo branding
- **Syntax highlighting** for code blocks

### Interactive Features
- **Live search** filters navigation
- **Smooth scrolling** to sections
- **Active section tracking** in sidebar
- **Copy code buttons** for easy copying

### Performance
- **Static HTML** - no backend required
- **Fast loading** - optimized assets
- **Offline capable** - no external dependencies (except CDN for highlight.js)

## Building Documentation

### Manual Updates
```bash
# Edit markdown files
vim docs/index.md
vim DAEMON.md

# HTML is static - no rebuild needed
# Just refresh browser
```

### Adding New Pages
1. Create markdown file in `docs/`
2. Add link to `docs/index.md`
3. Update `html/index.html` navigation if needed

### Customizing HTML
```bash
# Edit styles
vim docs/html/css/documentation.css

# Edit interactivity
vim docs/html/js/documentation.js

# Edit content
vim docs/html/index.html
```

## Documentation Standards

### Writing Guidelines
- **Clear and concise** - Get to the point quickly
- **Code examples** - Show, don't just tell
- **Step-by-step** - Break down complex tasks
- **Troubleshooting** - Anticipate common issues

### Markdown Formatting
```markdown
# Main Heading (H1)
## Section Heading (H2)
### Subsection (H3)

**Bold** for emphasis
`code` for inline code

```bash
# Code blocks with language
apilo daemon start
```

- Bullet lists
- For related items

1. Numbered lists
2. For sequential steps
```

### Code Examples
- Include complete, runnable examples
- Show expected output
- Explain what each part does
- Provide context for when to use

## Documentation Maintenance

### Regular Updates
- Review quarterly for accuracy
- Update version numbers
- Add new features as released
- Archive outdated content

### Version Control
All documentation is version controlled with Git:
```bash
git status docs/
git diff docs/
git commit -m "docs: update installation guide"
```

### Testing Documentation
- Test all code examples
- Verify all links work
- Check on mobile devices
- Validate HTML/CSS

## Contributing to Documentation

### Process
1. Identify gap or error
2. Create/edit markdown file
3. Test locally
4. Submit pull request
5. Review and merge

### Guidelines
- Follow existing style
- Keep it simple
- Add examples
- Test instructions
- Update index/navigation

## Documentation Metrics

### Coverage
- ✅ Installation - Complete
- ✅ Quick Start - Complete
- ✅ Daemon Service - Complete
- ✅ Hooks - Complete
- ✅ CLI Reference - In Progress
- ✅ API Reference - Planned
- ✅ Architecture - Complete
- ✅ Performance - Complete

### Quality Checklist
- [ ] All features documented
- [x] Code examples tested
- [x] Screenshots current
- [x] Links validated
- [x] Mobile responsive
- [x] Search functional
- [ ] Accessibility compliant
- [x] Load time < 1s

## Troubleshooting Documentation

### HTML Not Rendering Correctly
```bash
# Check file permissions
ls -la docs/html/

# Verify all assets present
ls docs/html/css/
ls docs/html/js/

# Test in different browser
open -a "Firefox" docs/html/index.html
```

### Search Not Working
```bash
# Check JavaScript loaded
# Open browser console (F12)
# Look for errors

# Verify file paths in HTML
grep -r "documentation.js" docs/html/
```

### Broken Links
```bash
# Find all markdown links
grep -r "\[.*\](.*)" docs/

# Test each link manually
```

## Future Enhancements

### Planned Features
- [ ] PDF export functionality
- [ ] Dark/light theme toggle
- [ ] Advanced search with fuzzy matching
- [ ] Versioned documentation
- [ ] API documentation generator
- [ ] Automated screenshots
- [ ] Interactive tutorials
- [ ] Video walkthroughs

### Enhancement Ideas
- AI-powered search
- Community contributions
- Multi-language support
- Progressive web app
- Offline mode
- Print-optimized CSS

## Support

For documentation issues:
- GitHub Issues: Report errors or gaps
- Pull Requests: Contribute improvements
- Discussions: Ask questions

## License

Documentation licensed same as apilo project.

---

**Last Updated**: 2025-10-03
**Version**: 2.0.0
**Status**: Production Ready ✅
