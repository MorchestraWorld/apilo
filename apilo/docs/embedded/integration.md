# Integration Guide

Step-by-step guide to integrate API Latency Optimizer into your application.

## Go Integration

### Basic Integration

```go
package main

import (
    "github.com/yourorg/api-latency-optimizer/src"
)

func main() {
    config := src.DefaultIntegratedConfig()
    optimizer, err := src.NewIntegratedOptimizer(config)
    if err != nil {
        panic(err)
    }
    
    optimizer.Start()
    defer optimizer.Stop()
    
    client := optimizer.GetClient()
    resp, err := client.Get("https://api.example.com")
    // Handle response...
}
```

### Advanced Configuration

See `apilo docs configuration` for complete configuration options.

## Other Languages

Integration examples coming soon for:
- Python
- Node.js
- Ruby
- Java

---

**See Also**: `apilo docs quickstart`
