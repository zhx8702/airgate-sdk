import type {
  AppShellTokens,
  ElevationContext,
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

/** 亮色主题 — iOS-inspired Calm Glass */
export const lightTheme: ThemeTokens = {
  primary: '#14b8a6',
  primaryHover: '#0d9488',
  primarySubtle: 'rgba(20, 184, 166, 0.08)',
  primaryGlow: 'rgba(20, 184, 166, 0.16)',

  success: '#16a34a',
  successSubtle: 'rgba(22, 163, 74, 0.07)',
  warning: '#d97706',
  warningSubtle: 'rgba(217, 119, 6, 0.07)',
  danger: '#dc2626',
  dangerSubtle: 'rgba(220, 38, 38, 0.07)',
  info: '#0284c7',
  infoSubtle: 'rgba(2, 132, 199, 0.07)',

  // 背景：青白色调，通透轻盈
  bgDeep: '#e6eff4',
  bg: '#f2f7fa',
  bgElevated: 'rgba(255, 255, 255, 0.72)',
  bgSurface: 'rgba(255, 255, 255, 0.62)',
  bgHover: 'rgba(29, 39, 52, 0.04)',
  bgActive: 'rgba(29, 39, 52, 0.07)',

  // 边框：浅灰色描边，提供清晰边界
  border: 'rgba(29, 39, 52, 0.08)',
  borderSubtle: 'rgba(29, 39, 52, 0.05)',
  borderFocus: 'rgba(20, 184, 166, 0.50)',

  // 文字：单一色源 rgb(29,39,52)，加强对比度
  text: 'rgba(29, 39, 52, 0.95)',
  textSecondary: 'rgba(29, 39, 52, 0.58)',
  textTertiary: 'rgba(29, 39, 52, 0.38)',
  textInverse: '#ffffff',

  glass: 'rgba(255, 255, 255, 0.72)',
  glassBorder: 'rgba(29, 39, 52, 0.06)',

  // 阴影：蓝灰调 rgba(85,102,122) + inset 白光（内敛风格）
  shadowSm: '0 2px 8px rgba(85, 102, 122, 0.05), inset 0 1px 0 rgba(255, 255, 255, 0.82)',
  shadowMd: '0 4px 16px rgba(85, 102, 122, 0.06), inset 0 1px 0 rgba(255, 255, 255, 0.76)',
  shadowLg: '0 18px 48px rgba(85, 102, 122, 0.08), inset 0 1px 0 rgba(255, 255, 255, 0.72)',
  shadowGlow: '0 10px 28px rgba(20, 184, 166, 0.08), inset 0 1px 0 rgba(255, 255, 255, 0.70)',
};

/** 通用基础 token */
export const foundationTokens: FoundationTokens = {
  radiusSm: '12px',
  radiusMd: '18px',
  radiusLg: '22px',
  radiusXl: '28px',
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

/**
 * 亮色主题 elevation 上下文覆盖
 * 不同 UI 层级（页面 → 弹窗 → 下拉）需要不同的背景/边框/阴影值。
 * 宿主在容器上设置 .ag-elevation-{context} class，子组件自动继承正确的 token 值。
 */
export const lightElevationContexts: Record<ElevationContext, Partial<ThemeTokens>> = {
  modal: {
    bgElevated: 'rgba(29, 39, 52, 0.024)',
    bgSurface: 'rgba(29, 39, 52, 0.04)',
    bgHover: 'rgba(29, 39, 52, 0.06)',
    glassBorder: 'rgba(29, 39, 52, 0.05)',
    border: 'rgba(29, 39, 52, 0.10)',
    shadowSm: 'none',
    shadowMd: 'none',
  },
  dropdown: {
    // dropdown 背景由宿主的 .ag-glass-dropdown 容器类处理
    // 预留空位，未来可扩展
  },
};
