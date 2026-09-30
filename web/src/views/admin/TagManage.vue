<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { createTag, deleteTag, fetchTags, updateTag } from '@/api'
import type { Tag } from '@/types'
import AdminPagination from '@/components/AdminPagination.vue'
import { usePagination } from '@/composables/usePagination'

const list = ref<Tag[]>([])
const { page, pagedList } = usePagination(list)

const loading = ref(false)
const creating = ref('')
const editing = reactive({ visible: false, id: 0, name: '' })

async function load() {
  loading.value = true
  try {
    list.value = (await fetchTags()) || []
  } finally {
    loading.value = false
  }
}

async function add() {
  const name = creating.value.trim()
  if (!name) {
    ElMessage.warning('请输入标签名')
    return
  }
  await createTag({ name })
  ElMessage.success('已添加')
  creating.value = ''
  load()
}

function openEdit(t: Tag) {
  editing.id = t.id
  editing.name = t.name
  editing.visible = true
}

async function saveEdit() {
  if (!editing.name.trim()) {
    ElMessage.warning('标签名不能为空')
    return
  }
  await updateTag(editing.id, { name: editing.name.trim() })
  ElMessage.success('已更新')
  editing.visible = false
  load()
}

async function remove(t: Tag) {
  await deleteTag(t.id)
  ElMessage.success('已删除')
  load()
}

onMounted(load)
</script>

<template>
  <div>
    <div class="toolbar card">
      <el-button type="primary" @click="add">＋ 添加标签</el-button>
      <el-input v-model="creating" placeholder="新标签名称" style="width: 240px" @keyup.enter="add" />
    </div>

    <el-table :data="pagedList" v-loading="loading" stripe class="card table-card">
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column prop="name" label="标签名" min-width="200" />
      <el-table-column label="文章数" width="110">
        <template #default="{ row }">{{ row.articleCount || 0 }}</template>
      </el-table-column>
      <el-table-column label="操作" width="150">
        <template #default="{ row }">
          <el-button link type="primary" size="small" @click="openEdit(row)">重命名</el-button>
          <el-popconfirm :title="`删除标签「${row.name}」？`" @confirm="remove(row)">
            <template #reference>
              <el-button link type="danger" size="small">删除</el-button>
            </template>
          </el-popconfirm>
        </template>
      </el-table-column>
    </el-table>
    <AdminPagination v-model:page="page" :total="list.length" />

    <el-dialog v-model="editing.visible" title="重命名标签" width="400px">
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
