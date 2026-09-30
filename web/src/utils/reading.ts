/** 学习模块阅读页的文本切分工具 */

/**
 * 段落切分：优先按空行分段；
 * 若整章没有空行（常见于每行一段的粘贴文本），退化为按单换行分段
 */
export function splitParagraphs(content: string): string[] {
  const normalized = (content || '').replace(/^\uFEFF/, '').replace(/\r\n?/g, '\n').trim()
  if (!normalized) return []

  let paras = normalized
    .split(/\n\s*\n/)
    .map((p) => p.trim())
    .filter(Boolean)

  if (paras.length <= 1 && normalized.includes('\n')) {
    paras = normalized
      .split('\n')
      .map((p) => p.trim())
      .filter(Boolean)
  }
  return paras
}

/** 若正文第一段就是章节标题（用户粘贴时连标题一起贴进来），渲染时去掉避免重复 */
export function dropLeadingTitle(paragraphs: string[], title: string): string[] {
  if (!paragraphs.length) return paragraphs
  const first = paragraphs[0]
  const t = (title || '').trim()
  const isHeading =
    /^(第\s*[0-9一二三四五六七八九十百千零两]+\s*[章回节卷篇话]|(chapter|kapitel)\s+\d+)\b/i.test(first) ||
    (t.length >= 6 && first.slice(0, 12) === t.slice(0, 12))
  return isHeading ? paragraphs.slice(1) : paragraphs
}

/** 句子切分：按 . ! ? … 及收尾引号/括号 */
export function splitSentences(paragraph: string): string[] {
  const parts = paragraph.match(/[^.!?…]+[.!?…]+["')\]”’]*|[^.!?…]+$/g)
  return (parts || [paragraph]).map((s) => s.trim()).filter(Boolean)
}

/** 句内分词：保留空白 token，供模板按序渲染 */
export function tokenizeWords(sentence: string): string[] {
  return sentence.split(/(\s+)/)
}

/** 词语归一化：去掉词首尾标点，仅保留字母（含 äöüß）、撇号、连字符 */
export function normalizeWord(token: string): string {
  return token.replace(/^[^\p{L}']+|[^\p{L}']+$/gu, '')
}

/** 句子 key：段落号-句子号 */
export function sentKey(p: number, s: number): string {
  return `${p}:${s}`
}
