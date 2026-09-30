export interface User {
  id: number
  username: string
  nickname: string
  avatar: string
}

export interface Category {
  id: number
  name: string
  articleCount?: number
}

export interface Tag {
  id: number
  name: string
  articleCount?: number
}

/** type: 0 博客文章 / 1 学习文章 */
export interface Article {
  id: number
  title: string
  summary: string
  cover: string
  content?: string
  translation?: string
  categoryId: number
  type: 0 | 1
  language: string
  status: 0 | 1
  views: number
  createdAt: string
  updatedAt: string
  category?: Category | null
  tags?: Tag[]
}

export interface ArticleNeighbor {
  id: number
  title: string
}

export interface ArticleDetailData {
  article: Article
  prev: ArticleNeighbor | null
  next: ArticleNeighbor | null
}

export interface Annotation {
  id: number
  articleId: number
  paragraphIndex: number
  quote: string
  note: string
  createdAt: string
  updatedAt: string
}

export interface CommentItem {
  id: number
  articleId: number
  nickname: string
  email: string
  content: string
  createdAt: string
  articleTitle?: string
}

export interface LinkItem {
  id: number
  name: string
  url: string
  logo: string
  description: string
  sort: number
}

export interface SinglePage {
  id: number
  slug: string
  title: string
  content: string
  createdAt: string
  updatedAt: string
}

export interface PageResult<T> {
  list: T[]
  total: number
}

export interface ArchiveItem {
  id: number
  title: string
  type: number
  language: string
  createdAt: string
}

export interface ArchiveGroup {
  month: string
  items: ArchiveItem[]
}

export interface DashboardStats {
  articleCount: number
  publishedCount: number
  learningCount: number
  commentCount: number
  categoryCount: number
  tagCount: number
  viewTotal: number
  bookCount: number
  chapterCount: number
}

export interface ArticleForm {
  title: string
  summary: string
  cover: string
  content: string
  translation: string
  categoryId: number
  type: 0 | 1
  language: string
  status: 0 | 1
  tagIds: number[]
}

export interface LoginResp {
  token: string
  user: User
}

/** ===== 学习模块：书籍 / 章节 ===== */
/** type: 0 学习书房 / 1 小说专栏 */
export interface Book {
  id: number
  title: string
  subtitle: string
  author: string
  language: string // en / de / zh
  cover: string
  description: string
  type: 0 | 1
  status: 0 | 1
  sort: number
  createdAt: string
  updatedAt: string
  chapterCount?: number
}

export interface ChapterItem {
  id: number
  title: string
  sort: number
  createdAt: string
}

export interface Chapter {
  id: number
  bookId: number
  title: string
  content: string
  translation: string
  sort: number
  status: 0 | 1
  createdAt: string
  updatedAt: string
}

export interface ChapterNeighbor {
  id: number
  title: string
}

export interface ChapterDetailData {
  chapter: Chapter
  book: Book
  prev: ChapterNeighbor | null
  next: ChapterNeighbor | null
}

export interface BookDetailData {
  book: Book
  chapters: ChapterItem[]
}

export interface TranslateResp {
  text: string
  cached: boolean
}

/** 章节阅读笔记 */
export interface ChapterNote {
  id: number
  chapterId: number
  paragraphIndex: number
  quote: string
  note: string
  createdAt: string
  updatedAt: string
}

/** 生词本 */
export interface WordItem {
  id: number
  word: string
  language: string
  translation: string
  sourceText: string
  createdAt: string
  updatedAt: string
}

/** 动漫排行榜条目 */
export interface AnimeItem {
  id: number
  title: string
  cover: string
  category: string
  region: string
  episodes: string
  playCount: number
  description: string
  rank: number
  status: 0 | 1
  createdAt: string
  updatedAt: string
}

/** AI 工具导航条目 */
export interface AiToolItem {
  id: number
  name: string
  url: string
  icon: string
  section: string
  category: string
  description: string
  tags: string
  sort: number
  status: 0 | 1
  createdAt: string
  updatedAt: string
}

/** AI 提示词 */
export interface AiPromptItem {
  id: number
  title: string
  section: string
  category: string
  content: string
  description: string
  sort: number
  status: 0 | 1
  createdAt: string
  updatedAt: string
}

/** 游戏板块条目 */
export interface GameItem {
  id: number
  title: string
  cover: string
  category: string
  platform: string
  description: string
  content?: string
  tags: string
  hot: number
  sort: number
  status: 0 | 1
  createdAt: string
  updatedAt: string
}

/** 接口调用日志 */
export interface ApiLogItem {
  id: number
  method: string
  path: string
  ip: string
  status: number
  durationMs: number
  userAgent: string
  createdAt: string
}

/** 游戏攻略/资讯帖子 */
export interface GamePostItem {
  id: number
  gameId: number
  title: string
  type: string
  summary: string
  content?: string
  views: number
  status: 0 | 1
  createdAt: string
  updatedAt: string
}

/** 游戏圈评论 */
export interface GameCommentItem {
  id: number
  postId: number
  gameId: number
  nickname: string
  email: string
  content: string
  createdAt: string
  postTitle?: string
  gameTitle?: string
}

/** 艺术鉴赏作品 */
export interface ArtworkItem {
  id: number
  title: string
  image: string
  description: string
  author: string
  sort: number
  status: 0 | 1
  createdAt: string
  updatedAt: string
}
