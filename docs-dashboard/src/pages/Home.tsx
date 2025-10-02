import { Link } from 'react-router-dom';
import {
  Zap,
  TrendingUp,
  Shield,
  Gauge,
  ArrowRight,
  Database,
  Activity,
  Cpu,
  GitBranch
} from 'lucide-react';
import { MetricsCard } from '../components/MetricsCard';
import { FeatureCard } from '../components/FeatureCard';

export function Home() {
  return (
    <div className="min-h-screen">
      {/* Hero Section */}
      <section className="relative overflow-hidden py-20 px-4">
        <div className="absolute inset-0 gradient-bg opacity-10 animate-gradient bg-[length:200%_200%]" />
        <div className="relative max-w-5xl mx-auto text-center">
          <h1 className="text-5xl md:text-6xl font-bold mb-6">
            API Latency Optimizer
          </h1>
          <p className="text-3xl md:text-4xl font-bold gradient-text mb-8">
            93.69% Faster API Performance
          </p>
          <p className="text-xl text-muted-foreground max-w-2xl mx-auto mb-8">
            Production-ready API optimization achieving 515ms → 33ms average latency
            with memory-bounded caching, circuit breaker protection, and real-time monitoring.
          </p>
          <div className="flex flex-wrap gap-4 justify-center">
            <Link to="/docs/quickstart" className="btn-primary">
              Get Started
              <ArrowRight className="w-4 h-4 ml-2" />
            </Link>
            <Link to="/docs/performance" className="btn-secondary">
              View Performance
            </Link>
          </div>
        </div>
      </section>

      {/* Quick Stats */}
      <section className="py-16 px-4 bg-muted/30">
        <div className="max-w-6xl mx-auto">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            <MetricsCard
              title="Latency Reduction"
              value="93.69%"
              description="Average latency: 515ms → 33ms"
              icon={Zap}
              trend="down"
              trendValue="482ms"
            />
            <MetricsCard
              title="Throughput Increase"
              value="15.8x"
              description="2.1 RPS → 33.5 RPS"
              icon={TrendingUp}
              trend="up"
              trendValue="31.4 RPS"
            />
            <MetricsCard
              title="Cache Hit Ratio"
              value="98%"
              description="Sustained under load"
              icon={Database}
              trend="up"
              trendValue="98%"
            />
          </div>
        </div>
      </section>

      {/* Features Grid */}
      <section className="py-16 px-4">
        <div className="max-w-6xl mx-auto">
          <div className="text-center mb-12">
            <h2 className="text-3xl font-bold mb-4">Production-Ready Features</h2>
            <p className="text-muted-foreground max-w-2xl mx-auto">
              Built with enterprise-grade reliability, monitoring, and performance optimization
            </p>
          </div>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            <FeatureCard
              icon={Database}
              title="Memory-Bounded Cache"
              description="Hard memory limits with configurable MB maximum and automatic GC optimization"
            />
            <FeatureCard
              icon={Shield}
              title="Circuit Breaker"
              description="Three-state protection with automatic failover and health checking"
            />
            <FeatureCard
              icon={Activity}
              title="Real-Time Monitoring"
              description="Production dashboard with metrics, alerts, and Prometheus integration"
            />
            <FeatureCard
              icon={GitBranch}
              title="Advanced Invalidation"
              description="Tag, pattern, dependency, and version-based cache invalidation strategies"
            />
            <FeatureCard
              icon={Cpu}
              title="HTTP/2 Optimization"
              description="Connection pooling, multiplexing, and TLS optimization for maximum throughput"
            />
            <FeatureCard
              icon={Gauge}
              title="Alert Management"
              description="Multi-level severity alerts with cooldown management and acknowledgment"
            />
          </div>
        </div>
      </section>

      {/* Performance Highlights */}
      <section className="py-16 px-4 bg-muted/30">
        <div className="max-w-6xl mx-auto">
          <div className="text-center mb-12">
            <h2 className="text-3xl font-bold mb-4">Validated Performance</h2>
            <p className="text-muted-foreground">
              Comprehensive testing and statistical validation
            </p>
          </div>
          <div className="card p-8">
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead>
                  <tr className="border-b">
                    <th className="text-left py-3 px-4">Metric</th>
                    <th className="text-right py-3 px-4">Baseline</th>
                    <th className="text-right py-3 px-4">Optimized</th>
                    <th className="text-right py-3 px-4">Improvement</th>
                  </tr>
                </thead>
                <tbody>
                  <tr className="border-b hover:bg-accent/50">
                    <td className="py-3 px-4 font-medium">Average Latency</td>
                    <td className="text-right py-3 px-4">515ms</td>
                    <td className="text-right py-3 px-4">33ms</td>
                    <td className="text-right py-3 px-4 text-green-500 font-bold">93.69%</td>
                  </tr>
                  <tr className="border-b hover:bg-accent/50">
                    <td className="py-3 px-4 font-medium">P50 Latency</td>
                    <td className="text-right py-3 px-4">460ms</td>
                    <td className="text-right py-3 px-4">29ms</td>
                    <td className="text-right py-3 px-4 text-green-500 font-bold">93.7%</td>
                  </tr>
                  <tr className="border-b hover:bg-accent/50">
                    <td className="py-3 px-4 font-medium">P95 Latency</td>
                    <td className="text-right py-3 px-4">850ms</td>
                    <td className="text-right py-3 px-4">75ms</td>
                    <td className="text-right py-3 px-4 text-green-500 font-bold">91.2%</td>
                  </tr>
                  <tr className="border-b hover:bg-accent/50">
                    <td className="py-3 px-4 font-medium">Throughput</td>
                    <td className="text-right py-3 px-4">2.1 RPS</td>
                    <td className="text-right py-3 px-4">33.5 RPS</td>
                    <td className="text-right py-3 px-4 text-green-500 font-bold">15.8x</td>
                  </tr>
                  <tr className="hover:bg-accent/50">
                    <td className="py-3 px-4 font-medium">Cache Hit Ratio</td>
                    <td className="text-right py-3 px-4">0%</td>
                    <td className="text-right py-3 px-4">98%</td>
                    <td className="text-right py-3 px-4 text-green-500 font-bold">N/A</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </section>

      {/* CTA Section */}
      <section className="py-16 px-4">
        <div className="max-w-4xl mx-auto text-center">
          <div className="gradient-bg rounded-2xl p-12 text-white">
            <h2 className="text-3xl font-bold mb-4">Ready to Optimize Your APIs?</h2>
            <p className="text-lg mb-8 opacity-90">
              Get started in 5 minutes with our Quick Start guide or explore the full documentation
            </p>
            <div className="flex flex-wrap gap-4 justify-center">
              <Link
                to="/docs/quickstart"
                className="inline-flex items-center justify-center px-6 py-3 rounded-lg bg-white text-primary-600 font-medium hover:bg-gray-100 transition-colors"
              >
                Quick Start Guide
                <ArrowRight className="w-4 h-4 ml-2" />
              </Link>
              <Link
                to="/docs/features"
                className="inline-flex items-center justify-center px-6 py-3 rounded-lg border-2 border-white text-white font-medium hover:bg-white/10 transition-colors"
              >
                Explore Features
              </Link>
            </div>
          </div>
        </div>
      </section>
    </div>
  );
}
