<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import {
  adminFetchAiPrompts, adminFetchAiTools, createAiPrompt, createAiTool,
  deleteAiPrompt, deleteAiTool, updateAiPrompt, updateAiTool,
} from '@/api'
import type { AiPromptItem, AiToolItem } from '@/types'
import AdminPagination from '@/components/AdminPagination.vue'
import { usePagination } from '@/composables/usePagination'

const tools = ref<AiToolItem[]>([])
const { page: toolPage, pagedList: pagedTools } = usePagination(tools)
const prompts = ref<AiPromptItem[]>([])
const { page: promptPage, pagedList: pagedPrompts } = usePagination(prompts)
const loading = ref(false)

/** 下拉选项：已有值去重 + 本次新建 */
const newSections = ref<string[]>([])
const newCats = ref<string[]>([])
const sectionOptions = computed(() => {
  const set = new Set<string>()
  for (const t of tools.value) if (t.section) set.add(t.section)
  for (const p of prompts.value) if (p.section) set.add(p.section)
  return [...new Set([...set, ...newSections.value])]
})
const catOptions = computed(() => {
  const set = new Set<string>()
  for (const t of tools.value) if (t.category) set.add(t.category)
  return [...new Set([...set, ...newCats.value])]
})

async function load() {
  loading.value = true
  try {
    const [t, p] = await Promise.all([adminFetchAiTools(), adminFetchAiPrompts()])
    tools.value = t || []
    prompts.value = p || []
  } finally {
    loading.value = false
  }
}

// ---------- 工具弹窗 ----------
const toolDialog = reactive({
  visible: false,
  mode: 'create' as 'create' | 'edit',
  editId: 0,
  data: {
    name: '', url: '', icon: '', section: '', category: '',
    description: '', tags: '', sort: 0, status: 1 as 0 | 1,
  },
})

function openToolCreate() {
  const maxSort = tools.value.reduce((m, t) => Math.max(m, t.sort), 0)
  toolDialog.mode = 'create'
  toolDialog.editId = 0
  toolDialog.data = {
    name: '', url: '', icon: '', section: sectionOptions.value[0] || 'AI编程',
    category: '', description: '', tags: '', sort: maxSort + 1, status: 1,
  }
  toolDialog.visible = true
}

function openToolEdit(row: AiToolItem) {
  toolDialog.mode = 'edit'
  toolDialog.editId = row.id
  toolDialog.data = {
    name: row.name, url: row.url, icon: row.icon, section: row.section,
    category: row.category, description: row.description, tags: row.tags,
    sort: row.sort, status: row.status,
  }
  toolDialog.visible = true
}

async function saveTool() {
  if (!toolDialog.data.name.trim() || !toolDialog.data.url.trim()) {
    ElMessage.warning('名称和网址不能为空')
    return
  }
  if (toolDialog.mode === 'create') {
    await createAiTool({ ...toolDialog.data })
    ElMessage.success('已添加')
  } else {
    await updateAiTool(toolDialog.editId, { ...toolDialog.data })
    ElMessage.success('已更新')
  }
  toolDialog.visible = false
  load()
}

async function removeTool(row: AiToolItem) {
  await deleteAiTool(row.id)
  ElMessage.success('已删除')
  load()
}

// ---------- 提示词弹窗 ----------
const promptDialog = reactive({
  visible: false,
  mode: 'create' as 'create' | 'edit',
  editId: 0,
  data: {
    title: '', section: '', category: '', content: '', description: '',
    sort: 0, status: 1 as 0 | 1,
  },
})

function openPromptCreate() {
  const maxSort = prompts.value.reduce((m, p) => Math.max(m, p.sort), 0)
  promptDialog.mode = 'create'
  promptDialog.editId = 0
  promptDialog.data = {
    title: '', section: sectionOptions.value[0] || 'AI编程', category: '',
    content: '', description: '', sort: maxSort + 1, status: 1,
  }
  promptDialog.visible = true
}

function openPromptEdit(row: AiPromptItem) {
  promptDialog.mode = 'edit'
  promptDialog.editId = row.id
  promptDialog.data = {
    title: row.title, section: row.section, category: row.category,
    content: row.content, description: row.description,
    sort: row.sort, status: row.status,
  }
  promptDialog.visible = true
}

async function savePrompt() {
  if (!promptDialog.data.title.trim() || !promptDialog.data.content.trim()) {
    ElMessage.warning('标题和内容不能为空')
    return
  }
  if (promptDialog.mode === 'create') {
    await createAiPrompt({ ...promptDialog.data })
    ElMessage.success('已添加')
  } else {
    await updateAiPrompt(promptDialog.editId, { ...promptDialog.data })
    ElMessage.success('已更新')
  }
  promptDialog.visible = false
  load()
}

async function removePrompt(row: AiPromptItem) {
  await deleteAiPrompt(row.id)
  ElMessage.success('已删除')
  load()
}

onMounted(load)
</script>

