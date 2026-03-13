import type {
  AppShellTokens,
  FoundationTokens,
  StaticTokenGroups,
  StaticTokens,
  ThemeName,
  ThemeTokens,
} from './types.js';

/** 暗色主题 — Obsidian Terminal */
export const darkTheme: ThemeTokens = {
  primary: '#00d4aa',
  primaryHover: '#00e6b8',
  primarySubtle: 'rgba(0, 212, 170, 0.12)',
  primaryGlow: 'rgba(0, 212, 170, 0.20)',

  success: '#22c55e',
  successSubtle: 'rgba(34, 197, 94, 0.14)',
  warning: '#f59e0b',
  warningSubtle: 'rgba(245, 158, 11, 0.14)',
  danger: '#ef4444',
  dangerSubtle: 'rgba(239, 68, 68, 0.14)',
  info: '#a78bfa',
  infoSubtle: 'rgba(167, 139, 250, 0.14)',

  bgDeep: '#09090b',
  bg: '#0f0f12',
  bgElevated: '#16161a',
  bgSurface: '#1c1c21',
  bgHover: '#25252b',
  bgActive: '#2e2e36',

  border: 'rgba(255, 255, 255, 0.08)',
  borderSubtle: 'rgba(255, 255, 255, 0.05)',
  borderFocus: 'rgba(0, 212, 170, 0.40)',

  text: '#ececf0',
  textSecondary: '#a1a1aa',
  textTertiary: '#63636e',
  textInverse: '#09090b',

  glass: 'rgba(255, 255, 255, 0.03)',
  glassBorder: 'rgba(255, 255, 255, 0.07)',

  shadowSm: '0 2px 8px rgba(0, 0, 0, 0.32)',
  shadowMd: '0 8px 24px rgba(0, 0, 0, 0.44)',
  shadowLg: '0 20px 48px rgba(0, 0, 0, 0.56)',
  shadowGlow: '0 0 0 1px rgba(0, 212, 170, 0.08), 0 8px 32px rgba(0, 212, 170, 0.12)',
};

/** 亮色主题 — Warm Paper */
export const lightTheme: ThemeTokens = {
  primary: '#0d9373',
  primaryHover: '#0b7d62',
  primarySubtle: 'rgba(13, 147, 115, 0.10)',
  primaryGlow: 'rgba(13, 147, 115, 0.14)',

  success: '#16a34a',
  successSubtle: 'rgba(22, 163, 74, 0.10)',
  warning: '#d97706',
  warningSubtle: 'rgba(217, 119, 6, 0.10)',
  danger: '#dc2626',
  dangerSubtle: 'rgba(220, 38, 38, 0.10)',
  info: '#7c3aed',
  infoSubtle: 'rgba(124, 58, 237, 0.10)',

  bgDeep: '#f3f2ef',
  bg: '#fafaf8',
  bgElevated: '#ffffff',
  bgSurface: '#f5f4f1',
  bgHover: '#eeedea',
  bgActive: '#e5e4e0',

  border: 'rgba(28, 25, 23, 0.08)',
  borderSubtle: 'rgba(28, 25, 23, 0.05)',
  borderFocus: 'rgba(13, 147, 115, 0.36)',

  text: '#1c1917',
  textSecondary: '#57534e',
  textTertiary: '#a8a29e',
  textInverse: '#fafaf8',

  glass: 'rgba(255, 255, 255, 0.72)',
  glassBorder: 'rgba(28, 25, 23, 0.08)',

  shadowSm: '0 1px 3px rgba(28, 25, 23, 0.06)',
  shadowMd: '0 4px 16px rgba(28, 25, 23, 0.08)',
  shadowLg: '0 12px 40px rgba(28, 25, 23, 0.12)',
  shadowGlow: '0 0 0 1px rgba(13, 147, 115, 0.06), 0 4px 20px rgba(13, 147, 115, 0.08)',
};

/** 通用基础 token */
export const foundationTokens: FoundationTokens = {
  radiusSm: '6px',
  radiusMd: '10px',
  radiusLg: '14px',
  radiusXl: '20px',
  fontSans: "'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif",
  fontMono: "'JetBrains Mono', 'SF Mono', 'Cascadia Code', monospace",
  transition: '200ms cubic-bezier(0.4, 0, 0.2, 1)',
  transitionSlow: '400ms cubic-bezier(0.4, 0, 0.2, 1)',
};

/** 应用壳层 token */
export const appShellTokens: AppShellTokens = {
  sidebarWidth: '260px',
  sidebarCollapsed: '72px',
  topbarHeight: '64px',
};

/** 分组后的静态 token */
export const staticTokenGroups: StaticTokenGroups = {
  foundation: foundationTokens,
  appShell: appShellTokens,
};

/** 不随主题变化的静态 token（向后兼容的扁平导出） */
export const staticTokens: StaticTokens = {
  ...foundationTokens,
  ...appShellTokens,
};

/** 主题集合 */
export const themes: Record<ThemeName, ThemeTokens> = {
  dark: darkTheme,
  light: lightTheme,
};
