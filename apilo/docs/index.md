# Apilo Documentation Portal

## API Latency Optimizer - Production-Ready Performance Tool

Welcome to the comprehensive documentation for Apilo, the API Latency Optimizer that achieves 93.69% latency reduction through intelligent caching, HTTP/2 optimization, and background daemon services.

---

## Quick Navigation

### 🚀 Getting Started
- [Installation Guide](INSTALLATION.md) - Complete setup instructions
- [Quick Start](quickstart.md) - Get running in 5 minutes
- [Configuration](configuration.md) - Customize your setup

### 🔧 Core Features
- [Daemon Service](DAEMON.md) - Background optimization service
- [Claude Code Integration](HOOK_GUIDE.md) - Automatic API optimization
- [Performance](performance.md) - Benchmarks and metrics
- [Cache System](cache.md) - Memory-bounded caching

### 📖 User Guides
- [CLI Commands](cli-reference.md) - Complete command reference
- [Makefile Targets](makefile-guide.md) - Build and deployment
- [Troubleshooting](troubleshooting.md) - Common issues and solutions
- [Best Practices](best-practices.md) - Optimization tips

### 🛠️ Development
- [Architecture](architecture.md) - System design and components
- [API Reference](api-reference.md) - Internal APIs
- [Contributing](../CONTRIBUTING.md) - How to contribute
- [Testing](testing.md) - Test suite and coverage

### 🔒 Advanced Topics
- [Security](security.md) - Security considerations
- [Performance Tuning](tuning.md) - Advanced optimization
- [Deployment](deployment.md) - Production deployment
- [Monitoring](monitoring.md) - Observability and metrics

---

## System Overview

```
┌─────────────────────────────────────────────────────────────┐
│                    APILO ARCHITECTURE                        │
└─────────────────────────────────────────────────────────────┘

User Request → Claude Code
                    ↓
            Hook Intercepts
                    ↓
         Daemon (Port 9876) ←→ Cache
                    ↓              ↓
            Optimization      Hit: <10ms
                    ↓         Miss: Fetch
            HTTP/2 Pool           ↓
                    ↓         Store in Cache
            Response ←────────────┘
```

## Key Features

### 🚀 Performance
- **93.69% Latency Reduction** - 515ms → 33ms
- **15.8x Throughput Increase** - 2.1 → 33.5 RPS
- **98% Cache Hit Ratio** - Intelligent caching
- **HTTP/2 Optimization** - Connection pooling and multiplexing

### 🔧 Daemon Service
- **Background Processing** - No user intervention
- **Resource Efficient** - <50MB RAM, <2% CPU
- **IPC Server** - HTTP API on localhost:9876
- **Auto-Recovery** - Graceful degradation

### 🪝 Claude Code Integration
- **Automatic Detection** - Pattern-based API call identification
- **Transparent Optimization** - Zero configuration
- **Hook System** - Installed via `~/.claude/hooks/`
- **Real-time Metrics** - Performance tracking

### 💾 Intelligent Caching
- **Memory-Bounded** - Configurable limits (default 500MB)
- **LRU Eviction** - Automatic memory management
- **TTL Support** - Time-based expiration (default 10min)
- **Cache Invalidation** - Manual and pattern-based

---

## Installation

### Quick Install
```bash
cd apilo
make all
```

This installs:
- ✅ Apilo CLI globally (`~/go/bin/apilo`)
- ✅ Claude Code hooks (`~/.claude/hooks/apilo-optimizer.sh`)
- ✅ Configuration directory (`~/.apilo/`)
- ✅ Daemon ready to start

### Start Using
```bash
# Start daemon
make daemon-start

# Check status
apilo daemon status

# Use Claude Code normally - API calls auto-optimized!
```

---

## Documentation Sections

### For Users
- **[Quick Start](quickstart.md)** - Get up and running
- **[CLI Reference](cli-reference.md)** - All commands explained
- **[Configuration](configuration.md)** - Customize behavior
- **[Troubleshooting](troubleshooting.md)** - Fix common issues

### For Developers
- **[Architecture](architecture.md)** - System design
- **[API Reference](api-reference.md)** - Internal APIs
- **[Contributing](../CONTRIBUTING.md)** - Development guide
- **[Testing](testing.md)** - Test procedures

### For Operators
- **[Deployment](deployment.md)** - Production setup
- **[Monitoring](monitoring.md)** - Metrics and alerts
- **[Security](security.md)** - Security best practices
- **[Tuning](tuning.md)** - Performance optimization

---

## Support & Community

- **GitHub Issues**: Report bugs and feature requests
- **Documentation**: This comprehensive guide
- **Examples**: See `examples/` directory
- **Changelog**: Track updates in `CHANGELOG.md`

---

## Version Information

**Current Version**: 2.0.0
**Status**: Production Ready ✅
**Last Updated**: 2025-10-03

---

## License

See [LICENSE](../LICENSE) file for details.

---

## Quick Links

| Resource | Location |
|----------|----------|
| Binary Installation | `~/go/bin/apilo` |
| Hook Script | `~/.claude/hooks/apilo-optimizer.sh` |
| Configuration | `~/.apilo/config/daemon.yaml` |
| PID File | `~/.apilo/daemon.pid` |
| Logs | `~/.apilo/logs/daemon.log` |
| Source Code | `apilo/` directory |

---

**Ready to optimize?** Start with the [Quick Start Guide](quickstart.md) or jump to [Installation](INSTALLATION.md).
