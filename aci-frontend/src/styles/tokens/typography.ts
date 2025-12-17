/**
 * Typography Design Tokens
 *
 * All typography values use CSS custom properties (var(--typography-*)).
 * Never hardcode font-size, font-weight, line-height, or letter-spacing in components.
 *
 * Typography System:
 * - H1: Neue Haas Grotesk Display (Helvetica Neue fallback)
 * - H2: Roboto Bold
 * - Paragraph: Roboto Regular, color #303030 or #000000
 * - Intro: 20px, font-weight 300
 * - Overline: All caps, Roboto Regular, color #757575
 */

export interface TypographyTokens {
  readonly fontFamily: {
    readonly display: string;
    readonly heading: string;
    readonly body: string;
    readonly sans: string;
    readonly mono: string;
  };
  readonly fontSize: {
    readonly xs: string;
    readonly sm: string;
    readonly base: string;
    readonly lg: string;
    readonly xl: string;
    readonly '2xl': string;
    readonly '3xl': string;
    readonly '4xl': string;
    readonly '5xl': string;
  };
  readonly fontWeight: {
    readonly light: string;
    readonly normal: string;
    readonly medium: string;
    readonly semibold: string;
    readonly bold: string;
  };
  readonly lineHeight: {
    readonly tight: string;
    readonly normal: string;
    readonly relaxed: string;
  };
  readonly letterSpacing: {
    readonly tight: string;
    readonly normal: string;
    readonly wide: string;
    readonly caps: string;
  };
  readonly color: {
    readonly primary: string;
    readonly secondary: string;
    readonly muted: string;
  };
}

export const typography: TypographyTokens = {
  fontFamily: {
    display: 'var(--typography-font-family-display)',
    heading: 'var(--typography-font-family-heading)',
    body: 'var(--typography-font-family-body)',
    sans: 'var(--typography-font-family-sans)',
    mono: 'var(--typography-font-family-mono)',
  },
  fontSize: {
    xs: 'var(--typography-font-size-xs)',
    sm: 'var(--typography-font-size-sm)',
    base: 'var(--typography-font-size-base)',
    lg: 'var(--typography-font-size-lg)',
    xl: 'var(--typography-font-size-xl)',
    '2xl': 'var(--typography-font-size-2xl)',
    '3xl': 'var(--typography-font-size-3xl)',
    '4xl': 'var(--typography-font-size-4xl)',
    '5xl': 'var(--typography-font-size-5xl)',
  },
  fontWeight: {
    light: 'var(--typography-font-weight-light)',
    normal: 'var(--typography-font-weight-normal)',
    medium: 'var(--typography-font-weight-medium)',
    semibold: 'var(--typography-font-weight-semibold)',
    bold: 'var(--typography-font-weight-bold)',
  },
  lineHeight: {
    tight: 'var(--typography-line-height-tight)',
    normal: 'var(--typography-line-height-normal)',
    relaxed: 'var(--typography-line-height-relaxed)',
  },
  letterSpacing: {
    tight: 'var(--typography-letter-spacing-tight)',
    normal: 'var(--typography-letter-spacing-normal)',
    wide: 'var(--typography-letter-spacing-wide)',
    caps: 'var(--typography-letter-spacing-caps)',
  },
  color: {
    primary: 'var(--typography-color-primary)',
    secondary: 'var(--typography-color-secondary)',
    muted: 'var(--typography-color-muted)',
  },
} as const;

/**
 * Semantic Typography Styles
 * Use these for consistent heading, paragraph, and text styling.
 */
export interface SemanticTypographyStyle {
  readonly fontFamily: string;
  readonly fontSize: string;
  readonly fontWeight: string;
  readonly lineHeight: string;
  readonly letterSpacing: string;
  readonly color: string;
  readonly textTransform?: string;
}

export interface SemanticTypography {
  readonly h1: SemanticTypographyStyle;
  readonly h2: SemanticTypographyStyle;
  readonly h3: SemanticTypographyStyle;
  readonly paragraph: SemanticTypographyStyle;
  readonly intro: SemanticTypographyStyle;
  readonly overline: SemanticTypographyStyle & { readonly textTransform: string };
}

export const semanticTypography: SemanticTypography = {
  /** H1 - Display heading: Neue Haas Grotesk Display */
  h1: {
    fontFamily: 'var(--typography-h1-font-family)',
    fontSize: 'var(--typography-h1-font-size)',
    fontWeight: 'var(--typography-h1-font-weight)',
    lineHeight: 'var(--typography-h1-line-height)',
    letterSpacing: 'var(--typography-h1-letter-spacing)',
    color: 'var(--typography-h1-color)',
  },
  /** H2 - Section heading: Roboto Bold */
  h2: {
    fontFamily: 'var(--typography-h2-font-family)',
    fontSize: 'var(--typography-h2-font-size)',
    fontWeight: 'var(--typography-h2-font-weight)',
    lineHeight: 'var(--typography-h2-line-height)',
    letterSpacing: 'var(--typography-h2-letter-spacing)',
    color: 'var(--typography-h2-color)',
  },
  /** H3 - Subsection heading: Roboto Bold */
  h3: {
    fontFamily: 'var(--typography-h3-font-family)',
    fontSize: 'var(--typography-h3-font-size)',
    fontWeight: 'var(--typography-h3-font-weight)',
    lineHeight: 'var(--typography-h3-line-height)',
    letterSpacing: 'var(--typography-h3-letter-spacing)',
    color: 'var(--typography-h3-color)',
  },
  /** Paragraph - Body text: Roboto Regular, #303030 */
  paragraph: {
    fontFamily: 'var(--typography-paragraph-font-family)',
    fontSize: 'var(--typography-paragraph-font-size)',
    fontWeight: 'var(--typography-paragraph-font-weight)',
    lineHeight: 'var(--typography-paragraph-line-height)',
    letterSpacing: 'var(--typography-paragraph-letter-spacing)',
    color: 'var(--typography-paragraph-color)',
  },
  /** Intro - Lead paragraph after title: 20px, font-weight 300 */
  intro: {
    fontFamily: 'var(--typography-intro-font-family)',
    fontSize: 'var(--typography-intro-font-size)',
    fontWeight: 'var(--typography-intro-font-weight)',
    lineHeight: 'var(--typography-intro-line-height)',
    letterSpacing: 'var(--typography-intro-letter-spacing)',
    color: 'var(--typography-intro-color)',
  },
  /** Overline - Small title above H1/H2: All caps, Roboto Regular, #757575 */
  overline: {
    fontFamily: 'var(--typography-overline-font-family)',
    fontSize: 'var(--typography-overline-font-size)',
    fontWeight: 'var(--typography-overline-font-weight)',
    lineHeight: 'var(--typography-overline-line-height)',
    letterSpacing: 'var(--typography-overline-letter-spacing)',
    textTransform: 'var(--typography-overline-text-transform)',
    color: 'var(--typography-overline-color)',
  },
} as const;

// Type exports for consumer convenience
export type FontFamily = keyof typeof typography.fontFamily;
export type FontSize = keyof typeof typography.fontSize;
export type FontWeight = keyof typeof typography.fontWeight;
export type LineHeight = keyof typeof typography.lineHeight;
export type LetterSpacing = keyof typeof typography.letterSpacing;
export type SemanticStyle = keyof typeof semanticTypography;
