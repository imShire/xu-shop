<script setup lang="ts">
import { useRouter } from 'vue-router'
const router = useRouter()
</script>

<template>
  <div style="display: flex; flex-direction: column; align-items: center; justify-content: center; height: 100vh; gap: 16px">
    <el-icon size="80" color="#f59e0b"><Warning /></el-icon>
    <h1 style="font-size: 32px; font-weight: 700; color: var(--text-primary)">403</h1>
    <p style="color: var(--text-secondary)">抱歉，您没有权限访问该页面</p>
    <div style="display: flex; gap: 12px">
      <el-button @click="router.back()">返回上页</el-button>
      <el-button type="primary" @click="router.push('/workbench')">回到工作台</el-button>
    </div>
  </div>
</template>
