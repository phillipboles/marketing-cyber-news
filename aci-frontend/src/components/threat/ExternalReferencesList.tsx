/**
 * ExternalReferencesList Component
 * Displays external references with collapsible functionality
 *
 * Features:
 * - Source icon/badge, title as link, reference type badge
 * - Collapsible if > 5 references
 * - Empty state handling
 *
 * Used in: ThreatDetail
 */

import React, { useState } from 'react';
import { ExternalLink, ChevronDown, ChevronUp, FileText, Shield, AlertTriangle } from 'lucide-react';
import { cn } from '@/lib/utils';
import type { ExternalReference } from '@/types/threat';
import { Badge } from '@/components/ui/badge';
import { colors } from '@/styles/tokens/colors';
import { spacing } from '@/styles/tokens/spacing';
import { typography } from '@/styles/tokens/typography';

export interface ExternalReferencesListProps {
  /**
   * Array of external references to display
   */
  readonly references: readonly ExternalReference[];
  /**
   * Maximum number of references to show before collapsing
   * @default 5
   */
  readonly maxVisible?: number;
  /**
   * Additional CSS classes
   */
  readonly className?: string;
}

const DEFAULT_MAX_VISIBLE = 5;

/**
 * Maps reference types to badge variants
 */
const REFERENCE_TYPE_CONFIG: Record<
  ExternalReference['type'],
  { variant: 'default' | 'secondary' | 'critical' | 'warning' | 'success' | 'info'; label: string }
> = {
  advisory: { variant: 'warning', label: 'Advisory' },
  article: { variant: 'default', label: 'Article' },
  report: { variant: 'info', label: 'Report' },
  cve: { variant: 'critical', label: 'CVE' },
  mitre: { variant: 'secondary', label: 'MITRE' },
  other: { variant: 'default', label: 'Reference' },
};

/**
 * Gets icon component for reference source
 */
function getSourceIcon(source: string): React.ComponentType<{ size?: number; 'aria-hidden'?: boolean }> {
  const lowerSource = source.toLowerCase();

  if (lowerSource.includes('mitre') || lowerSource.includes('att&ck')) {
    return Shield;
  }

  if (lowerSource.includes('cisa') || lowerSource.includes('advisory')) {
    return AlertTriangle;
  }

  return FileText;
}

/**
 * ExternalReferencesList - Collapsible list of external reference links
 *
 * @example
 * ```tsx
 * const references: ExternalReference[] = [
 *   {
 *     title: 'CVE-2024-1234 Detail',
 *     url: 'https://nvd.nist.gov/vuln/detail/CVE-2024-1234',
 *     source: 'NVD',
 *     type: 'cve',
 *   },
 * ];
 *
 * <ExternalReferencesList references={references} />
 * ```
 */
