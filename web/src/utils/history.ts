/** 阅读历史（localStorage，本浏览器内有效，无需登录） */

export interface ReadingRecord {
  chapterId: number
  chapterTitle: string
  bookId: number
  bookTitle: string
  language: string
  kind?: 'learn' | 'novel' // 学习书房 / 小说专栏（旧记录缺省视为 learn）
  at: number // 时间戳 ms
}

const KEY = 'readingHistory'
const MAX = 20

export function getReadingHistory(kind?: 'learn' | 'novel'): ReadingRecord[] {
  try {
    const list = JSON.parse(localStorage.getItem(KEY) || '[]')
    if (!Array.isArray(list)) return []
    if (!kind) return list
    return list.filter((x) => (x.kind || 'learn') === kind)
  } catch {
    return []
  }
}

/** 记录一次阅读：同章去重置顶，最多保留 MAX 条 */
export function pushReadingRecord(r: Omit<ReadingRecord, 'at'>) {
  const list = getReadingHistory().filter((x) => x.chapterId !== r.chapterId)
  list.unshift({ ...r, at: Date.now() })
  try {
    localStorage.setItem(KEY, JSON.stringify(list.slice(0, MAX)))
  } catch {
    /* 存储失败忽略 */
  }
}

export function removeReadingRecord(chapterId: number) {
  const list = getReadingHistory().filter((x) => x.chapterId !== chapterId)
  try {
    localStorage.setItem(KEY, JSON.stringify(list))
  } catch {
    /* 忽略 */
  }
}

/** 相对时间：刚刚 / N 分钟前 / N 小时前 / N 天前 / 日期 */
export function timeAgo(ts: number): string {
  const diff = Date.now() - ts
  const min = Math.floor(diff / 60000)
  if (min < 1) return '刚刚'
  if (min < 60) return `${min} 分钟前`
  const hour = Math.floor(min / 60)
  if (hour < 24) return `${hour} 小时前`
  const day = Math.floor(hour / 24)
  if (day < 30) return `${day} 天前`
  const d = new Date(ts)
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}
