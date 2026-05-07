import { computed } from 'vue';
import { useRoute } from 'vue-router';
import { useAuthStore } from '@/stores/auth';
const route = useRoute();
const auth = useAuthStore();
const activeMenu = computed(() => route.path);
const menus = [
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
];
function canAccess(item) {
    if (item.superAdminOnly && !auth.isSuperAdmin)
        return false;
    if (!item.perm)
        return true;
    return auth.isSuperAdmin || auth.perms.includes(item.perm);
}
const visibleMenus = computed(() => menus
    .map((item) => {
    if (!item.children)
        return item;
    const children = item.children.filter(canAccess);
    return { ...item, children };
})
    .filter((item) => (item.children ? item.children.length > 0 : canAccess(item))));
const __VLS_props = defineProps();
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
/** @type {__VLS_StyleScopedClasses['sidebar-menu']} */ ;
// CSS variable injection 
// CSS variable injection end 
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "sidebar" },
    ...{ class: ({ collapsed: __VLS_ctx.collapsed }) },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "sidebar-logo" },
});
const __VLS_0 = {}.ElIcon;
/** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
// @ts-ignore
const __VLS_1 = __VLS_asFunctionalComponent(__VLS_0, new __VLS_0({
    size: "24",
    color: "#f59e0b",
}));
const __VLS_2 = __VLS_1({
    size: "24",
    color: "#f59e0b",
}, ...__VLS_functionalComponentArgsRest(__VLS_1));
__VLS_3.slots.default;
const __VLS_4 = {}.Shop;
/** @type {[typeof __VLS_components.Shop, ]} */ ;
// @ts-ignore
const __VLS_5 = __VLS_asFunctionalComponent(__VLS_4, new __VLS_4({}));
const __VLS_6 = __VLS_5({}, ...__VLS_functionalComponentArgsRest(__VLS_5));
var __VLS_3;
if (!__VLS_ctx.collapsed) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ class: "logo-text" },
    });
}
const __VLS_8 = {}.ElMenu;
/** @type {[typeof __VLS_components.ElMenu, typeof __VLS_components.elMenu, typeof __VLS_components.ElMenu, typeof __VLS_components.elMenu, ]} */ ;
// @ts-ignore
const __VLS_9 = __VLS_asFunctionalComponent(__VLS_8, new __VLS_8({
    defaultActive: (__VLS_ctx.activeMenu),
    collapse: (__VLS_ctx.collapsed),
    collapseTransition: (false),
    defaultOpeneds: ([]),
    router: true,
    backgroundColor: "#1c1c27",
    textColor: "#a0a0b8",
    activeTextColor: "#f59e0b",
    ...{ class: "sidebar-menu" },
}));
const __VLS_10 = __VLS_9({
    defaultActive: (__VLS_ctx.activeMenu),
    collapse: (__VLS_ctx.collapsed),
    collapseTransition: (false),
    defaultOpeneds: ([]),
    router: true,
    backgroundColor: "#1c1c27",
    textColor: "#a0a0b8",
    activeTextColor: "#f59e0b",
    ...{ class: "sidebar-menu" },
}, ...__VLS_functionalComponentArgsRest(__VLS_9));
__VLS_11.slots.default;
for (const [item] of __VLS_getVForSourceType((__VLS_ctx.visibleMenus))) {
    (item.path || item.title);
    if (!item.children) {
        const __VLS_12 = {}.ElMenuItem;
        /** @type {[typeof __VLS_components.ElMenuItem, typeof __VLS_components.elMenuItem, typeof __VLS_components.ElMenuItem, typeof __VLS_components.elMenuItem, ]} */ ;
        // @ts-ignore
        const __VLS_13 = __VLS_asFunctionalComponent(__VLS_12, new __VLS_12({
            index: (item.path),
        }));
        const __VLS_14 = __VLS_13({
            index: (item.path),
        }, ...__VLS_functionalComponentArgsRest(__VLS_13));
        __VLS_15.slots.default;
        if (item.icon) {
            const __VLS_16 = {}.ElIcon;
            /** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
            // @ts-ignore
            const __VLS_17 = __VLS_asFunctionalComponent(__VLS_16, new __VLS_16({}));
            const __VLS_18 = __VLS_17({}, ...__VLS_functionalComponentArgsRest(__VLS_17));
            __VLS_19.slots.default;
            const __VLS_20 = ((item.icon));
            // @ts-ignore
            const __VLS_21 = __VLS_asFunctionalComponent(__VLS_20, new __VLS_20({}));
            const __VLS_22 = __VLS_21({}, ...__VLS_functionalComponentArgsRest(__VLS_21));
            var __VLS_19;
        }
        {
            const { title: __VLS_thisSlot } = __VLS_15.slots;
            (item.title);
        }
        var __VLS_15;
    }
    else {
        const __VLS_24 = {}.ElSubMenu;
        /** @type {[typeof __VLS_components.ElSubMenu, typeof __VLS_components.elSubMenu, typeof __VLS_components.ElSubMenu, typeof __VLS_components.elSubMenu, ]} */ ;
        // @ts-ignore
        const __VLS_25 = __VLS_asFunctionalComponent(__VLS_24, new __VLS_24({
            index: (item.title),
        }));
        const __VLS_26 = __VLS_25({
            index: (item.title),
        }, ...__VLS_functionalComponentArgsRest(__VLS_25));
        __VLS_27.slots.default;
        {
            const { title: __VLS_thisSlot } = __VLS_27.slots;
            if (item.icon) {
                const __VLS_28 = {}.ElIcon;
                /** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
                // @ts-ignore
                const __VLS_29 = __VLS_asFunctionalComponent(__VLS_28, new __VLS_28({}));
                const __VLS_30 = __VLS_29({}, ...__VLS_functionalComponentArgsRest(__VLS_29));
                __VLS_31.slots.default;
                const __VLS_32 = ((item.icon));
                // @ts-ignore
                const __VLS_33 = __VLS_asFunctionalComponent(__VLS_32, new __VLS_32({}));
                const __VLS_34 = __VLS_33({}, ...__VLS_functionalComponentArgsRest(__VLS_33));
                var __VLS_31;
            }
            __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
            (item.title);
        }
        for (const [child] of __VLS_getVForSourceType((item.children))) {
            const __VLS_36 = {}.ElMenuItem;
            /** @type {[typeof __VLS_components.ElMenuItem, typeof __VLS_components.elMenuItem, typeof __VLS_components.ElMenuItem, typeof __VLS_components.elMenuItem, ]} */ ;
            // @ts-ignore
            const __VLS_37 = __VLS_asFunctionalComponent(__VLS_36, new __VLS_36({
                key: (child.path),
                index: (child.path),
            }));
            const __VLS_38 = __VLS_37({
                key: (child.path),
                index: (child.path),
            }, ...__VLS_functionalComponentArgsRest(__VLS_37));
            __VLS_39.slots.default;
            (child.title);
            var __VLS_39;
        }
        var __VLS_27;
    }
}
var __VLS_11;
/** @type {__VLS_StyleScopedClasses['sidebar']} */ ;
/** @type {__VLS_StyleScopedClasses['sidebar-logo']} */ ;
/** @type {__VLS_StyleScopedClasses['logo-text']} */ ;
/** @type {__VLS_StyleScopedClasses['sidebar-menu']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            activeMenu: activeMenu,
            visibleMenus: visibleMenus,
        };
    },
    __typeProps: {},
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
    __typeProps: {},
});
; /* PartiallyEnd: #4569/main.vue */
