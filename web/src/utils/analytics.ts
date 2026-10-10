/// <reference types="vite/client" />
import { nextTick } from 'vue'
import type { Router } from 'vue-router'

/**
 * 前端埋点阅读路线：
 * main.ts 安装 installAnalytics → 路由切换准备新页面 → Axios 主请求成功
 * → analyticsResponse 等待渲染 → start 记录访问 → tick 累计停留
 * → enqueue 放入队列 → flush 发送给 Go 后端。
 *
 * 三类事件：page_view（访问）、page_engagement（累计停留）、click（外链点击）。
 * 本文件只负责采集与发送；模块归属、标题、去重和最终统计由后端处理。
 */
type Event = {
  // 每条事件自己的编号；生成后重试沿用它，后端据此避免重复入账。
  id: string
  // 一次页面访问的编号；这次访问的 PV、停留、点击共享同一个 viewId。
  viewId: string
  // 浏览器随机标识，保存在 localStorage；不是账号，也不是设备指纹。
  visitor: string
  // 一段连续访问的编号，超过 30 分钟无活动后重新生成。
  session: string
  // 只保存路由路径，例如 /game/12，不包含查询参数和 # 锚点。
  path: string
  // 外部入口站点的 origin；后端进一步只保存域名。
  source: string
  kind: 'page_view' | 'page_engagement' | 'click'
  // 从这次访问开始累计的有效秒数，不是“距离上次上报的秒数”。
  seconds: number
  // 浏览器侧访问开始时间，毫秒时间戳；后端用来约束时长，不用它划分统计日期。
  startedAt: number
  // 点击才需要：操作分类（ai_tool / watch_link / outbound）。
  action?: string
  // 点击目标 origin，例如 https://example.com，不携带链接参数。
  target?: string
}
// 当前页面的计时上下文，只驻留内存。token 和管理员开关在开始访问时固定下来。
type Page = {
  event: Event
  token: string
  includeAdmin: boolean
  // 上次计时检查时间，用于计算这次增加多少毫秒。
  lastTick: number
  // 当前页面最后一次操作时间，用于判断是否需要重新开始一次访问。
  activeAt: number
  // 前端保留毫秒精度，上报时再向下取整为秒，减少频繁取整的误差。
  milliseconds: number
}
// 等待发送的事件快照。失败时保留原对象，因此重试不会生成新的事件 ID。
type Queued = {
  event: Event
  token: string
  includeAdmin: boolean
  attempts: number
}
let router: Router | undefined
// generation 是页面“版本号”：A 页请求很慢、已切换到 B 页时，忽略迟到的 A 页响应。
// ready 表示主请求已成功并等待过渲染；sending 防止普通 fetch 同时发送同一队列。
let generation = 0,
  path = '',
  ready = false,
  sending = false