<template>
  <div>
    <!-- 工具管理 -->
    <div class="toolbar card">
      <el-button type="primary" @click="openToolCreate">＋ 添加工具</el-button>
      <span class="tip">大类（如：AI编程 / AI绘画）决定前台一级 Tab；分类（如：代码补全 / 对话助手）决定二级筛选</span>
    </div>

    <el-table :data="pagedTools" v-loading="loading" stripe class="card table-card">
      <el-table-column prop="sort" label="排序" width="65" />
      <el-table-column prop="name" label="名称" width="140" />
      <el-table-column prop="url" label="网址" min-width="200" show-overflow-tooltip />
      <el-table-column prop="section" label="大类" width="95" />
      <el-table-column prop="category" label="分类" width="95" />
      <el-table-column prop="tags" label="标签" width="120" />
      <el-table-column label="状态" width="80">
        <template #default="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small">
            {{ row.status === 1 ? '展示' : '隐藏' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="145" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" size="small" @click="openToolEdit(row)">编辑</el-button>
          <el-popconfirm :title="`删除「${row.name}」？`" @confirm="removeTool(row)">
            <template #reference>
              <el-button link type="danger" size="small">删除</el-button>
            </template>
          </el-popconfirm>
        </template>
      </el-table-column>
    </el-table>
    <AdminPagination v-model:page="toolPage" :total="tools.length" />

    <!-- 提示词管理 -->
    <div class="toolbar card prompt-toolbar">
      <el-button type="primary" @click="openPromptCreate">＋ 添加提示词</el-button>
      <span class="tip">提示词按大类跟随对应 Tab 展示，前台支持一键复制</span>
    </div>

    <el-table :data="pagedPrompts" v-loading="loading" stripe class="card table-card">
      <el-table-column prop="sort" label="排序" width="65" />
      <el-table-column prop="title" label="标题" width="160" />
      <el-table-column prop="section" label="大类" width="95" />
      <el-table-column prop="description" label="简介" min-width="220" show-overflow-tooltip />
      <el-table-column label="状态" width="80">
        <template #default="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small">
            {{ row.status === 1 ? '展示' : '隐藏' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="145" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" size="small" @click="openPromptEdit(row)">编辑</el-button>
          <el-popconfirm :title="`删除「${row.title}」？`" @confirm="removePrompt(row)">
            <template #reference>
              <el-button link type="danger" size="small">删除</el-button>
            </template>
          </el-popconfirm>
        </template>
      </el-table-column>
    </el-table>
    <AdminPagination v-model:page="promptPage" :total="prompts.length" />


    <!-- 工具弹窗 -->
    <el-dialog v-model="toolDialog.visible" :title="toolDialog.mode === 'create' ? '添加工具' : '编辑工具'" width="560px">
      <el-form label-width="90px">
        <el-form-item label="名称 *">
          <el-input v-model="toolDialog.data.name" placeholder="如：ChatGPT" />
        </el-form-item>
        <el-form-item label="网址 *">
          <el-input v-model="toolDialog.data.url" placeholder="https://…" />
        </el-form-item>
        <el-form-item label="图标URL">
          <el-input v-model="toolDialog.data.icon" placeholder="选填，留空显示首字母" />
        </el-form-item>
        <el-form-item label="大类 *">
          <el-select v-model="toolDialog.data.section" filterable allow-create default-first-option placeholder="下拉选或输入新大类（如 AI绘画）">
            <el-option v-for="s in sectionOptions" :key="s" :label="s" :value="s" />
          </el-select>
        </el-form-item>
        <el-form-item label="分类">
          <el-select v-model="toolDialog.data.category" filterable allow-create default-first-option clearable placeholder="下拉选或输入（如：代码补全）">
            <el-option v-for="c in catOptions" :key="c" :label="c" :value="c" />
          </el-select>
        </el-form-item>
        <el-form-item label="简介">
          <el-input v-model="toolDialog.data.description" type="textarea" :rows="2" />
        </el-form-item>
        <el-form-item label="标签">
          <el-input v-model="toolDialog.data.tags" placeholder="逗号分隔，如：免费,开源" />
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="toolDialog.data.sort" :min="0" />
        </el-form-item>
        <el-form-item label="状态">
          <el-radio-group v-model="toolDialog.data.status">
            <el-radio-button :value="1">展示</el-radio-button>
            <el-radio-button :value="0">隐藏</el-radio-button>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="toolDialog.visible = false">取消</el-button>
        <el-button type="primary" @click="saveTool">保存</el-button>
      </template>
    </el-dialog>

    <!-- 提示词弹窗 -->
    <el-dialog v-model="promptDialog.visible" :title="promptDialog.mode === 'create' ? '添加提示词' : '编辑提示词'" width="560px">
      <el-form label-width="90px">
        <el-form-item label="标题 *">
          <el-input v-model="promptDialog.data.title" />
        </el-form-item>
        <el-form-item label="大类 *">
          <el-select v-model="promptDialog.data.section" filterable allow-create default-first-option>
            <el-option v-for="s in sectionOptions" :key="s" :label="s" :value="s" />
          </el-select>
        </el-form-item>
        <el-form-item label="分类">
          <el-input v-model="promptDialog.data.category" placeholder="选填" />
        </el-form-item>
        <el-form-item label="提示词 *">
          <el-input v-model="promptDialog.data.content" type="textarea" :rows="6" />
        </el-form-item>
        <el-form-item label="简介">
          <el-input v-model="promptDialog.data.description" />
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="promptDialog.data.sort" :min="0" />
        </el-form-item>
        <el-form-item label="状态">
          <el-radio-group v-model="promptDialog.data.status">
            <el-radio-button :value="1">展示</el-radio-button>
            <el-radio-button :value="0">隐藏</el-radio-button>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="promptDialog.visible = false">取消</el-button>
        <el-button type="primary" @click="savePrompt">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.toolbar {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 14px 18px;
  margin-bottom: 14px;
}

.prompt-toolbar {
  margin-top: 24px;
}

.tip {
  font-size: 12.5px;
  color: var(--muted);
}

.table-card {
  width: 100%;
}
</style>
