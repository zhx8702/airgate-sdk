// 公共工具函数

export const API = '';

export const typeColors = [
  { bg: 'var(--accent-light)', fg: 'var(--accent)' },
  { bg: 'var(--teal-light)', fg: 'var(--teal)' },
  { bg: 'var(--amber-light)', fg: 'var(--amber)' },
  { bg: 'var(--violet-light)', fg: 'var(--violet)' },
];

/** 全局插件信息（init 后赋值） */
export let pluginInfo = null;

export function setPluginInfo(info) {
  pluginInfo = info;
}

export function getBadgeStyle(typeKey) {
  if (!pluginInfo?.account_types) return typeColors[0];
  const idx = pluginInfo.account_types.findIndex(t => t.key === typeKey);
  return typeColors[Math.max(0, idx) % typeColors.length];
}

export function getTypeLabel(typeKey) {
  if (!pluginInfo?.account_types) return typeKey;
  return pluginInfo.account_types.find(t => t.key === typeKey)?.label || typeKey;
}

export function maskKey(key) {
  if (!key) return '';
  if (key.length <= 8) return '····';
  return key.slice(0, 4) + '···' + key.slice(-4);
}

export function iconLetter(name, typeKey) {
  if (name) return name[0].toUpperCase();
  return (typeKey || '?')[0].toUpperCase();
}
