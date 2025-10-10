#!/bin/bash
#
# Apilo Claude Code Hook - Automatic API Optimization
#
# This hook integrates with Claude Code to automatically optimize API calls
# through the apilo daemon running in the background.
#
# Installation:
#   cp apilo-optimizer.sh ~/.claude/hooks/
#   chmod +x ~/.claude/hooks/apilo-optimizer.sh
#
# The hook will automatically optimize API requests when:
# - Apilo daemon is running (apilo daemon start)
# - API calls are detected in Claude Code queries
# - Cache can be leveraged for performance improvement

# Configuration
DAEMON_PORT=${APILO_DAEMON_PORT:-9876}
DAEMON_HOST="localhost"
TIMEOUT=5

# Check if daemon is running
check_daemon() {
    curl -s --max-time 1 "http://${DAEMON_HOST}:${DAEMON_PORT}/health" > /dev/null 2>&1
    return $?
}

# Optimize API request through daemon
optimize_request() {
    local url="$1"
    local method="${2:-GET}"

    # Prepare JSON request
    local json_payload=$(cat <<EOF
{
    "url": "$url",
    "method": "$method",
    "timeout": "${TIMEOUT}s"
}
EOF
)

    # Send to daemon for optimization
    response=$(curl -s --max-time "$TIMEOUT" \
        -X POST \
        -H "Content-Type: application/json" \
        -d "$json_payload" \
        "http://${DAEMON_HOST}:${DAEMON_PORT}/optimize")

    echo "$response"
}

# Detect API calls in query
detect_api_call() {
    local query="$1"

    # Simple pattern matching for common API URLs
    if echo "$query" | grep -qE "https?://[^/]+/(api|v[0-9]+)/"; then
        return 0
    fi

    # Check for api.* domains
    if echo "$query" | grep -qE "https?://api\.[^/]+/"; then
        return 0
    fi

    return 1
}

# Extract URL from query
extract_url() {
    local query="$1"
    echo "$query" | grep -oE "https?://[^ \"\']+" | head -1
}

# Main hook logic
main() {
    # Get query from environment or stdin
    local query="${CLAUDE_QUERY:-$(cat)}"

    # Check if daemon is running
    if ! check_daemon; then
        # Daemon not running - pass through without optimization
        echo "$query"
        exit 0
    fi

    # Detect if query contains API call
    if detect_api_call "$query"; then
        url=$(extract_url "$query")

        if [ -n "$url" ]; then
            # Log optimization attempt
            echo "[apilo] Optimizing API call: $url" >&2

            # Optimize through daemon
            result=$(optimize_request "$url")

            # Check if optimization succeeded
            if [ -n "$result" ] && echo "$result" | jq -e '.optimized' > /dev/null 2>&1; then
                cache_hit=$(echo "$result" | jq -r '.cache_hit')
                latency=$(echo "$result" | jq -r '.latency')

                if [ "$cache_hit" = "true" ]; then
                    echo "[apilo] ✅ Cache hit - optimized response (${latency})" >&2
                else
                    echo "[apilo] 🔄 Cached for future requests (${latency})" >&2
                fi
            fi
        fi
    fi

    # Pass query through (potentially with optimization metadata)
    echo "$query"
}

# Run main if executed directly
if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
    main "$@"
fi
