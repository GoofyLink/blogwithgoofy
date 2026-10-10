<script setup lang="ts">
import { computed, ref } from 'vue'
import { MdEditor } from 'md-editor-v3'
import { ElMessage } from 'element-plus'
import { uploadImage, type AnimeForm } from '@/api'
import { watchStatuses, airStatuses } from '@/utils/anime'
const form = defineModel<AnimeForm>({ required: true })
const props = defineProps<{ categories: string[] }>()
const emit = defineEmits<{ uploading: [value: boolean] }>()
const uploads = ref(0)
const options = computed(() => [
  ...new Set([
    '热血',
    '奇幻',
    '日常',
    '悬疑',
    '科幻',
    '恋爱',
    ...props.categories,
  ]),
])
async function upload(files: File[], callback: (urls: string[]) => void) {
  uploads.value++
  emit('uploading', true)
  try {
    callback(await Promise.all(files.map(uploadImage)))
  } catch {
    ElMessage.error('图片上传失败，请重试')
  } finally {
    uploads.value--
    emit('uploading', uploads.value > 0)
  }
}
function coverChange(file: { raw?: File }) {
  if (file.raw)
    void upload([file.raw], (urls) => {
      form.value.cover = urls[0] || ''
    })
}
</script>
<template>
  <el-form label-width="100px" class="anime-form">
    <h3>作品信息</h3>
    <el-form-item label="名称 *"
      ><el-input v-model="form.title" maxlength="191" placeholder="动漫名称"
    /></el-form-item>
    <el-form-item label="海报"
      ><div class="cover-row">
        <img v-if="form.cover" :src="form.cover" alt="海报预览" /><el-input
          v-model="form.cover"
          maxlength="500"
          placeholder="竖版海报地址"
        /><el-upload
          accept="image/*"
          :show-file-list="false"
          :auto-upload="false"
          :on-change="coverChange"
          ><el-button :loading="uploads > 0">上传</el-button></el-upload
        >
      </div></el-form-item
    >
    <div class="form-grid">
      <el-form-item label="题材"
        ><el-select v-model="form.category" allow-create filterable clearable
          ><el-option
            v-for="c in options"
            :key="c"
            :value="c"
            :label="c" /></el-select
      ></el-form-item>
      <el-form-item label="地区"
        ><el-input
          v-model="form.region"
          maxlength="100"
          placeholder="如：中国 / 日本"
      /></el-form-item>
      <el-form-item label="年份"
        ><el-input-number
          v-model="form.year"
          :min="0"
          :max="2100"
          :precision="0"
        /><span class="hint">0 表示未填写</span></el-form-item
      >
      <el-form-item label="作品来源"
        ><el-select v-model="form.source" allow-create filterable clearable
          ><el-option
            v-for="s in ['原创', '漫画改', '小说改', '游戏改']"
            :key="s"
            :value="s"
            :label="s" /></el-select
      ></el-form-item>
      <el-form-item label="播出状态"
        ><el-select v-model="form.airStatus" clearable
          ><el-option
            v-for="s in airStatuses"
            :key="s.value"
            :value="s.value"
            :label="s.label" /></el-select
      ></el-form-item>
      <el-form-item label="总集数"
        ><el-input-number
          :model-value="form.totalEpisodes ?? undefined"
          @update:model-value="form.totalEpisodes = $event ?? null"
          :min="1"
          :precision="0"
          placeholder="未知留空"
      /></el-form-item>
    </div>
    <el-form-item label="作品简介"
      ><el-input
        v-model="form.description"
        type="textarea"
        :rows="3"
        maxlength="1000"
        show-word-limit
    /></el-form-item>
    <h3>我的追番记录</h3>
    <div class="form-grid">
      <el-form-item label="观看状态"
        ><el-select v-model="form.watchStatus" clearable
          ><el-option
            v-for="s in watchStatuses"
            :key="s.value"
            :value="s.value"
            :label="s.label" /></el-select
      ></el-form-item>
      <el-form-item label="已看集数"
        ><el-input-number
          v-model="form.watched"
          :min="0"
          :max="form.totalEpisodes ?? undefined"
          :precision="0"
      /></el-form-item>
      <el-form-item label="个人评分"
        ><el-input-number
          :model-value="form.rating ?? undefined"
          @update:model-value="form.rating = $event ?? null"
          :min="0"
          :max="10"
          :step="0.5"
          :precision="1"
          placeholder="未评分"
      /></el-form-item>
      <el-form-item label="喜爱排名"
        ><el-input-number v-model="form.rank" :min="0" :precision="0" /><span
          class="hint"
          >0 不上榜，数字越小越靠前</span
        ></el-form-item
      >
    </div>
    <el-form-item label="短评 / 推荐理由"
      ><el-input
        v-model="form.review"
        type="textarea"
        maxlength="300"
        :rows="2"
        show-word-limit
        placeholder="这部作品打动我的地方"
    /></el-form-item>
    <el-form-item label="私藏推荐"
      ><el-switch v-model="form.recommended"
    /></el-form-item>
    <el-form-item v-if="form.recommended" label="推荐顺序"
      ><el-input-number
        v-model="form.recommendOrder"
        :min="0"
        :precision="0"
      /><span class="hint">数字越小越靠前</span></el-form-item
    >
    <el-form-item label="观后感"
      ><MdEditor
        v-model="form.content"
        :preview="true"
        @on-upload-img="upload"
        style="height: 420px"
    /></el-form-item>
    <el-form-item label="包含剧透"
      ><el-switch v-model="form.spoiler" /><span class="hint"
        >开启后，详情页默认折叠观后感</span
      ></el-form-item
    >
    <h3>观看入口与发布</h3>
    <el-form-item label="正版平台"
      ><el-input
        v-model="form.watchPlatform"
        maxlength="100"
        placeholder="如：哔哩哔哩"
    /></el-form-item>
    <el-form-item label="观看链接"
      ><el-input
        v-model="form.watchUrl"
        maxlength="1000"
        placeholder="https://…（选填）"
    /></el-form-item>
    <el-form-item label="展示状态"
      ><el-radio-group v-model="form.status"
        ><el-radio-button :value="1">上架</el-radio-button
        ><el-radio-button :value="0">下架</el-radio-button></el-radio-group
      ></el-form-item
    >
  </el-form>
</template>
<style scoped>
.form-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0 16px;
}
.cover-row {
  display: flex;
  align-items: center;
  width: 100%;
  gap: 12px;
}
.cover-row img {
  width: 48px;
  height: 72px;
  object-fit: cover;
  border-radius: 4px;
}
.hint {
  color: var(--muted);
  font-size: 12px;
  margin-left: 8px;
}
h3 {
  border-bottom: 1px solid var(--border);
  padding-bottom: 10px;
  margin: 20px 0;
}
@media (max-width: 700px) {
  .form-grid {
    grid-template-columns: 1fr;
  }
  .cover-row {
    flex-wrap: wrap;
  }
}
</style>
