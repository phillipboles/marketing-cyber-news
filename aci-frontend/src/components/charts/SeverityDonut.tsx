/**
 * SeverityDonut Chart Component
 *
 * Displays threat severity distribution as an interactive donut chart using Recharts.
 * Uses design tokens for colors and animations - NO hardcoded values.
 *
 * @see /aci-frontend/src/styles/tokens/colors.ts
 * @see /aci-frontend/src/styles/tokens/motion.ts
 */

import React, { useMemo } from 'react';
import { PieChart, Pie, Cell, Tooltip, ResponsiveContainer } from 'recharts';
import { colors } from '@/styles/tokens/colors';
import type { Severity } from '@/types/threat';

/**
 * Props for SeverityDonut component
 */
export interface SeverityDonutProps {
  /** Severity distribution counts */
  data: {
    critical: number;
    high: number;
    medium: number;
    low: number;
  };
  /** Chart size preset */
  size?: 'sm' | 'md' | 'lg';
  /** Show legend with severity labels */
  showLegend?: boolean;
  /** Enable entrance animation */
  animated?: boolean;
}

/**
 * Size configuration mapping
 */
const SIZE_MAP = {
  sm: { width: 150, height: 150, innerRadius: 45, outerRadius: 65 },
  md: { width: 200, height: 200, innerRadius: 60, outerRadius: 90 },
  lg: { width: 300, height: 300, innerRadius: 90, outerRadius: 135 },
} as const;

/**
 * Severity metadata for rendering
 */
const SEVERITY_CONFIG: Record<Severity, { label: string; color: string }> = {
  critical: { label: 'critical', color: colors.severity.critical },
  high: { label: 'high', color: colors.severity.high },
  medium: { label: 'medium', color: colors.severity.medium },
  low: { label: 'low', color: colors.severity.low },
};

/**
 * Get CSS variable value from computed styles
 * Required because recharts needs actual color values, not CSS variables
 */
const getCSSVariableValue = (cssVar: string): string => {
  if (typeof window === 'undefined') return '#000000';

  // Remove 'var(' and ')' if present
  const varName = cssVar.replace(/^var\(/, '').replace(/\)$/, '');
  const value = getComputedStyle(document.documentElement)
    .getPropertyValue(varName)
    .trim();

  return value || '#000000';
};

/**
 * Props for custom tooltip
 */
interface TooltipPayload {
  payload: {
    label: string;
    value: number;
  };
  value: number;
}

interface CustomTooltipProps {
  active?: boolean;
  payload?: TooltipPayload[];
  total: number;
}

/**
 * Custom tooltip renderer - DEFINED OUTSIDE COMPONENT
 */
const CustomTooltip: React.FC<CustomTooltipProps> = ({ active, payload, total }) => {
  if (!active || !payload || !payload[0]) return null;

  const item = payload[0];
  const percentage = ((item.value / total) * 100).toFixed(1);

  return (
    <div
      style={{
        background: 'var(--color-bg-elevated)',
        border: `1px solid var(--color-border-default)`,
        borderRadius: '6px',
        padding: '8px 12px',
        fontSize: '14px',
        color: 'var(--color-text-primary)',
      }}
      data-testid="severity-donut-tooltip"
      role="tooltip"
      aria-live="polite"
    >
      <div style={{ fontWeight: 600, marginBottom: '4px' }}>
        {item.payload.label}
      </div>
      <div style={{ color: 'var(--color-text-secondary)' }}>
        Count: <strong>{item.value}</strong>
      </div>
      <div style={{ color: 'var(--color-text-secondary)' }}>
        Percentage: <strong>{percentage}%</strong>
      </div>
    </div>
  );
};

/**
 * Chart data entry type
 * Index signature required for Recharts compatibility (ChartDataInput = Record<string, unknown>)
 */
interface ChartDataEntry {
  label: string;
  value: number;
  severity: Severity;
  color: string;
  [key: string]: unknown;
}

/**
 * Props for manual legend
 */
interface ManualLegendProps {
  chartData: ChartDataEntry[];
}

/**
 * Manual legend renderer - DEFINED OUTSIDE COMPONENT
 */
