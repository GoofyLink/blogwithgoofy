<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { fetchAnalytics, fetchDashboard } from '@/api'
import type { AnalyticsData } from '@/types/analytics'
import type { DashboardStats } from '@/types'
import AnalyticsTrend from '@/components/AnalyticsTrend.vue'
import { analyticsDebugState, setAnalyticsDebug } from '@/utils/analytics'
const day = (offset = 0) =>
  new Date(Date.now() + 8 * 3600000 + offset * 86400000)
    .toISOString()
    .slice(0, 10)
const dates = ref<[string, string]>([day(-6), day()])
const module = ref(''),
  includeAdmin = ref(false),
  sort = ref<'uv' | 'pv' | 'averageSeconds'>('uv')
const data = ref<AnalyticsData | null>(null),
  content = ref<DashboardStats | null>(null)
const loading = ref(false),
  failed = ref(false),
  debug = ref(analyticsDebugState())
let version = 0
const duration = (v: number) =>
  v < 60
    ? `${Math.round(v)} 秒`
    : `${Math.floor(v / 60)} 分 ${Math.round(v % 60)} 秒`
const modules = computed(() =>
  [...(data.value?.modules || [])].sort(
    (a, b) => b[sort.value] - a[sort.value],
  ),
)
const max = computed(() =>
  Math.max(1, ...modules.value.map((r) => r[sort.value])),
)
const labels: Record<string, string> = {
  home: '首页',
  essays: '随笔',
  learn: '学习',
  novel: '小说',
  anime: '动漫',
  game: '游戏',
  art: '艺术',
  ai: 'AI',
  archives: '归档',
  links: '友链',
  about: '关于',
  page: '单页',
}
const actions: Record<string, string> = {
  ai_tool: 'AI 工具跳转',
  watch_link: '正版观看链接',
  outbound: '其他外链',
}
const metrics = computed(() =>
  data.value
    ? [
        { label: '浏览次数 PV', value: data.value.metrics.pv },
        { label: '独立访客 UV', value: data.value.metrics.uv },
        { label: '访问会话', value: data.value.metrics.sessions },
        {
          label: '平均有效停留',
          value: duration(data.value.metrics.averageSeconds),
        },
      ]
    : [],
)
const contentCards = [
  { key: 'articleCount', label: '文章' },
  { key: 'bookCount', label: '书籍' },
  { key: 'chapterCount', label: '章节' },
  { key: 'commentCount', label: '文章评论' },
] as const
async function load() {
  if (!dates.value?.[0] || !dates.value?.[1]) return
  const v = ++version
  loading.value = true
  failed.value = false
  try {
    const result = await fetchAnalytics({
      start: dates.value[0],
      end: dates.value[1],
      module: module.value,
      includeAdmin: includeAdmin.value,
    })
    if (v === version) data.value = result
  } catch {
    if (v === version) failed.value = true
  } finally {
    if (v === version) loading.value = false
  }
}
function range(days: number, yesterday = false) {
  dates.value = yesterday ? [day(-1), day(-1)] : [day(1 - days), day()]
  void load()
}
function drillRow(row: { key: string }) {
  drill(row.key)
}
function drill(key: string) {
  module.value = key
  void load()
}
function updateDebug() {
  setAnalyticsDebug(debug.value.enabled, debug.value.includeAdmin)
}
function disabledDate(d: Date) {
  const value =
    d.getFullYear() +
    '-' +
    String(d.getMonth() + 1).padStart(2, '0') +
    '-' +
    String(d.getDate()).padStart(2, '0')
  return value > day() || value < day(-364)
}
onMounted(() => {
  void load()
  void fetchDashboard()
    .then((r) => {
      content.value = r
    })
    .catch(() => {})
})
</script>
<template>
  <div class="analytics-dashboard">
    <header class="heading">
      <div>
        <h1>访问与内容</h1>
        <p>了解读者在看什么，以及哪些内容值得继续写。</p>
      </div>
      <router-link to="/admin/articles/new"
        ><el-button type="primary">写新文章</el-button></router-link
      >
    </header>
    <div class="card filters">
      <div class="shortcuts">
        <el-button @click="range(1)">今天</el-button
        ><el-button @click="range(1, true)">昨天</el-button
        ><el-button @click="range(7)">近 7 天</el-button
        ><el-button @click="range(30)">近 30 天</el-button>
      </div>
      <el-date-picker
        v-model="dates"
        type="daterange"
        value-format="YYYY-MM-DD"
        start-placeholder="开始日期"
        end-placeholder="结束日期"
        :clearable="false"
        :disabled-date="disabledDate"
        @change="load"
      /><el-select
        v-model="module"
        placeholder="全部模块"
        clearable
        aria-label="统计模块"
        @change="load"
        ><el-option
          v-for="(label, key) in labels"
          :key="key"
          :value="key"
          :label="label" /></el-select
      ><el-checkbox v-model="includeAdmin" @change="load"
        >包含管理员测试访问</el-checkbox
      ><el-button :loading="loading" @click="load">刷新</el-button>
    </div>
    <div v-if="failed" class="card error" role="alert">
      统计加载失败，请确认后端已启动新接口。<el-button @click="load"
        >重新加载</el-button
      >
    </div>
    <div v-else v-loading="loading">
      <template v-if="data">
        <p class="scope">
          {{ dates[0] }} 至 {{ dates[1] }} · 北京时间 ·
          {{ module ? labels[module] : '全站'
          }}<el-button v-if="module" text @click="drill('')"
            >返回全站</el-button
          >
        </p>
        <div class="metric-grid">
          <section v-for="m in metrics" :key="m.label" class="card metric">
            <span>{{ m.label }}</span
            ><strong>{{ m.value }}</strong>
          </section>
        </div>
        <p v-if="!data.metrics.pv" class="empty-note">
          所选范围暂无访问。统计从启用后开始记录；开发环境和管理员浏览默认不计入。
        </p>
        <AnalyticsTrend :rows="data.trend" />
        <section class="card panel">
          <header>
            <div>
              <h3>模块访问排行</h3>
              <p>点击模块查看该栏目内容；占比按浏览次数计算。</p>
            </div>
            <el-select v-model="sort" aria-label="模块排序"
              ><el-option value="uv" label="按访客数" /><el-option
                value="pv"
                label="按浏览次数" /><el-option
                value="averageSeconds"
                label="按平均停留"
            /></el-select>
          </header>
          <el-table :data="modules" @row-click="drillRow" class="module-table"
            ><el-table-column label="模块" min-width="220"
              ><template #default="{ row }"
                ><button class="module-button" @click.stop="drill(row.key)">
                  {{ row.title }}
                </button>
                <div class="bar">
                  <span
                    :style="{ width: `${(row[sort] / max) * 100}%` }"
                  /></div></template></el-table-column
            ><el-table-column
              prop="uv"
              label="访客 UV"
              sortable
              width="115"
            /><el-table-column
              prop="pv"
              label="浏览 PV"
              sortable
              width="115"
            /><el-table-column label="PV 占比" width="100"
              ><template #default="{ row }"
                >{{ row.share.toFixed(1) }}%</template
              ></el-table-column
            ><el-table-column label="平均有效停留" min-width="135"
              ><template #default="{ row }">{{
                duration(row.averageSeconds)
              }}</template></el-table-column
            ></el-table
          >
        </section>
        <section class="card panel">
          <header>
            <div>
              <h3>热门页面与内容</h3>
              <p>按独立访客排序，最多展示 100 个页面。</p>
            </div>
          </header>
          <el-table :data="data.pages" max-height="460"
            ><el-table-column label="内容" min-width="240"
              ><template #default="{ row }"
                ><router-link :to="row.key" target="_blank">{{
                  row.title
                }}</router-link
                ><small class="path">{{ row.key }}</small></template
              ></el-table-column
            ><el-table-column
              prop="uv"
              label="访客 UV"
              sortable
              width="115"
            /><el-table-column
              prop="pv"
              label="浏览 PV"
              sortable
              width="115"
            /><el-table-column label="平均有效停留" min-width="135"
              ><template #default="{ row }">{{
                duration(row.averageSeconds)
              }}</template></el-table-column
            ></el-table
          >
        </section>
        <div class="split">
          <section class="card panel">
            <h3>访问来源</h3>
            <el-table :data="data.sources"
              ><el-table-column
                prop="key"
                label="来源站点"
                min-width="180" /><el-table-column
                prop="uv"
                label="UV"
                width="70" /><el-table-column prop="pv" label="PV" width="70"
            /></el-table>
          </section>
          <section class="card panel">
            <h3>设备分布</h3>
            <el-table :data="data.devices"
              ><el-table-column prop="key" label="设备" /><el-table-column
                prop="uv"
                label="UV"
                width="70"
              /><el-table-column label="PV 占比" width="100"
                ><template #default="{ row }"
                  >{{ row.share.toFixed(1) }}%</template
                ></el-table-column
              ></el-table
            >
          </section>
        </div>
        <section class="card panel">
          <h3>关键链接点击</h3>
          <el-table :data="data.actions"
            ><el-table-column label="操作" min-width="160"
              ><template #default="{ row }">{{
                actions[row.action] || row.action
              }}</template></el-table-column
            ><el-table-column
              prop="target"
              label="目标站点"
              min-width="200" /><el-table-column
              prop="count"
              label="点击次数"
              width="100" /><el-table-column
              prop="uv"
              label="点击访客"
              width="100"
          /></el-table>
        </section>
        <details class="card notes">
          <summary>统计口径与调试</summary>
          <p>
            启用时间：{{
              new Date(data.startedAt).toLocaleString('zh-CN', {
                timeZone: 'Asia/Shanghai',
              })
            }}。页面与详情成功加载后记录 PV；刷新计次，筛选和锚点不重复计次。UV
            按随机浏览器标识在所选时间内去重，各模块 UV 不能相加当作全站 UV。
          </p>
          <p>
            会话以 30 分钟无活动为界；平均有效停留 = 页面可见且 60
            秒内有操作的累计时长 ÷
            PV，按页面访问开始日期归属。来源取访问入口的站点域名，设备由浏览器信息估计。
          </p>
          <p>
            访问明细保留 90 天，匿名汇总保留 365
            天。浏览器拦截、关闭页面和网络中断可能造成少量漏报，机器人过滤不保证覆盖全部自动化流量。
          </p>
          <div class="debug">
            <el-checkbox
              v-if="debug.development"
              v-model="debug.enabled"
              @change="updateDebug"
              >此浏览器启用本地开发埋点</el-checkbox
            ><el-checkbox v-model="debug.includeAdmin" @change="updateDebug"
              >此浏览器上报管理员测试访问</el-checkbox
            >
          </div>
          <p>
            调试开关仅影响当前浏览器，开启后在前台新开页面测试；查看测试数据时，请勾选上方“包含管理员测试访问”。后台页面始终不计入。
          </p>
        </details>
      </template>
    </div>
    <section v-if="content" class="content-section">
      <h3>内容概况</h3>
      <div class="content-grid">
        <div v-for="c in contentCards" :key="c.key" class="card">
          <strong>{{ content[c.key] }}</strong
          ><span>{{ c.label }}</span>
        </div>
      </div>
    </section>
  </div>
