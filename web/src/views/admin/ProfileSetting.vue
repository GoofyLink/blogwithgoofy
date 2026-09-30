<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { fetchSettings, saveSettings, uploadImage } from '@/api'

const loading = ref(false)
const saving = ref(false)
const uploading = ref(false)

const form = reactive({
  avatar: '',
  nickname: '',
  bio: '',
  github: '',
  email: '',
  now_dev: '',
  now_learning: '',
  now_reading: '',
  now_update: '',
})

const FIELDS: { key: keyof typeof form; label: string; placeholder?: string; type?: 'input' | 'textarea' }[] = [
  { key: 'nickname', label: '昵称', placeholder: '首页展示的名字，如：小天' },
  { key: 'bio', label: '个人简介', placeholder: '首页自我介绍下方的那段话', type: 'textarea' },
  { key: 'github', label: 'GitHub', placeholder: 'https://github.com/你的用户名' },
  { key: 'email', label: '邮箱', placeholder: 'mailto 展示用，如：hello@example.com' },
  { key: 'now_dev', label: '正在开发', placeholder: '如：小天 Goofy Blog · Vue 3 + Go' },
  { key: 'now_learning', label: '正在学习', placeholder: '如：英语 · 德语' },
  { key: 'now_reading', label: '正在阅读', placeholder: '如：《三体》 刘慈欣' },
  { key: 'now_update', label: '最近更新', placeholder: '如：博客上线了多主题切换' },
]

async function load() {
  loading.value = true
  try {
    const s = (await fetchSettings()) || {}
    form.avatar = s.avatar || ''
    for (const f of FIELDS) form[f.key] = s[f.key] || ''
  } finally {
    loading.value = false
  }
}

async function onAvatarChange(file: { raw?: File }) {
  if (!file.raw) return
  uploading.value = true
  try {
    const url = await uploadImage(file.raw)
    form.avatar = url
    ElMessage.success('图片已上传，地址已填入')
  } finally {
    uploading.value = false
  }
}

async function save() {
  saving.value = true
  try {
    await saveSettings({ ...form })
    ElMessage.success('已保存，前台立即生效')
  } finally {
    saving.value = false
  }
}

onMounted(load)
</script>

<template>
  <div v-loading="loading">
    <div class="card page">
      <h2 class="page-title">个人信息配置</h2>
      <p class="page-sub">配置前台首页与页脚展示的内容，保存后立即生效。</p>

      <el-form label-width="110px" class="form">
        <!-- 头像 -->
        <el-form-item label="头像">
          <div class="avatar-row">
            <div class="avatar-preview">
              <img v-if="form.avatar" :src="form.avatar" alt="头像预览" />
              <span v-else class="placeholder">天</span>
            </div>
            <div class="avatar-ops">
              <el-input
                v-model="form.avatar"
                placeholder="粘贴图片地址（如 /uploads/xxx.png），或点右侧按钮直接上传"
              />
              <el-upload
                accept="image/*"
                :show-file-list="false"
                :auto-upload="false"
                :on-change="onAvatarChange"
              >
                <el-button :loading="uploading" class="up-btn">上传图片</el-button>
              </el-upload>
            </div>
          </div>
        </el-form-item>

        <el-form-item v-for="f in FIELDS" :key="f.key" :label="f.label">
          <el-input
            v-if="f.type === 'textarea'"
            v-model="form[f.key]"
            type="textarea"
            :rows="3"
            :placeholder="f.placeholder"
          />
          <el-input v-else v-model="form[f.key]" :placeholder="f.placeholder" />
        </el-form-item>

        <el-form-item>
          <el-button type="primary" :loading="saving" @click="save">保存配置</el-button>
        </el-form-item>
      </el-form>
    </div>
  </div>
</template>

<style scoped>
.page {
  padding: 26px 32px;
  max-width: 720px;
}

.page-title {
  margin: 0 0 6px;
  font-family: var(--font-heading);
}

.page-sub {
  color: var(--muted);
  font-size: 13.5px;
  margin: 0 0 22px;
}

.avatar-row {
  display: flex;
  align-items: center;
  gap: 18px;
  width: 100%;
}

.avatar-preview {
  width: 84px;
  height: 84px;
  flex-shrink: 0;
  border-radius: 50%;
  overflow: hidden;
  border: 1px solid var(--border);
  background: var(--accent);
  display: grid;
  place-items: center;
}

.avatar-preview img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.avatar-preview .placeholder {
  color: #fff;
  font-family: var(--font-heading);
  font-size: 36px;
  font-weight: 700;
}

.avatar-ops {
  flex: 1;
  display: flex;
  gap: 10px;
}

.up-btn {
  white-space: nowrap;
}
</style>
