<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { adminFetchBooks, createBook, deleteBook, updateBook } from '@/api'
import type { Book } from '@/types'

const router = useRouter()

const list = ref<Book[]>([])
const loading = ref(false)

type BookForm = Pick<Book, 'title' | 'subtitle' | 'author' | 'language' | 'cover' | 'description' | 'type' | 'status' | 'sort'>

const emptyForm = (): BookForm => ({
  title: '',
  subtitle: '',
  author: '',
  language: 'en',
  cover: '',
  description: '',
  type: 0,
  status: 1,
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
    list.value = (await adminFetchBooks()) || []
  } finally {
    loading.value = false
  }
}

function openCreate() {
  dialog.mode = 'create'
  dialog.data = emptyForm()
  dialog.visible = true
}

function openEdit(row: Book) {
  dialog.mode = 'edit'
  dialog.editId = row.id
  dialog.data = {
    title: row.title,
    subtitle: row.subtitle,
    author: row.author,
    language: row.language,
    cover: row.cover,
    description: row.description,
    type: row.type ?? 0,
    status: row.status,
    sort: row.sort,
  }
  dialog.visible = true
}

async function save() {
  if (!dialog.data.title.trim()) {
    ElMessage.warning('请填写书名')
    return
  }
  if (dialog.mode === 'create') {
    const created = await createBook({ ...dialog.data })
    ElMessage.success('已添加')
    dialog.visible = false
    load()
    try {
      await ElMessageBox.confirm(
        `《${created.title}》已创建。现在去批量导入章节吗？（整本 TXT 一次导入，自动按章节切分）`,
        '添加章节',
        { confirmButtonText: '去导入章节', cancelButtonText: '稍后' },
      )
      router.push(`/admin/books/${created.id}/chapters`)
    } catch {
      /* 稍后 */
    }
  } else {
    await updateBook(dialog.editId, { ...dialog.data })
    ElMessage.success('已更新')
    dialog.visible = false
    load()
  }
}

async function remove(row: Book) {
  await deleteBook(row.id)
  ElMessage.success('已删除（含全部章节）')
  load()
}

function manageChapters(row: Book) {
  router.push(`/admin/books/${row.id}/chapters`)
}

onMounted(load)
</script>

<template>
  <div>
    <div class="toolbar">
      <el-button type="primary" @click="openCreate">＋ 添加书籍</el-button>
    </div>

    <el-table :data="list" v-loading="loading" stripe class="card table-card">
      <el-table-column label="书名" min-width="260">
        <template #default="{ row }">
          <b>{{ row.title }}</b>
          <div v-if="row.subtitle" class="sub">{{ row.subtitle }}</div>
          <div v-if="row.description" class="desc">{{ row.description }}</div>
        </template>
      </el-table-column>
      <el-table-column prop="author" label="作者" width="130" />
      <el-table-column label="类型" width="95">
        <template #default="{ row }">
          <el-tag size="small" :type="row.type === 1 ? 'success' : 'warning'" effect="plain">
            {{ row.type === 1 ? '小说专栏' : '学习书房' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="语言" width="80">
        <template #default="{ row }">
          <el-tag size="small" :type="row.language === 'de' ? 'warning' : 'primary'" effect="plain">
            {{ row.language === 'de' ? '德语' : '英语' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="章节" width="70">
        <template #default="{ row }">{{ row.chapterCount || 0 }}</template>
      </el-table-column>
      <el-table-column label="状态" width="85">
        <template #default="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small">
            {{ row.status === 1 ? '上架' : '下架' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="sort" label="排序" width="70" />
      <el-table-column label="操作" width="210" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" size="small" @click="manageChapters(row)">章节</el-button>
          <el-button link type="primary" size="small" @click="openEdit(row)">编辑</el-button>
          <el-popconfirm :title="`删除《${row.title}》及其全部章节？`" @confirm="remove(row)">
            <template #reference>
              <el-button link type="danger" size="small">删除</el-button>
            </template>
          </el-popconfirm>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="dialog.visible" :title="dialog.mode === 'create' ? '添加书籍' : '编辑书籍'" width="560px">
      <el-form label-width="90px">
        <el-form-item label="书名 *">
          <el-input v-model="dialog.data.title" placeholder="中文书名，如：伊索寓言精选" />
        </el-form-item>
        <el-form-item label="原文书名">
          <el-input v-model="dialog.data.subtitle" placeholder="如：Aesop's Fables" />
        </el-form-item>
        <el-form-item label="作者">
          <el-input v-model="dialog.data.author" />
        </el-form-item>
        <el-form-item label="类型 *">
          <el-radio-group v-model="dialog.data.type">
            <el-radio-button :value="0">学习书房</el-radio-button>
            <el-radio-button :value="1">小说专栏</el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="语言 *">
          <el-select v-model="dialog.data.language" style="width: 160px">
            <el-option label="英语 en" value="en" />
            <el-option label="德语 de" value="de" />
            <el-option label="中文 zh" value="zh" />
          </el-select>
        </el-form-item>
        <el-form-item label="封面 URL">
          <el-input v-model="dialog.data.cover" placeholder="选填，留空显示自动封面" />
        </el-form-item>
        <el-form-item label="简介">
          <el-input v-model="dialog.data.description" type="textarea" :rows="3" />
        </el-form-item>
        <div class="form-row">
          <el-form-item label="状态">
            <el-radio-group v-model="dialog.data.status">
              <el-radio-button :value="1">上架</el-radio-button>
              <el-radio-button :value="0">下架</el-radio-button>
            </el-radio-group>
          </el-form-item>
          <el-form-item label="排序">
            <el-input-number v-model="dialog.data.sort" :min="0" />
          </el-form-item>
        </div>
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

.sub {
  font-size: 12px;
  color: var(--muted);
  font-style: italic;
}

.desc {
  font-size: 12px;
  color: var(--muted);
  margin-top: 2px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 420px;
}

.form-row {
  display: flex;
  gap: 20px;
}
</style>
