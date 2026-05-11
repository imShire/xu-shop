<script setup lang="ts">
import { ref, onMounted, shallowRef, onBeforeUnmount } from 'vue'
import { Editor, Toolbar } from '@wangeditor/editor-for-vue'
import '@wangeditor/editor/dist/css/style.css'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { getArticle, createArticle, updateArticle } from '@/api/cms'
import UploadImage from '@/components/UploadImage/index.vue'

const route = useRoute()
const router = useRouter()
const isEdit = !!route.params.id
const loading = ref(false)
const saving = ref(false)

const form = ref({
  title: '',
  cover: '',
  content: '',
  status: 'draft' as 'draft' | 'published',
  sort: 0,
})

// wangEditor
const editorRef = shallowRef()
const editorConfig = { MENU_CONF: {} }
onBeforeUnmount(() => editorRef.value?.destroy())

onMounted(async () => {
  if (isEdit) {
    loading.value = true
    try {
      const data = await getArticle(route.params.id as string)
      form.value = {
        title: data.title ?? '',
        cover: data.cover ?? '',
        content: data.content ?? '',
        status: data.status ?? 'draft',
        sort: data.sort ?? 0,
      }
    } finally {
      loading.value = false
    }
  }
})

async function save() {
  if (!form.value.title.trim()) {
    ElMessage.error('请输入文章标题')
    return
  }
  saving.value = true
  try {
    if (isEdit) {
      await updateArticle(route.params.id as string, form.value)
    } else {
      await createArticle(form.value)
    }
    ElMessage.success('保存成功')
    router.push('/cms/articles')
  } catch {
    ElMessage.error('保存失败')
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <div class="page-card" v-loading="loading">
    <div style="margin-bottom: 16px">
      <h3>{{ isEdit ? '编辑文章' : '新增文章' }}</h3>
    </div>
    <el-form :model="form" label-width="80px" style="max-width: 800px">
      <el-form-item label="标题" required>
        <el-input v-model="form.title" placeholder="请输入文章标题" />
      </el-form-item>
      <el-form-item label="封面">
        <UploadImage v-model="form.cover" />
      </el-form-item>
      <el-form-item label="内容">
        <div class="editor-wrapper">
          <Toolbar
            :editor="editorRef"
            :defaultConfig="{}"
            mode="default"
            style="border-bottom: 1px solid #ccc"
          />
          <Editor
            v-model="form.content"
            :defaultConfig="editorConfig"
            mode="default"
            style="height: 400px; overflow-y: hidden"
            @onCreated="(e: any) => (editorRef = e)"
          />
        </div>
      </el-form-item>
      <el-form-item label="状态">
        <el-radio-group v-model="form.status">
          <el-radio value="draft">草稿</el-radio>
          <el-radio value="published">发布</el-radio>
        </el-radio-group>
      </el-form-item>
      <el-form-item label="排序">
        <el-input-number v-model="form.sort" :min="0" />
      </el-form-item>
      <el-form-item>
        <el-button type="primary" :loading="saving" @click="save">保存</el-button>
        <el-button @click="router.push('/cms/articles')">取消</el-button>
      </el-form-item>
    </el-form>
  </div>
</template>

<style scoped>
.editor-wrapper {
  border: 1px solid #ccc;
  width: 100%;
}
</style>
