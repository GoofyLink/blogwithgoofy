/** 浏览器原生语音朗读封装（Web Speech API） */

export interface SpeakHandle {
  pause(): void
  resume(): void
  stop(): void
}

export interface SpeakOptions {
  lang: string // en / de
  rate: number // 语速，朗读中修改会在下一句生效
  onStart?: (index: number) => void
  onDone?: () => void
}

/** 触发浏览器加载语音列表（Chrome 首次异步） */
export function warmUpVoices() {
  if (typeof window !== 'undefined' && 'speechSynthesis' in window) {
    window.speechSynthesis.getVoices()
  }
}

/** 按语言挑选最优语音：优先在线神经语音（Google / Natural / Online） */
export function pickVoice(lang: string): SpeechSynthesisVoice | null {
  if (!('speechSynthesis' in window)) return null
  const target = lang === 'de' ? 'de' : lang === 'zh' ? 'zh' : 'en'
  const preferred = target === 'de' ? 'de-de' : target === 'zh' ? 'zh-cn' : 'en-us'
  const voices = window.speechSynthesis.getVoices().filter((v) =>
    v.lang.toLowerCase().startsWith(target),
  )
  if (!voices.length) return null

  const score = (v: SpeechSynthesisVoice) => {
    let s = 0
    const n = v.name.toLowerCase()
    if (v.lang.toLowerCase().startsWith(preferred)) s += 3
    if (n.includes('natural')) s += 5
    if (n.includes('google')) s += 4
    if (n.includes('online')) s += 2
    if (n.includes('microsoft')) s += 1
    if (v.localService) s -= 1 // 在线语音通常更好
    return s
  }
  return [...voices].sort((a, b) => score(b) - score(a))[0]
}

/**
 * 依次朗读一组文本（通常为句子）。
 * onStart 回调当前句索引，用于页面高亮跟随；stop 后不会触发后续回调。
 */
export function speakTexts(texts: string[], opts: SpeakOptions): SpeakHandle {
  let stopped = false
  let paused = false
  let idx = 0

  const speakNext = () => {
    if (stopped) return
    if (idx >= texts.length) {
      opts.onDone?.()
      return
    }
    const i = idx
    const u = new SpeechSynthesisUtterance(texts[i])
    const voice = pickVoice(opts.lang)
    if (voice) {
      u.voice = voice
      u.lang = voice.lang
    } else {
      u.lang = opts.lang === 'de' ? 'de-DE' : opts.lang === 'zh' ? 'zh-CN' : 'en-US'
    }
    u.rate = opts.rate
    u.onstart = () => {
      if (!stopped) opts.onStart?.(i)
    }
    u.onend = () => {
      if (stopped) return
      idx++
      speakNext()
    }
    u.onerror = () => {
      if (stopped) return
      idx++
      speakNext()
    }
    window.speechSynthesis.speak(u)
  }
  speakNext()

  return {
    pause() {
      if (!stopped && !paused) {
        paused = true
        window.speechSynthesis.pause()
      }
    },
    resume() {
      if (paused && !stopped) {
        paused = false
        window.speechSynthesis.resume()
      }
    },
    stop() {
      stopped = true
      window.speechSynthesis.cancel()
    },
  }
}