let current: Page | undefined
let lastActivity = Date.now()
// 队列不写入本地存储；关闭浏览器后无法保证恢复未成功发送的事件。
const queue: Queued[] = []
// WeakMap 把 Axios 请求配置对象与发起时的页面版本关联，对象释放后不会一直占内存。
const requestGenerations = new WeakMap<object, number>()
const memory = new Map<string, string>()
// 浏览器禁用 localStorage 时回退到内存；这种情况下刷新会丢失身份，UV 可能增加。
function read(key: string): string {
  try {
    return localStorage.getItem(key) || memory.get(key) || ''
  } catch {
    return memory.get(key) || ''
  }
}
function write(key: string, value: string) {
  memory.set(key, value)
  try {
    localStorage.setItem(key, value)
  } catch {
    /* 本地持久化失败时，仍保留上面写入的内存值。 */
  }
}
// 优先用浏览器生成 UUID；备用分支用安全随机字节并设置 UUID v4 的版本/变体位。
const uuid = () => {
  if (crypto.randomUUID) return crypto.randomUUID()
  const bytes = crypto.getRandomValues(new Uint8Array(16))
  bytes[6] = (bytes[6] & 15) | 64
  bytes[8] = (bytes[8] & 63) | 128
  const h = Array.from(bytes, (b) => b.toString(16).padStart(2, '0')).join('')
  return `${h.slice(0, 8)}-${h.slice(8, 12)}-${h.slice(12, 16)}-${h.slice(16, 20)}-${h.slice(20)}`
}
// 多次访问复用同一个 visitor；清除网站存储或换浏览器后会被视为新访客。
function identity() {
  let id = read('analyticsVisitor')
  if (!/^[a-f0-9-]{36}$/.test(id)) {
    id = uuid()
    write('analyticsVisitor', id)
  }
  return id
}
// at 是会话最近续期时间。有效计时期间会续期，同源标签页也可以读取这个会话。
function session() {
  let state: { id: string; at: number } | undefined
  try {
    state = JSON.parse(read('analyticsSession'))
  } catch {
    /* 首次访问或存储值损坏时，下面会创建新会话。 */
  }
  if (
    !state ||
    Date.now() - state.at > 30 * 60 * 1000 ||
    !/^[a-f0-9-]{36}$/.test(state.id)
  )
    state = { id: uuid(), at: Date.now() }
  state.at = Date.now()
  write('analyticsSession', JSON.stringify(state))
  return state.id
}
// 开发模式必须主动开启调试；存在登录 token 时，还必须允许管理员测试上报。
// 这里只决定是否采集，管理员身份最终由服务端验证，不能靠客户端自称。
function enabled() {
  return (
    (!import.meta.env.DEV || read('analyticsDebug') === '1') &&
    (!read('token') || read('analyticsIncludeAdmin') === '1')
  )
}
// 给后台“统计口径与调试”的复选框读取当前浏览器配置。
export function analyticsDebugState() {
  return {
    development: import.meta.env.DEV,
    enabled: read('analyticsDebug') === '1',
    includeAdmin: read('analyticsIncludeAdmin') === '1',
  }
}
// 只修改当前浏览器配置，不修改服务器全局设置，也不补报已经错过的访问。
export function setAnalyticsDebug(enabled: boolean, includeAdmin: boolean) {
  write('analyticsDebug', enabled ? '1' : '0')
  write('analyticsIncludeAdmin', includeAdmin ? '1' : '0')
}
// 路径白名单 + 主数据接口对应表。例如 /game 等待 /games 成功才计 PV。
// 未列出的路径（包括 /admin）返回 undefined，因此不会开始采集。
// 新增前台模块时，这里和后端 resolveAnalyticsPage 都要补充。
function expectedAPI(p: string): string | undefined {
  const staticPaths: Record<string, string> = {
    '/': '/articles',
    '/essays': '/articles',
    '/learn': '/books',
    '/novel': '/books',
    '/anime': '/animes',
    '/game': '/games',
    '/art': '/arts',
    '/ai': '/ai-tools',
    '/archives': '/articles/archives',
    '/links': '/links',
    '/about': '/pages/about',
    '/learn/words': '/words',
    '/learn/words/review': '/words',
  }
  if (staticPaths[p]) return staticPaths[p]
  if (/^\/(category|tag)\/\d+$/.test(p)) return '/articles'
  if (/^\/p\/[^/]+$/.test(p)) return '/pages/' + p.slice(3)
  const m = p.match(
    /^\/(article|anime|art|game|game\/post|learn\/book|novel\/book|learn\/chapter|novel\/chapter)\/(\d+)$/,
  )
  if (!m) return undefined
  const resources: Record<string, string> = {
    article: 'articles',
    anime: 'animes',
    art: 'arts',
    game: 'games',
    'game/post': 'game-posts',
    'learn/book': 'books',
    'novel/book': 'books',
    'learn/chapter': 'chapters',
    'novel/chapter': 'chapters',
  }
  return `/${resources[m[1]]}/${m[2]}`
}
// 由 Axios 请求拦截器调用，记录该请求属于哪次路由切换。
export function analyticsRequest(config: object) {
  requestGenerations.set(config, generation)
}
// 由 Axios 成功响应拦截器调用；只有当前页面的主接口才允许启动计次。
export function analyticsResponse(config: { url?: string }) {
  if (
    !router ||
    !expectedAPI(path) ||
    requestGenerations.get(config) !== generation ||
    config.url?.split('?')[0] !== expectedAPI(path)
  )
    return
  const captured = generation
  // 等待 Vue 更新和两帧浏览器绘制，让页面先消费返回的数据。
  // 等待期间仍可能切换路由，因此回调内再检查一次 generation。
  void nextTick().then(() =>
    requestAnimationFrame(() =>
      requestAnimationFrame(() => {
        if (generation !== captured) return
        ready = true
        start()
      }),
    ),
  )
}
// 开始一次访问：已开始、数据未就绪、禁用采集、页面隐藏时都直接返回。
// current 的存在保证同一页面多个成功响应不会重复触发 PV。
function start() {
  if (
    current ||
    !ready ||
    !enabled() ||
    document.visibilityState !== 'visible' ||
    !expectedAPI(path)
  )
    return
  const now = Date.now()
  let source = ''
  try {
    const u = new URL(document.referrer)
    if (u.hostname !== location.hostname) source = u.origin
  } catch {
    /* 没有有效 referrer 时保持空字符串，后端归为直接访问。 */
  }
  const event: Event = {
    id: uuid(),
    viewId: uuid(),
    visitor: identity(),
    session: session(),
    path,
    source,
    kind: 'page_view',
    seconds: 0,
    startedAt: now,
  }
  current = {
    event,
    token: read('token'),
    includeAdmin: read('analyticsIncludeAdmin') === '1',
    lastTick: now,
    activeAt: now,
    milliseconds: 0,
  }
  lastActivity = now
  enqueue(current, 'page_view')
  void flush()
}
// 将事件放入发送队列：继承 visitor/session/viewId，但每条新事件单独生成 id。
// extra 用于给点击事件补充 action 和 target。
function enqueue(page: Page, kind: Event['kind'], extra: Partial<Event> = {}) {
  const event = {
    ...page.event,
    kind,
    id: uuid(),
    seconds: Math.floor(page.milliseconds / 1000),
    ...extra,
  }
  // 同一次访问只保留最新待发的累计停留。例如 15 秒尚未发出，又到了 30 秒，
  // 直接保留 30 秒即可；后端会基于已保存的最大秒数算差值。
  if (kind === 'page_engagement') {
    const i = queue.findIndex(
      (q) => q.event.viewId === event.viewId && q.event.kind === kind,
    )
    if (i >= 0) queue.splice(i, 1)
  }
  queue.push({
    event,
    token: page.token,
    includeAdmin: page.includeAdmin,
    attempts: 0,
  })
  // 断网时限制内存占用；超过 100 条会丢弃最旧的一条，这是一种有界的尽力上报。
  if (queue.length > 100) queue.shift()
}
// 每秒检查一次，但不能直接“每调用一次就加 1 秒”，因为定时器可能延迟。
// 按实际间隔累加，每次最多加 2 秒，避免电脑休眠后把整段休眠时间算进去。
// 可见且最近 60 秒有操作才累计；单次页面访问最多累计 4 小时。
function tick() {
  if (!current) return
  const now = Date.now(),
    delta = Math.min(2000, Math.max(0, now - current.lastTick))
  current.lastTick = now
  if (
    document.visibilityState === 'visible' &&
    now - lastActivity < 60000 &&
    enabled()
  ) {
    current.milliseconds = Math.min(14400000, current.milliseconds + delta)
    session()
  }
}
// 离开页面时补最后一段停留并清空 current；实际发送由调用者触发。
function finish() {
  if (current) {
    tick()
    enqueue(current, 'page_engagement')
    current = undefined
  }
}
// 普通情况用 fetch 获取响应；页面隐藏/离开时用 beacon 尝试继续交给浏览器发送。
// 每批只取身份和管理员开关相同的最多 10 条，避免把不同登录状态混为一批。
async function flush(beacon = false) {
  if ((sending && !beacon) || !queue.length) return
  const first = queue[0],
    batch = queue
      .filter(
        (q) => q.token === first.token && q.includeAdmin === first.includeAdmin,
      )
      .slice(0, 10)
  const payload = JSON.stringify({
    events: batch.map((q) => q.event),
    token: first.token,
    includeAdmin: first.includeAdmin,
  })
  if (beacon && navigator.sendBeacon) {
    // beacon 不提供服务端处理结果，所以不删除队列。页面恢复后仍可用相同 ID 重试。
    // 允许它与正在进行的 fetch 重叠，重复请求由后端事件 ID 去重。
    navigator.sendBeacon(
      '/api/v1/analytics/events',
      new Blob([payload], { type: 'application/json' }),
    )
    return
  }
  sending = true
  try {
    // 直接用 fetch，不经过业务 Axios 拦截器，避免统计失败弹出提示影响读者。
    // keepalive 尝试让页面离开后请求继续，8 秒没有完成则中止并保留待重试事件。
    const response = await fetch('/api/v1/analytics/events', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: payload,
      keepalive: true,
      signal: AbortSignal.timeout(8000),
    })
    // 成功则移出队列；多数 4xx 代表事件不合法，重试无意义，也移出。
    // 429（限流）、5xx 和网络异常保留重试，下一次定时 flush 会再次发送。
    if (
      response.ok ||
      (response.status >= 400 &&
        response.status < 500 &&
        response.status !== 429)
    ) {
      for (const item of batch) {
        const i = queue.indexOf(item)
        if (i >= 0) queue.splice(i, 1)
      }
    } else throw new Error('retry')
  } catch {
    for (const item of batch) {
      item.attempts++
      // 连续发送失败达到 5 次后丢弃，避免一直重试拖累前台。
      if (item.attempts >= 5) {
        const i = queue.indexOf(item)
        if (i >= 0) queue.splice(i, 1)
      }
    }
  } finally {
    sending = false
  }
}
// main.ts 在 app.use(router) 前调用一次，以便监听首次路由进入。
// 这里统一安装全站监听器；各业务页面无需重复注册计时器和点击监听。
export function installAnalytics(r: Router) {
  router = r
  r.afterEach((to, _from, failure) => {
    // 导航失败不计次；只比较 path，所以同页 query/hash 变化不会重复计次。
    if (failure || to.path === path) return
    finish()
    void flush()
    path = to.path
    generation++
    ready = false
  })
  // 长时间没有操作后再次活动，结束旧访问并尝试开始新的访问上下文。
  const activity = () => {
    lastActivity = Date.now()
    if (current && Date.now() - current.activeAt > 30 * 60 * 1000) {
      finish()
      start()
    }
    if (current) current.activeAt = Date.now()
  }
  for (const name of [
    'pointerdown',
    'pointermove',
    'keydown',
    'scroll',
    'touchstart',
  ])
    window.addEventListener(name, activity, { passive: true })
  // 切到其他标签页时尝试补报；切回来重置计时基准，避免累加隐藏期间的时间。
  document.addEventListener('visibilitychange', () => {
    if (document.visibilityState === 'hidden') {
      tick()
      if (current) enqueue(current, 'page_engagement')
      void flush(true)
    } else {
      lastActivity = Date.now()
      if (current && Date.now() - current.activeAt > 30 * 60 * 1000) finish()
      if (current) current.lastTick = Date.now()
      start()
      void flush()
    }
  })
  // 关闭、刷新或离开当前文档时最后尝试补报；浏览器强制结束仍可能丢失。
  window.addEventListener('pagehide', () => {
    if (current) {
      tick()
      enqueue(current, 'page_engagement')
    }
    void flush(true)
  })
  // 事件委托：监听 document 就能捕获动态渲染的链接，无需每个链接绑定埋点。
  // closest 允许点击链接内的图片、图标时也找到外层 <a>。
  document.addEventListener('click', (e) => {
    if (!current || !enabled()) return
    const link = (e.target as Element)?.closest?.(
      'a[href]',
    ) as HTMLAnchorElement | null
    if (!link) return
    try {
      const u = new URL(link.href)
      if (
        !['http:', 'https:'].includes(u.protocol) ||
        u.origin === location.origin
      )
        return
      const action =
        link.dataset.analyticsAction ||
        (path === '/ai' ? 'ai_tool' : 'outbound')
      tick()
      enqueue(current, 'click', { action, target: u.origin })
      void flush(true)
      void flush()
    } catch {
      /* 无法解析的链接忽略，不影响原来的点击行为。 */
    }
  })
  // 计时频率与网络上报频率分开：每秒计时，每 15 秒批量发送以减少请求数。
  setInterval(tick, 1000)
  setInterval(() => {
    if (current && enabled()) enqueue(current, 'page_engagement')
    void flush()
  }, 15000)
}
