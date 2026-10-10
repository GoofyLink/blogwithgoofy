<script setup lang="ts">
import { computed, ref } from 'vue'
const props = defineProps<{
  rows: { time: string; pv: number; uv: number }[]
}>()
const selected = ref<number | null>(null)
const table = ref(false)
const max = computed(() => Math.max(1, ...props.rows.map((r) => r.pv)))
const x = (i: number) => 50 + (i * 820) / Math.max(1, props.rows.length - 1)
const y = (v: number) => 220 - (v / max.value) * 185
const line = (key: 'pv' | 'uv') =>
  props.rows.map((r, i) => `${x(i)},${y(r[key])}`).join(' ')
const point = computed(() =>
  selected.value == null ? null : props.rows[selected.value],
)
</script>
<template>
  <section class="card trend">
    <header>
      <h3>访问趋势</h3>
      <div>
        <span class="pv">● 浏览次数 PV</span
        ><span class="uv">● 独立访客 UV</span
        ><el-button text @click="table = !table">{{
          table ? '显示图表' : '查看数据表'
        }}</el-button>
      </div>
    </header>
    <div v-if="!table" class="chart">
      <svg
        viewBox="0 0 900 260"
        role="img"
        aria-label="访问趋势，实线为浏览次数，虚线为独立访客"
      >
        <g v-for="n in [0, 0.5, 1]" :key="n">
          <line
            x1="50"
            x2="870"
            :y1="y(max * n)"
            :y2="y(max * n)"
            class="grid"
          />
          <text x="42" :y="y(max * n) + 4" text-anchor="end">
            {{ Math.round(max * n) }}
          </text>
        </g>
        <polyline :points="line('pv')" class="pv-line" />
        <polyline :points="line('uv')" class="uv-line" />
        <g v-for="(r, i) in rows" :key="r.time">
          <circle
            :cx="x(i)"
            :cy="y(r.pv)"
            r="8"
            class="hit"
            tabindex="0"
            @mouseenter="selected = i"
            @focus="selected = i"
          >
            <title>{{ r.time }}：PV {{ r.pv }}，UV {{ r.uv }}</title>
          </circle>
          <text
            v-if="
              i === 0 ||
              i === rows.length - 1 ||
              i % Math.max(1, Math.ceil(rows.length / 6)) === 0
            "
            :x="x(i)"
            y="246"
            text-anchor="middle"
          >
            {{ r.time.length > 5 ? r.time.slice(5) : r.time }}
          </text>
        </g>
      </svg>
      <p aria-live="polite">
        {{
          point
            ? `${point.time}：浏览 ${point.pv} 次，访客 ${point.uv} 位`
            : '将鼠标移到数据点，或切换数据表查看详情'
        }}
      </p>
    </div>
    <el-table v-else :data="rows" max-height="320"
      ><el-table-column prop="time" label="时间" /><el-table-column
        prop="pv"
        label="浏览次数 PV" /><el-table-column prop="uv" label="独立访客 UV"
    /></el-table>
  </section>
</template>
<style scoped>
.trend {
  padding: 22px;
  min-width: 0;
}
header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  flex-wrap: wrap;
}
h3 {
  margin: 0;
  font-size: 17px;
}
header > div {
  display: flex;
  gap: 16px;
  align-items: center;
  flex-wrap: wrap;
  font-size: 12px;
}
.pv {
  color: var(--accent);
}
.uv {
  color: var(--muted);
}
svg {
  width: 100%;
  min-height: 180px;
  max-height: 340px;
}
svg text {
  fill: var(--muted);
  font-size: 11px;
}
.grid {
  stroke: var(--border);
}
.pv-line,
.uv-line {
  fill: none;
  stroke-width: 2.5;
  stroke: var(--accent);
}
.uv-line {
  stroke: var(--muted);
  stroke-dasharray: 6 4;
}
.hit {
  fill: transparent;
  cursor: crosshair;
}
.hit:hover,
.hit:focus {
  fill: var(--accent);
  outline: none;
}
.chart p {
  font-size: 12px;
  color: var(--muted);
  margin: 0;
}
</style>
