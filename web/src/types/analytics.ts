export interface AnalyticsGroup {
  key: string
  title: string
  pv: number
  uv: number
  seconds: number
  averageSeconds: number
  share: number
}
export interface AnalyticsData {
  metrics: {
    pv: number
    uv: number
    sessions: number
    seconds: number
    averageSeconds: number
  }
  trend: { time: string; pv: number; uv: number }[]
  modules: AnalyticsGroup[]
  pages: AnalyticsGroup[]
  sources: AnalyticsGroup[]
  devices: AnalyticsGroup[]
  actions: { action: string; target: string; count: number; uv: number }[]
  startedAt: string
  timezone: string
  retentionDays: number
}
