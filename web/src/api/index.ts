import { get, post, put, del } from './request'
import type {
  ArchiveGroup, Article, ArticleDetailData, ArticleForm, Book, BookDetailData,
  Category, Chapter, ChapterDetailData, ChapterNote, CommentItem, DashboardStats,
  LinkItem, LoginResp, PageResult, SinglePage, Tag, TranslateResp, WordItem, AnimeItem, ArtworkItem, AiToolItem, AiPromptItem, GameItem,
} from '@/types'

// ---- 认证 ----
export const login = (data: { username: string; password: string }) =>
  post<LoginResp>('/auth/login', data)

export const changePassword = (data: { oldPassword: string; newPassword: string }) =>
  put<null>('/admin/password', data)

// ---- 文章 ----
export interface ArticleQuery {
  page?: number
  size?: number
  type?: number | string
  language?: string
  categoryId?: number | string
  tagId?: number | string
  keyword?: string
}

export const fetchArticles = (params: ArticleQuery) =>
  get<PageResult<Article>>('/articles', params)

export const fetchArticle = (id: number | string) =>
  get<ArticleDetailData>(`/articles/${id}`)

export const fetchArchives = () => get<ArchiveGroup[]>('/articles/archives')

export const adminFetchArticles = (params: ArticleQuery) =>
  get<PageResult<Article>>('/admin/articles', params)

export const createArticle = (data: ArticleForm) => post<Article>('/admin/articles', data)

export const updateArticle = (id: number, data: ArticleForm) =>
  put<Article>(`/admin/articles/${id}`, data)

export const deleteArticle = (id: number) => del<null>(`/admin/articles/${id}`)

// ---- 评论 ----
export const fetchComments = (articleId: number | string) =>
  get<CommentItem[]>(`/articles/${articleId}/comments`)

export const createComment = (
  articleId: number | string,
  data: { nickname: string; email?: string; content: string },
) => post<CommentItem>(`/articles/${articleId}/comments`, data)

export const adminFetchComments = (params: { page?: number; size?: number; articleId?: number | string }) =>
  get<PageResult<CommentItem>>('/admin/comments', params)

export const deleteComment = (id: number) => del<null>(`/admin/comments/${id}`)

// ---- 学习模块：书籍 / 章节 / 翻译 ----
export const fetchBooks = (params?: { language?: string; type?: number | string }) =>
  get<Book[]>('/books', params)

export const fetchBook = (id: number | string) =>
  get<BookDetailData>(`/books/${id}`)

export const fetchChapter = (id: number | string) =>
  get<ChapterDetailData>(`/chapters/${id}`)

export const translateText = (text: string, from: string) =>
  get<TranslateResp>('/translate', { text, from })

export const fetchChapterNotes = (chapterId: number | string) =>
  get<ChapterNote[]>(`/chapters/${chapterId}/notes`)

export const createChapterNote = (data: {
  chapterId: number
  paragraphIndex: number
  quote: string
  note: string
}) => post<ChapterNote>('/admin/chapter-notes', data)

export const updateChapterNote = (
  id: number,
  data: { paragraphIndex: number; quote: string; note: string },
) => put<ChapterNote>(`/admin/chapter-notes/${id}`, data)

export const deleteChapterNote = (id: number) => del<null>(`/admin/chapter-notes/${id}`)

// ---- 生词本 ----
export const fetchWords = (params?: { language?: string; keyword?: string }) =>
  get<WordItem[]>('/words', params)

export const addWord = (data: {
  word: string
  language: string
  translation: string
  sourceText?: string
}) => post<WordItem>('/admin/words', data)

export const deleteWord = (id: number) => del<null>(`/admin/words/${id}`)

export const adminFetchBooks = () => get<Book[]>('/admin/books')

export const createBook = (data: Omit<Book, 'id' | 'createdAt' | 'updatedAt' | 'chapterCount'>) =>
  post<Book>('/admin/books', data)

export const updateBook = (id: number, data: Omit<Book, 'id' | 'createdAt' | 'updatedAt' | 'chapterCount'>) =>
  put<Book>(`/admin/books/${id}`, data)

export const deleteBook = (id: number) => del<null>(`/admin/books/${id}`)

export const adminFetchChapters = (bookId: number | string) =>
  get<Chapter[]>(`/admin/books/${bookId}/chapters`)

export const createChapter = (data: {
  bookId: number
  title: string
  content: string
  translation: string
  sort: number
  status: 0 | 1
}) => post<Chapter>('/admin/chapters', data)

export const updateChapter = (
  id: number,
  data: { title: string; content: string; translation: string; sort: number; status: 0 | 1 },
) => put<Chapter>(`/admin/chapters/${id}`, data)

export const deleteChapter = (id: number) => del<null>(`/admin/chapters/${id}`)

export const batchCreateChapters = (data: {
  bookId: number
  chapters: { title: string; content: string; translation?: string }[]
}) => post<{ created: number }>('/admin/chapters/batch', data)

// ---- 分类 / 标签 ----
export const fetchCategories = () => get<Category[]>('/categories')
export const createCategory = (data: { name: string }) => post<Category>('/admin/categories', data)
export const updateCategory = (id: number, data: { name: string }) =>
  put<Category>(`/admin/categories/${id}`, data)
export const deleteCategory = (id: number) => del<null>(`/admin/categories/${id}`)

export const fetchTags = () => get<Tag[]>('/tags')
export const createTag = (data: { name: string }) => post<Tag>('/admin/tags', data)
export const updateTag = (id: number, data: { name: string }) =>
  put<Tag>(`/admin/tags/${id}`, data)
