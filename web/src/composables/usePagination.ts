import { computed, ref, watch, type Ref } from 'vue'

/**
 * 后台客户端分页：传入完整列表，返回当前页数据
 * 列表变化（删除/重载）时自动把页码钳制回有效范围
 */
export function usePagination<T>(list: Ref<T[]>, size = 10) {
  const page = ref(1)
  const pagedList = computed(() => list.value.slice((page.value - 1) * size, page.value * size))
  watch(
    () => list.value.length,
    (len) => {
      const max = Math.max(1, Math.ceil(len / size))
      if (page.value > max) page.value = max
    },
  )
  return { page, pagedList }
}
