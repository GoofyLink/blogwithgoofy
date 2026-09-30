import MarkdownIt from 'markdown-it'
import hljs from 'highlight.js'
import DOMPurify from 'dompurify'

const md = new MarkdownIt({
  html: false,
  linkify: true,
  highlight(code, lang) {
    if (lang && hljs.getLanguage(lang)) {
      try {
        return hljs.highlight(code, { language: lang }).value
      } catch {
        /* 忽略高亮错误 */
      }
    }
    return ''
  },
})

export interface TocItem {
  id: string
  text: string
  level: number
}

/** 从 Markdown 源码提取目录（与渲染时的标题编号保持一致） */
export function extractToc(src: string): TocItem[] {
  const tokens = md.parse(src || '', {})
  const toc: TocItem[] = []
  let counter = 0
  for (const tok of tokens) {
    if (tok.type === 'heading_open') {
      const text = (tok.children || [])
        .filter((t) => t.type === 'text' || t.type === 'code_inline')
        .map((t) => t.content)
        .join('')
      toc.push({ id: `h${counter++}`, text, level: Number(tok.tag.slice(1)) })
    }
  }
  return toc
}

/** 渲染 Markdown 为经过消毒的 HTML，并为标题注入 h0/h1... id 供 TOC 定位 */
export function renderMarkdown(src: string): string {
  let counter = 0
  const defaultRender =
    md.renderer.rules.heading_open ||
    ((tokens, idx, options, _env, self) => self.renderToken(tokens, idx, options))
  md.renderer.rules.heading_open = (tokens, idx, options, env, self) => {
    tokens[idx].attrSet('id', `h${counter++}`)
    return defaultRender(tokens, idx, options, env, self)
  }
  const html = md.render(src || '')
  return DOMPurify.sanitize(html)
}

/** 学习文章：按空行拆分段落（段落内保留原样），逐段渲染 */
export function splitParagraphs(content: string): string[] {
  return (content || '')
    .split(/\n\s*\n/)
    .map((p) => p.trim())
    .filter(Boolean)
}
