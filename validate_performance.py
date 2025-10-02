#!/usr/bin/env python3
"""
Statistical Performance Validation Tool
Enforces statistical rigor for all performance claims
Prevents false claims through automated validation
"""

import json
import subprocess
import statistics
import math
from typing import List, Dict, Tuple, Optional
from dataclasses import dataclass
from datetime import datetime
import argparse
import sys

@dataclass
class PerformanceResult:
    """Single performance measurement result"""
    latency_ms: float
    throughput_rps: float
    timestamp: str
    success: bool
    metadata: Dict = None

@dataclass
class StatisticalSummary:
    """Statistical analysis of performance results"""
    sample_size: int
    mean: float
    median: float
    std_dev: float
    min_val: float
    max_val: float
    p25: float
    p75: float
    p95: float
    p99: float
    confidence_interval_95: Tuple[float, float]
    coefficient_variation: float

@dataclass
class ComparisonResult:
    """Statistical comparison between baseline and optimized"""
    baseline_stats: StatisticalSummary
    optimized_stats: StatisticalSummary
    improvement_percent: float
    improvement_absolute: float
    p_value: float
    cohens_d: float
    confidence_interval_diff: Tuple[float, float]
    statistical_significance: bool
    practical_significance: bool
    sample_adequate: bool

class PerformanceValidator:
    """Enforces statistical validation for performance claims"""

    # Statistical Requirements (from STATISTICAL_VALIDATION_PROTOCOL.md)
    MIN_SAMPLE_SIZE = 30
    SIGNIFICANCE_LEVEL = 0.05
    MIN_EFFECT_SIZE = 0.5  # Cohen's d for medium effect
    CONFIDENCE_LEVEL = 0.95
    MAX_CV = 0.3  # Maximum coefficient of variation (30%)

    def __init__(self, strict_mode: bool = True):
        self.strict_mode = strict_mode
        self.violations = []

    def validate_sample_size(self, n: int) -> bool:
        """Validate adequate sample size"""
        if n < self.MIN_SAMPLE_SIZE:
            self.violations.append(f"Insufficient sample size: {n} < {self.MIN_SAMPLE_SIZE}")
            return False
        return True

    def calculate_statistics(self, values: List[float]) -> StatisticalSummary:
        """Calculate comprehensive descriptive statistics"""
        if not values:
            raise ValueError("Cannot calculate statistics for empty dataset")

        n = len(values)
        mean_val = statistics.mean(values)
        median_val = statistics.median(values)

        if n > 1:
            std_dev = statistics.stdev(values)
        else:
            std_dev = 0.0

        # Calculate percentiles
        sorted_vals = sorted(values)
        p25 = self._percentile(sorted_vals, 25)
        p75 = self._percentile(sorted_vals, 75)
        p95 = self._percentile(sorted_vals, 95)
        p99 = self._percentile(sorted_vals, 99)

        # Calculate 95% confidence interval for mean
        if n > 1:
            margin_error = 1.96 * (std_dev / math.sqrt(n))  # Approximate for large n
            ci_lower = mean_val - margin_error
            ci_upper = mean_val + margin_error
        else:
            ci_lower = ci_upper = mean_val

        # Coefficient of variation
        cv = (std_dev / mean_val) if mean_val != 0 else 0

        return StatisticalSummary(
            sample_size=n,
            mean=mean_val,
            median=median_val,
            std_dev=std_dev,
            min_val=min(values),
            max_val=max(values),
            p25=p25,
            p75=p75,
            p95=p95,
            p99=p99,
            confidence_interval_95=(ci_lower, ci_upper),
            coefficient_variation=cv
        )

    def _percentile(self, sorted_values: List[float], percentile: int) -> float:
        """Calculate percentile value"""
        if not sorted_values:
            return 0.0

        n = len(sorted_values)
        if n == 1:
            return sorted_values[0]

        rank = (percentile / 100.0) * (n - 1)
        lower_idx = int(rank)
        upper_idx = min(lower_idx + 1, n - 1)

        if lower_idx == upper_idx:
            return sorted_values[lower_idx]

        weight = rank - lower_idx
        return sorted_values[lower_idx] * (1 - weight) + sorted_values[upper_idx] * weight

    def two_sample_ttest(self, baseline: List[float], optimized: List[float]) -> Tuple[float, float]:
        """Perform two-sample t-test (simplified version)"""
        n1, n2 = len(baseline), len(optimized)

        if n1 < 2 or n2 < 2:
            return 0.0, 1.0  # No significance possible

        mean1, mean2 = statistics.mean(baseline), statistics.mean(optimized)
        std1, std2 = statistics.stdev(baseline), statistics.stdev(optimized)

        # Pooled standard error
        pooled_se = math.sqrt((std1**2 / n1) + (std2**2 / n2))

        if pooled_se == 0:
            return 0.0, 1.0

        # T-statistic
        t_stat = (mean2 - mean1) / pooled_se

        # Approximate p-value (simplified - would use proper t-distribution in production)
        # For large samples, t-distribution approaches normal distribution
        p_value = 2 * (1 - self._norm_cdf(abs(t_stat)))

        return t_stat, min(p_value, 1.0)

    def _norm_cdf(self, x: float) -> float:
        """Approximate normal CDF (simplified)"""
        # Very rough approximation - would use proper implementation in production
        if x > 6:
            return 1.0
        elif x < -6:
            return 0.0
        else:
            # Rough approximation
            return 0.5 + 0.5 * math.tanh(x * 0.7)

    def cohens_d(self, baseline: List[float], optimized: List[float]) -> float:
        """Calculate Cohen's d effect size"""
        if len(baseline) < 2 or len(optimized) < 2:
            return 0.0

        mean1, mean2 = statistics.mean(baseline), statistics.mean(optimized)
        std1, std2 = statistics.stdev(baseline), statistics.stdev(optimized)

        # Pooled standard deviation
        n1, n2 = len(baseline), len(optimized)
        pooled_std = math.sqrt(((n1 - 1) * std1**2 + (n2 - 1) * std2**2) / (n1 + n2 - 2))

        if pooled_std == 0:
            return 0.0

        return (mean2 - mean1) / pooled_std

    def validate_performance_comparison(self, baseline_data: List[float],
                                      optimized_data: List[float]) -> ComparisonResult:
        """Comprehensive statistical validation of performance comparison"""

        # Reset violations
        self.violations = []

        # Validate sample sizes
        sample_adequate = (self.validate_sample_size(len(baseline_data)) and
                          self.validate_sample_size(len(optimized_data)))

        # Calculate statistics for both conditions
        baseline_stats = self.calculate_statistics(baseline_data)
        optimized_stats = self.calculate_statistics(optimized_data)

        # Check coefficient of variation
        if baseline_stats.coefficient_variation > self.MAX_CV:
            self.violations.append(f"High baseline variability: CV={baseline_stats.coefficient_variation:.3f} > {self.MAX_CV}")
        if optimized_stats.coefficient_variation > self.MAX_CV:
            self.violations.append(f"High optimized variability: CV={optimized_stats.coefficient_variation:.3f} > {self.MAX_CV}")

        # Calculate improvement
        improvement_absolute = optimized_stats.mean - baseline_stats.mean
        improvement_percent = (improvement_absolute / baseline_stats.mean) * 100 if baseline_stats.mean != 0 else 0

        # Statistical tests
        t_stat, p_value = self.two_sample_ttest(baseline_data, optimized_data)
        effect_size = self.cohens_d(baseline_data, optimized_data)

        # Significance tests
        statistical_significance = p_value < self.SIGNIFICANCE_LEVEL
        practical_significance = abs(effect_size) >= self.MIN_EFFECT_SIZE

        # Confidence interval for difference
        if len(baseline_data) > 1 and len(optimized_data) > 1:
            se_diff = math.sqrt((baseline_stats.std_dev**2 / len(baseline_data)) +
                               (optimized_stats.std_dev**2 / len(optimized_data)))
            margin = 1.96 * se_diff
            ci_lower = improvement_absolute - margin
            ci_upper = improvement_absolute + margin
        else:
            ci_lower = ci_upper = improvement_absolute

        return ComparisonResult(
            baseline_stats=baseline_stats,
            optimized_stats=optimized_stats,
            improvement_percent=improvement_percent,
            improvement_absolute=improvement_absolute,
            p_value=p_value,
            cohens_d=effect_size,
            confidence_interval_diff=(ci_lower, ci_upper),
            statistical_significance=statistical_significance,
            practical_significance=practical_significance,
            sample_adequate=sample_adequate
        )

    def run_benchmark(self, binary_path: str, url: str, requests: int = 50,
                     concurrency: int = 10) -> List[PerformanceResult]:
        """Run performance benchmark with statistical validation"""

        if requests < self.MIN_SAMPLE_SIZE:
            print(f"WARNING: Request count {requests} < minimum {self.MIN_SAMPLE_SIZE}")

        results = []
        for i in range(requests):
            try:
                # Run single request and parse output
                cmd = [binary_path, "-url", url, "-requests", "1", "-concurrency", "1"]
                result = subprocess.run(cmd, capture_output=True, text=True, timeout=30)

                if result.returncode == 0:
                    # Parse output for latency and throughput
                    # This is a simplified parser - would need to match actual output format
                    latency = self._extract_latency_from_output(result.stdout)
                    throughput = self._extract_throughput_from_output(result.stdout)

                    results.append(PerformanceResult(
                        latency_ms=latency,
                        throughput_rps=throughput,
                        timestamp=datetime.now().isoformat(),
                        success=True
                    ))
                else:
                    results.append(PerformanceResult(
                        latency_ms=0,
                        throughput_rps=0,
                        timestamp=datetime.now().isoformat(),
                        success=False,
                        metadata={"error": result.stderr}
                    ))

            except subprocess.TimeoutExpired:
                results.append(PerformanceResult(
                    latency_ms=0,
                    throughput_rps=0,
                    timestamp=datetime.now().isoformat(),
                    success=False,
                    metadata={"error": "timeout"}
                ))

        return results

    def _extract_latency_from_output(self, output: str) -> float:
        """Extract latency from benchmark output - simplified parser"""
        # This would need to be implemented based on actual output format
        # For now, return a placeholder
        return 150.0  # ms

    def _extract_throughput_from_output(self, output: str) -> float:
        """Extract throughput from benchmark output - simplified parser"""
        # This would need to be implemented based on actual output format
        # For now, return a placeholder
        return 30.0  # req/s

    def generate_validation_report(self, comparison: ComparisonResult) -> str:
        """Generate comprehensive validation report"""

        report = f"""
# 📊 Statistical Performance Validation Report

**Generated**: {datetime.now().isoformat()}
**Validation Standard**: {self.CONFIDENCE_LEVEL*100}% confidence, p<{self.SIGNIFICANCE_LEVEL}

## 🎯 Executive Summary

**Performance Change**: {comparison.improvement_percent:+.1f}% ({comparison.improvement_absolute:+.1f}ms)
**Statistical Significance**: {'✅ YES' if comparison.statistical_significance else '❌ NO'} (p={comparison.p_value:.4f})
**Practical Significance**: {'✅ YES' if comparison.practical_significance else '❌ NO'} (d={comparison.cohens_d:.3f})
**Sample Size Adequate**: {'✅ YES' if comparison.sample_adequate else '❌ NO'}

## 📈 Statistical Results

### Baseline Performance
- **Mean**: {comparison.baseline_stats.mean:.1f}ms
- **Median**: {comparison.baseline_stats.median:.1f}ms
- **Std Dev**: {comparison.baseline_stats.std_dev:.1f}ms
- **95% CI**: [{comparison.baseline_stats.confidence_interval_95[0]:.1f}, {comparison.baseline_stats.confidence_interval_95[1]:.1f}]ms
- **Sample Size**: n={comparison.baseline_stats.sample_size}
- **Coefficient of Variation**: {comparison.baseline_stats.coefficient_variation:.3f}

### Optimized Performance
- **Mean**: {comparison.optimized_stats.mean:.1f}ms
- **Median**: {comparison.optimized_stats.median:.1f}ms
- **Std Dev**: {comparison.optimized_stats.std_dev:.1f}ms
- **95% CI**: [{comparison.optimized_stats.confidence_interval_95[0]:.1f}, {comparison.optimized_stats.confidence_interval_95[1]:.1f}]ms
- **Sample Size**: n={comparison.optimized_stats.sample_size}
- **Coefficient of Variation**: {comparison.optimized_stats.coefficient_variation:.3f}

### Statistical Comparison
- **Improvement**: {comparison.improvement_percent:+.1f}% ({comparison.improvement_absolute:+.1f}ms)
- **95% CI for Difference**: [{comparison.confidence_interval_diff[0]:.1f}, {comparison.confidence_interval_diff[1]:.1f}]ms
- **P-value**: {comparison.p_value:.4f}
- **Effect Size (Cohen's d)**: {comparison.cohens_d:.3f}
- **Effect Interpretation**: {self._interpret_effect_size(comparison.cohens_d)}

## ✅ Validation Status

### Statistical Requirements
- **Sample Size**: {'✅ PASS' if comparison.sample_adequate else '❌ FAIL'} (n≥{self.MIN_SAMPLE_SIZE})
- **Statistical Significance**: {'✅ PASS' if comparison.statistical_significance else '❌ FAIL'} (p<{self.SIGNIFICANCE_LEVEL})
- **Effect Size**: {'✅ PASS' if comparison.practical_significance else '❌ FAIL'} (|d|≥{self.MIN_EFFECT_SIZE})
- **Baseline Stability**: {'✅ PASS' if comparison.baseline_stats.coefficient_variation <= self.MAX_CV else '❌ FAIL'} (CV≤{self.MAX_CV})

### Overall Validation
{'✅ VALIDATED' if self._is_claim_valid(comparison) else '❌ NOT VALIDATED'}: Performance improvement claim

## 🚨 Violations Found
"""

        if self.violations:
            for violation in self.violations:
                report += f"- ❌ {violation}\n"
        else:
            report += "- ✅ No violations found\n"

        report += f"""
## 🎯 Recommendations

{self._generate_recommendations(comparison)}

---
*This report meets statistical validation protocol requirements.*
*Performance claims are only valid if all validation criteria pass.*
"""

        return report

    def _interpret_effect_size(self, d: float) -> str:
        """Interpret Cohen's d effect size"""
        abs_d = abs(d)
        if abs_d < 0.2:
            return "Negligible effect"
        elif abs_d < 0.5:
            return "Small effect"
        elif abs_d < 0.8:
            return "Medium effect"
        else:
            return "Large effect"

    def _is_claim_valid(self, comparison: ComparisonResult) -> bool:
        """Determine if performance claim is statistically valid"""
        return (comparison.sample_adequate and
                comparison.statistical_significance and
                comparison.practical_significance and
                len(self.violations) == 0)

    def _generate_recommendations(self, comparison: ComparisonResult) -> str:
        """Generate recommendations based on validation results"""
        recommendations = []

        if not comparison.sample_adequate:
            recommendations.append(f"Increase sample size to at least {self.MIN_SAMPLE_SIZE} per condition")

        if not comparison.statistical_significance:
            recommendations.append("No statistical significance detected - improvement may be due to chance")

        if not comparison.practical_significance:
            recommendations.append("Effect size is small - improvement may not be practically meaningful")

        if comparison.baseline_stats.coefficient_variation > self.MAX_CV:
            recommendations.append("High baseline variability - consider controlling external factors")

        if self.violations:
            recommendations.append("Address all validation violations before making performance claims")

        if not recommendations:
            recommendations.append("All validation criteria met - performance improvement is statistically supported")

        return "\n".join(f"- {rec}" for rec in recommendations)

