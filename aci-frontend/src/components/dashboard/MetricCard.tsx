/**
 * MetricCard Component
 *
 * Reusable dashboard metric card for displaying KPIs with optional trend indicators.
 * Follows design token standards - NO hardcoded CSS values.
 *
 * @example
 * ```tsx
 * <MetricCard
 *   title="Total Threats"
 *   value={2847}
 *   icon={<AlertTriangle />}
 *   trend={{ direction: 'up', percentage: 12.5 }}
 *   variant="critical"
 * />
 * ```
 */

import React from 'react';
import { Card, CardContent } from '@/components/ui/card';
import { cn } from '@/lib/utils';
import { colors } from '@/styles/tokens/colors';
import { spacing, componentSpacing } from '@/styles/tokens/spacing';
import { typography } from '@/styles/tokens/typography';
import { shadows } from '@/styles/tokens/shadows';
import { borders } from '@/styles/tokens/borders';

export interface MetricCardProps {
  /** Metric title (e.g., "Total Threats") */
  title: string;
  /** Numeric value to display */
  value: number;
  /** Optional icon element */
  icon?: React.ReactNode;
  /** Optional trend indicator */
  trend?: {
    direction: 'up' | 'down' | 'neutral';
    percentage: number;
  };
  /** Visual variant affecting border and accent colors */
  variant?: 'default' | 'critical' | 'warning' | 'success';
}

/**
 * Formats large numbers with K/M suffixes
 * @example
 * formatNumber(1500) => "1,500"
 * formatNumber(1500000) => "1.5M"
 */
function formatNumber(value: number): string {
  if (value >= 1_000_000) {
    return new Intl.NumberFormat('en-US', {
      notation: 'compact',
      maximumFractionDigits: 1,
    }).format(value);
  }

  return new Intl.NumberFormat('en-US').format(value);
}

/**
 * Returns the appropriate border color CSS variable based on variant
 */
function getVariantBorderColor(variant: MetricCardProps['variant']): string {
  switch (variant) {
    case 'critical':
      return colors.severity.critical;
    case 'warning':
      return colors.severity.medium;
    case 'success':
      return colors.severity.low;
    case 'default':
    default:
      return colors.brand.primary;
  }
}

/**
 * Returns the appropriate accent color for trend indicators
 */
function getTrendColor(direction: 'up' | 'down' | 'neutral'): string {
  switch (direction) {
    case 'up':
      return colors.semantic.success;
    case 'down':
      return colors.semantic.error;
    case 'neutral':
    default:
      return colors.text.muted;
  }
}

/**
 * Returns the appropriate arrow icon for trend direction
 */
function getTrendArrow(direction: 'up' | 'down' | 'neutral'): string {
  switch (direction) {
    case 'up':
      return '↑';
    case 'down':
      return '↓';
    case 'neutral':
    default:
      return '→';
  }
}

export function MetricCard({
  title,
  value,
  icon,
  trend,
  variant = 'default',
}: MetricCardProps) {
  const borderColor = getVariantBorderColor(variant);
  const formattedValue = formatNumber(value);

  // Build accessible label
  const ariaLabel = trend
    ? `${title}: ${formattedValue}, trending ${trend.direction} by ${trend.percentage}%`
    : `${title}: ${formattedValue}`;

  return (
    <Card
      role="region"
      aria-label={ariaLabel}
      data-testid="metric-card"
      data-variant={variant}
      className={cn('relative overflow-hidden')}
      style={{
        borderLeftWidth: borders.width.thick,
        borderLeftColor: borderColor,
        borderRadius: borders.radius.lg,
        boxShadow: shadows.md,
      }}
    >
      <CardContent
        style={{
          padding: componentSpacing.lg,
        }}
      >
        {/* Header: Title and Icon */}
        <div
          className="flex items-start justify-between"
          style={{
            marginBottom: spacing[3],
          }}
        >
          <h3
            style={{
              fontSize: typography.fontSize.sm,
              fontWeight: typography.fontWeight.medium,
              color: colors.text.muted,
              textTransform: 'uppercase',
              letterSpacing: typography.letterSpacing.wide,
            }}
          >
            {title}
          </h3>
          {icon && (
            <div
              style={{
                color: borderColor,
                opacity: 0.7,
              }}
              aria-hidden="true"
            >
              {icon}
            </div>
          )}
        </div>

        {/* Main Value */}
        <div
          data-testid="metric-value"
          style={{
            fontSize: typography.fontSize['3xl'],
            fontWeight: typography.fontWeight.bold,
            color: colors.text.primary,
            lineHeight: typography.lineHeight.tight,
            marginBottom: trend ? spacing[3] : '0',
          }}
        >
          {formattedValue}
        </div>

        {/* Trend Indicator */}
        {trend && (
          <div
            data-testid="trend-indicator"
            data-direction={trend.direction}
            className="flex items-center"
            style={{
              fontSize: typography.fontSize.sm,
              fontWeight: typography.fontWeight.medium,
              color: getTrendColor(trend.direction),
              gap: spacing[1],
            }}
          >
            <span
              aria-hidden="true"
              style={{
                fontSize: typography.fontSize.base,
              }}
            >
              {getTrendArrow(trend.direction)}
            </span>
            <span>
              {trend.percentage.toFixed(1)}%
            </span>
            <span
              style={{
                color: colors.text.muted,
                fontWeight: typography.fontWeight.normal,
              }}
            >
              vs. previous period
            </span>
          </div>
        )}
      </CardContent>
    </Card>
  );
}

MetricCard.displayName = 'MetricCard';
