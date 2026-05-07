import { createRouter, createWebHistory } from 'vue-router';
import { useAuthStore } from '@/stores/auth';
import { useAppStore } from '@/stores/app';
const staticRoutes = [
    {
        path: '/login',
        component: () => import('@/views/login/Login.vue'),
        meta: { title: '登录' },
    },
    { path: '/403', component: () => import('@/views/error/403.vue') },
    { path: '/404', component: () => import('@/views/error/404.vue') },
    { path: '/:pathMatch(.*)*', redirect: '/404' },
];
export const asyncRoutes = [
    {
        path: '/',
        component: () => import('@/layouts/BasicLayout.vue'),
        redirect: '/workbench',
        children: [
            {
                path: 'workbench',
                component: () => import('@/views/workbench/Workbench.vue'),
                meta: { title: '工作台', icon: 'HomeFilled' },
            },
            // 商品管理
            {
                path: 'product/list',
                component: () => import('@/views/product/ProductList.vue'),
                meta: { title: '商品列表', perm: 'product.view', icon: 'Goods' },
            },
            {
                path: 'product/edit/:id?',
                component: () => import('@/views/product/ProductEdit.vue'),
                meta: { title: '编辑商品', perm: 'product.edit', hidden: true },
            },
            {
                path: 'product/category',
                component: () => import('@/views/product/CategoryList.vue'),
                meta: { title: '商品分类', perm: 'category.view' },
            },
            // 库存
            {
                path: 'inventory/alerts',
                component: () => import('@/views/inventory/AlertList.vue'),
                meta: { title: '库存预警', perm: 'inventory.view' },
            },
            {
                path: 'inventory/logs',
                component: () => import('@/views/inventory/LogList.vue'),
                meta: { title: '库存日志', perm: 'inventory.view' },
            },
            // 订单
            {
                path: 'order/list',
                component: () => import('@/views/order/OrderList.vue'),
                meta: { title: '全部订单', perm: 'order.view', icon: 'List' },
            },
            {
                path: 'order/detail/:id',
                component: () => import('@/views/order/OrderDetail.vue'),
                meta: { title: '订单详情', hidden: true },
            },
            // 发货
            {
                path: 'shipping/pending',
                component: () => import('@/views/shipping/PendingList.vue'),
                meta: { title: '待发货', perm: 'shipment.view' },
            },
            {
                path: 'shipping/shipped',
                component: () => import('@/views/shipping/ShippedList.vue'),
                meta: { title: '已发货', perm: 'shipment.view' },
            },
            {
                path: 'shipping/sender',
                component: () => import('@/views/shipping/SenderAddress.vue'),
                meta: { title: '发件地址', perm: 'system.setting.view' },
            },
            {
                path: 'shipping/carrier',
                component: () => import('@/views/shipping/CarrierConfig.vue'),
                meta: { title: '快递配置', perm: 'system.setting.view' },
            },
            // 售后
            {
                path: 'aftersale/list',
                component: () => import('@/views/aftersale/AftersaleList.vue'),
                meta: { title: '售后管理', perm: 'aftersale.view', icon: 'Service' },
            },
            // 支付
            {
                path: 'payment/list',
                component: () => import('@/views/payment/PaymentList.vue'),
                meta: { title: '支付记录', perm: 'payment.view' },
            },
            {
                path: 'payment/refunds',
                component: () => import('@/views/payment/RefundList.vue'),
                meta: { title: '退款记录', perm: 'payment.view' },
            },
            {
                path: 'payment/reconcile',
                component: () => import('@/views/payment/ReconcileList.vue'),
                meta: { title: '对账管理', perm: 'reconcile.view' },
            },
            // 用户
            {
                path: 'user/list',
                component: () => import('@/views/user/UserList.vue'),
                meta: { title: '用户管理', perm: 'user.view', icon: 'User' },
            },
            {
                path: 'user/:id',
                component: () => import('@/views/user/UserDetail.vue'),
                meta: { title: '用户详情', perm: 'user.view', hidden: true },
            },
            // 私域
            {
                path: 'private-domain/channel',
                component: () => import('@/views/private-domain/ChannelCodeList.vue'),
                meta: { title: '渠道码', perm: 'channel.view' },
            },
            {
                path: 'private-domain/tags',
                component: () => import('@/views/private-domain/TagList.vue'),
                meta: { title: '客户标签', perm: 'tag.view' },
            },
            // 内容管理
            {
                path: 'content/banners',
                component: () => import('@/views/content/BannerList.vue'),
                meta: { title: 'Banner管理', perm: 'banner.view', icon: 'Picture' },
            },
            {
                path: 'content/nav-icons',
                component: () => import('@/views/content/NavIconList.vue'),
                meta: { title: '金刚区管理', perm: 'nav_icon.view', icon: 'Grid' },
            },
            // 统计
            {
                path: 'stats/overview',
                component: () => import('@/views/stats/SalesOverview.vue'),
                meta: { title: '销售概览', perm: 'stats.view', icon: 'TrendCharts' },
            },
            {
                path: 'stats/products',
                component: () => import('@/views/stats/ProductSales.vue'),
                meta: { title: '商品销量', perm: 'stats.view' },
            },
            {
                path: 'stats/channels',
                component: () => import('@/views/stats/ChannelStats.vue'),
                meta: { title: '渠道分析', perm: 'stats.view' },
            },
            {
                path: 'stats/users',
                component: () => import('@/views/stats/UserStats.vue'),
                meta: { title: '用户分析', perm: 'stats.view' },
            },
            // 通知
            {
                path: 'notification/list',
                component: () => import('@/views/notification/NotificationList.vue'),
                meta: { title: '通知记录', perm: 'notif.view' },
            },
            {
                path: 'notification/templates',
                component: () => import('@/views/notification/TemplateList.vue'),
                meta: { title: '通知模板', perm: 'notif.config' },
            },
            // 系统
            {
                path: 'system/admins',
                component: () => import('@/views/system/AdminList.vue'),
                meta: { title: '员工管理', perm: 'system.admin.view', icon: 'Setting' },
            },
            {
                path: 'system/roles',
                component: () => import('@/views/system/RoleList.vue'),
                meta: { title: '角色权限', perm: 'system.role.view' },
            },
            {
                path: 'system/audit',
                component: () => import('@/views/system/AuditLog.vue'),
                meta: { title: '操作日志', perm: 'system.audit.view' },
            },
            {
                path: 'system/freight',
                component: () => import('@/views/system/FreightTemplate.vue'),
                meta: { title: '运费模板', perm: 'system.setting.view' },
            },
            {
                path: 'system/upload',
                component: () => import('@/views/system/UploadSetting.vue'),
                meta: { title: '上传设置', perm: 'system.upload.view' },
            },
            {
                path: 'system/settings',
                component: () => import('@/views/system/SystemSetting.vue'),
                meta: { title: '系统设置', perm: 'system.setting.view' },
            },
        ],
    },
];
export const router = createRouter({
    history: createWebHistory(),
    routes: [...staticRoutes, ...asyncRoutes],
});
router.beforeEach(async (to, _from, next) => {
    const auth = useAuthStore();
    // 等待 auth 初始化完成（刷新场景下 init() 是异步的）
    if (!auth.ready) {
        await auth.init();
    }
    if (to.path === '/login') {
        if (auth.isLoggedIn)
            return next('/');
        return next();
    }
    if (!auth.isLoggedIn)
        return next('/login');
    const perm = to.meta?.perm;
    if (perm && !auth.isSuperAdmin && !auth.perms.includes(perm)) {
        return next('/403');
    }
    // 同步 tags
    const appStore = useAppStore();
    const title = to.meta?.title || to.path;
    if (!to.meta?.hidden) {
        appStore.addTag({ path: to.path, title });
    }
    next();
});