def main():
    """Command-line interface for performance validation"""
    parser = argparse.ArgumentParser(description="Statistical Performance Validation Tool")
    parser.add_argument("--baseline-data", type=str, help="JSON file with baseline measurements")
    parser.add_argument("--optimized-data", type=str, help="JSON file with optimized measurements")
    parser.add_argument("--run-benchmark", type=str, help="Path to benchmark binary")
    parser.add_argument("--url", type=str, default="https://api.anthropic.com", help="Target URL")
    parser.add_argument("--requests", type=int, default=50, help="Number of requests per condition")
    parser.add_argument("--output", type=str, help="Output file for validation report")

    args = parser.parse_args()

    validator = PerformanceValidator()

    if args.baseline_data and args.optimized_data:
        # Validate existing data
        with open(args.baseline_data) as f:
            baseline = json.load(f)
        with open(args.optimized_data) as f:
            optimized = json.load(f)

        # Extract latency values (assuming JSON format with latency_ms field)
        baseline_latencies = [r['latency_ms'] for r in baseline if r.get('success', True)]
        optimized_latencies = [r['latency_ms'] for r in optimized if r.get('success', True)]

        comparison = validator.validate_performance_comparison(baseline_latencies, optimized_latencies)

    elif args.run_benchmark:
        # Run new benchmark
        print("Running baseline benchmark...")
        baseline_results = validator.run_benchmark(args.run_benchmark, args.url, args.requests)

        print("Running optimized benchmark...")
        optimized_results = validator.run_benchmark(args.run_benchmark, args.url, args.requests)

        baseline_latencies = [r.latency_ms for r in baseline_results if r.success]
        optimized_latencies = [r.latency_ms for r in optimized_results if r.success]

        comparison = validator.validate_performance_comparison(baseline_latencies, optimized_latencies)

    else:
        print("Error: Must provide either --baseline-data and --optimized-data, or --run-benchmark")
        sys.exit(1)

    # Generate validation report
    report = validator.generate_validation_report(comparison)

    if args.output:
        with open(args.output, 'w') as f:
            f.write(report)
        print(f"Validation report written to {args.output}")
    else:
        print(report)

    # Exit with error code if validation fails
    if not validator._is_claim_valid(comparison):
        print("\n🚨 VALIDATION FAILED: Performance claims are not statistically supported")
        sys.exit(1)
    else:
        print("\n✅ VALIDATION PASSED: Performance claims are statistically supported")

if __name__ == "__main__":
    main()