export function ExternalReferencesList({
  references,
  maxVisible = DEFAULT_MAX_VISIBLE,
  className,
}: ExternalReferencesListProps): React.JSX.Element {
  const [isExpanded, setIsExpanded] = useState<boolean>(false);

  // Guard: Empty state
  if (references.length === 0) {
    return (
      <div
        data-testid="external-references-empty"
        style={{
          padding: spacing[6],
          textAlign: 'center',
          color: colors.text.muted,
          fontSize: typography.fontSize.sm,
        }}
        className={className}
      >
        No external references available.
      </div>
    );
  }

  const effectiveMaxVisible = Math.max(1, maxVisible);
  const shouldCollapse = references.length > effectiveMaxVisible;
  const displayedReferences = isExpanded || !shouldCollapse ? references : references.slice(0, effectiveMaxVisible);
  const hiddenCount = Math.max(0, references.length - effectiveMaxVisible);

  const handleToggle = (): void => {
    setIsExpanded((prev) => !prev);
  };

  return (
    <div
      data-testid="external-references-list"
      aria-label="External references and sources"
      className={cn('flex flex-col', className)}
      style={{
        gap: spacing[4],
      }}
    >
      {/* Reference List */}
      <ul
        role="list"
        aria-label="External reference links"
        style={{
          display: 'flex',
          flexDirection: 'column',
          gap: spacing[3],
          listStyle: 'none',
          padding: 0,
          margin: 0,
        }}
      >
        {displayedReferences.map((reference, index) => {
          const Icon = getSourceIcon(reference.source);
          const typeConfig = REFERENCE_TYPE_CONFIG[reference.type];

          return (
            <li
              key={`${reference.url}-${index}`}
              role="listitem"
              style={{
                display: 'flex',
                alignItems: 'flex-start',
                gap: spacing[3],
                padding: spacing[3],
                backgroundColor: colors.background.elevated,
                borderRadius: 'var(--border-radius-md)',
                border: `1px solid ${colors.border.default}`,
              }}
            >
              {/* Source Icon */}
              <div
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  width: '32px',
                  height: '32px',
                  borderRadius: 'var(--border-radius-sm)',
                  backgroundColor: colors.background.secondary,
                  flexShrink: 0,
                  color: colors.text.secondary,
                }}
              >
                <Icon size={16} aria-hidden={true} />
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
                {/* Title as Link */}
                <a
                  href={reference.url}
                  target="_blank"
                  rel="noopener noreferrer"
                  aria-label={`${reference.title} from ${reference.source} (opens in new tab)`}
                  style={{
                    display: 'inline-flex',
                    alignItems: 'center',
                    gap: spacing[2],
                    fontSize: typography.fontSize.base,
                    fontWeight: typography.fontWeight.medium,
                    color: colors.text.primary,
                    textDecoration: 'none',
                    transition: `color var(--motion-duration-fast) var(--motion-easing-default)`,
                  }}
                  className="hover:text-primary"
                >
                  <span style={{ overflow: 'hidden', textOverflow: 'ellipsis' }}>
                    {reference.title}
                  </span>
                  <ExternalLink size={14} aria-hidden={true} style={{ flexShrink: 0 }} />
                </a>

                {/* Source and Type Badges */}
                <div
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: spacing[2],
                    flexWrap: 'wrap',
                  }}
                >
                  {/* Source Badge */}
                  <span
                    style={{
                      fontSize: typography.fontSize.xs,
                      color: colors.text.muted,
                      fontWeight: typography.fontWeight.medium,
                    }}
                  >
                    {reference.source}
                  </span>

                  {/* Type Badge */}
                  <Badge variant={typeConfig.variant} data-testid="reference-type-badge">
                    {typeConfig.label}
                  </Badge>
                </div>
              </div>
            </li>
          );
        })}
      </ul>

      {/* Show More/Less Button */}
      {shouldCollapse && (
        <button
          data-testid="external-references-toggle-button"
          type="button"
          onClick={handleToggle}
          aria-expanded={isExpanded}
          aria-controls="external-references-list"
          aria-label={isExpanded ? 'Show fewer references' : `Show ${hiddenCount} more references`}
          style={{
            display: 'inline-flex',
            alignItems: 'center',
            alignSelf: 'flex-start',
            gap: spacing[2],
            padding: `${spacing[2]} ${spacing[3]}`,
            fontSize: typography.fontSize.sm,
            fontWeight: typography.fontWeight.medium,
            color: colors.text.secondary,
            backgroundColor: 'transparent',
            border: 'none',
            cursor: 'pointer',
            transition: `color var(--motion-duration-fast) var(--motion-easing-default)`,
          }}
          className="hover:text-primary"
        >
          {isExpanded ? (
            <>
              <ChevronUp size={16} aria-hidden={true} />
              <span>Show Less</span>
            </>
          ) : (
            <>
              <ChevronDown size={16} aria-hidden={true} />
              <span>Show {hiddenCount} more</span>
            </>
          )}
        </button>
      )}
    </div>
  );
}

/**
 * Accessibility Notes:
 * - Semantic <ul> and <li> elements
 * - ARIA labels for list and individual items
 * - aria-expanded attribute on toggle button
 * - aria-controls links button to content
 * - External links have descriptive aria-labels
 * - All icons hidden from screen readers
 * - Keyboard accessible (native elements)
 *
 * Performance Notes:
 * - Only renders visible references
 * - Efficient array slicing
 * - Minimal re-renders (useState for expansion only)
 * - Suitable for large reference lists (50+)
 *
 * Design Token Usage:
 * - Colors: colors.text.*, colors.background.*, colors.border.*
 * - Spacing: spacing[2-6]
 * - Typography: typography.fontSize.*, typography.fontWeight.*
 *
 * Testing:
 * - data-testid="external-references-list" for container
 * - data-testid="external-references-empty" for empty state
 * - data-testid="external-references-toggle-button" for expand/collapse
 * - data-testid="reference-type-badge" for type badges
 */
