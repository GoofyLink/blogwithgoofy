import type { AnimeForm } from '@/api'
export const watchStatuses = [
  { value: 'planned', label: '想看' },
  { value: 'watching', label: '在看' },
  { value: 'completed', label: '已看完' },
  { value: 'paused', label: '暂时搁置' },
  { value: 'dropped', label: '弃番' },
]
export const airStatuses = [
  { value: 'upcoming', label: '未开播' },
  { value: 'airing', label: '连载中' },
  { value: 'finished', label: '已完结' },
]
export const watchLabel = (value: string) =>
  watchStatuses.find((s) => s.value === value)?.label || '未记录'
export const airLabel = (value: string) =>
  airStatuses.find((s) => s.value === value)?.label || '未记录'
export const progressLabel = (a: {
  watched: number
  totalEpisodes: number | null
}) => `${a.watched} / ${a.totalEpisodes ?? '?'} 集`
export const newAnimeForm = (): AnimeForm => ({
  title: '',
  cover: '',
  category: '',
  region: '',
  description: '',
  rank: 0,
  status: 1,
  watchStatus: '',
  airStatus: '',
  watched: 0,
  totalEpisodes: null,
  rating: null,
  year: 0,
  source: '',
  review: '',
  content: '',
  spoiler: false,
  recommended: false,
  recommendOrder: 0,
  watchUrl: '',
  watchPlatform: '',
})
