<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { UploadFilled } from '@element-plus/icons-vue'
import {
  adminFetchBooks, adminFetchChapters, batchCreateChapters, createChapter,
  deleteChapter, updateChapter,
} from '@/api'
import type { Book, Chapter } from '@/types'
import AdminPagination from '@/components/AdminPagination.vue'
import { usePagination } from '@/composables/usePagination'
import { alignTranslations, parseChapters, parseTitleList, readFileText, type ParsedChapter } from '@/utils/chapterImport'

const route = useRoute()
const router = useRouter()

const bookId = computed(() => Number(route.params.id))
const book = ref<Book | null>(null)
const list = ref<Chapter[]>([])
const { page, pagedList } = usePagination(list)

const loading = ref(false)

const dialog = reactive({
  visible: false,
  mode: 'create' as 'create' | 'edit',
  editId: 0,
  data: {
    title: '',
    content: '',
    translation: '',
    sort: 1,
    status: 1 as 0 | 1,
  },
})

async function load() {
  loading.value = true
  try {
    const [books, chapters] = await Promise.all([adminFetchBooks(), adminFetchChapters(bookId.value)])
    book.value = books.find((b) => b.id === bookId.value) || null
    list.value = chapters || []
  } finally {
    loading.value = false
  }
}

function openCreate() {
  dialog.mode = 'create'
  dialog.data = {
    title: '',
    content: '',
    translation: '',
    sort: (list.value[list.value.length - 1]?.sort || 0) + 1,
    status: 1,
  }
  dialog.visible = true
}

function openEdit(row: Chapter) {
  dialog.mode = 'edit'
  dialog.editId = row.id
  dialog.data = {
    title: row.title,
    content: row.content,
    translation: row.translation || '',
    sort: row.sort,
    status: row.status,
  }
  dialog.visible = true
}

async function save() {
  if (!dialog.data.title.trim() || !dialog.data.content.trim()) {
    ElMessage.warning('请填写章节标题和内容')
    return
  }
  if (dialog.mode === 'create') {
    await createChapter({ bookId: bookId.value, ...dialog.data })
    ElMessage.success('已添加')
  } else {
    await updateChapter(dialog.editId, { ...dialog.data })
    ElMessage.success('已更新')
  }
  dialog.visible = false
  load()
}

async function remove(row: Chapter) {
  await deleteChapter(row.id)
  ElMessage.success('已删除')
  load()
}

const wordCount = (s: string) => (s ? s.trim().split(/\s+/).length : 0)

// ==================== 批量导入 ====================
const importDialog = reactive({
  visible: false,
  // full = 完整导入（原文+可选译文）；titles = 仅标题快速建章
  importMode: 'full' as 'full' | 'titles',
  sourceMode: 'paste' as 'paste' | 'file',
  splitMode: 'auto' as 'auto' | 'separator' | 'size',
  separator: '---',
  chapterSize: 3000,
  rawText: '',
  rawFileName: '',
  transMode: 'none' as 'none' | 'paste' | 'file',
  transText: '',
  transFileName: '',
  titleText: '',
  preview: [] as ParsedChapter[],
  warning: '',
  importing: false,
})

function openImport() {
  importDialog.visible = true
  importDialog.importMode = 'full'
  importDialog.sourceMode = 'paste'
  importDialog.splitMode = 'auto'
  importDialog.separator = '---'
  importDialog.chapterSize = 3000
  importDialog.rawText = ''
  importDialog.rawFileName = ''
  importDialog.transMode = 'none'
  importDialog.transText = ''
  importDialog.transFileName = ''
  importDialog.titleText = ''
  importDialog.preview = []
  importDialog.warning = ''
}

async function onRawFileChange(file: { raw?: File; name?: string }) {
  if (!file.raw) return
  try {
    importDialog.rawText = await readFileText(file.raw)
    importDialog.rawFileName = file.name || file.raw.name
    ElMessage.success(`已读取 ${importDialog.rawFileName}`)
  } catch {
    ElMessage.error('文件读取失败')
  }
}

async function onTransFileChange(file: { raw?: File; name?: string }) {
  if (!file.raw) return
  try {
    importDialog.transText = await readFileText(file.raw)
    importDialog.transFileName = file.name || file.raw.name
    ElMessage.success(`已读取 ${importDialog.transFileName}`)
  } catch {
    ElMessage.error('文件读取失败')
  }
}

