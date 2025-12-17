/**
 * ThreatTimeline Chart Component
 *
 * Line/area chart showing threat count over time with optional severity breakdown.
 * Uses Reviz (reaviz) library with design tokens for colors and motion.
 *
 * Design Requirements:
 * - NO HARDCODED COLORS: All colors from @/styles/tokens/colors
 * - NO HARDCODED ANIMATIONS: All motion from @/styles/tokens/motion
 * - Responsive width (100% of container)
 * - Interactive tooltips with date and counts
 * - Grid lines for readability
 * - Smooth curve interpolation
 * - Accessibility: ARIA labels, keyboard navigation, data-testid attributes
 */

import React from 'react';
import { LineChart, AreaChart, LinearXAxis, LinearYAxis, GridlineSeries, Area, StackedAreaSeries, LineSeries, Line, TooltipArea, ChartTooltip } from 'reaviz';
import { colors } from '@/styles/tokens/colors';
import type { ChartDataShape } from 'reaviz';

/**
 * Data point representing threat count for a specific date
 */
export interface TimelineDataPoint {
  date: string; // ISO date string (YYYY-MM-DD)
  count: number; // Total threat count for that day
  critical?: number; // Optional severity breakdown
  high?: number;
  medium?: number;
  low?: number;
}

/**
 * Props for ThreatTimeline component
 */
export interface ThreatTimelineProps {
  /** Array of timeline data points */
  data: TimelineDataPoint[];
  /** Date range filter - defaults to 7 days */
  dateRange?: '7d' | '30d' | '90d';
  /** Show stacked area chart by severity instead of simple line */
  showBreakdown?: boolean;
  /** Chart height in pixels - defaults to 300 */
  height?: number;
  /** Enable/disable animations - defaults to true */
  animated?: boolean;
}

/**
 * Format date string to readable format (e.g., "Mon 12")
 */
function formatDate(dateString: string): string {
  try {
    const date = new Date(dateString);
    const weekday = date.toLocaleDateString('en-US', { weekday: 'short' });
    const day = date.getDate();
    return `${weekday} ${day}`;
  } catch {
    return dateString;
  }
}

/**
 * Transform data for Reviz line chart format (single series)
 */
function transformDataForLineChart(data: TimelineDataPoint[]): ChartDataShape[] {
  if (!data || data.length === 0) {
    return [];
  }

  return data.map((point) => ({
    key: formatDate(point.date),
    data: point.count,
    metadata: {
      date: point.date,
      critical: point.critical || 0,
      high: point.high || 0,
      medium: point.medium || 0,
      low: point.low || 0,
    },
  }));
}

/**
 * Transform data for Reviz stacked area chart format (multi-series)
 */
function transformDataForAreaChart(data: TimelineDataPoint[]): ChartDataShape[] {
  if (!data || data.length === 0) {
    return [];
  }

  return data.map((point) => ({
    key: formatDate(point.date),
    data: [
      { key: 'Critical', data: point.critical || 0 },
      { key: 'High', data: point.high || 0 },
      { key: 'Medium', data: point.medium || 0 },
      { key: 'Low', data: point.low || 0 },
    ],
    metadata: { date: point.date },
  }));
}

/**
 * Tooltip data structure
 */
interface TooltipData {
  key?: string;
  data?: number | Array<{ key: string; data: number }>;
  metadata?: {
    date?: string;
    critical?: number;
    high?: number;
    medium?: number;
    low?: number;
  };
}

/**
 * Custom tooltip component with severity breakdown
 */
const CustomTooltipContent: React.FC<{ data: TooltipData }> = ({ data }) => {
  const metadata = data?.metadata || {};
  const isLineChart = typeof data?.data === 'number';

  return (
    <div
      style={{
        background: colors.background.elevated,
        border: `1px solid ${colors.border.default}`,
        borderRadius: '4px',
        padding: '8px 12px',
        color: colors.text.primary,
        fontSize: '14px',
        minWidth: '120px',
      }}
      role="tooltip"
    >
      <div style={{ fontWeight: 600, marginBottom: '4px' }}>{data?.key || 'Unknown'}</div>
      {isLineChart && typeof data.data === 'number' && (
        <>
          <div style={{ color: colors.text.secondary }}>
            Total: <strong>{data.data as number}</strong>
          </div>
          {(metadata.critical ?? 0) > 0 && (
            <div style={{ color: colors.severity.critical, fontSize: '12px' }}>
              Critical: {metadata.critical}
            </div>
          )}
          {(metadata.high ?? 0) > 0 && (
            <div style={{ color: colors.severity.high, fontSize: '12px' }}>
              High: {metadata.high}
            </div>
          )}
          {(metadata.medium ?? 0) > 0 && (
            <div style={{ color: colors.severity.medium, fontSize: '12px' }}>
              Medium: {metadata.medium}
            </div>
          )}
          {(metadata.low ?? 0) > 0 && (
            <div style={{ color: colors.severity.low, fontSize: '12px' }}>
              Low: {metadata.low}
            </div>
          )}
        </>
      )}
      {!isLineChart && Array.isArray(data?.data) && (
        <>
          {data.data.map((item: { key: string; data: number }) => (
            <div
              key={item.key}
              style={{
                color: colors.text.secondary,
                fontSize: '12px',
                marginTop: '2px',
              }}
            >
              {item.key}: <strong>{item.data}</strong>
            </div>
          ))}
        </>
      )}
    </div>
  );
};

