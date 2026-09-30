<script setup lang="ts">
import { onMounted, onUnmounted, ref, watch } from 'vue'
import Matter from 'matter-js'

/**
 * 右侧萌宠挂绳：Matter.js 物理（长绳从页面顶部垂下 + 拖拽回弹 + 睡眠省电）
 * - 物理坐标为整个视口：绳子 SVG 铺满全屏，拖到哪都不会被裁剪
 * - 悬空停在右下角（页面 70% 高度处），松手后加大阻尼快速摆回原位
 * - 关闭后记忆（localStorage），窄屏（<1240px）自动隐藏
 */

const CLOSED_KEY = 'petRopeClosed'
const MIN_W = 1240
const STRIP_W = 240 // 右侧物理区域宽度（拖拽边界）
const LINKS = 6
const PET_W = 84
const PET_H = 80

const closed = ref(localStorage.getItem(CLOSED_KEY) === '1')
const wide = ref(window.innerWidth >= MIN_W)
const ropePath = ref('')
const petTransform = ref('')
const dragging = ref(false)

let engine: Matter.Engine | null = null
let petBody: Matter.Body | null = null
let linkBodies: Matter.Body[] = []
let cons: Matter.Constraint[] = []
let wallBodies: Matter.Body[] = []
let rafId = 0
let mouse = { x: 0, y: 0 }
let isDragging = false

let anchor = { x: 0, y: 64 } // 视口坐标
let stripLeft = 0
let restY = 0 // 悬空静止中心 y（页面 70% 高度处）
let segLen = 0

const reduced = window.matchMedia('(prefers-reduced-motion: reduce)').matches

function computeLayout() {
  stripLeft = window.innerWidth - STRIP_W
  anchor = { x: stripLeft + STRIP_W - 90, y: 64 }
  restY = window.innerHeight * 0.7
  segLen = (restY - anchor.y) / (LINKS + 1)
}

function close() {
  closed.value = true
  localStorage.setItem(CLOSED_KEY, '1')
  destroy()
}

function reopen() {
  closed.value = false
  localStorage.removeItem(CLOSED_KEY)
  if (wide.value && !reduced) initPhysics()
  else renderStatic()
}

// ---------- 物理世界 ----------
function initPhysics() {
  computeLayout()
  const eng = Matter.Engine.create({ enableSleeping: true })
  eng.gravity.y = 1
  eng.positionIterations = 10
  eng.constraintIterations = 10 // 让绳链近似刚性，悬空时绷直不松弛
  engine = eng

  // 绳链：锚点（页面顶部）+ 若干小圆环
  const anchorBody = Matter.Bodies.circle(anchor.x, anchor.y, 3, { isStatic: true })
  linkBodies = []
  const parts = [anchorBody]
  for (let i = 1; i <= LINKS; i++) {
    const b = Matter.Bodies.circle(anchor.x, anchor.y + i * segLen, 2.5, {
      frictionAir: 0.03,
      render: { visible: false },
    })
    linkBodies.push(b)
    parts.push(b)
  }

  // 宠物刚体：初始从导航下方掉落，被长绳拉住荡到右下角悬空停稳
  petBody = Matter.Bodies.rectangle(anchor.x, anchor.y + 60, PET_W, PET_H, {
    chamfer: { radius: 16 },
    density: 0.0012,
    frictionAir: 0.02,
    restitution: 0,
  })

  cons = []
  const all = [anchorBody, ...linkBodies, petBody]
  for (let i = 0; i < all.length - 1; i++) {
    cons.push(
      Matter.Constraint.create({
        bodyA: all[i],
        bodyB: all[i + 1],
        length: segLen,
        stiffness: 1, // 刚性绳：悬空时绷直
        damping: 0.1,
        pointB: i === all.length - 2 ? { x: 0, y: -PET_H / 2 } : undefined,
      }),
    )
  }

  // 左右墙：拖拽边界（拖拽坐标另有钳制兜底）
  wallBodies = makeWalls()

  Matter.Composite.add(eng.world, [...parts, petBody, ...cons, ...wallBodies])

  // 拖拽时每帧把宠物固定到鼠标位置（已钳制在区域内）
  Matter.Events.on(eng, 'beforeUpdate', () => {
    if (isDragging && petBody) {
      Matter.Body.setPosition(petBody, { x: mouse.x, y: mouse.y })
      Matter.Body.setVelocity(petBody, { x: 0, y: 0 })
      Matter.Body.setAngularVelocity(petBody, 0)
      Matter.Sleeping.set(petBody, false)
    }
  })

  rafId = requestAnimationFrame(frame)
}

function makeWalls(): Matter.Body[] {
  const h = window.innerHeight
  const opts = { isStatic: true }
  return [
    Matter.Bodies.rectangle(stripLeft - 40, h / 2, 80, h * 4, opts), // 左墙
    Matter.Bodies.rectangle(stripLeft + STRIP_W + 40, h / 2, 80, h * 4, opts), // 右墙
    Matter.Bodies.rectangle(stripLeft + STRIP_W / 2, h + 500, STRIP_W * 6, 80, opts), // 兜底（放得很远）
  ]
}

