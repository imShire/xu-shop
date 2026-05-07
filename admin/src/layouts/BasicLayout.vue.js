import { computed, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useAppStore } from '@/stores/app';
import { useAuthStore } from '@/stores/auth';
import Sidebar from './components/Sidebar.vue';
import { Fold, Expand, Bell, ArrowDown, Close, } from '@element-plus/icons-vue';
const appStore = useAppStore();
const authStore = useAuthStore();
const route = useRoute();
const router = useRouter();
const collapsed = computed(() => appStore.sidebarCollapsed);
const tags = computed(() => appStore.tags);
const user = computed(() => authStore.user);
watch(() => route.path, () => {
    if (!route.meta?.hidden && route.meta?.title) {
        appStore.setBreadcrumbs([{ title: route.meta.title, path: route.path }]);
    }
}, { immediate: true });
function closeTag(path) {
    const currentPath = route.path;
    appStore.removeTag(path);
    if (currentPath === path) {
        const remaining = tags.value;
        if (remaining.length > 0) {
            router.push(remaining[remaining.length - 1].path);
        }
        else {
            router.push('/workbench');
        }
    }
}
async function handleLogout() {
    await authStore.logout();
    router.push('/login');
}
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
// CSS variable injection 
// CSS variable injection end 
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "layout-container" },
});
/** @type {[typeof Sidebar, ]} */ ;
// @ts-ignore
const __VLS_0 = __VLS_asFunctionalComponent(Sidebar, new Sidebar({
    collapsed: (__VLS_ctx.collapsed),
}));
const __VLS_1 = __VLS_0({
    collapsed: (__VLS_ctx.collapsed),
}, ...__VLS_functionalComponentArgsRest(__VLS_0));
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "main-container" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.header, __VLS_intrinsicElements.header)({
    ...{ class: "header" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "header-left" },
});
const __VLS_3 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_4 = __VLS_asFunctionalComponent(__VLS_3, new __VLS_3({
    ...{ 'onClick': {} },
    text: true,
    icon: (__VLS_ctx.collapsed ? __VLS_ctx.Expand : __VLS_ctx.Fold),
    size: "large",
}));
const __VLS_5 = __VLS_4({
    ...{ 'onClick': {} },
    text: true,
    icon: (__VLS_ctx.collapsed ? __VLS_ctx.Expand : __VLS_ctx.Fold),
    size: "large",
}, ...__VLS_functionalComponentArgsRest(__VLS_4));
let __VLS_7;
let __VLS_8;
let __VLS_9;
const __VLS_10 = {
    onClick: (...[$event]) => {
        __VLS_ctx.appStore.toggleSidebar();
    }
};
var __VLS_6;
const __VLS_11 = {}.ElBreadcrumb;
/** @type {[typeof __VLS_components.ElBreadcrumb, typeof __VLS_components.elBreadcrumb, typeof __VLS_components.ElBreadcrumb, typeof __VLS_components.elBreadcrumb, ]} */ ;
// @ts-ignore
const __VLS_12 = __VLS_asFunctionalComponent(__VLS_11, new __VLS_11({
    separator: "/",
    ...{ style: {} },
}));
const __VLS_13 = __VLS_12({
    separator: "/",
    ...{ style: {} },
}, ...__VLS_functionalComponentArgsRest(__VLS_12));
__VLS_14.slots.default;
const __VLS_15 = {}.ElBreadcrumbItem;
/** @type {[typeof __VLS_components.ElBreadcrumbItem, typeof __VLS_components.elBreadcrumbItem, typeof __VLS_components.ElBreadcrumbItem, typeof __VLS_components.elBreadcrumbItem, ]} */ ;
// @ts-ignore
const __VLS_16 = __VLS_asFunctionalComponent(__VLS_15, new __VLS_15({}));
const __VLS_17 = __VLS_16({}, ...__VLS_functionalComponentArgsRest(__VLS_16));
__VLS_18.slots.default;
var __VLS_18;
for (const [crumb] of __VLS_getVForSourceType((__VLS_ctx.appStore.breadcrumbs))) {
    const __VLS_19 = {}.ElBreadcrumbItem;
    /** @type {[typeof __VLS_components.ElBreadcrumbItem, typeof __VLS_components.elBreadcrumbItem, typeof __VLS_components.ElBreadcrumbItem, typeof __VLS_components.elBreadcrumbItem, ]} */ ;
    // @ts-ignore
    const __VLS_20 = __VLS_asFunctionalComponent(__VLS_19, new __VLS_19({
        key: (crumb.path),
    }));
    const __VLS_21 = __VLS_20({
        key: (crumb.path),
    }, ...__VLS_functionalComponentArgsRest(__VLS_20));
    __VLS_22.slots.default;
    (crumb.title);
    var __VLS_22;
}
var __VLS_14;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "header-right" },
});
const __VLS_23 = {}.ElBadge;
/** @type {[typeof __VLS_components.ElBadge, typeof __VLS_components.elBadge, typeof __VLS_components.ElBadge, typeof __VLS_components.elBadge, ]} */ ;
// @ts-ignore
const __VLS_24 = __VLS_asFunctionalComponent(__VLS_23, new __VLS_23({
    value: (0),
    hidden: (true),
}));
const __VLS_25 = __VLS_24({
    value: (0),
    hidden: (true),
}, ...__VLS_functionalComponentArgsRest(__VLS_24));
__VLS_26.slots.default;
const __VLS_27 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_28 = __VLS_asFunctionalComponent(__VLS_27, new __VLS_27({
    text: true,
    icon: (__VLS_ctx.Bell),
    size: "large",
}));
const __VLS_29 = __VLS_28({
    text: true,
    icon: (__VLS_ctx.Bell),
    size: "large",
}, ...__VLS_functionalComponentArgsRest(__VLS_28));
var __VLS_26;
const __VLS_31 = {}.ElDropdown;
/** @type {[typeof __VLS_components.ElDropdown, typeof __VLS_components.elDropdown, typeof __VLS_components.ElDropdown, typeof __VLS_components.elDropdown, ]} */ ;
// @ts-ignore
const __VLS_32 = __VLS_asFunctionalComponent(__VLS_31, new __VLS_31({
    ...{ 'onCommand': {} },
    trigger: "click",
}));
const __VLS_33 = __VLS_32({
    ...{ 'onCommand': {} },
    trigger: "click",
}, ...__VLS_functionalComponentArgsRest(__VLS_32));
let __VLS_35;
let __VLS_36;
let __VLS_37;
const __VLS_38 = {
    onCommand: ((cmd) => cmd === 'logout' && __VLS_ctx.handleLogout())
};
__VLS_34.slots.default;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "user-info" },
});
const __VLS_39 = {}.ElAvatar;
/** @type {[typeof __VLS_components.ElAvatar, typeof __VLS_components.elAvatar, typeof __VLS_components.ElAvatar, typeof __VLS_components.elAvatar, ]} */ ;
// @ts-ignore
const __VLS_40 = __VLS_asFunctionalComponent(__VLS_39, new __VLS_39({
    size: (28),
    ...{ style: {} },
}));
const __VLS_41 = __VLS_40({
    size: (28),
    ...{ style: {} },
}, ...__VLS_functionalComponentArgsRest(__VLS_40));
__VLS_42.slots.default;
(__VLS_ctx.user?.real_name?.charAt(0) || __VLS_ctx.user?.username?.charAt(0) || 'A');
var __VLS_42;
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
    ...{ class: "username" },
});
(__VLS_ctx.user?.real_name || __VLS_ctx.user?.username);
const __VLS_43 = {}.ElIcon;
/** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
// @ts-ignore
const __VLS_44 = __VLS_asFunctionalComponent(__VLS_43, new __VLS_43({}));
const __VLS_45 = __VLS_44({}, ...__VLS_functionalComponentArgsRest(__VLS_44));
__VLS_46.slots.default;
const __VLS_47 = {}.ArrowDown;
/** @type {[typeof __VLS_components.ArrowDown, ]} */ ;
// @ts-ignore
const __VLS_48 = __VLS_asFunctionalComponent(__VLS_47, new __VLS_47({}));
const __VLS_49 = __VLS_48({}, ...__VLS_functionalComponentArgsRest(__VLS_48));
var __VLS_46;
{
    const { dropdown: __VLS_thisSlot } = __VLS_34.slots;
    const __VLS_51 = {}.ElDropdownMenu;
    /** @type {[typeof __VLS_components.ElDropdownMenu, typeof __VLS_components.elDropdownMenu, typeof __VLS_components.ElDropdownMenu, typeof __VLS_components.elDropdownMenu, ]} */ ;
    // @ts-ignore
    const __VLS_52 = __VLS_asFunctionalComponent(__VLS_51, new __VLS_51({}));
    const __VLS_53 = __VLS_52({}, ...__VLS_functionalComponentArgsRest(__VLS_52));
    __VLS_54.slots.default;
    const __VLS_55 = {}.ElDropdownItem;
    /** @type {[typeof __VLS_components.ElDropdownItem, typeof __VLS_components.elDropdownItem, typeof __VLS_components.ElDropdownItem, typeof __VLS_components.elDropdownItem, ]} */ ;
    // @ts-ignore
    const __VLS_56 = __VLS_asFunctionalComponent(__VLS_55, new __VLS_55({
        disabled: true,
    }));
    const __VLS_57 = __VLS_56({
        disabled: true,
    }, ...__VLS_functionalComponentArgsRest(__VLS_56));
    __VLS_58.slots.default;
    (__VLS_ctx.user?.username);
    var __VLS_58;
    const __VLS_59 = {}.ElDropdownItem;
    /** @type {[typeof __VLS_components.ElDropdownItem, typeof __VLS_components.elDropdownItem, typeof __VLS_components.ElDropdownItem, typeof __VLS_components.elDropdownItem, ]} */ ;
    // @ts-ignore
    const __VLS_60 = __VLS_asFunctionalComponent(__VLS_59, new __VLS_59({
        divided: true,
        command: "logout",
    }));
    const __VLS_61 = __VLS_60({
        divided: true,
        command: "logout",
    }, ...__VLS_functionalComponentArgsRest(__VLS_60));
    __VLS_62.slots.default;
    var __VLS_62;
    var __VLS_54;
}
var __VLS_34;
if (__VLS_ctx.tags.length > 0) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "tags-bar" },
    });
    for (const [tag] of __VLS_getVForSourceType((__VLS_ctx.tags))) {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ onClick: (...[$event]) => {
                    if (!(__VLS_ctx.tags.length > 0))
                        return;
                    __VLS_ctx.router.push(tag.path);
                } },
            key: (tag.path),
            ...{ class: "tag-item" },
            ...{ class: ({ active: __VLS_ctx.route.path === tag.path }) },
        });
        (tag.title);
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
            ...{ onClick: (...[$event]) => {
                    if (!(__VLS_ctx.tags.length > 0))
                        return;
                    __VLS_ctx.closeTag(tag.path);
                } },
            ...{ class: "close-icon" },
        });
        const __VLS_63 = {}.ElIcon;
        /** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
        // @ts-ignore
        const __VLS_64 = __VLS_asFunctionalComponent(__VLS_63, new __VLS_63({
            size: "10",
        }));
        const __VLS_65 = __VLS_64({
            size: "10",
        }, ...__VLS_functionalComponentArgsRest(__VLS_64));
        __VLS_66.slots.default;
        const __VLS_67 = {}.Close;
        /** @type {[typeof __VLS_components.Close, ]} */ ;
        // @ts-ignore
        const __VLS_68 = __VLS_asFunctionalComponent(__VLS_67, new __VLS_67({}));
        const __VLS_69 = __VLS_68({}, ...__VLS_functionalComponentArgsRest(__VLS_68));
        var __VLS_66;
    }
}
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "content-wrapper" },
});
const __VLS_71 = {}.RouterView;
/** @type {[typeof __VLS_components.RouterView, typeof __VLS_components.routerView, typeof __VLS_components.RouterView, typeof __VLS_components.routerView, ]} */ ;
// @ts-ignore
const __VLS_72 = __VLS_asFunctionalComponent(__VLS_71, new __VLS_71({}));
const __VLS_73 = __VLS_72({}, ...__VLS_functionalComponentArgsRest(__VLS_72));
{
    const { default: __VLS_thisSlot } = __VLS_74.slots;
    const [{ Component }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_75 = {}.transition;
    /** @type {[typeof __VLS_components.Transition, typeof __VLS_components.transition, typeof __VLS_components.Transition, typeof __VLS_components.transition, ]} */ ;
    // @ts-ignore
    const __VLS_76 = __VLS_asFunctionalComponent(__VLS_75, new __VLS_75({
        name: "slide-fade",
        mode: "out-in",
    }));
    const __VLS_77 = __VLS_76({
        name: "slide-fade",
        mode: "out-in",
    }, ...__VLS_functionalComponentArgsRest(__VLS_76));
    __VLS_78.slots.default;
    const __VLS_79 = ((Component));
    // @ts-ignore
    const __VLS_80 = __VLS_asFunctionalComponent(__VLS_79, new __VLS_79({}));
    const __VLS_81 = __VLS_80({}, ...__VLS_functionalComponentArgsRest(__VLS_80));
    var __VLS_78;
    __VLS_74.slots['' /* empty slot name completion */];
}
var __VLS_74;
/** @type {__VLS_StyleScopedClasses['layout-container']} */ ;
/** @type {__VLS_StyleScopedClasses['main-container']} */ ;
/** @type {__VLS_StyleScopedClasses['header']} */ ;
/** @type {__VLS_StyleScopedClasses['header-left']} */ ;
/** @type {__VLS_StyleScopedClasses['header-right']} */ ;
/** @type {__VLS_StyleScopedClasses['user-info']} */ ;
/** @type {__VLS_StyleScopedClasses['username']} */ ;
/** @type {__VLS_StyleScopedClasses['tags-bar']} */ ;
/** @type {__VLS_StyleScopedClasses['tag-item']} */ ;
/** @type {__VLS_StyleScopedClasses['close-icon']} */ ;
/** @type {__VLS_StyleScopedClasses['content-wrapper']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            Sidebar: Sidebar,
            Fold: Fold,
            Expand: Expand,
            Bell: Bell,
            ArrowDown: ArrowDown,
            Close: Close,
            appStore: appStore,
            route: route,
            router: router,
            collapsed: collapsed,
            tags: tags,
            user: user,
            closeTag: closeTag,
            handleLogout: handleLogout,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
