<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { createCategory, deleteCategory, fetchCategories, updateCategory } from '@/api'
import type { Category } from '@/types'

const list = ref<Category[]>([])
const loading = ref(false)
const creating = ref('')
const editing = reactive({ visible: false, id: 0, name: '' })

async function load() {
  loading.value = true
  try {
    list.value = (await fetchCategories()) || []
  } finally {
    loading.value = false
  }
}

async function add() {
  const name = creating.value.trim()
  if (!name) {
    ElMessage.warning('请输入分类名')
    return
  }
  await createCategory({ name })
  ElMessage.success('已添加')
  creating.value = ''
  load()
}

function openEdit(c: Category) {
  editing.id = c.id
  editing.name = c.name
  editing.visible = true
}

async function saveEdit() {
  if (!editing.name.trim()) {
    ElMessage.warning('分类名不能为空')
    return
  }
  await updateCategory(editing.id, { name: editing.name.trim() })
  ElMessage.success('已更新')
  editing.visible = false
  load()
}

async function remove(c: Category) {
  await deleteCategory(c.id)
  ElMessage.success('已删除')
  load()
}

onMounted(load)
</script>

<template>
  <div>
    <div class="toolbar card">
      <el-button type="primary" @click="add">＋ 添加分类</el-button>
      <el-input v-model="creating" placeholder="新分类名称" style="width: 240px" @keyup.enter="add" />
    </div>

    <el-table :data="list" v-loading="loading" stripe class="card table-card">
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column prop="name" label="分类名" min-width="200" />
      <el-table-column label="文章数" width="110">
        <template #default="{ row }">{{ row.articleCount || 0 }}</template>
      </el-table-column>
      <el-table-column label="操作" width="150">
        <template #default="{ row }">
          <el-button link type="primary" size="small" @click="openEdit(row)">重命名</el-button>
          <el-popconfirm :title="`删除分类「${row.name}」？`" @confirm="remove(row)">
            <template #reference>
              <el-button link type="danger" size="small">删除</el-button>
            </template>
          </el-popconfirm>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="editing.visible" title="重命名分类" width="400px">
      <el-input v-model="editing.name" @keyup.enter="saveEdit" />
      <template #footer>
        <el-button @click="editing.visible = false">取消</el-button>
        <el-button type="primary" @click="saveEdit">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.toolbar {
  display: flex;
  gap: 10px;
  padding: 14px 18px;
  margin-bottom: 14px;
}

.table-card {
  width: 100%;
}
</style>
