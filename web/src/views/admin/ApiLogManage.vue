<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import dayjs from 'dayjs'
import { ElMessage } from 'element-plus'
import { Search } from '@element-plus/icons-vue'
import { adminFetchApiLogs, clearApiLogs, fetchApiLogStats } from '@/api'

import type { ApiLogItem } from '@/types'

const list = ref<ApiLogItem[]>([])
const total = ref(0)
const page = ref(1)
const size = 20
const loading = ref(false)

const stats = reactive({ total: 0, today: 0, ips: 0 })

const filters = reactive({
  keyword: '',
  method: '',
  status: '',
})

async function load() {
  loading.value = true
  try {
    const data = await adminFetchApiLogs({
      page: page.value,
      size,
      keyword: filters.keyword || undefined,
      method: filters.method || undefined,
      status: filters.status || undefined,
    })
    list.value = data.list || []
    total.value = data.total
  } finally {
    loading.value = false
  }
}

async function loadStats() {
  const s = await fetchApiLogStats()
  stats.total = s.total
  stats.today = s.today
  stats.ips = s.ips
}

function search() {
  page.value = 1
  load()
  loadStats()
}

async function clearAll() {
  await clearApiLogs()
  ElMessage.success('日志已清空')
  load()
  loadStats()
}

const statusType = (s: number): string => {
  const code = s
  if (code >= 500) return 'danger'
  if (code >= 400) return 'warning'
  return 'success'
}

const fmtTime = (v: string): string => dayjs(v).format('YYYY-MM-DD HH:mm:ss')

onMounted(() => {
  load()
  loadStats()
})
</script>

<template>
  <div>
    <!-- 概览 -->
    <div class="stat-row">
      <div class="stat card"><span class="stat-num">{{ stats.total }}</span><span class="stat-label">总调用</span></div>
      <div class="stat card"><span class="stat-num">{{ stats.today }}</span><span class="stat-label">今日调用</span></div>
      <div class="stat card"><span class="stat-num">{{ stats.ips }}</span><span class="stat-label">独立 IP</span></div>
      <div class="stat tip-stat card"><span class="tip">异步落库 · 自动清理 7 天前记录 · 日志查询接口本身不记录</span></div>
    </div>

    <!-- 筛选 -->
    <div class="toolbar card">
      <el-input
        v-model="filters.keyword"
        placeholder="搜索路径或 IP…"
        clearable
        style="width: 260px"
        :prefix-icon="Search"
        @keyup.enter="search"
        @clear="search"
      />
      <el-select v-model="filters.method" placeholder="全部方法" clearable style="width: 120px" @change="search">
        <el-option label="GET" value="GET" />
        <el-option label="POST" value="POST" />
        <el-option label="PUT" value="PUT" />
        <el-option label="DELETE" value="DELETE" />
      </el-select>
      <el-select v-model="filters.status" placeholder="全部状态" clearable style="width: 140px" @change="search">
        <el-option label="2xx 成功" value="200" />
        <el-option label="401 未授权" value="401" />
        <el-option label="404 未找到" value="404" />
        <el-option label="500 服务器错误" value="500" />
      </el-select>
      <el-button type="danger" plain class="clear-btn" @click="clearAll">清空日志</el-button>
    </div>

    <el-table :data="list" v-loading="loading" stripe class="card table-card">
      <el-table-column label="时间" width="165">
        <template #default="{ row }">{{ fmtTime(row.createdAt) }}</template>
      </el-table-column>
      <el-table-column prop="method" label="方法" width="75">
        <template #default="{ row }">
          <el-tag size="small" :type="row.method === 'GET' ? 'info' : 'primary'" effect="plain">{{ row.method }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="path" label="接口路径" min-width="280" show-overflow-tooltip />
      <el-table-column prop="ip" label="IP 地址" width="130" />
      <el-table-column label="状态" width="85">
        <template #default="{ row }">
          <el-tag size="small" :type="statusType(row.status)">{{ row.status }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="durationMs" label="耗时" width="80">
        <template #default="{ row }">{{ row.durationMs }}ms</template>
      </el-table-column>
      <el-table-column prop="userAgent" label="User-Agent" min-width="200" show-overflow-tooltip />
    </el-table>

    <div class="pagination-wrap" v-if="total > size">
      <el-pagination
        background
        layout="total, prev, pager, next"
        :total="total"
        :page-size="size"
        :current-page="page"
        @current-change="(p: number) => { page = p; load() }"
      />
    </div>
  </div>
</template>

<style scoped>
.stat-row {
  display: grid;
  grid-template-columns: repeat(3, 140px) 1fr;
  gap: 12px;
  margin-bottom: 14px;
  align-items: stretch;
}

.stat {
  padding: 16px 20px;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.stat-num {
  font-family: var(--font-heading);
  font-size: 24px;
  font-weight: 800;
  color: var(--heading);
}

.stat-label {
  font-size: 12.5px;
  color: var(--muted);
}

.tip-stat {
  justify-content: center;
}

.tip {
  font-size: 12.5px;
  color: var(--muted);
}

.toolbar {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 16px;
  margin-bottom: 14px;
}

.clear-btn {
  margin-left: auto;
}

.table-card {
  width: 100%;
}

.pagination-wrap {
  display: flex;
  justify-content: flex-end;
  margin-top: 14px;
}

@media (max-width: 900px) {
  .stat-row {
    grid-template-columns: repeat(3, 1fr);
  }

  .tip-stat {
    display: none;
  }
}
</style>
