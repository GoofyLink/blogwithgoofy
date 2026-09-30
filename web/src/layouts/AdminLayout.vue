<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { changePassword } from '@/api'
import { useAuthStore } from '@/stores/auth'
import ThemeSwitcher from '@/components/ThemeSwitcher.vue'

const router = useRouter()
const route = useRoute()
const auth = useAuthStore()

const menu = [
  { path: '/admin/dashboard', title: '仪表盘', icon: 'Odometer' },
  { path: '/admin/profile', title: '个人信息', icon: 'User' },
  { path: '/admin/articles', title: '文章管理', icon: 'Document' },
  { path: '/admin/books', title: '书籍管理', icon: 'Reading' },
  { path: '/admin/animes', title: '动漫管理', icon: 'VideoPlay' },
  { path: '/admin/arts', title: '艺术鉴赏', icon: 'Picture' },
  { path: '/admin/ai', title: 'AI 百宝箱', icon: 'Cpu' },
  { path: '/admin/games', title: '游戏管理', icon: 'GamePad' },
  { path: '/admin/categories', title: '分类管理', icon: 'FolderOpened' },
  { path: '/admin/tags', title: '标签管理', icon: 'PriceTag' },
  { path: '/admin/comments', title: '评论管理', icon: 'ChatDotRound' },
  { path: '/admin/links', title: '友链管理', icon: 'Link' },
  { path: '/admin/pages', title: '单页管理', icon: 'Files' },
]

const pwdDialog = reactive({ visible: false, oldPassword: '', newPassword: '' })
const pwdSaving = ref(false)

async function savePwd() {
  if (!pwdDialog.oldPassword || pwdDialog.newPassword.length < 6) {
    ElMessage.warning('请填写原密码，新密码至少 6 位')
    return
  }
  pwdSaving.value = true
  try {
    await changePassword({ oldPassword: pwdDialog.oldPassword, newPassword: pwdDialog.newPassword })
    ElMessage.success('密码已修改')
    pwdDialog.visible = false
    pwdDialog.oldPassword = ''
    pwdDialog.newPassword = ''
  } finally {
    pwdSaving.value = false
  }
}

function logout() {
  auth.logout()
  ElMessage.success('已退出登录')
  router.push('/admin/login')
}
</script>

<template>
  <el-container class="admin-shell">
    <el-aside width="230px" class="admin-aside">
      <router-link to="/" class="admin-brand">
        <span class="brand-mark">天</span>
        <span>小天 Goofy</span>
      </router-link>

      <el-menu :default-active="route.path" router class="admin-menu">
        <el-menu-item v-for="m in menu" :key="m.path" :index="m.path">
          <el-icon><component :is="m.icon" /></el-icon>
          <span>{{ m.title }}</span>
        </el-menu-item>
      </el-menu>

      <div class="aside-foot">
        <router-link to="/" class="front-link">↗ 查看博客前台</router-link>
      </div>
    </el-aside>

    <el-container>
      <el-header class="admin-header">
        <div class="header-left">
          <ThemeSwitcher />
        </div>
        <el-dropdown @command="(cmd: string) => cmd === 'logout' ? logout() : (pwdDialog.visible = true)">
          <span class="user-chip">
            <span class="user-avatar">{{ (auth.user?.nickname || auth.user?.username || 'A').slice(0, 1) }}</span>
            {{ auth.user?.nickname || auth.user?.username }}
            <el-icon><ArrowDown /></el-icon>
          </span>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="password">修改密码</el-dropdown-item>
              <el-dropdown-item command="logout" divided>退出登录</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </el-header>

      <el-main class="admin-main">
        <router-view />
      </el-main>
    </el-container>

    <el-dialog v-model="pwdDialog.visible" title="修改密码" width="420px">
      <el-form label-width="80px">
        <el-form-item label="原密码">
          <el-input v-model="pwdDialog.oldPassword" type="password" show-password />
        </el-form-item>
        <el-form-item label="新密码">
          <el-input v-model="pwdDialog.newPassword" type="password" show-password placeholder="至少 6 位" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="pwdDialog.visible = false">取消</el-button>
        <el-button type="primary" :loading="pwdSaving" @click="savePwd">保存</el-button>
      </template>
    </el-dialog>
  </el-container>
</template>

<style scoped>
.admin-shell {
  height: 100vh;
}

.admin-aside {
  display: flex;
  flex-direction: column;
  background: var(--admin-aside-bg);
  color: var(--admin-aside-fg);
  border-right: var(--admin-aside-border);
}

.admin-brand {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 20px 22px 16px;
  color: var(--admin-aside-fg);
  font-family: var(--font-heading);
  font-size: 17px;
  font-weight: 700;
}

.brand-mark {
  width: 32px;
  height: 32px;
  display: grid;
  place-items: center;
  background: var(--accent);
  border-radius: 8px;
  font-size: 17px;
  color: #fff;
}

.admin-menu {
  border-right: none;
  background: transparent;
  flex: 1;
}

.admin-menu :deep(.el-menu-item) {
  color: var(--admin-aside-item);
  height: 46px;
}

.admin-menu :deep(.el-menu-item:hover) {
  background: var(--admin-aside-hover-bg);
  color: var(--admin-aside-hover-fg);
}

.admin-menu :deep(.el-menu-item.is-active) {
  background: var(--admin-aside-active-bg);
  color: var(--admin-aside-active-fg);
}

.aside-foot {
  padding: 16px 22px;
}

.front-link {
  font-size: 13px;
  color: var(--admin-aside-foot);
}

.front-link:hover {
  color: var(--admin-aside-foot-hover);
}

.admin-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: var(--admin-header-bg);
  border-bottom: 1px solid var(--border);
  height: 58px;
}

.header-left {
  display: flex;
  align-items: center;
}

.user-chip {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  font-size: 14px;
  color: var(--heading);
  outline: none;
}

.user-avatar {
  width: 30px;
  height: 30px;
  border-radius: 50%;
  background: var(--accent-soft);
  color: var(--accent);
  display: grid;
  place-items: center;
  font-weight: 700;
}

.admin-main {
  background: var(--admin-main-bg);
  overflow-y: auto;
}
</style>
