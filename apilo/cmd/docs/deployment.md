# Deployment Guide

Production deployment guide.

## Build

```bash
cd apilo
go build -ldflags="-w -s" -o apilo
```

## Install

```bash
go install
```

## Configuration

Create production config:

```bash
apilo config init > /etc/apilo/config.yaml
```

## Monitoring

Start with monitoring:

```bash
apilo monitor <url> --port 8080
```

## Docker Deployment

```dockerfile
FROM golang:1.24 as builder
WORKDIR /app
COPY . .
RUN go build -o apilo

FROM alpine:latest
COPY --from=builder /app/apilo /usr/local/bin/
CMD ["apilo", "monitor", "${API_URL}"]
```

---

**See Also**: `apilo docs configuration`