const ManualLegend: React.FC<ManualLegendProps> = ({ chartData }) => {
  return (
    <div
      style={{
        display: 'flex',
        justifyContent: 'center',
        gap: '16px',
        flexWrap: 'wrap',
        marginTop: '12px',
      }}
      data-testid="reviz-legend"
      role="region"
      aria-label="Severity legend"
    >
      {chartData.map((entry) => {
        return (
          <div
            key={entry.label}
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: '6px',
              fontSize: '13px',
              color: 'var(--color-text-secondary)',
            }}
            data-testid={`legend-item-${entry.label}`}
          >
            <span
              className="legend-color"
              style={{
                width: '12px',
                height: '12px',
                borderRadius: '2px',
                backgroundColor: entry.color,
                display: 'inline-block',
              }}
              aria-hidden="true"
            />
            <span className="legend-label">{entry.label}</span>
          </div>
        );
      })}
    </div>
  );
};

/**
 * SeverityDonut Component
 *
 * Renders a donut chart showing threat severity distribution.
 * Includes tooltips, optional legend, and accessibility features.
 *
 * @example
 * ```tsx
 * <SeverityDonut
 *   data={{ critical: 12, high: 45, medium: 78, low: 120 }}
 *   size="md"
 *   showLegend={true}
 *   animated={true}
 * />
 * ```
 */
export const SeverityDonut: React.FC<SeverityDonutProps> = ({
  data,
  size = 'md',
  showLegend = false,
  animated = true,
}) => {
  const sizeConfig = SIZE_MAP[size];

  // Transform data to recharts format
  const chartData = useMemo(() => {
    const entries: ChartDataEntry[] = [];

    (Object.keys(data) as Severity[]).forEach((severity) => {
      const count = data[severity];
      if (count > 0) {
        entries.push({
          label: SEVERITY_CONFIG[severity].label,
          value: count,
          severity,
          color: SEVERITY_CONFIG[severity].color,
        });
      }
    });

    return entries;
  }, [data]);

  // Calculate total for percentage display
  const total = useMemo(() => {
    return Object.values(data).reduce((sum, count) => sum + count, 0);
  }, [data]);

  // Handle empty state
  const isEmpty = total === 0;

  return (
    <div
      data-testid="severity-donut"
      data-size={size}
      style={{
        width: `${sizeConfig.width}px`,
      }}
    >
      {/* Hidden data for testing/accessibility */}
      <div
        data-testid="reviz-pie-chart"
        data-total={total}
        data-item-count={chartData.length}
        data-animated={animated}
        role="img"
        aria-label={isEmpty ? 'No severity data available' : `Severity distribution: ${chartData.map(d => `${d.label} ${d.value}`).join(', ')}`}
        style={{ position: 'relative' }}
      >
        {/* Empty state */}
        {isEmpty && (
          <div
            data-testid="empty-state"
            style={{
              width: `${sizeConfig.width}px`,
              height: `${sizeConfig.height}px`,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              border: `1px dashed var(--color-border-default)`,
              borderRadius: '8px',
              color: 'var(--color-text-muted)',
              fontSize: '14px',
            }}
          >
            No data available
          </div>
        )}

        {/* Render hidden test markers for each slice */}
        {!isEmpty && chartData.map((item) => (
          <div
            key={item.label}
            data-testid={`pie-slice-${item.label}`}
            data-severity={item.label}
            data-value={item.value}
            data-color={item.color}
            style={{ display: 'none' }}
            aria-hidden="true"
          />
        ))}

        {/* Chart rendering */}
        {!isEmpty && (
          <ResponsiveContainer width="100%" height={sizeConfig.height}>
            <PieChart>
              <Pie
                data={chartData}
                dataKey="value"
                nameKey="label"
                cx="50%"
                cy="50%"
                innerRadius={sizeConfig.innerRadius}
                outerRadius={sizeConfig.outerRadius}
                paddingAngle={2}
                animationDuration={animated ? 800 : 0}
                animationEasing="ease-out"
                label={false}
              >
                {chartData.map((entry, index) => {
                  const resolvedColor = getCSSVariableValue(entry.color);
                  return (
                    <Cell
                      key={`cell-${index}`}
                      fill={resolvedColor}
                      stroke="var(--color-bg-primary)"
                      strokeWidth={2}
                    />
                  );
                })}
              </Pie>
              <Tooltip content={<CustomTooltip total={total} />} />
            </PieChart>
          </ResponsiveContainer>
        )}
      </div>

      {/* Manual legend outside Recharts for better test compatibility */}
      {showLegend && !isEmpty && <ManualLegend chartData={chartData} />}
    </div>
  );
};

/**
 * Display name for React DevTools
 */
SeverityDonut.displayName = 'SeverityDonut';

export default SeverityDonut;