</template>
<style scoped>
.analytics-dashboard {
  max-width: 1500px;
  margin: auto;
}
.heading {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
  margin-bottom: 22px;
}
h1 {
  font: 700 28px var(--font-heading);
  margin: 0 0 8px;
}
.heading p,
.scope {
  color: var(--muted);
  font-size: 13px;
}
.filters {
  padding: 16px;
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}
.filters > .el-select {
  width: 150px;
}
.shortcuts {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}
.shortcuts .el-button {
  margin: 0;
}
.metric-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  margin: 18px 0;
}
.metric {
  padding: 22px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.metric span {
  font-size: 13px;
  color: var(--muted);
}
.metric strong {
  font: 700 30px var(--font-heading);
}
.panel {
  padding: 22px;
  margin-top: 20px;
  min-width: 0;
}
.panel header {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  align-items: center;
  flex-wrap: wrap;
  margin-bottom: 12px;
}
.panel header > .el-select {
  width: 160px;
}
h3 {
  margin: 0 0 8px;
  font-size: 17px;
}
.panel header p {
  font-size: 12px;
  color: var(--muted);
  margin: 0;
}
.bar {
  height: 5px;
  background: var(--surface);
  border-radius: 3px;
  margin: 8px 0;
}
.bar span {
  height: 100%;
  display: block;
  background: var(--accent);
  border-radius: 3px;
}
.module-button {
  border: 0;
  background: none;
  padding: 0;
  color: var(--accent);
  cursor: pointer;
  font: inherit;
}
.module-table {
  cursor: pointer;
}
.path {
  display: block;
  color: var(--muted);
}
.split {
  display: grid;
  grid-template-columns: 1.4fr 1fr;
  gap: 20px;
}
.notes {
  padding: 20px;
  margin-top: 20px;
  color: var(--muted);
  font-size: 13px;
  line-height: 1.8;
}
.notes summary {
  cursor: pointer;
  color: var(--text);
}
.debug {
  display: flex;
  gap: 16px;
  flex-wrap: wrap;
}
.content-section {
  margin-top: 26px;
}
.content-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
}
.content-grid > div {
  padding: 18px;
  display: flex;
  align-items: baseline;
  gap: 12px;
}
.content-grid strong {
  font-size: 24px;
}
.content-grid span {
  color: var(--muted);
}
.error,
.empty-note {
  padding: 20px;
  color: var(--muted);
}
@media (max-width: 1000px) {
  .split {
    grid-template-columns: 1fr;
  }
  .metric-grid,
  .content-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}
@media (max-width: 600px) {
  .heading {
    align-items: flex-start;
  }
  .metric strong {
    font-size: 23px;
  }
  .panel {
    padding: 14px;
  }
  .filters :deep(.el-date-editor) {
    max-width: 100%;
  }
}
</style>
