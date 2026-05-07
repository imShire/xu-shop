import { ref, onMounted } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { ElMessage, ElMessageBox } from 'element-plus';
import { getUserDetail, disableUser, enableUser } from '@/api/user';
import { getOrderList } from '@/api/order';
import { formatAmount, formatTime, orderStatusMap } from '@/utils/format';
const route = useRoute();
const router = useRouter();
const userId = String(route.params.id);
const loading = ref(false);
const user = ref(null);
const actionLoading = ref(false);
const ordersLoading = ref(false);
const orders = ref([]);
async function loadUser() {
    loading.value = true;
    try {
        user.value = await getUserDetail(userId);
    }
    finally {
        loading.value = false;
    }
}
async function loadOrders() {
    ordersLoading.value = true;
    try {
        const res = await getOrderList({ user_id: userId, page: 1, page_size: 10 });
        orders.value = res?.list ?? res ?? [];
    }
    catch {
        orders.value = [];
    }
    finally {
        ordersLoading.value = false;
    }
}
async function handleToggleStatus() {
    if (!user.value)
        return;
    const isActive = user.value.status === 'active';
    const action = isActive ? '禁用' : '启用';
    try {
        await ElMessageBox.confirm(`确认${action}该用户？`, '提示', { type: 'warning' });
        actionLoading.value = true;
        if (isActive) {
            await disableUser(userId);
        }
        else {
            await enableUser(userId);
        }
        ElMessage.success(`${action}成功`);
        await loadUser();
    }
    catch (e) {
        if (e !== 'cancel')
            ElMessage.error(e?.message || '操作失败');
    }
    finally {
        actionLoading.value = false;
    }
}
onMounted(async () => {
    await loadUser();
    loadOrders();
});
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({});
__VLS_asFunctionalDirective(__VLS_directives.vLoading)(null, { ...__VLS_directiveBindingRestFields, value: (__VLS_ctx.loading) }, null, null);
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ style: {} },
});
const __VLS_0 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_1 = __VLS_asFunctionalComponent(__VLS_0, new __VLS_0({
    ...{ 'onClick': {} },
    text: true,
    icon: ('ArrowLeft'),
}));
const __VLS_2 = __VLS_1({
    ...{ 'onClick': {} },
    text: true,
    icon: ('ArrowLeft'),
}, ...__VLS_functionalComponentArgsRest(__VLS_1));
let __VLS_4;
let __VLS_5;
let __VLS_6;
const __VLS_7 = {
    onClick: (...[$event]) => {
        __VLS_ctx.router.push('/user/list');
    }
};
__VLS_3.slots.default;
var __VLS_3;
__VLS_asFunctionalElement(__VLS_intrinsicElements.h3, __VLS_intrinsicElements.h3)({
    ...{ style: {} },
});
if (__VLS_ctx.user) {
    const __VLS_8 = {}.ElCard;
    /** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
    // @ts-ignore
    const __VLS_9 = __VLS_asFunctionalComponent(__VLS_8, new __VLS_8({
        ...{ style: {} },
    }));
    const __VLS_10 = __VLS_9({
        ...{ style: {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_9));
    __VLS_11.slots.default;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ style: {} },
    });
    const __VLS_12 = {}.ElAvatar;
    /** @type {[typeof __VLS_components.ElAvatar, typeof __VLS_components.elAvatar, typeof __VLS_components.ElAvatar, typeof __VLS_components.elAvatar, ]} */ ;
    // @ts-ignore
    const __VLS_13 = __VLS_asFunctionalComponent(__VLS_12, new __VLS_12({
        size: (72),
        src: (__VLS_ctx.user.avatar),
    }));
    const __VLS_14 = __VLS_13({
        size: (72),
        src: (__VLS_ctx.user.avatar),
    }, ...__VLS_functionalComponentArgsRest(__VLS_13));
    __VLS_15.slots.default;
    (__VLS_ctx.user.nickname?.charAt(0) || '?');
    var __VLS_15;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ style: {} },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ style: {} },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ style: {} },
    });
    (__VLS_ctx.user.nickname || '未设置昵称');
    const __VLS_16 = {}.ElTag;
    /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
    // @ts-ignore
    const __VLS_17 = __VLS_asFunctionalComponent(__VLS_16, new __VLS_16({
        type: (__VLS_ctx.user.status === 'active' ? 'success' : 'danger'),
        size: "small",
    }));
    const __VLS_18 = __VLS_17({
        type: (__VLS_ctx.user.status === 'active' ? 'success' : 'danger'),
        size: "small",
    }, ...__VLS_functionalComponentArgsRest(__VLS_17));
    __VLS_19.slots.default;
    (__VLS_ctx.user.status === 'active' ? '正常' : '禁用');
    var __VLS_19;
    const __VLS_20 = {}.ElDescriptions;
    /** @type {[typeof __VLS_components.ElDescriptions, typeof __VLS_components.elDescriptions, typeof __VLS_components.ElDescriptions, typeof __VLS_components.elDescriptions, ]} */ ;
    // @ts-ignore
    const __VLS_21 = __VLS_asFunctionalComponent(__VLS_20, new __VLS_20({
        column: (2),
        size: "small",
    }));
    const __VLS_22 = __VLS_21({
        column: (2),
        size: "small",
    }, ...__VLS_functionalComponentArgsRest(__VLS_21));
    __VLS_23.slots.default;
    const __VLS_24 = {}.ElDescriptionsItem;
    /** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
    // @ts-ignore
    const __VLS_25 = __VLS_asFunctionalComponent(__VLS_24, new __VLS_24({
        label: "手机号",
    }));
    const __VLS_26 = __VLS_25({
        label: "手机号",
    }, ...__VLS_functionalComponentArgsRest(__VLS_25));
    __VLS_27.slots.default;
    (__VLS_ctx.user.phone || '-');
    var __VLS_27;
    const __VLS_28 = {}.ElDescriptionsItem;
    /** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
    // @ts-ignore
    const __VLS_29 = __VLS_asFunctionalComponent(__VLS_28, new __VLS_28({
        label: "用户 ID",
    }));
    const __VLS_30 = __VLS_29({
        label: "用户 ID",
    }, ...__VLS_functionalComponentArgsRest(__VLS_29));
    __VLS_31.slots.default;
    (__VLS_ctx.user.id);
    var __VLS_31;
    const __VLS_32 = {}.ElDescriptionsItem;
    /** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
    // @ts-ignore
    const __VLS_33 = __VLS_asFunctionalComponent(__VLS_32, new __VLS_32({
        label: "注册时间",
    }));
    const __VLS_34 = __VLS_33({
        label: "注册时间",
    }, ...__VLS_functionalComponentArgsRest(__VLS_33));
    __VLS_35.slots.default;
    (__VLS_ctx.formatTime(__VLS_ctx.user.created_at));
    var __VLS_35;
    const __VLS_36 = {}.ElDescriptionsItem;
    /** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
    // @ts-ignore
    const __VLS_37 = __VLS_asFunctionalComponent(__VLS_36, new __VLS_36({
        label: "最近登录",
    }));
    const __VLS_38 = __VLS_37({
        label: "最近登录",
    }, ...__VLS_functionalComponentArgsRest(__VLS_37));
    __VLS_39.slots.default;
    (__VLS_ctx.formatTime(__VLS_ctx.user.last_login_at));
    var __VLS_39;
    var __VLS_23;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({});
    const __VLS_40 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_41 = __VLS_asFunctionalComponent(__VLS_40, new __VLS_40({
        ...{ 'onClick': {} },
        type: (__VLS_ctx.user.status === 'active' ? 'danger' : 'success'),
        loading: (__VLS_ctx.actionLoading),
    }));
    const __VLS_42 = __VLS_41({
        ...{ 'onClick': {} },
        type: (__VLS_ctx.user.status === 'active' ? 'danger' : 'success'),
        loading: (__VLS_ctx.actionLoading),
    }, ...__VLS_functionalComponentArgsRest(__VLS_41));
    let __VLS_44;
    let __VLS_45;
    let __VLS_46;
    const __VLS_47 = {
        onClick: (__VLS_ctx.handleToggleStatus)
    };
    __VLS_43.slots.default;
    (__VLS_ctx.user.status === 'active' ? '禁用用户' : '启用用户');
    var __VLS_43;
    var __VLS_11;
    const __VLS_48 = {}.ElCard;
    /** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
    // @ts-ignore
    const __VLS_49 = __VLS_asFunctionalComponent(__VLS_48, new __VLS_48({}));
    const __VLS_50 = __VLS_49({}, ...__VLS_functionalComponentArgsRest(__VLS_49));
    __VLS_51.slots.default;
    const __VLS_52 = {}.ElTabs;
    /** @type {[typeof __VLS_components.ElTabs, typeof __VLS_components.elTabs, typeof __VLS_components.ElTabs, typeof __VLS_components.elTabs, ]} */ ;
    // @ts-ignore
    const __VLS_53 = __VLS_asFunctionalComponent(__VLS_52, new __VLS_52({}));
    const __VLS_54 = __VLS_53({}, ...__VLS_functionalComponentArgsRest(__VLS_53));
    __VLS_55.slots.default;
    const __VLS_56 = {}.ElTabPane;
    /** @type {[typeof __VLS_components.ElTabPane, typeof __VLS_components.elTabPane, typeof __VLS_components.ElTabPane, typeof __VLS_components.elTabPane, ]} */ ;
    // @ts-ignore
    const __VLS_57 = __VLS_asFunctionalComponent(__VLS_56, new __VLS_56({
        label: "订单记录",
        name: "orders",
    }));
    const __VLS_58 = __VLS_57({
        label: "订单记录",
        name: "orders",
    }, ...__VLS_functionalComponentArgsRest(__VLS_57));
    __VLS_59.slots.default;
    const __VLS_60 = {}.ElTable;
    /** @type {[typeof __VLS_components.ElTable, typeof __VLS_components.elTable, typeof __VLS_components.ElTable, typeof __VLS_components.elTable, ]} */ ;
    // @ts-ignore
    const __VLS_61 = __VLS_asFunctionalComponent(__VLS_60, new __VLS_60({
        data: (__VLS_ctx.orders),
        ...{ style: {} },
    }));
    const __VLS_62 = __VLS_61({
        data: (__VLS_ctx.orders),
        ...{ style: {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_61));
    __VLS_asFunctionalDirective(__VLS_directives.vLoading)(null, { ...__VLS_directiveBindingRestFields, value: (__VLS_ctx.ordersLoading) }, null, null);
    __VLS_63.slots.default;
    const __VLS_64 = {}.ElTableColumn;
    /** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
    // @ts-ignore
    const __VLS_65 = __VLS_asFunctionalComponent(__VLS_64, new __VLS_64({
        label: "订单号",
        prop: "order_no",
        minWidth: "160",
    }));
    const __VLS_66 = __VLS_65({
        label: "订单号",
        prop: "order_no",
        minWidth: "160",
    }, ...__VLS_functionalComponentArgsRest(__VLS_65));
    const __VLS_68 = {}.ElTableColumn;
    /** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
    // @ts-ignore
    const __VLS_69 = __VLS_asFunctionalComponent(__VLS_68, new __VLS_68({
        label: "状态",
        width: "100",
    }));
    const __VLS_70 = __VLS_69({
        label: "状态",
        width: "100",
    }, ...__VLS_functionalComponentArgsRest(__VLS_69));
    __VLS_71.slots.default;
    {
        const { default: __VLS_thisSlot } = __VLS_71.slots;
        const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
        const __VLS_72 = {}.ElTag;
        /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
        // @ts-ignore
        const __VLS_73 = __VLS_asFunctionalComponent(__VLS_72, new __VLS_72({
            type: (__VLS_ctx.orderStatusMap[row.status]?.type || ''),
            size: "small",
        }));
        const __VLS_74 = __VLS_73({
            type: (__VLS_ctx.orderStatusMap[row.status]?.type || ''),
            size: "small",
        }, ...__VLS_functionalComponentArgsRest(__VLS_73));
        __VLS_75.slots.default;
        (__VLS_ctx.orderStatusMap[row.status]?.label || row.status);
        var __VLS_75;
    }
    var __VLS_71;
    const __VLS_76 = {}.ElTableColumn;
    /** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
    // @ts-ignore
    const __VLS_77 = __VLS_asFunctionalComponent(__VLS_76, new __VLS_76({
        label: "实付金额",
        width: "110",
    }));
    const __VLS_78 = __VLS_77({
        label: "实付金额",
        width: "110",
    }, ...__VLS_functionalComponentArgsRest(__VLS_77));
    __VLS_79.slots.default;
    {
        const { default: __VLS_thisSlot } = __VLS_79.slots;
        const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
        (__VLS_ctx.formatAmount(row.pay_amount ?? row.amount_cents ?? 0));
    }
    var __VLS_79;
    const __VLS_80 = {}.ElTableColumn;
    /** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
    // @ts-ignore
    const __VLS_81 = __VLS_asFunctionalComponent(__VLS_80, new __VLS_80({
        label: "下单时间",
        width: "160",
    }));
    const __VLS_82 = __VLS_81({
        label: "下单时间",
        width: "160",
    }, ...__VLS_functionalComponentArgsRest(__VLS_81));
    __VLS_83.slots.default;
    {
        const { default: __VLS_thisSlot } = __VLS_83.slots;
        const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
        (__VLS_ctx.formatTime(row.created_at));
    }
    var __VLS_83;
    const __VLS_84 = {}.ElTableColumn;
    /** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
    // @ts-ignore
    const __VLS_85 = __VLS_asFunctionalComponent(__VLS_84, new __VLS_84({
        label: "操作",
        width: "80",
        fixed: "right",
    }));
    const __VLS_86 = __VLS_85({
        label: "操作",
        width: "80",
        fixed: "right",
    }, ...__VLS_functionalComponentArgsRest(__VLS_85));
    __VLS_87.slots.default;
    {
        const { default: __VLS_thisSlot } = __VLS_87.slots;
        const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
        const __VLS_88 = {}.ElButton;
        /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
        // @ts-ignore
        const __VLS_89 = __VLS_asFunctionalComponent(__VLS_88, new __VLS_88({
            ...{ 'onClick': {} },
            text: true,
            type: "primary",
            size: "small",
        }));
        const __VLS_90 = __VLS_89({
            ...{ 'onClick': {} },
            text: true,
            type: "primary",
            size: "small",
        }, ...__VLS_functionalComponentArgsRest(__VLS_89));
        let __VLS_92;
        let __VLS_93;
        let __VLS_94;
        const __VLS_95 = {
            onClick: (...[$event]) => {
                if (!(__VLS_ctx.user))
                    return;
                __VLS_ctx.router.push(`/order/detail/${row.id}`);
            }
        };
        __VLS_91.slots.default;
        var __VLS_91;
    }
    var __VLS_87;
    {
        const { empty: __VLS_thisSlot } = __VLS_63.slots;
        const __VLS_96 = {}.ElEmpty;
        /** @type {[typeof __VLS_components.ElEmpty, typeof __VLS_components.elEmpty, ]} */ ;
        // @ts-ignore
        const __VLS_97 = __VLS_asFunctionalComponent(__VLS_96, new __VLS_96({
            description: "暂无订单记录",
            imageSize: (60),
        }));
        const __VLS_98 = __VLS_97({
            description: "暂无订单记录",
            imageSize: (60),
        }, ...__VLS_functionalComponentArgsRest(__VLS_97));
    }
    var __VLS_63;
    var __VLS_59;
    const __VLS_100 = {}.ElTabPane;
    /** @type {[typeof __VLS_components.ElTabPane, typeof __VLS_components.elTabPane, typeof __VLS_components.ElTabPane, typeof __VLS_components.elTabPane, ]} */ ;
    // @ts-ignore
    const __VLS_101 = __VLS_asFunctionalComponent(__VLS_100, new __VLS_100({
        label: "收货地址",
        name: "addresses",
    }));
    const __VLS_102 = __VLS_101({
        label: "收货地址",
        name: "addresses",
    }, ...__VLS_functionalComponentArgsRest(__VLS_101));
    __VLS_103.slots.default;
    const __VLS_104 = {}.ElEmpty;
    /** @type {[typeof __VLS_components.ElEmpty, typeof __VLS_components.elEmpty, ]} */ ;
    // @ts-ignore
    const __VLS_105 = __VLS_asFunctionalComponent(__VLS_104, new __VLS_104({
        description: "暂无数据",
        imageSize: (60),
    }));
    const __VLS_106 = __VLS_105({
        description: "暂无数据",
        imageSize: (60),
    }, ...__VLS_functionalComponentArgsRest(__VLS_105));
    var __VLS_103;
    var __VLS_55;
    var __VLS_51;
}
else if (!__VLS_ctx.loading) {
    const __VLS_108 = {}.ElEmpty;
    /** @type {[typeof __VLS_components.ElEmpty, typeof __VLS_components.elEmpty, ]} */ ;
    // @ts-ignore
    const __VLS_109 = __VLS_asFunctionalComponent(__VLS_108, new __VLS_108({
        description: "用户不存在或已被删除",
    }));
    const __VLS_110 = __VLS_109({
        description: "用户不存在或已被删除",
    }, ...__VLS_functionalComponentArgsRest(__VLS_109));
}
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            formatAmount: formatAmount,
            formatTime: formatTime,
            orderStatusMap: orderStatusMap,
            router: router,
            loading: loading,
            user: user,
            actionLoading: actionLoading,
            ordersLoading: ordersLoading,
            orders: orders,
            handleToggleStatus: handleToggleStatus,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