function frame() {
  rafId = requestAnimationFrame(frame)
  if (engine) {
    if (!document.hidden) Matter.Engine.update(engine, 1000 / 60)
    syncDom()
  }
}

function syncDom() {
  if (!engine || !petBody) return
  const pts: Matter.Vector[] = [anchor]
  for (const l of linkBodies) pts.push(l.position)
  const petTop = {
    x: petBody.position.x + Math.sin(petBody.angle) * -(PET_H / 2),
    y: petBody.position.y - Math.cos(petBody.angle) * (PET_H / 2),
  }
  pts.push(petTop)
  ropePath.value = 'M' + pts.map((p) => `${p.x.toFixed(1)},${p.y.toFixed(1)}`).join(' L')

  petTransform.value = `translate(${(petBody.position.x - PET_W / 2).toFixed(1)}px, ${(
    petBody.position.y -
    PET_H / 2
  ).toFixed(1)}px) rotate(${petBody.angle.toFixed(3)}rad)`
}

// ---------- 静态渲染（减少动态偏好时：宠物悬空停在右下角） ----------
function renderStatic() {
  computeLayout()
  ropePath.value = `M${anchor.x},${anchor.y} L${anchor.x},${restY - PET_H / 2}`
  petTransform.value = `translate(${anchor.x - PET_W / 2}px, ${restY - PET_H / 2}px)`
}

// ---------- 拖拽（指针事件 → 视口坐标，并钳制在边界内） ----------
function clampMouse() {
  mouse.x = Math.min(Math.max(mouse.x, stripLeft + 12), stripLeft + STRIP_W - 12)
  mouse.y = Math.min(Math.max(mouse.y, anchor.y + 20), window.innerHeight - 12)
}

function onPetDown(e: PointerEvent) {
  if (!engine) return
  isDragging = true
  dragging.value = true
  mouse = { x: e.clientX, y: e.clientY }
  clampMouse()
  ;(e.currentTarget as HTMLElement).setPointerCapture(e.pointerId)
  for (const b of linkBodies) Matter.Sleeping.set(b, false)
  Matter.Sleeping.set(petBody!, false)
}
function onPetMove(e: PointerEvent) {
  if (!isDragging) return
  mouse = { x: e.clientX, y: e.clientY }
  clampMouse()
}
function onPetUp() {
  if (!isDragging) return
  isDragging = false
  dragging.value = false
  // 松手后加大空气阻力，让钟摆快速衰减、摆回原悬空位
  if (petBody) petBody.frictionAir = 0.12
  linkBodies.forEach((b) => (b.frictionAir = 0.08))
  const restore = () => {
    const v = petBody ? Math.abs(petBody.velocity.x) + Math.abs(petBody.velocity.y) : 0
    if (v < 0.06) {
      if (petBody) petBody.frictionAir = 0.02
      linkBodies.forEach((b) => (b.frictionAir = 0.03))
      return
    }
    requestAnimationFrame(restore)
  }
  requestAnimationFrame(restore)
}

// ---------- 生命周期 ----------
function onResize() {
  wide.value = window.innerWidth >= MIN_W
  if (engine && wallBodies.length) {
    // 重算布局：绳长、锚点、边界全部复位到右下角悬空位
    computeLayout()
    for (const c of cons) c.length = segLen
    linkBodies.forEach((b, i) => {
      Matter.Body.setPosition(b, { x: anchor.x, y: anchor.y + (i + 1) * segLen })
      Matter.Body.setVelocity(b, { x: 0, y: 0 })
    })
    if (petBody) {
      Matter.Body.setPosition(petBody, { x: anchor.x, y: restY })
      Matter.Body.setVelocity(petBody, { x: 0, y: 0 })
      Matter.Body.setAngle(petBody, 0)
    }
    for (const w of wallBodies) Matter.Composite.remove(engine.world, w)
    wallBodies = makeWalls()
    Matter.Composite.add(engine.world, wallBodies)
  }
}

function destroy() {
  cancelAnimationFrame(rafId)
  rafId = 0
  if (engine) {
    Matter.Events.off(engine, 'beforeUpdate')
    Matter.Composite.clear(engine.world, false)
    Matter.Engine.clear(engine)
    engine = null
  }
}

onMounted(() => {
  window.addEventListener('resize', onResize)
  if (reduced) {
    renderStatic()
    return
  }
  if (wide.value) initPhysics()
  else renderStatic()
})

onUnmounted(() => {
  window.removeEventListener('resize', onResize)
  destroy()
})
</script>