function doParse() {
  // 仅标题模式：每行一个标题，正文留空后续补充
  if (importDialog.importMode === 'titles') {
    const { titles, warning } = parseTitleList(importDialog.titleText)
    if (warning || !titles.length) {
      importDialog.preview = []
      importDialog.warning = warning || '没有解析到标题'
      return
    }
    importDialog.preview = titles.map((t) => ({ title: t, content: '' }))
    importDialog.warning = ''
    return
  }

  const result = parseChapters(
    importDialog.rawText,
    importDialog.splitMode,
    importDialog.separator,
    importDialog.chapterSize,
  )
  if (result.warning || !result.chapters.length) {
    importDialog.preview = []
    importDialog.warning = result.warning || '没有解析到章节'
    return
  }
  if (importDialog.transMode !== 'none' && importDialog.transText.trim()) {
    const aligned = alignTranslations(
      result.chapters,
      importDialog.transText,
      importDialog.splitMode,
      importDialog.separator,
      importDialog.chapterSize,
    )
    importDialog.preview = aligned.aligned
    importDialog.warning = aligned.warning || ''
  } else {
    importDialog.preview = result.chapters
    importDialog.warning = ''
  }
}

function removePreviewRow(idx: number) {
  importDialog.preview.splice(idx, 1)
}

async function doImport() {
  if (!importDialog.preview.length) {
    ElMessage.warning('请先解析并确认章节')
    return
  }
  importDialog.importing = true
  try {
    const res = await batchCreateChapters({
      bookId: bookId.value,
      chapters: importDialog.preview.map((c) => ({
        title: c.title.trim() || '未命名章节',
        content: c.content,
        translation: c.translation || '',
      })),
    })
    ElMessage.success(`成功导入 ${res.created} 章`)
    importDialog.visible = false
    load()
  } finally {
    importDialog.importing = false
  }
}

onMounted(load)
</script>

