<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { User, Lock } from '@element-plus/icons-vue'
import { login } from '@/api'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const route = useRoute()
const auth = useAuthStore()

const form = reactive({ username: 'admin', password: '' })
const loading = ref(false)

async function submit() {
  if (!form.username || !form.password) {
    ElMessage.warning('请输入用户名和密码')
    return
  }
  loading.value = true
  try {
    const data = await login({ username: form.username, password: form.password })
    auth.setAuth(data.token, data.user)
    ElMessage.success('登录成功')
    router.push((route.query.redirect as string) || '/admin/dashboard')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="login-page">
    <div class="login-card card">
      <div class="login-logo">
        <span class="logo-mark">天</span>
      </div>
      <h1 class="login-title">小天 Goofy · 后台</h1>
      <el-form @submit.prevent>
        <el-form-item>
          <el-input v-model="form.username" placeholder="用户名" size="large" :prefix-icon="User" @keyup.enter="submit" />
        </el-form-item>
        <el-form-item>
          <el-input
            v-model="form.password"
            type="password"
            placeholder="密码"
            size="large"
            :prefix-icon="Lock"
            show-password
            @keyup.enter="submit"
          />
        </el-form-item>
        <el-button type="primary" size="large" class="login-btn" :loading="loading" @click="submit">
          登 录
        </el-button>
      </el-form>
      <router-link to="/" class="back-home">← 回到博客前台</router-link>
    </div>
  </div>
</template>

<style scoped>
.login-page {
  min-height: 100vh;
  display: grid;
  place-items: center;
  background:
    radial-gradient(circle at 20% 10%, var(--accent-soft), transparent 45%),
    radial-gradient(circle at 85% 85%, var(--surface), transparent 40%),
    var(--bg);
}

.login-card {
  width: 380px;
  padding: 40px 38px 30px;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.login-logo .logo-mark {
  width: 52px;
  height: 52px;
  display: grid;
  place-items: center;
  background: var(--accent);
  color: #fff;
  font-family: var(--font-heading);
  font-size: 28px;
  border-radius: 14px;
  box-shadow: var(--shadow);
}

.login-title {
  font-family: var(--font-heading);
  font-size: 20px;
  margin: 16px 0 24px;
}

.login-btn {
  width: 100%;
  margin-top: 4px;
}

.back-home {
  margin-top: 18px;
  font-size: 13px;
  color: var(--muted);
}

.back-home:hover {
  color: var(--accent);
}
</style>
