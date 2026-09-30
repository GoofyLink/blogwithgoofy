<script setup lang="ts">
/**
 * 后台公共分页组件：统一样式（总数 + 上一页/页码/下一页）
 * 数据量不足一页时自动隐藏
 */
const props = withDefaults(defineProps<{ page: number; total: number; size?: number }>(), {
  size: 10,
})
const emit = defineEmits<{ 'update:page': [page: number] }>()
</script>

<template>
  <div v-if="total > size" class="admin-pagination">
    <el-pagination
      background
      layout="total, prev, pager, next"
      :total="total"
      :page-size="size"
      :current-page="page"
      @current-change="(p: number) => emit('update:page', p)"
    />
  </div>
</template>

<style scoped>
.admin-pagination {
  display: flex;
  justify-content: flex-end;
  margin-top: 14px;
}
</style>
