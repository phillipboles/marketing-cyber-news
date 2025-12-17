/**
 * RecommendationsList Component
 * Displays actionable recommendations grouped by priority
 *
 * Features:
 * - Groups recommendations by priority (immediate, short_term, long_term)
 * - Shows priority badge, title, description, category icon
 * - Collapsible groups
 * - Empty state handling
 *
 * Used in: ThreatDetail
 */

import React, { useState } from 'react';
import {
  AlertCircle,
  Clock,
  Calendar,
  Shield,
  Settings,
  Eye,
  Lock,
  Database,
  Users,
  CheckCircle2,
  ChevronDown,
  ChevronUp,
} from 'lucide-react';
import { cn } from '@/lib/utils';
import type { Recommendation, RecommendationPriority } from '@/types/threat';
import { Badge } from '@/components/ui/badge';
import { colors } from '@/styles/tokens/colors';
import { spacing } from '@/styles/tokens/spacing';
import { typography } from '@/styles/tokens/typography';

export interface RecommendationsListProps {
  /**
   * Array of recommendations to display
   */
  readonly recommendations: readonly Recommendation[];
  /**
   * Additional CSS classes
   */
  readonly className?: string;
}

/**
 * Priority configuration
 */
const PRIORITY_CONFIG: Record<
  RecommendationPriority,
  {
    label: string;
    icon: React.ComponentType<{ size?: number; 'aria-hidden'?: boolean }>;
    variant: 'critical' | 'warning' | 'info';
    description: string;
  }
> = {
  immediate: {
    label: 'Immediate Action',
    icon: AlertCircle,
    variant: 'critical',
    description: 'Take action now to prevent exploitation',
  },
  short_term: {
    label: 'Short Term',
    icon: Clock,
    variant: 'warning',
    description: 'Address within the next few days',
  },
  long_term: {
    label: 'Long Term',
    icon: Calendar,
    variant: 'info',
    description: 'Strategic improvements for future resilience',
  },
};

/**
 * Category icon mapping
 */
const CATEGORY_ICONS: Record<
  Recommendation['category'],
  React.ComponentType<{ size?: number; 'aria-hidden'?: boolean }>
> = {
  patch: Shield,
  configuration: Settings,
  monitoring: Eye,
  access_control: Lock,
  backup: Database,
  training: Users,
  other: CheckCircle2,
};

/**
 * Groups recommendations by priority
 */
function groupByPriority(
  recommendations: readonly Recommendation[]
): Record<RecommendationPriority, readonly Recommendation[]> {
  const groups: Record<RecommendationPriority, Recommendation[]> = {
    immediate: [],
    short_term: [],
    long_term: [],
  };

  for (const rec of recommendations) {
    groups[rec.priority].push(rec);
  }

  return groups;
}

/**
 * Priority group component
 */
interface PriorityGroupProps {
  readonly priority: RecommendationPriority;
  readonly recommendations: readonly Recommendation[];
  readonly isDefaultOpen?: boolean;
}

