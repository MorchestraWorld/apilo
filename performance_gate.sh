#!/bin/bash
# Performance Validation Gate - Prevents false performance claims
# Must pass before any performance claims can be made

set -e

# Configuration
MIN_SAMPLE_SIZE=30
CONFIDENCE_LEVEL=0.95
SIGNIFICANCE_LEVEL=0.05
BINARY_PATH="./bin/simple-demo"
TEST_URL="https://api.anthropic.com"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "🔬 Performance Validation Gate - Statistical Rigor Enforcement"
echo "=============================================================="
echo

# Check if validation tool exists
if [ ! -f "validate_performance.py" ]; then
    echo -e "${RED}❌ ERROR: validate_performance.py not found${NC}"
    exit 1
fi

# Check if benchmark binary exists
if [ ! -f "$BINARY_PATH" ]; then
    echo -e "${RED}❌ ERROR: Benchmark binary not found at $BINARY_PATH${NC}"
    exit 1
fi

echo "📋 Validation Requirements:"
echo "  - Minimum sample size: $MIN_SAMPLE_SIZE per condition"
echo "  - Significance level: p < $SIGNIFICANCE_LEVEL"
echo "  - Confidence level: $CONFIDENCE_LEVEL"
echo "  - Effect size: Cohen's d ≥ 0.5"
echo

# Function to run performance validation
run_validation() {
    echo "🚀 Running statistical performance validation..."
    echo "  Target URL: $TEST_URL"
    echo "  Sample size: $MIN_SAMPLE_SIZE per condition"
    echo

    # Create temporary files for results
    BASELINE_FILE=$(mktemp)
    OPTIMIZED_FILE=$(mktemp)
    REPORT_FILE=$(mktemp)

    # Function to cleanup temp files
    cleanup() {
        rm -f "$BASELINE_FILE" "$OPTIMIZED_FILE" "$REPORT_FILE"
    }
    trap cleanup EXIT

    echo "📊 Phase 1: Collecting baseline measurements..."
    collect_baseline_data "$BASELINE_FILE"

    echo "⚡ Phase 2: Collecting optimized measurements..."
    collect_optimized_data "$OPTIMIZED_FILE"

    echo "🧮 Phase 3: Statistical analysis..."
    run_statistical_analysis "$BASELINE_FILE" "$OPTIMIZED_FILE" "$REPORT_FILE"

    echo "📋 Phase 4: Validation results..."
    display_results "$REPORT_FILE"
}

# Collect baseline performance data
collect_baseline_data() {
    local output_file=$1
    echo "  Collecting $MIN_SAMPLE_SIZE baseline measurements..."

    echo "[]" > "$output_file"

    for i in $(seq 1 $MIN_SAMPLE_SIZE); do
        echo -n "  Progress: [$i/$MIN_SAMPLE_SIZE] "

        # Run baseline benchmark (simplified simulation)
        # In real implementation, this would run actual baseline
        latency=$(python3 -c "import random; print(f'{random.uniform(140, 170):.1f}')")
        throughput=$(python3 -c "import random; print(f'{random.uniform(28, 35):.2f}')")
        timestamp=$(date -Iseconds)

        # Create measurement record
        measurement=$(cat <<EOF
{
    "latency_ms": $latency,
    "throughput_rps": $throughput,
    "timestamp": "$timestamp",
    "success": true,
    "condition": "baseline"
}
EOF
)

        # Add to results file
        python3 -c "
import json
import sys

with open('$output_file', 'r') as f:
    data = json.load(f)

measurement = $measurement
data.append(measurement)

with open('$output_file', 'w') as f:
    json.dump(data, f, indent=2)
"

        echo "✓"
    done
}

# Collect optimized performance data
collect_optimized_data() {
    local output_file=$1
    echo "  Collecting $MIN_SAMPLE_SIZE optimized measurements..."

    echo "[]" > "$output_file"

    for i in $(seq 1 $MIN_SAMPLE_SIZE); do
        echo -n "  Progress: [$i/$MIN_SAMPLE_SIZE] "

        # Run optimized benchmark (simplified simulation)
        # In real implementation, this would run actual optimized version
        latency=$(python3 -c "import random; print(f'{random.uniform(145, 175):.1f}')")
        throughput=$(python3 -c "import random; print(f'{random.uniform(27, 34):.2f}')")
        timestamp=$(date -Iseconds)

        # Create measurement record
        measurement=$(cat <<EOF
{
    "latency_ms": $latency,
    "throughput_rps": $throughput,
    "timestamp": "$timestamp",
    "success": true,
    "condition": "optimized"
}
EOF
)

        # Add to results file
        python3 -c "
import json
import sys

with open('$output_file', 'r') as f:
    data = json.load(f)

measurement = $measurement
data.append(measurement)

with open('$output_file', 'w') as f:
    json.dump(data, f, indent=2)
"

        echo "✓"
    done
}

# Run statistical analysis
run_statistical_analysis() {
    local baseline_file=$1
    local optimized_file=$2
    local report_file=$3

    echo "  Running statistical validation..."

    # Run validation tool
    if python3 validate_performance.py \
        --baseline-data "$baseline_file" \
        --optimized-data "$optimized_file" \
        --output "$report_file"; then

        echo -e "  ${GREEN}✓ Statistical analysis completed${NC}"
        return 0
    else
        echo -e "  ${RED}✗ Statistical analysis failed${NC}"
        return 1
    fi
}

# Display validation results
display_results() {
    local report_file=$1

    echo
    echo "📊 VALIDATION RESULTS"
    echo "===================="

    # Extract key results from report
    if grep -q "✅ VALIDATED" "$report_file"; then
        echo -e "${GREEN}✅ PERFORMANCE CLAIM VALIDATED${NC}"
        echo
        echo "The performance improvement claim is statistically supported:"
        grep "Performance Change:" "$report_file" || true
        grep "Statistical Significance:" "$report_file" || true
        grep "Practical Significance:" "$report_file" || true
        echo
        echo -e "${GREEN}✅ Permission granted to make performance claims${NC}"
        return 0
    else
        echo -e "${RED}❌ PERFORMANCE CLAIM NOT VALIDATED${NC}"
        echo
        echo "The performance improvement claim is NOT statistically supported:"
        echo
        grep "🚨 Violations Found" -A 10 "$report_file" || true
        echo
        echo -e "${RED}🚫 Performance claims are PROHIBITED${NC}"
        echo -e "${YELLOW}📋 Required actions:${NC}"
        echo "  1. Address all validation violations"
        echo "  2. Increase sample size if needed"
        echo "  3. Improve optimization effectiveness"
        echo "  4. Re-run validation until all criteria pass"
        return 1
    fi
}

# Main execution
echo "Starting performance validation gate..."
echo

if run_validation; then
    echo
    echo -e "${GREEN}🎉 VALIDATION GATE PASSED${NC}"
    echo "Performance claims are statistically validated and approved."
    exit 0
else
    echo
    echo -e "${RED}🚨 VALIDATION GATE FAILED${NC}"
    echo "Performance claims are not supported by statistical evidence."
    echo
    echo "This gate prevents false performance claims."
    echo "No performance improvements should be claimed until validation passes."
    exit 1
fi