/**
 * Empty state component when no data available
 */
const EmptyState: React.FC<{ height: number }> = ({ height }) => (
  <div
    data-testid="empty-state"
    style={{
      height: `${height}px`,
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      color: colors.text.muted,
      fontSize: '14px',
    }}
    role="status"
    aria-label="No threat data available"
  >
    No data available
  </div>
);

/**
 * ThreatTimeline Component
 *
 * Displays threat count over time as either:
 * - Line chart (default): Single line showing total count
 * - Stacked area chart (showBreakdown=true): Areas for each severity level
 *
 * @example
 * ```tsx
 * <ThreatTimeline
 *   data={[
 *     { date: '2024-01-01', count: 10, critical: 2, high: 3, medium: 3, low: 2 },
 *     { date: '2024-01-02', count: 15, critical: 3, high: 4, medium: 5, low: 3 },
 *   ]}
 *   showBreakdown={true}
 * />
 * ```
 */
export const ThreatTimeline: React.FC<ThreatTimelineProps> = ({
  data,
  dateRange = '7d',
  showBreakdown = false,
  height = 300,
  animated = true,
}) => {
  // Handle empty data
  if (!data || data.length === 0) {
    return <EmptyState height={height} />;
  }

  const chartData = showBreakdown
    ? transformDataForAreaChart(data)
    : transformDataForLineChart(data);

  // Chart dimensions
  const width = 800; // Will be made responsive by container

  return (
    <div
      data-testid="threat-timeline"
      data-date-range={dateRange}
      data-show-breakdown={showBreakdown}
      data-animated={animated}
      role="img"
      aria-label={`Threat timeline chart showing ${data.length} days of data`}
      style={{ width: '100%', height: `${height}px` }}
    >
      {showBreakdown ? (
        // Stacked area chart with severity breakdown
        <AreaChart
          width={width}
          height={height}
          data={chartData}
          xAxis={
            <LinearXAxis
              type="category"
            />
          }
          yAxis={
            <LinearYAxis
              type="value"
            />
          }
          gridlines={<GridlineSeries />}
          series={
            <StackedAreaSeries
              animated={animated}
              colorScheme={(_data: ChartDataShape, index: number) => {
                // Map severity levels to colors
                const severityColors = [
                  colors.severity.critical,
                  colors.severity.high,
                  colors.severity.medium,
                  colors.severity.low,
                ];
                return severityColors[index] || colors.brand.primary;
              }}
              tooltip={
                <TooltipArea
                  tooltip={
                    <ChartTooltip
                      content={(point: ChartDataShape) => <CustomTooltipContent data={point as unknown as TooltipData} />}
                    />
                  }
                />
              }
              area={
                <Area
                  gradient={<></>}
                  style={() => ({
                    fillOpacity: 0.6,
                  })}
                />
              }
            />
          }
        />
      ) : (
        // Simple line chart for total count
        <LineChart
          width={width}
          height={height}
          data={chartData}
          xAxis={
            <LinearXAxis
              type="category"
            />
          }
          yAxis={
            <LinearYAxis
              type="value"
            />
          }
          gridlines={<GridlineSeries />}
          series={
            <LineSeries
              animated={animated}
              colorScheme={colors.brand.primary}
              tooltip={
                <TooltipArea
                  tooltip={
                    <ChartTooltip
                      content={(point: ChartDataShape) => <CustomTooltipContent data={point as unknown as TooltipData} />}
                    />
                  }
                />
              }
              line={
                <Line
                  strokeWidth={2}
                  style={{
                    stroke: colors.brand.primary,
                  }}
                />
              }
            />
          }
        />
      )}
    </div>
  );
};

export default ThreatTimeline;
