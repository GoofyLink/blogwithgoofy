/** 主题注册表：新增主题 = 在 style.css 加变量块 + 在这里加一行 */

export interface ThemeDef {
  key: string
  label: string
  hint?: string
}

export const THEMES: ThemeDef[] = [
  { key: 'paper', label: '米纸陶土', hint: '暖纸色 · 默认' },
  { key: 'cool', label: '冷白科技蓝', hint: '现代工具风' },
  { key: 'tomato', label: '番茄橙', hint: '轻快阅读风' },
  { key: 'mimo', label: '极简墨白', hint: 'MiMo 极简技术风' },
]

const STORAGE_KEY = 'theme'

export function currentTheme(): string {
  return document.documentElement.dataset.theme || 'paper'
}

export function applyTheme(key: string) {
  document.documentElement.dataset.theme = key
  try {
    localStorage.setItem(STORAGE_KEY, key)
  } catch {
    /* 忽略 */
  }
}

/** 启动时恢复上次选择的主题 */
export function initTheme() {
  try {
    const saved = localStorage.getItem(STORAGE_KEY)
    if (saved) document.documentElement.dataset.theme = saved
  } catch {
    /* 忽略 */
  }
}