<template>
  <div>
    <div class="toolbar">
      <div class="toolbar-left">
        <el-button @click="router.push('/admin/books')">← 返回书籍列表</el-button>
        <span class="book-name" v-if="book">《{{ book.title }}》章节管理（{{ book.language === 'de' ? '德语' : '英语' }}）</span>
      </div>
      <div>
        <el-button type="success" plain @click="openImport">📥 批量导入章节</el-button>
        <el-button type="primary" @click="openCreate">＋ 添加单章</el-button>
      </div>
    </div>

    <el-table :data="pagedList" v-loading="loading" stripe class="card table-card">
      <el-table-column prop="sort" label="序号" width="70" />
      <el-table-column prop="title" label="章节标题" min-width="240" />
      <el-table-column label="字数" width="90">
        <template #default="{ row }">{{ wordCount(row.content) }}</template>
      </el-table-column>
      <el-table-column label="中文对照" width="90">
        <template #default="{ row }">
          <el-tag :type="row.translation ? 'success' : 'info'" size="small" effect="plain">
            {{ row.translation ? '有' : '无' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="状态" width="85">
        <template #default="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small">
            {{ row.status === 1 ? '已发布' : '草稿' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="150" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" size="small" @click="openEdit(row)">编辑</el-button>
          <el-popconfirm :title="`删除「${row.title}」？`" @confirm="remove(row)">
            <template #reference>
              <el-button link type="danger" size="small">删除</el-button>
            </template>
          </el-popconfirm>
        </template>
      </el-table-column>
    </el-table>
    <AdminPagination v-model:page="page" :total="list.length" />

    <!-- 单章编辑 -->
    <el-dialog
      v-model="dialog.visible"
      :title="dialog.mode === 'create' ? '添加章节' : '编辑章节'"
      width="820px"
      top="4vh"
    >
      <el-form label-width="90px">
        <div class="form-row">
          <el-form-item label="章节标题 *">
            <el-input v-model="dialog.data.title" placeholder="如：Chapter 1 · The Ant and the Grasshopper" />
          </el-form-item>
          <el-form-item label="排序" label-width="50px">
            <el-input-number v-model="dialog.data.sort" :min="0" />
          </el-form-item>
          <el-form-item label="状态" label-width="50px">
            <el-radio-group v-model="dialog.data.status" size="small">
              <el-radio-button :value="1">发布</el-radio-button>
              <el-radio-button :value="0">草稿</el-radio-button>
            </el-radio-group>
          </el-form-item>
        </div>
        <el-form-item label="原文内容 *">
          <el-input
            v-model="dialog.data.content"
            type="textarea"
            :rows="16"
            placeholder="纯文本，段落之间用空行分隔。句子翻译按 . ! ? 自动切分。"
          />
        </el-form-item>
        <el-form-item label="中文对照">
          <el-input
            v-model="dialog.data.translation"
            type="textarea"
            :rows="8"
            placeholder="选填：整章中文译文，段落之间用空行分隔（阅读页可开关显示）。没有也可以留空，读者可点击句子实时翻译。"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialog.visible = false">取消</el-button>
        <el-button type="primary" @click="save">保存</el-button>
      </template>
    </el-dialog>

    <!-- 批量导入 -->
    <el-dialog v-model="importDialog.visible" title="📥 批量导入章节" width="900px" top="3vh">
      <!-- 导入模式 -->
      <el-radio-group v-model="importDialog.importMode" size="small" class="mode-switch">
        <el-radio-button value="full">完整导入（原文 + 可选译文）</el-radio-button>
        <el-radio-button value="titles">仅标题快速建章</el-radio-button>
      </el-radio-group>

      <el-alert
        v-if="!importDialog.preview.length"
        type="info"
        :closable="false"
        show-icon
        class="import-tip"
        :title="importDialog.importMode === 'full'
          ? '整本书一次导入：粘贴全文或上传 TXT，自动按章节标题切分；中文译文（可选）可一起对齐导入。'
          : '只有目录？每行粘贴一个标题（如：第一篇: 年轻军官与老兵 A Young Officer and an Old Soldier），一次建好全部章节，正文以后在列表里逐章补充。'"
      />

      <!-- ===== 仅标题模式 ===== -->
      <template v-if="importDialog.importMode === 'titles'">
        <h4 class="sec-title">① 章节标题列表 <span class="sec-sub">（每行一个标题）</span></h4>
        <el-input
          v-model="importDialog.titleText"
          type="textarea"
          :rows="12"
          placeholder="每行粘贴一个章节标题，例如：&#10;第一篇: 年轻军官与老兵 A Young Officer and an Old Soldier&#10;第二篇: 您是要感谢她吗？ Are You Going to Thank Her?&#10;第三篇: 萨姆与托德 Sam and Tod&#10;……"
        />
      </template>

      <!-- ===== 完整导入模式 ===== -->
      <template v-else>
      <!-- 原文输入 -->
      <h4 class="sec-title">① 原文 <span class="sec-sub">（TXT 纯文本）</span></h4>
      <el-radio-group v-model="importDialog.sourceMode" size="small">
        <el-radio-button value="paste">粘贴文本</el-radio-button>
        <el-radio-button value="file">上传 TXT 文件</el-radio-button>
      </el-radio-group>

      <el-input
        v-if="importDialog.sourceMode === 'paste'"
        v-model="importDialog.rawText"
        type="textarea"
        :rows="8"
        class="import-area"
        placeholder="粘贴整本书内容，章节标题独占一行，例如：&#10;Chapter 1&#10;正文……&#10;&#10;Chapter 2&#10;正文……&#10;&#10;也支持「第1章」「Kapitel 1」等标题。"
      />
      <el-upload
        v-else
        drag
        accept=".txt"
        :limit="1"
        :auto-upload="false"
        :on-change="onRawFileChange"
        class="import-upload"
      >
        <el-icon size="36"><UploadFilled /></el-icon>
        <div class="el-upload__text">拖入或点击选择 .txt 文件</div>
        <template #tip>
          <div class="el-upload__tip">{{ importDialog.rawFileName ? '已读取：' + importDialog.rawFileName : '自动识别 UTF-8 / GBK 编码' }}</div>
        </template>
      </el-upload>

      <!-- 分割方式 -->
      <h4 class="sec-title">② 章节分割方式</h4>
      <el-radio-group v-model="importDialog.splitMode" size="small">
        <el-radio-button value="auto">自动识别标题</el-radio-button>
        <el-radio-button value="size">按长度自动分章</el-radio-button>
        <el-radio-button value="separator">自定义分隔符</el-radio-button>
      </el-radio-group>
      <div v-if="importDialog.splitMode === 'size'" class="sep-row">
        <span class="sep-label">每章约：</span>
        <el-input-number v-model="importDialog.chapterSize" :min="800" :max="20000" :step="500" size="small" />
        <span class="sep-tip">字。若文中带有「第X章」等章节标题，会自动优先按标题切分并提取原标题</span>
      </div>
      <div v-if="importDialog.splitMode === 'separator'" class="sep-row">
        <span class="sep-label">分隔符（独占一行的文本）：</span>
        <el-input v-model="importDialog.separator" style="width: 180px" placeholder="如 --- 或 ===" />
        <span class="sep-tip">每块第一行作为章节标题</span>
      </div>

      <!-- 译文 -->
      <h4 class="sec-title">③ 中文译文 <span class="sec-sub">（可选，同样规则切分后按顺序对齐）</span></h4>
      <el-radio-group v-model="importDialog.transMode" size="small">
        <el-radio-button value="none">不导入</el-radio-button>
        <el-radio-button value="paste">粘贴文本</el-radio-button>
        <el-radio-button value="file">上传 TXT 文件</el-radio-button>
      </el-radio-group>
      <el-input
        v-if="importDialog.transMode === 'paste'"
        v-model="importDialog.transText"
        type="textarea"
        :rows="5"
        class="import-area"
        placeholder="粘贴整本书的中文译文，章节划分与原文一致即可"
      />
      <el-upload
        v-else-if="importDialog.transMode === 'file'"
        drag
        accept=".txt"
        :limit="1"
        :auto-upload="false"
        :on-change="onTransFileChange"
        class="import-upload"
      >
        <el-icon size="36"><UploadFilled /></el-icon>
        <div class="el-upload__text">拖入或点击选择中文译文 .txt</div>
        <template #tip>
          <div class="el-upload__tip">{{ importDialog.transFileName ? '已读取：' + importDialog.transFileName : '' }}</div>
        </template>
      </el-upload>
      </template><!-- 完整导入模式结束 -->

      <div class="parse-actions">
        <el-button @click="doParse">🔍 解析预览</el-button>
      </div>

      <!-- 预览 -->
      <template v-if="importDialog.preview.length">
        <el-alert v-if="importDialog.warning" type="warning" :closable="false" show-icon class="import-tip" :title="importDialog.warning" />
        <h4 class="sec-title">
          ④ 预览（共 {{ importDialog.preview.length }} 章，标题可直接修改，确认后导入）
        </h4>
        <el-table :data="importDialog.preview" max-height="300" size="small" stripe>
          <el-table-column type="index" label="#" width="45" />
          <el-table-column label="章节标题" min-width="240">
            <template #default="{ row }">
              <el-input v-model="row.title" size="small" />
            </template>
          </el-table-column>
          <el-table-column label="字数" width="70">
            <template #default="{ row }">{{ wordCount(row.content) }}</template>
          </el-table-column>
          <el-table-column label="译文" width="70">
            <template #default="{ row }">
              <el-tag :type="row.translation ? 'success' : 'info'" size="small" effect="plain">
                {{ row.translation ? '有' : '无' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="内容预览" min-width="180" show-overflow-tooltip>
            <template #default="{ row }">
              {{ row.content ? row.content.slice(0, 80) + '…' : '（正文待录入）' }}
            </template>
          </el-table-column>
          <el-table-column label="" width="60">
            <template #default="{ $index }">
              <el-button link type="danger" size="small" @click="removePreviewRow($index)">移除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </template>

      <template #footer>
        <el-button @click="importDialog.visible = false">取消</el-button>
        <el-button
          type="primary"
          :loading="importDialog.importing"
          :disabled="!importDialog.preview.length"
          @click="doImport"
        >
          导入 {{ importDialog.preview.length || '' }} 章
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 14px;
}

.toolbar-left {
  display: flex;
  align-items: center;
  gap: 14px;
}

.book-name {
  font-size: 15px;
  font-weight: 600;
}

.table-card {
  width: 100%;
}

.form-row {
  display: flex;
  gap: 12px;
  align-items: center;
  flex-wrap: wrap;
}

.form-row > .el-form-item {
  flex: 1;
  min-width: 220px;
}

.sec-title {
  margin: 18px 0 10px;
  font-size: 14.5px;
}

.sec-sub {
  font-weight: 400;
  color: var(--muted);
  font-size: 12.5px;
}

.import-area {
  margin-top: 10px;
}

.import-upload {
  margin-top: 10px;
  width: 100%;
}

.sep-row {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-top: 10px;
  flex-wrap: wrap;
}

.sep-label {
  font-size: 13px;
  color: var(--muted);
}

.sep-tip {
  font-size: 12px;
  color: var(--muted);
}

.parse-actions {
  margin-top: 16px;
  display: flex;
  justify-content: center;
}

.mode-switch {
  margin-bottom: 14px;
}

.import-tip {
  margin-bottom: 6px;
}
</style>
