<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const auth = useAuthStore()
const activeMenu = computed(() => route.path)

interface MenuItem {
  title: string
  path?: string
  icon?: string
  perm?: string
  superAdminOnly?: boolean
  children?: MenuItem[]
}

const menus: MenuItem[] = [
  { title: '工作台', path: '/workbench', icon: 'HomeFilled' },
  {
    title: '商品管理', icon: 'Goods',
    children: [
      { title: '商品列表', path: '/product/list', perm: 'product.view' },
      { title: '商品分类', path: '/product/category', perm: 'category.view' },
    ],
  },
  {
    title: '库存管理', icon: 'Box',
    children: [
      { title: '库存预警', path: '/inventory/alerts', perm: 'inventory.view' },
      { title: '库存日志', path: '/inventory/logs', perm: 'inventory.view' },
    ],
  },
  {
    title: '订单管理', icon: 'List',
    children: [
      { title: '全部订单', path: '/order/list', perm: 'order.view' },
    ],
  },
  {
    title: '发货管理', icon: 'Van',
    children: [
      { title: '待发货', path: '/shipping/pending', perm: 'shipment.view' },
      { title: '已发货', path: '/shipping/shipped', perm: 'shipment.view' },
      { title: '发件地址', path: '/shipping/sender', perm: 'system.setting.view' },
      { title: '快递配置', path: '/shipping/carrier', perm: 'system.setting.view' },
    ],
  },
  {
    title: '售后管理', icon: 'Service',
    children: [
      { title: '售后列表', path: '/aftersale/list', perm: 'aftersale.view' },
    ],
  },
  {
    title: '支付管理', icon: 'CreditCard',
    children: [
      { title: '支付记录', path: '/payment/list', perm: 'payment.view' },
      { title: '退款记录', path: '/payment/refunds', perm: 'payment.view' },
      { title: '对账管理', path: '/payment/reconcile', perm: 'reconcile.view' },
    ],
  },
  { title: '用户管理', path: '/user/list', icon: 'User', perm: 'user.view' },
  {
    title: '私域运营', icon: 'Connection',
    children: [
      { title: '渠道码', path: '/private-domain/channel', perm: 'channel.view' },
      { title: '客户标签', path: '/private-domain/tags', perm: 'tag.view' },
    ],
  },
  {
    title: '内容管理', icon: 'Picture',
    children: [
      { title: 'Banner管理', path: '/content/banners', perm: 'banner.view' },
      { title: '金刚区', path: '/content/nav-icons', perm: 'nav_icon.view' },
      { title: '首页装修', path: '/decorate/home', perm: 'decorate.view' },
      { title: '文章列表', path: '/cms/articles', perm: 'cms.article.view' },
    ],
  },
  {
    title: '数据统计', icon: 'TrendCharts',
    children: [
      { title: '销售概览', path: '/stats/overview', perm: 'stats.view' },
      { title: '商品销量', path: '/stats/products', perm: 'stats.view' },
      { title: '渠道分析', path: '/stats/channels', perm: 'stats.view' },
      { title: '用户分析', path: '/stats/users', perm: 'stats.view' },
    ],
  },
  {
    title: '通知管理', icon: 'Bell',
    children: [
      { title: '通知记录', path: '/notification/list', perm: 'notif.view' },
      { title: '通知模板', path: '/notification/templates', perm: 'notif.config' },
    ],
  },
  {
    title: '系统管理', icon: 'Setting',
    children: [
      { title: '员工管理', path: '/system/admins', perm: 'system.admin.view' },
      { title: '角色权限', path: '/system/roles', perm: 'system.role.view' },
      { title: '运费模板', path: '/system/freight', perm: 'system.setting.view' },
      { title: '上传设置', path: '/system/upload', perm: 'system.upload.view' },
      { title: '系统设置', path: '/system/settings', perm: 'system.setting.view' },
      { title: '操作日志', path: '/system/audit', perm: 'system.audit.view' },
    ],
  },
]

function canAccess(item: MenuItem) {
  if (item.superAdminOnly && !auth.isSuperAdmin) return false
  if (!item.perm) return true
  return auth.isSuperAdmin || auth.perms.includes(item.perm)
}

const visibleMenus = computed(() =>
  menus
    .map((item) => {
      if (!item.children) return item
      const children = item.children.filter(canAccess)
      return { ...item, children }
    })
    .filter((item) => (item.children ? item.children.length > 0 : canAccess(item)))
)

interface Props {
  collapsed: boolean
}
defineProps<Props>()
</script>

<template>
  <div class="sidebar" :class="{ collapsed }">
    <!-- Logo -->
    <div class="sidebar-logo">
      <el-icon size="24" color="#f59e0b"><Shop /></el-icon>
      <span v-if="!collapsed" class="logo-text">xu-shop</span>
    </div>

    <!-- 菜单 -->
    <el-menu
      :default-active="activeMenu"
      :collapse="collapsed"
      :collapse-transition="false"
      :default-openeds="[]"
      router
      background-color="#1c1c27"
      text-color="#a0a0b8"
      active-text-color="#f59e0b"
      class="sidebar-menu"
    >
      <template v-for="item in visibleMenus" :key="item.path || item.title">
        <!-- 无子菜单 -->
        <el-menu-item v-if="!item.children" :index="item.path!">
          <el-icon v-if="item.icon"><component :is="item.icon" /></el-icon>
          <template #title>{{ item.title }}</template>
        </el-menu-item>

        <!-- 有子菜单 -->
        <el-sub-menu v-else :index="item.title">
          <template #title>
            <el-icon v-if="item.icon"><component :is="item.icon" /></el-icon>
            <span>{{ item.title }}</span>
          </template>
          <el-menu-item
            v-for="child in item.children"
            :key="child.path"
            :index="child.path!"
          >
            {{ child.title }}
          </el-menu-item>
        </el-sub-menu>
      </template>
    </el-menu>
  </div>
</template>

<style scoped>
.sidebar-logo {
  height: var(--header-height);
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  padding: 0 16px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
  flex-shrink: 0;
}

.logo-text {
  font-size: 18px;
  font-weight: 700;
  color: #fff;
  letter-spacing: 1px;
}

.sidebar-menu {
  flex: 1;
  overflow-y: auto;
  overflow-x: hidden;
  padding: 8px 0;
}

.sidebar-menu:not(.el-menu--collapse) {
  width: var(--sidebar-width);
}
</style>
