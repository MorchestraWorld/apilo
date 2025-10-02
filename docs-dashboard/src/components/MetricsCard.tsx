import type { LucideIcon } from 'lucide-react';

interface MetricsCardProps {
  title: string;
  value: string;
  description?: string;
  icon?: LucideIcon;
  trend?: 'up' | 'down';
  trendValue?: string;
}

export function MetricsCard({
  title,
  value,
  description,
  icon: Icon,
  trend,
  trendValue
}: MetricsCardProps) {
  return (
    <div className="metric-card">
      <div className="flex items-start justify-between">
        <div className="flex-1">
          <p className="text-sm text-muted-foreground font-medium">{title}</p>
          <div className="mt-2 flex items-baseline gap-2">
            <h3 className="text-3xl font-bold gradient-text">{value}</h3>
            {trend && trendValue && (
              <span className={`text-sm font-medium ${trend === 'up' ? 'text-green-500' : 'text-red-500'}`}>
                {trend === 'up' ? '↑' : '↓'} {trendValue}
              </span>
            )}
          </div>
          {description && (
            <p className="mt-2 text-sm text-muted-foreground">{description}</p>
          )}
        </div>
        {Icon && (
          <div className="gradient-bg p-3 rounded-lg">
            <Icon className="w-6 h-6 text-white" />
          </div>
        )}
      </div>
    </div>
  );
}
