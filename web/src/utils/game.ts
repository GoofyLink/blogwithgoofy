export const playStatuses = [
  { value: 'wishlist', label: '想玩' },
  { value: 'playing', label: '在玩' },
  { value: 'completed', label: '已通关' },
  { value: 'paused', label: '暂时搁置' },
] as const
export const playLabel = (value: string) => playStatuses.find(s => s.value === value)?.label || '未记录'
export const splitGameTags = (value: string) => [...new Set((value || '').split(/[,，/、]/).map(s => s.trim()).filter(Boolean))]
