<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { createLink, deleteLink, fetchLinks, updateLink } from '@/api'
import type { LinkItem } from '@/types'

const list = ref<LinkItem[]>([])
const loading = ref(false)

const emptyForm = (): Omit<LinkItem, 'id'> => ({
  name: '',
  url: '',
  logo: '',
  description: '',
  sort: 0,
})

const dialog = reactive({
  visible: false,
  mode: 'create' as 'create' | 'edit',
  editId: 0,
  data: emptyForm(),
})

async function load() {
  loading.value = true
  try {
    list.value = (await fetchLinks()) || []
  } finally {
    loading.value = false
  }
}

function openCreate() {
  dialog.mode = 'create'
  dialog.data = emptyForm()
  dialog.visible = true
}

function openEdit(row: LinkItem) {
  dialog.mode = 'edit'
  dialog.editId = row.id
  dialog.data = { name: row.name, url: row.url, logo: row.logo, description: row.description, sort: row.sort }
  dialog.visible = true
}

async function save() {
  if (!dialog.data.name.trim() || !dialog.data.url.trim()) {
    ElMessage.warning('名称和地址不能为空')
    return
  }
  if (dialog.mode === 'create') {
    await createLink({ ...dialog.data })
    ElMessage.success('已添加')
  } else {
    await updateLink(dialog.editId, { ...dialog.data })
    ElMessage.success('已更新')
  }
  dialog.visible = false
  load()
}

async function remove(row: LinkItem) {
  await deleteLink(row.id)
  ElMessage.success('已删除')
  load()
}

onMounted(load)
</script>

<template>
  <div>
    <div class="toolbar">
      <el-button type="primary" @click="openCreate">＋ 添加友链</el-button>
    </div>

    <el-table :data="list" v-loading="loading" stripe class="card table-card">
      <el-table-column prop="name" label="名称" width="160" />
      <el-table-column label="地址" min-width="220">
        <template #default="{ row }">
          <a :href="row.url" target="_blank" rel="noopener" class="url-link">{{ row.url }}</a>
        </template>
      </el-table-column>
      <el-table-column prop="description" label="描述" min-width="180" />
      <el-table-column prop="sort" label="排序" width="80" />
      <el-table-column label="操作" width="150" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" size="small" @click="openEdit(row)">编辑</el-button>
          <el-popconfirm :title="`删除友链「${row.name}」？`" @confirm="remove(row)">
            <template #reference>
              <el-button link type="danger" size="small">删除</el-button>
            </template>
          </el-popconfirm>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="dialog.visible" :title="dialog.mode === 'create' ? '添加友链' : '编辑友链'" width="480px">
      <el-form label-width="70px">
        <el-form-item label="名称">
          <el-input v-model="dialog.data.name" />
        </el-form-item>
        <el-form-item label="地址">
          <el-input v-model="dialog.data.url" placeholder="https://…" />
        </el-form-item>
        <el-form-item label="头像">
          <el-input v-model="dialog.data.logo" placeholder="头像/Logo URL（选填）" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="dialog.data.description" />
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="dialog.data.sort" :min="0" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialog.visible = false">取消</el-button>
        <el-button type="primary" @click="save">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.toolbar {
  margin-bottom: 14px;
}

.table-card {
  width: 100%;
}

.url-link {
  color: var(--el-color-primary);
  font-size: 13px;
}
</style>
