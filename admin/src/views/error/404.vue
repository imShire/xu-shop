<script setup lang="ts">
import { useRouter } from 'vue-router'
const router = useRouter()
</script>

<template>
  <div style="display: flex; flex-direction: column; align-items: center; justify-content: center; height: 100vh; gap: 16px">
    <el-icon size="80" color="#a0a0b8"><DocumentDelete /></el-icon>
    <h1 style="font-size: 32px; font-weight: 700; color: var(--text-primary)">404</h1>
    <p style="color: var(--text-secondary)">页面不存在，可能已被删除或地址有误</p>
    <el-button type="primary" @click="router.push('/workbench')">回到工作台</el-button>
  </div>
</template>
