<script setup lang="ts">
/**
 * 左侧装饰：淡淡的手绘星星（纯 SVG + CSS，无物理）
 * - 窄屏（<1240px）自动隐藏
 * - 尊重系统"减少动态效果"偏好（静态显示）
 */
const prefersReduced = window.matchMedia('(prefers-reduced-motion: reduce)').matches

interface Star {
  x: number // 百分比（相对左侧装饰带宽度）
  y: number // 视口百分比
  size: number
  delay: number
  rotate: number
}

const stars: Star[] = [
  { x: 30, y: 16, size: 22, delay: 0, rotate: -12 },
  { x: 62, y: 27, size: 14, delay: 1.2, rotate: 18 },
  { x: 22, y: 42, size: 17, delay: 2.1, rotate: 8 },
  { x: 55, y: 55, size: 12, delay: 0.6, rotate: -20 },
  { x: 34, y: 68, size: 20, delay: 1.7, rotate: 14 },
  { x: 64, y: 80, size: 13, delay: 2.8, rotate: -6 },
]
</script>

<template>
  <div class="star-field" aria-hidden="true">
    <svg
      v-for="(s, i) in stars"
      :key="i"
      class="star"
      :class="{ twinkle: !prefersReduced }"
      :style="{
        left: s.x + '%',
        top: s.y + 'vh',
        width: s.size + 'px',
        height: s.size + 'px',
        transform: `rotate(${s.rotate}deg)`,
        animationDelay: s.delay + 's',
      }"
      viewBox="0 0 24 24"
      fill="none"
    >
      <!-- 手绘感四角星光 -->
      <path
        d="M12 2 C12.6 7 13.5 9.5 22 12 C13.5 14.5 12.6 17 12 22 C11.4 17 10.5 14.5 2 12 C10.5 9.5 11.4 7 12 2 Z"
        stroke="currentColor"
        stroke-width="1.4"
        stroke-linejoin="round"
      />
    </svg>
  </div>
</template>

<style scoped>
.star-field {
  position: fixed;
  left: 0;
  top: 0;
  bottom: 0;
  width: 230px;
  pointer-events: none;
  color: var(--accent);
  z-index: 5;
  display: none;
}

@media (min-width: 1240px) {
  .star-field {
    display: block;
  }
}

.star {
  position: absolute;
  opacity: 0.4;
}

.star.twinkle {
  animation: twinkle 4.5s ease-in-out infinite;
}

@keyframes twinkle {
  0%,
  100% {
    opacity: 0.3;
    scale: 0.92;
  }
  50% {
    opacity: 0.9;
    scale: 1.08;
  }
}

@media (prefers-reduced-motion: reduce) {
  .star.twinkle {
    animation: none;
  }
}
</style>