<template>
  <!-- 关闭后：右缘小爪印，点击找回 -->
  <button v-if="closed && wide" class="reopen" title="找回小猫" @click="reopen">🐾</button>

  <div v-else-if="wide" class="pet-strip">
    <!-- 关闭按钮 -->
    <button class="pet-close" title="收起小猫" @click="close">✕</button>

    <!-- 绳子（铺满视口，拖到哪都可见） -->
    <svg class="rope-svg">
      <path :d="ropePath" stroke="currentColor" stroke-width="2" stroke-linecap="round" fill="none" />
      <circle :cx="anchor.x" :cy="anchor.y" r="3.5" fill="currentColor" />
    </svg>

    <!-- 宠物（物理体跟随） -->
    <div
      class="pet"
      :class="{ grabbing: dragging }"
      :style="{ width: PET_W + 'px', height: PET_H + 'px', transform: petTransform }"
      title="拖拽我玩"
      @pointerdown="onPetDown"
      @pointermove="onPetMove"
      @pointerup="onPetUp"
      @pointercancel="onPetUp"
    >
      <!-- 橘白小猫 SVG -->
      <svg viewBox="0 0 100 96" class="pet-svg">
        <path
          d="M78 66 Q94 62 90 46"
          stroke="#f0973f"
          stroke-width="7"
          stroke-linecap="round"
          fill="none"
        />
        <ellipse cx="50" cy="64" rx="30" ry="24" fill="#ffffff" stroke="#3b3630" stroke-width="2" />
        <path d="M32 52 Q40 44 50 48" stroke="#f0973f" stroke-width="6" stroke-linecap="round" fill="none" />
        <circle cx="50" cy="30" r="20" fill="#f9a03f" stroke="#3b3630" stroke-width="2" />
        <path d="M34 16 L32 4 L44 10 Z" fill="#f9a03f" stroke="#3b3630" stroke-width="2" stroke-linejoin="round" />
        <path d="M66 16 L68 4 L56 10 Z" fill="#f9a03f" stroke="#3b3630" stroke-width="2" stroke-linejoin="round" />
        <path d="M35.5 12 L35 7.5 L40 10 Z" fill="#ffd9b8" />
        <path d="M64.5 12 L65 7.5 L60 10 Z" fill="#ffd9b8" />
        <path d="M46 12 L46 17 M52 11 L52 16" stroke="#e07b1f" stroke-width="2.4" stroke-linecap="round" />
        <path d="M40 30 Q43 27 46 30" stroke="#26211b" stroke-width="2" stroke-linecap="round" fill="none" />
        <path d="M56 30 Q59 27 62 30" stroke="#26211b" stroke-width="2" stroke-linecap="round" fill="none" />
        <path d="M48.5 35 L51.5 35 L50 37 Z" fill="#3b3630" />
        <path d="M50 37 Q47 40 44.5 38 M50 37 Q53 40 55.5 38" stroke="#3b3630" stroke-width="1.4" fill="none" stroke-linecap="round" />
        <path d="M30 32 L20 30 M30 36 L21 37" stroke="#3b3630" stroke-width="1.1" stroke-linecap="round" />
        <path d="M70 32 L80 30 M70 36 L79 37" stroke="#3b3630" stroke-width="1.1" stroke-linecap="round" />
        <circle cx="38" cy="50" r="7" fill="#ffffff" stroke="#3b3630" stroke-width="2" />
        <circle cx="62" cy="50" r="7" fill="#ffffff" stroke="#3b3630" stroke-width="2" />
      </svg>
    </div>
  </div>
</template>

<style scoped>
.pet-strip {
  position: fixed;
  inset: 0;
  z-index: 5;
  pointer-events: none;
  color: var(--scroll-thumb);
  display: none;
}

@media (min-width: 1240px) {
  .pet-strip {
    display: block;
  }
}

.rope-svg {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  pointer-events: none;
}

.pet {
  position: absolute;
  left: 0;
  top: 0;
  pointer-events: auto;
  cursor: grab;
  touch-action: none;
  filter: drop-shadow(0 6px 10px rgb(0 0 0 / 0.12));
  will-change: transform;
}

.pet.grabbing {
  cursor: grabbing;
}

.pet-svg {
  width: 100%;
  height: 100%;
  display: block;
}

.pet-close {
  position: absolute;
  top: 20px;
  right: 14px;
  width: 22px;
  height: 22px;
  border-radius: 50%;
  border: 1px solid var(--border);
  background: var(--card);
  color: var(--muted);
  font-size: 11px;
  line-height: 1;
  cursor: pointer;
  pointer-events: auto;
  transition: all 0.15s;
}

.pet-close:hover {
  color: var(--accent);
  border-color: var(--accent);
}

.reopen {
  position: fixed;
  right: 6px;
  top: 45%;
  border: none;
  background: transparent;
  font-size: 18px;
  opacity: 0.35;
  cursor: pointer;
  pointer-events: auto;
  transition: opacity 0.2s;
}

.reopen:hover {
  opacity: 1;
}
</style>
