# API Latency Optimizer - Benchmark Setup Guide

**Version**: 2.0
**Last Updated**: 2025-10-02
**Analysis By**: Morchestrator + Analyst Agents

---

## Quick Start

### Option 1: FREE Testing (Recommended for Development)

```bash
# No authentication needed!
apilo benchmark --config config/benchmark_httpbin.yaml

# Or specific test run
apilo benchmark --config config/benchmark_httpbin.yaml --run quick_validation
```

### Option 2: Anthropic API Testing (Requires API Key)

```bash
# 1. Set your API key
export ANTHROPIC_API_KEY="sk-ant-api03-your-key-here"

# 2. Run minimal test (10 requests, ~$0.001 cost)
apilo benchmark --config config/benchmark_anthropic.yaml --run tier1_minimal

# 3. Run full tier1 test (50 requests, ~$0.015 cost)
apilo benchmark --config config/benchmark_anthropic.yaml --run tier1_safe_load
```

---

## Configuration Files

### 1. `benchmark_anthropic.yaml` - Production Anthropic API

**Use Case**: Real API performance testing with Claude
**Cost**: ~$0.001 - $0.060 depending on test profile
**Rate Limits**: Respects Tier 1 (50 RPM) and Tier 2 (1000 RPM) limits

**Test Profiles**:
- `tier1_minimal` - 10 requests, validation only
- `tier1_safe_load` - 50 requests, safe for Tier 1 accounts
- `tier2_moderate_load` - 200 requests (DISABLED by default, requires Tier 2+)

### 2. `benchmark_httpbin.yaml` - Free Test Endpoint

**Use Case**: Configuration validation, development testing, stress tests
**Cost**: FREE
**Rate Limits**: Generous (100-1000 RPM)

**Test Profiles**:
- `quick_validation` - 100 requests, configuration check
- `stress_test` - 1000 requests, high load testing
- `latency_baseline` - 50 requests, single concurrency

### 3. `benchmark_config.yaml` - Original (NOT RECOMMENDED)

**Status**: DEPRECATED - Contains multiple critical issues
**Issues**: Missing auth, invalid endpoints, cost explosion risk
**Use**: Reference only, do not use for production

---

## Authentication Setup

### Anthropic API Key

**Step 1**: Get API Key
1. Visit https://console.anthropic.com
2. Settings → API Keys → Create Key
3. Copy key (starts with `sk-ant-api03-...`)

**Step 2**: Set Environment Variable

**macOS/Linux**:
```bash
export ANTHROPIC_API_KEY="sk-ant-api03-your-key-here"

# Persist in shell config
echo 'export ANTHROPIC_API_KEY="sk-ant-api03-..."' >> ~/.zshrc
source ~/.zshrc
```

**Windows PowerShell**:
```powershell
$env:ANTHROPIC_API_KEY = "sk-ant-api03-your-key-here"

# Persist system-wide
[System.Environment]::SetEnvironmentVariable('ANTHROPIC_API_KEY', 'sk-ant-api03-...', 'User')
```

**Step 3**: Verify
```bash
curl https://api.anthropic.com/v1/messages \
  -H "x-api-key: $ANTHROPIC_API_KEY" \
  -H "anthropic-version: 2023-06-01" \
  -H "content-type: application/json" \
  -d '{
    "model": "claude-3-5-haiku-20250219",
    "max_tokens": 64,
    "messages": [{"role": "user", "content": "Hello"}]
  }'
```

---

## Rate Limits & Tiers

### Anthropic API Tiers

| Tier | Spent | RPM | Input TPM | Output TPM |
|------|-------|-----|-----------|------------|
| 1 | $0 | 50 | 40,000 | 8,000 |
| 2 | $5+ | 1,000 | 80,000 | 16,000 |
| 3 | $40+ | 2,000 | 160,000 | 32,000 |
| 4 | $1,000+ | 4,000 | 400,000 | 80,000 |

**Check Your Tier**:
```bash
curl -I https://api.anthropic.com/v1/messages \
  -H "x-api-key: $ANTHROPIC_API_KEY" \
  | grep ratelimit
```

