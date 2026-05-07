<script setup lang="ts">
import { computed, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAppStore } from '@/stores/app'
import { useAuthStore } from '@/stores/auth'
import Sidebar from './components/Sidebar.vue'
import {
  Fold,
  Expand,
  Bell,
  ArrowDown,
  Close,
} from '@element-plus/icons-vue'

const appStore = useAppStore()
const authStore = useAuthStore()
const route = useRoute()
const router = useRouter()

const collapsed = computed(() => appStore.sidebarCollapsed)
const tags = computed(() => appStore.tags)
const user = computed(() => authStore.user)

watch(
  () => route.path,
  () => {
    if (!(route.meta?.hidden as boolean) && route.meta?.title) {
      appStore.setBreadcrumbs([{ title: route.meta.title as string, path: route.path }])
    }
  },
  { immediate: true }
)

function closeTag(path: string) {
  const currentPath = route.path
  appStore.removeTag(path)
  if (currentPath === path) {
    const remaining = tags.value
    if (remaining.length > 0) {
      router.push(remaining[remaining.length - 1].path)
    } else {
      router.push('/workbench')
    }
  }
}

async function handleLogout() {
  await authStore.logout()
  router.push('/login')
}
</script>

<template>
  <div class="layout-container">
    <!-- 侧边栏 -->
    <Sidebar :collapsed="collapsed" />

    <!-- 主区域 -->
    <div class="main-container">
      <!-- 顶栏 -->
      <header class="header">
        <div class="header-left">
          <el-button
            text
            :icon="collapsed ? Expand : Fold"
            size="large"
            @click="appStore.toggleSidebar()"
          />
          <el-breadcrumb separator="/" style="margin-left: 8px">
            <el-breadcrumb-item>管理后台</el-breadcrumb-item>
            <el-breadcrumb-item
              v-for="crumb in appStore.breadcrumbs"
              :key="crumb.path"
            >
              {{ crumb.title }}
            </el-breadcrumb-item>
          </el-breadcrumb>
        </div>

        <div class="header-right">
          <el-badge :value="0" :hidden="true">
            <el-button text :icon="Bell" size="large" />
          </el-badge>

          <el-dropdown trigger="click" @command="(cmd: string) => cmd === 'logout' && handleLogout()">
            <div class="user-info">
              <el-avatar :size="28" style="background: #f59e0b; color: #fff; font-size: 12px">
                {{ user?.real_name?.charAt(0) || user?.username?.charAt(0) || 'A' }}
              </el-avatar>
              <span class="username">{{ user?.real_name || user?.username }}</span>
              <el-icon><ArrowDown /></el-icon>
            </div>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item disabled>
                  {{ user?.username }}
                </el-dropdown-item>
                <el-dropdown-item divided command="logout">退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </header>

      <!-- 标签页 -->
      <div v-if="tags.length > 0" class="tags-bar">
        <div
          v-for="tag in tags"
          :key="tag.path"
          class="tag-item"
          :class="{ active: route.path === tag.path }"
          @click="router.push(tag.path)"
        >
          {{ tag.title }}
          <span
            class="close-icon"
            @click.stop="closeTag(tag.path)"
          >
            <el-icon size="10"><Close /></el-icon>
          </span>
        </div>
      </div>

      <!-- 内容区 -->
      <div class="content-wrapper">
        <router-view v-slot="{ Component }">
          <transition name="slide-fade" mode="out-in">
            <component :is="Component" />
          </transition>
        </router-view>
      </div>
    </div>
  </div>
</template>

<style scoped>
.header-left {
  display: flex;
  align-items: center;
  flex: 1;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 4px;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 6px;
  transition: background 0.2s;

  &:hover {
    background: var(--content-bg);
  }
}

.username {
  font-size: 13px;
  color: var(--text-regular);
  max-width: 100px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