export const deleteTag = (id: number) => del<null>(`/admin/tags/${id}`)

// ---- 友链 / 单页 ----
export const fetchLinks = () => get<LinkItem[]>('/links')
export const createLink = (data: Omit<LinkItem, 'id'>) => post<LinkItem>('/admin/links', data)
export const updateLink = (id: number, data: Omit<LinkItem, 'id'>) =>
  put<LinkItem>(`/admin/links/${id}`, data)
export const deleteLink = (id: number) => del<null>(`/admin/links/${id}`)

export const fetchPage = (slug: string) => get<SinglePage>(`/pages/${slug}`)
export const adminFetchPages = () => get<SinglePage[]>('/admin/pages')
export const createPage = (data: { slug: string; title: string; content: string }) =>
  post<SinglePage>('/admin/pages', data)
export const updatePage = (id: number, data: { slug: string; title: string; content: string }) =>
  put<SinglePage>(`/admin/pages/${id}`, data)
export const deletePage = (id: number) => del<null>(`/admin/pages/${id}`)

// ---- 仪表盘 / 上传 ----
export const fetchDashboard = () => get<DashboardStats>('/admin/dashboard')

export async function uploadImage(file: File): Promise<string> {
  const fd = new FormData()
  fd.append('file', file)
  const data = await post<{ url: string }>('/admin/upload', fd)
  return data.url
}

// ---- AI 工具导航 ----
export const fetchAiTools = (params?: { section?: string }) =>
  get<AiToolItem[]>('/ai-tools', params)

export const fetchAiPrompts = (params?: { section?: string }) =>
  get<AiPromptItem[]>('/ai-prompts', params)

export const adminFetchAiTools = () => get<AiToolItem[]>('/admin/ai-tools')

export const createAiTool = (data: {
  name: string
  url: string
  icon: string
  section: string
  category: string
  description: string
  tags: string
  sort: number
  status: 0 | 1
}) => post<AiToolItem>('/admin/ai-tools', data)

export const updateAiTool = (id: number, data: {
  name: string
  url: string
  icon: string
  section: string
  category: string
  description: string
  tags: string
  sort: number
  status: 0 | 1
}) => put<AiToolItem>(`/admin/ai-tools/${id}`, data)

export const deleteAiTool = (id: number) => del<null>(`/admin/ai-tools/${id}`)

export const adminFetchAiPrompts = () => get<AiPromptItem[]>('/admin/ai-prompts')

export const createAiPrompt = (data: {
  title: string
  section: string
  category: string
  content: string
  description: string
  sort: number
  status: 0 | 1
}) => post<AiPromptItem>('/admin/ai-prompts', data)

export const updateAiPrompt = (id: number, data: {
  title: string
  section: string
  category: string
  content: string
  description: string
  sort: number
  status: 0 | 1
}) => put<AiPromptItem>(`/admin/ai-prompts/${id}`, data)

export const deleteAiPrompt = (id: number) => del<null>(`/admin/ai-prompts/${id}`)

// ---- 游戏板块 ----
export const fetchGames = (params?: { category?: string }) =>
  get<GameItem[]>('/games', params)

export const adminFetchGames = () => get<GameItem[]>('/admin/games')

export const createGame = (data: {
  title: string
  cover: string
  category: string
  platform: string
  description: string
  tags: string
  hot: number
  sort: number
  status: 0 | 1
}) => post<GameItem>('/admin/games', data)

export const updateGame = (id: number, data: {
  title: string
  cover: string
  category: string
  platform: string
  description: string
  tags: string
  hot: number
  sort: number
  status: 0 | 1
}) => put<GameItem>(`/admin/games/${id}`, data)

export const deleteGame = (id: number) => del<null>(`/admin/games/${id}`)

// ---- 站点配置 ----
export const fetchSettings = () => get<Record<string, string>>('/settings')

export const saveSettings = (data: Record<string, string>) =>
  put<null>('/admin/settings', data)

// ---- 动漫排行 ----
export const fetchAnimes = () => get<AnimeItem[]>('/animes')

export const adminFetchAnimes = () => get<AnimeItem[]>('/admin/animes')

export const createAnime = (data: {
  title: string
  cover: string
  region: string
  episodes: string
  description: string
  rank: number
  status: 0 | 1
}) => post<AnimeItem>('/admin/animes', data)

export const updateAnime = (id: number, data: {
  title: string
  cover: string
  region: string
  episodes: string
  description: string
  rank: number
  status: 0 | 1
}) => put<AnimeItem>(`/admin/animes/${id}`, data)

export const deleteAnime = (id: number) => del<null>(`/admin/animes/${id}`)

// ---- 艺术鉴赏 ----
export const fetchArts = () => get<ArtworkItem[]>('/arts')

export const adminFetchArts = () => get<ArtworkItem[]>('/admin/arts')

export const createArt = (data: {
  title: string
  image: string
  description: string
  author: string
  sort: number
  status: 0 | 1
}) => post<ArtworkItem>('/admin/arts', data)

export const updateArt = (id: number, data: {
  title: string
  image: string
  description: string
  author: string
  sort: number
  status: 0 | 1
}) => put<ArtworkItem>(`/admin/arts/${id}`, data)

export const deleteArt = (id: number) => del<null>(`/admin/arts/${id}`)

export const reorderAnimes = (ids: number[]) =>
  post<null>('/admin/animes/reorder', { ids })