function PriorityGroup({
  priority,
  recommendations,
  isDefaultOpen = false,
}: PriorityGroupProps): React.JSX.Element | null {
  const [isOpen, setIsOpen] = useState<boolean>(isDefaultOpen);

  if (recommendations.length === 0) {
    return null;
  }

  const config = PRIORITY_CONFIG[priority];
  const Icon = config.icon;

  const handleToggle = (): void => {
    setIsOpen((prev) => !prev);
  };

  return (
    <div
      data-testid={`priority-group-${priority}`}
      style={{
        display: 'flex',
        flexDirection: 'column',
        gap: spacing[3],
        padding: spacing[4],
        backgroundColor: colors.background.elevated,
        borderRadius: 'var(--border-radius-lg)',
        border: `1px solid ${colors.border.default}`,
      }}
    >
      {/* Group Header */}
      <button
        type="button"
        onClick={handleToggle}
        aria-expanded={isOpen}
        aria-controls={`recommendations-${priority}`}
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          width: '100%',
          padding: 0,
          backgroundColor: 'transparent',
          border: 'none',
          cursor: 'pointer',
          textAlign: 'left',
        }}
      >
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: spacing[3],
          }}
        >
          <div
            style={{
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              width: '32px',
              height: '32px',
              borderRadius: 'var(--border-radius-sm)',
              backgroundColor: colors.background.secondary,
              color: colors.text.secondary,
            }}
          >
            <Icon size={16} aria-hidden={true} />
          </div>

          <div
            style={{
              display: 'flex',
              flexDirection: 'column',
              gap: spacing[1],
            }}
          >
            <div
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: spacing[2],
              }}
            >
              <h3
                style={{
                  fontSize: typography.fontSize.base,
                  fontWeight: typography.fontWeight.semibold,
                  color: colors.text.primary,
                  margin: 0,
                }}
              >
                {config.label}
              </h3>
              <Badge variant={config.variant}>{recommendations.length}</Badge>
            </div>

            <p
              style={{
                fontSize: typography.fontSize.sm,
                color: colors.text.muted,
                margin: 0,
              }}
            >
              {config.description}
            </p>
          </div>
        </div>

        <span style={{ color: colors.text.secondary }}>
          {isOpen ? (
            <ChevronUp size={20} aria-hidden={true} />
          ) : (
            <ChevronDown size={20} aria-hidden={true} />
          )}
        </span>
      </button>

      {/* Group Content */}
      {isOpen && (
        <ul
          id={`recommendations-${priority}`}
          role="list"
          style={{
            display: 'flex',
            flexDirection: 'column',
            gap: spacing[3],
            listStyle: 'none',
            padding: 0,
            margin: 0,
          }}
        >
          {recommendations.map((recommendation) => {
            const CategoryIcon = CATEGORY_ICONS[recommendation.category];

            return (
              <li
                key={recommendation.id}
                role="listitem"
                data-testid="recommendation-item"
                style={{
                  display: 'flex',
                  gap: spacing[3],
                  padding: spacing[3],
                  backgroundColor: colors.background.primary,
                  borderRadius: 'var(--border-radius-md)',
                  border: `1px solid ${colors.border.default}`,
                }}
              >
                {/* Category Icon */}
                <div
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    width: '24px',
                    height: '24px',
                    borderRadius: 'var(--border-radius-sm)',
                    backgroundColor: colors.background.elevated,
                    flexShrink: 0,
                    color: colors.text.muted,
                  }}
                >
                  <CategoryIcon size={12} aria-hidden={true} />
                </div>

                {/* Content */}
                <div
                  style={{
                    display: 'flex',
                    flexDirection: 'column',
                    gap: spacing[2],
                    flex: 1,
                    minWidth: 0,
                  }}
                >
                  <h4
                    style={{
                      fontSize: typography.fontSize.base,
                      fontWeight: typography.fontWeight.medium,
                      color: colors.text.primary,
                      margin: 0,
                    }}
                  >
                    {recommendation.title}
                  </h4>

                  <p
                    style={{
                      fontSize: typography.fontSize.sm,
                      lineHeight: typography.lineHeight.relaxed,
                      color: colors.text.secondary,
                      margin: 0,
                    }}
                  >
                    {recommendation.description}
                  </p>

                  {/* Category Badge */}
                  <Badge variant="outline" style={{ alignSelf: 'flex-start' }}>
                    {recommendation.category.replace(/_/g, ' ')}
                  </Badge>
                </div>
              </li>
            );
          })}
        </ul>
      )}
    </div>
  );
}

/**
 * RecommendationsList - Grouped list of actionable recommendations
 *
 * Features:
 * - Groups by priority (immediate, short_term, long_term)
 * - Collapsible groups
 * - Category icons
 * - Priority badges
 * - Empty state
 *
 * @example
 * ```tsx
 * const recommendations: Recommendation[] = [
 *   {
 *     id: '1',
 *     title: 'Apply Security Patches',
 *     description: 'Update to the latest version immediately',
 *     priority: 'immediate',
 *     category: 'patch',
 *   },
 * ];
 *
 * <RecommendationsList recommendations={recommendations} />
 * ```
 */
export function RecommendationsList({
  recommendations,
  className,
}: RecommendationsListProps): React.JSX.Element {
  // Guard: Empty state
  if (recommendations.length === 0) {
    return (
      <div
        data-testid="recommendations-empty"
        style={{
          padding: spacing[6],
          textAlign: 'center',
          color: colors.text.muted,
          fontSize: typography.fontSize.sm,
        }}
        className={className}
      >
        No recommendations available.
      </div>
    );
  }

  const grouped = groupByPriority(recommendations);
  const priorityOrder: RecommendationPriority[] = ['immediate', 'short_term', 'long_term'];

  return (
    <div
      data-testid="recommendations-list"
      aria-label="Security recommendations"
      className={cn('flex flex-col', className)}
      style={{
        gap: spacing[4],
      }}
    >
      {priorityOrder.map((priority) => (
        <PriorityGroup
          key={priority}
          priority={priority}
          recommendations={grouped[priority]}
          isDefaultOpen={priority === 'immediate'}
        />
      ))}
    </div>
  );
}

/**
 * Accessibility Notes:
 * - Semantic <ul> and <li> elements
 * - ARIA labels for lists and items
 * - aria-expanded on collapse buttons
 * - aria-controls links buttons to content
 * - Proper heading hierarchy (h3, h4)
 * - All icons hidden from screen readers
 * - Keyboard accessible (native button elements)
 *
 * Performance Notes:
 * - Efficient grouping (single pass)
 * - Only renders expanded groups
 * - Minimal re-renders (local state per group)
 * - Suitable for large recommendation lists (50+)
 *
 * Design Token Usage:
 * - Colors: colors.text.*, colors.background.*, colors.border.*
 * - Spacing: spacing[1-6]
 * - Typography: typography.fontSize.*, typography.fontWeight.*, typography.lineHeight.*
 *
 * Testing:
 * - data-testid="recommendations-list" for container
 * - data-testid="recommendations-empty" for empty state
 * - data-testid="priority-group-{priority}" for each group
 * - data-testid="recommendation-item" for individual items
 */
