# Troubleshooting Guide

Common issues and solutions.

## High Memory Usage

**Issue**: Memory usage exceeds configured limits

**Solutions**:
1. Reduce `max_memory_mb` in config
2. Lower `default_ttl` 
3. Increase `gc_threshold_percent`

## Low Cache Hit Ratio

**Issue**: Cache hit ratio below 90%

**Solutions**:
1. Increase cache memory
2. Extend TTL values
3. Review invalidation patterns

## Circuit Breaker Tripping

**Issue**: Circuit breaker constantly opening

**Solutions**:
1. Increase `failure_threshold`
2. Check backend health
3. Review timeout settings

---

**Get Help**: GitHub Issues or `apilo docs`
