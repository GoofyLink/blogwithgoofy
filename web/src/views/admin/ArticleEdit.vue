<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { MdEditor } from 'md-editor-v3'
import { createArticle, fetchArticle, fetchCategories, fetchTags, updateArticle, uploadImage } from '@/api'
import type { ArticleForm as ArticleFormType, Category, Tag } from '@/types'

const route = useRoute()
const router = useRouter()

const editId = computed(() => Number(route.params.id) || 0)

const categories = ref<Category[]>([])
const tags = ref<Tag[]>([])
const saving = ref(false)

const form = reactive<ArticleFormType>({
  title: '',
  summary: '',
  cover: '',
  content: '',
  translation: '',
  categoryId: 0,
  type: 0,
  language: 'zh',
  status: 1,
  tagIds: [],
})

async function onUploadImg(files: File[], callback: (urls: string[]) => void) {
  try {
    const urls = await Promise.all(files.map((f) => uploadImage(f)))
    callback(urls)
  } catch {
    ElMessage.error('图片上传失败')
  }
}

async function load() {
  const [cats, tgs] = await Promise.all([fetchCategories(), fetchTags()])
  categories.value = cats || []
  tags.value = tgs || []

  if (editId.value) {
    const data = await fetchArticle(editId.value)
    const a = data.article
    form.title = a.title
    form.summary = a.summary
    form.cover = a.cover
    form.content = a.content || ''
    form.translation = a.translation || ''
    form.categoryId = a.categoryId
    form.type = a.type
    form.language = a.language
    form.status = a.status
    form.tagIds = (a.tags || []).map((t) => t.id)
  }
}

async function save() {
  if (!form.title.trim() || !form.content.trim()) {
    ElMessage.warning('标题和正文不能为空')
    return
  }
  saving.value = true
  try {
    if (editId.value) {
      await updateArticle(editId.value, { ...form })
      ElMessage.success('已保存')
    } else {
      await createArticle({ ...form })
      ElMessage.success('已创建')
    }
    router.push('/admin/articles')
  } finally {
    saving.value = false
  }
}

onMounted(load)
</script>

<template>
  <div class="edit-page">
    <div class="edit-head">
      <h2>{{ editId ? '编辑文章' : '新建文章' }}</h2>
      <div class="head-ops">
        <router-link v-if="editId" :to="`/article/${editId}`">
          <el-button>前台预览</el-button>
        </router-link>
        <el-button @click="router.back()">返回</el-button>
        <el-button type="primary" :loading="saving" @click="save">
          {{ form.status === 1 ? '发布' : '保存草稿' }}
        </el-button>
      </div>
    </div>

    <div class="edit-form card">
      <el-form label-width="90px">
        <el-form-item label="标题">
          <el-input v-model="form.title" placeholder="文章标题" maxlength="190" show-word-limit />
        </el-form-item>

        <div class="form-row2">
          <el-form-item label="类型">
            <el-radio-group v-model="form.type">
              <el-radio-button :value="0">博客文章</el-radio-button>
              <el-radio-button :value="1">学习文章</el-radio-button>
            </el-radio-group>
          </el-form-item>
          <el-form-item label="语言">
            <el-select v-model="form.language" :disabled="form.type === 0" style="width: 140px">
              <el-option label="中文 zh" value="zh" :disabled="true" />
              <el-option label="英语 en" value="en" />
              <el-option label="德语 de" value="de" />
            </el-select>
          </el-form-item>
          <el-form-item label="状态">
            <el-radio-group v-model="form.status">
              <el-radio-button :value="1">发布</el-radio-button>
              <el-radio-button :value="0">草稿</el-radio-button>
            </el-radio-group>
          </el-form-item>
        </div>

        <div class="form-row2">
          <el-form-item label="分类">
            <el-select v-model="form.categoryId" placeholder="选择分类" clearable style="width: 180px">
              <el-option v-for="c in categories" :key="c.id" :label="c.name" :value="c.id" />
            </el-select>
          </el-form-item>
          <el-form-item label="标签">
            <el-select v-model="form.tagIds" multiple placeholder="选择标签（在标签管理中创建）" style="width: 100%">
              <el-option v-for="t in tags" :key="t.id" :label="t.name" :value="t.id" />
            </el-select>
          </el-form-item>
        </div>

        <el-form-item label="封面图">
          <el-input v-model="form.cover" placeholder="封面图片 URL（选填，可直接粘贴 /uploads/xx.png）" />
        </el-form-item>

        <el-form-item label="摘要">
          <el-input v-model="form.summary" type="textarea" :rows="2" maxlength="1000" show-word-limit placeholder="列表页显示的摘要" />
        </el-form-item>

        <el-form-item label="正文">
          <div class="editor-wrap">
            <MdEditor
              v-model="form.content"
              :style="{ height: '460px' }"
              :toolbars-exclude="['github', 'htmlPreview', 'catalog']"
              placeholder="支持 Markdown，工具栏图片可直接上传…"
              @on-upload-img="onUploadImg"
            />
          </div>
        </el-form-item>

        <el-form-item v-if="form.type === 1" label="中文对照">
          <div class="editor-wrap">
            <MdEditor
              v-model="form.translation"
              :style="{ height: '280px' }"
              :toolbars-exclude="['github', 'htmlPreview', 'catalog']"
              placeholder="学习文章的中文全文对照，段落划分需与原文一致（阅读页可切换显示）"
              @on-upload-img="onUploadImg"
            />
          </div>
        </el-form-item>
      </el-form>
    </div>
  </div>
</template>

<style scoped>
.edit-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 14px;
}

.edit-head h2 {
  margin: 0;
  font-family: var(--font-heading);
}

.head-ops {
  display: flex;
  gap: 4px;
}

.edit-form {
  padding: 24px 28px;
}

.form-row2 {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}

.form-row2 > .el-form-item {
  flex: 1;
  min-width: 220px;
}

.editor-wrap {
  width: 100%;
}
</style>
