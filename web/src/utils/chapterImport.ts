/** 整本书文本 → 章节数组 的解析工具（批量导入用） */

export interface ParsedChapter {
  title: string
  content: string
  translation?: string
}

export interface ParseResult {
  chapters: ParsedChapter[]
  warning?: string
}

export type SplitMode = 'auto' | 'separator' | 'size'

/** 章节标题行：Chapter 1 / Kapitel 3 / 第十二章 / 第一篇 / 第三话 / 「1. 疯狂年代」（支持 markdown # 前缀） */
const EN_HEADING_RE = /^(?:chapter|kapitel)\s+(?:\d+|[ivxlcdm]+)\b/i

/** 「1. 标题」「2．标题」数字编号标题；排除 3.14 这类小数 */
function numHeading(line: string): { num: string; rest: string } | null {
  const m = line.match(/^(\d{1,4})\s*[．.、]\s*(.*)$/)
  if (!m) return null
  if (/^\d/.test(m[2] || '')) return null
  return { num: m[1], rest: (m[2] || '').trim() }
}

function isChapterHeading(line: string): boolean {
  const s = line.replace(/^#+\s*/, '').trim()
  if (!s || s.length > 80) return false
  if (EN_HEADING_RE.test(s)) return true
  if (numHeading(s)) return true
  return /^第\s*[0-9一二三四五六七八九十百千零两]+\s*[章回节卷篇话]/.test(s)
}

/** 标题行 → 章节标题：数字编号行提取编号与名称，无名称时用「第N章」 */
function headingToTitle(line: string): string {
  const s = line.replace(/^#+\s*/, '').trim()
  const nh = numHeading(s)
  if (nh) return nh.rest || `第${nh.num}章`
  return cleanTitle(line)
}

/** 全文分段：中文小说常见「每行一段、行首全角缩进」，优先识别；否则按空行分段 */
function splitToParagraphs(normalized: string): string[] {
  const nonEmpty = normalized.split('\n').filter((l) => l.trim())
  const indented = nonEmpty.filter((l) => /^[\s\u3000]{2,}/.test(l)).length
  if (nonEmpty.length > 1 && indented >= nonEmpty.length * 0.6) {
    return nonEmpty.map((l) => l.trim()).filter(Boolean)
  }
  let paras = normalized.split(/\n\s*\n/).map((p) => p.trim()).filter(Boolean)
  if (paras.length <= 1 && normalized.includes('\n')) {
    paras = nonEmpty.map((l) => l.trim()).filter(Boolean)
  }
  return paras
}

/** 去掉段首缩进空格（含全角空格 　） */
function cleanPara(p: string): string {
  return p.replace(/^[\s\u3000]+/, '')
}

/** 仅标题模式：每行一个标题，返回去空行、修剪后的标题列表 */
export function parseTitleList(text: string): { titles: string[]; warning?: string } {
  const lines = (text || '')
    .replace(/^\uFEFF/, '')
    .replace(/\r\n?/g, '\n')
    .split('\n')
    .map((l) => l.replace(/^#+\s*/, '').trim())
    .filter(Boolean)
    .map((l) => l.slice(0, 100))
  if (!lines.length) return { titles: [], warning: '内容为空，请每行粘贴一个标题' }
  return { titles: lines }
}

function cleanTitle(line: string): string {
  return line.replace(/^#+\s*/, '').trim().slice(0, 100)
}

function joinLines(lines: string[]): string {
  return lines.join('\n').replace(/\n{3,}/g, '\n\n').trim()
}

/**
 * 解析整书文本为章节列表
 * - auto 模式：按 Chapter/Kapitel/第X章 标题行切分，标题行即章节标题，首个标题前的内容作为「前言」
 * - separator 模式：按「整行等于分隔符」切分，每块第一行作为标题
 */
/** 按章节标题行切分（auto 模式 / size 模式检测到标题时自动升级） */
function parseByHeadings(normalized: string): ParseResult {
  const chapters: ParsedChapter[] = []
  const preface: string[] = []
  let current: { title: string; lines: string[] } | null = null

  for (const line of normalized.split('\n')) {
    if (isChapterHeading(line)) {
      if (current) chapters.push({ title: current.title, content: cleanPara(joinLines(current.lines)) })
      current = { title: headingToTitle(line), lines: [] }
    } else if (current) {
      current.lines.push(line)
    } else {
      preface.push(line)
    }
  }
  if (current) chapters.push({ title: current.title, content: cleanPara(joinLines(current.lines)) })

  // 标题前内容：只有实质内容（≥20 字）才值得单独成章，短的一两行（书名等）直接丢弃
  const prefaceContent = joinLines(preface)
  if (prefaceContent.replace(/\s/g, '').length >= 20) {
    chapters.unshift({ title: '前言 Preface', content: prefaceContent })
  }

  if (!chapters.length) {
    return {
      chapters: [],
      warning: '未识别到章节标题（支持 Chapter 1 / 第1章 / Kapitel 1）。可改用「自定义分隔符」模式。',
    }
  }
  return { chapters }
}

export function parseChapters(
  text: string,
  mode: SplitMode,
  separator = '---',
  sizePerChapter = 3000,
): ParseResult {
  const normalized = (text || '').replace(/^\uFEFF/, '').replace(/\r\n?/g, '\n').trim()
  if (!normalized) return { chapters: [], warning: '内容为空，请先粘贴文本或选择文件' }

  // 按长度分章模式：若文中已带有 2 个以上可识别的章节标题，
  // 自动升级为按标题切分，提取真实标题（如「第二章 夜空骄阳」）
  if (mode === 'size') {
    const headingCount = normalized.split('\n').filter(isChapterHeading).length
    if (headingCount >= 2) return parseByHeadings(normalized)
  }

  // 按长度自动分章：适合没有章节标题的整篇 txt（如每行一段的小说）
  if (mode === 'size') {
    const target = Math.max(800, Math.min(20000, Math.round(sizePerChapter) || 3000))
    const paras = splitToParagraphs(normalized).map(cleanPara)
    // 跳过开头的书名行（如《宇宙坍缩》）
    if (paras.length && paras[0].length <= 30 && /^《.+》$/.test(paras[0])) paras.shift()
    if (!paras.length) return { chapters: [], warning: '内容为空' }

    const chapters: ParsedChapter[] = []
    let cur: string[] = []
    let curLen = 0
    for (const p of paras) {
      cur.push(p)
      curLen += p.length
      if (curLen >= target) {
        chapters.push({ title: `第${chapters.length + 1}章`, content: cur.join('\n\n') })
        cur = []
        curLen = 0
      }
    }
    // 收尾：剩余内容太短则并入上一章，否则独立成章
    if (cur.length) {
      const tail = cur.join('\n\n')
      if (chapters.length && curLen < target * 0.2) {
        chapters[chapters.length - 1].content += '\n\n' + tail
      } else {
        chapters.push({ title: `第${chapters.length + 1}章`, content: tail })
      }
    }
    return { chapters }
  }

  if (mode === 'auto') return parseByHeadings(normalized)

  if (mode === 'separator') {
    const sep = (separator || '').trim()
    if (!sep) return { chapters: [], warning: '请填写分隔符' }

    const blocks: string[][] = [[]]
    for (const line of normalized.split('\n')) {
      if (line.trim() === sep) blocks.push([])
      else blocks[blocks.length - 1].push(line)
    }

    const chapters: ParsedChapter[] = []
    for (const block of blocks) {
      const nonEmpty = block.filter((l) => l.trim())
      if (!nonEmpty.length) continue
      const title = cleanTitle(nonEmpty[0]) || '未命名章节'
      const content = joinLines(block.slice(1)) || nonEmpty[0].trim()
      chapters.push({ title, content })
    }
    if (!chapters.length) return { chapters: [], warning: '没有解析到任何内容' }
    return { chapters }
  }

  return { chapters: [], warning: '未知的分割模式' }
}

/** 读取 txt 文件：UTF-8 优先，出现乱码字符时回退 GBK（中文 txt 常见编码） */
export async function readFileText(file: File): Promise<string> {
  const buf = await file.arrayBuffer()
  let text = new TextDecoder('utf-8').decode(buf)
  if (text.includes('\uFFFD')) {
    try {
      text = new TextDecoder('gbk').decode(buf)
    } catch {
      /* 浏览器不支持 gbk 时保留 utf-8 结果 */
    }
  }
  return text
}

/** 把译文按同样规则切分并与章节按序对齐 */
export function alignTranslations(
  chapters: ParsedChapter[],
  transText: string,
  mode: SplitMode,
  separator: string,
  sizePerChapter = 3000,
): { aligned: ParsedChapter[]; warning?: string } {
  const parsed = parseChapters(transText, mode, separator, sizePerChapter)
  if (!parsed.chapters.length || !transText.trim()) {
    return { aligned: chapters }
  }
  let warning: string | undefined
  let trans = parsed.chapters
  // auto 模式下若译文把「前言」识别进来了而原文没有（或反之），数量不齐时仍按下标对齐
  if (trans.length !== chapters.length) {
    warning = `原文 ${chapters.length} 章 / 译文 ${trans.length} 章，已按顺序对齐前 ${Math.min(
      chapters.length,
      trans.length,
    )} 章，其余可在章节编辑里单独补充`
  }
  const aligned = chapters.map((ch, i) => {
    if (i < trans.length) {
      // 译文块正文（auto 模式标题行无用，取 content；separator 模式首行被当标题，拼回）
      const body = trans[i].content || trans[i].title
      return { ...ch, translation: body }
    }
    return ch
  })
  return { aligned, warning }
}
