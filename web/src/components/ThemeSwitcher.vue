<script setup lang="ts">
import { ref } from 'vue'
import { Check } from '@element-plus/icons-vue'
import { THEMES, applyTheme, currentTheme } from '@/themes'

const current = ref(currentTheme())

function set(key: string | number | object) {
  applyTheme(String(key))
  current.value = String(key)
}
</script>

<template>
  <el-dropdown trigger="click" @command="set">
    <span class="theme-trigger" title="切换主题">🎨</span>
    <template #dropdown>
      <el-dropdown-menu>
        <el-dropdown-item v-for="t in THEMES" :key="t.key" :command="t.key">
          <span class="theme-row">
            <span class="theme-texts">
              <span class="theme-label">{{ t.label }}</span>
              <span v-if="t.hint" class="theme-hint">{{ t.hint }}</span>
            </span>
            <el-icon v-if="current === t.key" class="theme-check"><Check /></el-icon>
          </span>
        </el-dropdown-item>
      </el-dropdown-menu>
    </template>
  </el-dropdown>
</template>

<style scoped>
.theme-trigger {
  cursor: pointer;
  font-size: 17px;
  line-height: 1;
  user-select: none;
  outline: none;
}

.theme-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 18px;
  min-width: 150px;
}

.theme-texts {
  display: flex;
  flex-direction: column;
  line-height: 1.4;
}

.theme-label {
  font-size: 13.5px;
}

.theme-hint {
  font-size: 11.5px;
  color: var(--muted);
}

.theme-check {
  color: var(--accent);
}
</style>
