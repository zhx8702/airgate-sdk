// 插件 Widget 加载器
// 动态加载插件提供的 React 组件，挂载到表单的 plugin-widget-slot 中

import { pluginInfo, API } from './utils.js';

let pluginModule = null;
let widgetRoot = null;

/** 加载插件前端模块（仅加载一次） */
async function loadPluginModule() {
  if (pluginModule !== null) return pluginModule;
  const widget = pluginInfo?.frontend_widgets?.find(w => w.slot === 'account-form');
  if (!widget) {
    pluginModule = false;
    return false;
  }
  try {
    pluginModule = await import('/plugin-assets/' + widget.entry_file);
  } catch (e) {
    console.warn('加载插件前端模块失败:', e);
    pluginModule = false;
  }
  return pluginModule;
}

/**
 * 尝试加载并渲染插件 Widget
 * @param {string} typeKey 当前选中的账号类型
 * @param {object} formState 表单状态回调
 * @returns {boolean} 是否成功加载了 Widget（用于决定是否隐藏默认凭证字段）
 */
export async function tryLoadPluginWidget(typeKey, formState) {
  const slot = document.getElementById('plugin-widget-slot');
  const credFields = document.getElementById('cred-fields');

  const mod = await loadPluginModule();
  if (!mod || !mod.default?.accountForm) {
    // 无插件 Widget，显示默认凭证字段
    slot.style.display = 'none';
    slot.innerHTML = '';
    credFields.style.display = '';
    return false;
  }

  // 有插件 Widget，隐藏默认凭证字段
  slot.style.display = 'block';
  credFields.style.display = 'none';

  const FormComponent = mod.default.accountForm;

  // 初始化 React root（仅一次）
  if (!widgetRoot) {
    const ReactDOM = await import('react-dom/client');
    widgetRoot = ReactDOM.createRoot(slot);
  }

  const React = await import('react');

  // 构建 oauth prop
  const oauth = buildOAuthProp();

  widgetRoot.render(
    React.createElement(FormComponent, {
      credentials: formState.credentials,
      onChange: formState.onChange,
      mode: formState.mode,
      accountType: typeKey,
      onAccountTypeChange: formState.onAccountTypeChange,
      onSuggestedName: formState.onSuggestedName,
      oauth,
    })
  );

  return true;
}

/** 构建传递给 AccountForm 的 oauth prop */
function buildOAuthProp() {
  return {
    start: async () => {
      const res = await fetch(API + '/api/oauth/start', { method: 'POST' });
      if (!res.ok) {
        const err = await res.text();
        throw new Error(err || '生成授权链接失败');
      }
      return res.json();
    },
    exchange: async (callbackURL) => {
      // 从回调 URL 提取 code 和 state
      const url = new URL(callbackURL);
      const code = url.searchParams.get('code');
      const state = url.searchParams.get('state');
      if (!code || !state) {
        throw new Error('回调 URL 缺少 code 或 state 参数');
      }
      const res = await fetch(API + '/api/oauth/callback', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ code, state }),
      });
      if (!res.ok) {
        const err = await res.text();
        throw new Error(err || '授权交换失败');
      }
      const data = await res.json();
      return {
        accountType: data.account?.account_type || 'oauth',
        accountName: data.account?.name || '',
        credentials: data.account?.credentials || {},
      };
    },
  };
}

/** 卸载 Widget（关闭表单时调用） */
export function unmountWidget() {
  if (widgetRoot) {
    widgetRoot.unmount();
    widgetRoot = null;
  }
  const slot = document.getElementById('plugin-widget-slot');
  slot.style.display = 'none';
  slot.innerHTML = '';
}