Look for:
```
anthropic-ratelimit-requests-limit: 50
anthropic-ratelimit-tokens-limit: 40000
```

---

## Cost Estimates

### Claude 3.5 Haiku (Recommended for Benchmarks)

**Pricing**:
- Input: $0.80 per 1M tokens
- Output: $4.00 per 1M tokens

**Per Request Estimate**:
- Input tokens: ~20 (short prompt)
- Output tokens: ~50-256 (response)
- **Cost per request**: ~$0.0003 - $0.0015

**Benchmark Profile Costs**:
- `tier1_minimal` (10 requests): **$0.001**
- `tier1_safe_load` (50 requests × 3 iterations): **$0.015**
- `tier2_moderate_load` (200 requests × 3 iterations): **$0.060**

### Claude 3.5 Sonnet (More Expensive)

**Pricing**:
- Input: $3.00 per 1M tokens
- Output: $15.00 per 1M tokens
- **Cost per request**: ~$0.003 - $0.015 (5-10x more expensive)

---

## Best Practices

### 1. Start Small

```bash
# Always validate configuration first with minimal test
apilo benchmark --config benchmark_anthropic.yaml --run tier1_minimal
```

### 2. Use Free Endpoints for Development

```bash
# Test your benchmark configuration without cost
apilo benchmark --config benchmark_httpbin.yaml --run quick_validation
```

### 3. Monitor Rate Limits

Watch for 429 errors in output. If you see them:
1. Reduce `concurrency`
2. Increase `inter_request_delay`
3. Reduce `total_requests`

### 4. Check Your Tier Before Large Tests

```bash
# Verify your account tier supports the load
curl -I https://api.anthropic.com/v1/messages \
  -H "x-api-key: $ANTHROPIC_API_KEY" \
  | grep ratelimit-requests-limit
```

### 5. Use Haiku for Cost-Sensitive Benchmarks

In `benchmark_anthropic.yaml`, use:
```yaml
model: "claude-3-5-haiku-20250219"  # Not sonnet
```

---

## Troubleshooting

### "401 Unauthorized"
**Cause**: Missing or invalid API key
**Fix**:
```bash
export ANTHROPIC_API_KEY="sk-ant-api03-your-key"
echo $ANTHROPIC_API_KEY  # Verify it's set
```

### "429 Too Many Requests"
**Cause**: Rate limit exceeded
**Fix**:
1. Wait 60 seconds
2. Reduce concurrency in config
3. Check your tier limits

### "All requests failed"
**Cause**: Invalid endpoint or missing request body
**Fix**: Use provided `benchmark_anthropic.yaml` configuration

### "High costs"
**Cause**: Too many requests or using expensive model
**Fix**:
1. Use `tier1_minimal` profile
2. Switch to Haiku model
3. Use HTTPBin for development

---

## Example Commands

### Quick Validation (FREE)
```bash
apilo benchmark --config config/benchmark_httpbin.yaml --run quick_validation
```

### Minimal Anthropic Test ($0.001)
```bash
export ANTHROPIC_API_KEY="sk-ant-..."
apilo benchmark --config config/benchmark_anthropic.yaml --run tier1_minimal
```

### Full Tier 1 Test ($0.015)
```bash
apilo benchmark --config config/benchmark_anthropic.yaml --run tier1_safe_load
```

### Stress Test (FREE)
```bash
apilo benchmark --config config/benchmark_httpbin.yaml --run stress_test
```

---

## Analysis Report Reference

For complete technical analysis and recommendations, see:
- **Report**: Generated by Morchestrator + Analyst agents (2025-10-02)
- **Key Findings**: 7 critical issues identified in original config
- **Resolution**: New configurations created (`benchmark_anthropic.yaml`, `benchmark_httpbin.yaml`)

---

## Support

- **Documentation**: `apilo docs benchmark`
- **Issues**: Check benchmark output logs in `benchmarks/results/`
- **API Status**: https://status.anthropic.com
