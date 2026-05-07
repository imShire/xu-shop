import { ref, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { ElMessage, ElMessageBox } from 'element-plus';
import ProTable from '@/components/ProTable/index.vue';
import { useTable } from '@/composables/useTable';
import { getOrderList, exportOrders, cancelOrder } from '@/api/order';
import { shipOrder } from '@/api/shipping';
import { formatAmount, formatTime, orderStatusMap } from '@/utils/format';
const router = useRouter();
const searchForm = ref({
    order_no: '',
    phone: '',
    status: '',
    start_date: '',
    end_date: '',
});
const dateRange = ref(null);
const { list, total, page, pageSize, loading, fetch } = useTable((params) => getOrderList({ ...params, ...searchForm.value }));
const columns = [
    { label: '订单号', slot: 'order_no', width: 180 },
    { label: '用户', slot: 'user', width: 120 },
    { label: '商品', slot: 'items', minWidth: 160 },
    { label: '实付金额', slot: 'amount', width: 110, align: 'right' },
    { label: '状态', slot: 'status', width: 90, align: 'center' },
    { label: '下单时间', prop: 'created_at', width: 150, formatter: (row) => formatTime(row.created_at) },
];
const shipDialogVisible = ref(false);
const currentOrderId = ref('');
const shipForm = ref({ carrier_code: '', tracking_no: '' });
const shipLoading = ref(false);
function openShipDialog(orderId) {
    currentOrderId.value = orderId;
    shipForm.value = { carrier_code: '', tracking_no: '' };
    shipDialogVisible.value = true;
}
async function handleShip() {
    if (!shipForm.value.carrier_code || !shipForm.value.tracking_no) {
        return ElMessage.warning('请填写快递信息');
    }
    shipLoading.value = true;
    try {
        await ElMessageBox.confirm('确认发货？发货后无法撤回', '发货确认', { type: 'warning' });
        await shipOrder(currentOrderId.value, shipForm.value);
        ElMessage.success('发货成功');
        shipDialogVisible.value = false;
        fetch(searchForm.value);
    }
    catch (e) {
        if (e !== 'cancel')
            ElMessage.error(e?.message || '发货失败');
    }
    finally {
        shipLoading.value = false;
    }
}
async function handleCancel(row) {
    try {
        await ElMessageBox.confirm('确认取消该订单？', '提示', { type: 'warning' });
        await cancelOrder(row.id, '');
        ElMessage.success('取消成功');
        fetch(searchForm.value);
    }
    catch (e) {
        if (e !== 'cancel')
            ElMessage.error(e?.message || '取消失败');
    }
}
async function handleExport() {
    const blob = await exportOrders(searchForm.value);
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `orders_${Date.now()}.xlsx`;
    a.click();
    URL.revokeObjectURL(url);
}
function handleSearch() {
    if (dateRange.value) {
        searchForm.value.start_date = dateRange.value[0];
        searchForm.value.end_date = dateRange.value[1];
    }
    else {
        searchForm.value.start_date = '';
        searchForm.value.end_date = '';
    }
    page.value = 1;
    fetch(searchForm.value);
}
function handleReset() {
    searchForm.value = { order_no: '', phone: '', status: '', start_date: '', end_date: '' };
    dateRange.value = null;
    handleSearch();
}
onMounted(() => fetch(searchForm.value));
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "page-card" },
});
/** @type {[typeof ProTable, typeof ProTable, ]} */ ;
// @ts-ignore
const __VLS_0 = __VLS_asFunctionalComponent(ProTable, new ProTable({
    ...{ 'onRefresh': {} },
    columns: (__VLS_ctx.columns),
    data: (__VLS_ctx.list),
    total: (__VLS_ctx.total),
    loading: (__VLS_ctx.loading),
    page: (__VLS_ctx.page),
    pageSize: (__VLS_ctx.pageSize),
}));
const __VLS_1 = __VLS_0({
    ...{ 'onRefresh': {} },
    columns: (__VLS_ctx.columns),
    data: (__VLS_ctx.list),
    total: (__VLS_ctx.total),
    loading: (__VLS_ctx.loading),
    page: (__VLS_ctx.page),
    pageSize: (__VLS_ctx.pageSize),
}, ...__VLS_functionalComponentArgsRest(__VLS_0));
let __VLS_3;
let __VLS_4;
let __VLS_5;
const __VLS_6 = {
    onRefresh: (...[$event]) => {
        __VLS_ctx.fetch(__VLS_ctx.searchForm);
    }
};
__VLS_2.slots.default;
{
    const { search: __VLS_thisSlot } = __VLS_2.slots;
    const __VLS_7 = {}.ElInput;
    /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
    // @ts-ignore
    const __VLS_8 = __VLS_asFunctionalComponent(__VLS_7, new __VLS_7({
        modelValue: (__VLS_ctx.searchForm.order_no),
        placeholder: "订单号",
        clearable: true,
        ...{ style: {} },
    }));
    const __VLS_9 = __VLS_8({
        modelValue: (__VLS_ctx.searchForm.order_no),
        placeholder: "订单号",
        clearable: true,
        ...{ style: {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_8));
    const __VLS_11 = {}.ElInput;
    /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
    // @ts-ignore
    const __VLS_12 = __VLS_asFunctionalComponent(__VLS_11, new __VLS_11({
        modelValue: (__VLS_ctx.searchForm.phone),
        placeholder: "用户手机号",
        clearable: true,
        ...{ style: {} },
    }));
    const __VLS_13 = __VLS_12({
        modelValue: (__VLS_ctx.searchForm.phone),
        placeholder: "用户手机号",
        clearable: true,
        ...{ style: {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_12));
    const __VLS_15 = {}.ElSelect;
    /** @type {[typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, ]} */ ;
    // @ts-ignore
    const __VLS_16 = __VLS_asFunctionalComponent(__VLS_15, new __VLS_15({
        modelValue: (__VLS_ctx.searchForm.status),
        placeholder: "订单状态",
        clearable: true,
        ...{ style: {} },
    }));
    const __VLS_17 = __VLS_16({
        modelValue: (__VLS_ctx.searchForm.status),
        placeholder: "订单状态",
        clearable: true,
        ...{ style: {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_16));
    __VLS_18.slots.default;
    for (const [v, k] of __VLS_getVForSourceType((__VLS_ctx.orderStatusMap))) {
        const __VLS_19 = {}.ElOption;
        /** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
        // @ts-ignore
        const __VLS_20 = __VLS_asFunctionalComponent(__VLS_19, new __VLS_19({
            key: (k),
            label: (v.label),
            value: (k),
        }));
        const __VLS_21 = __VLS_20({
            key: (k),
            label: (v.label),
            value: (k),
        }, ...__VLS_functionalComponentArgsRest(__VLS_20));
    }
    var __VLS_18;
    const __VLS_23 = {}.ElDatePicker;
    /** @type {[typeof __VLS_components.ElDatePicker, typeof __VLS_components.elDatePicker, ]} */ ;
    // @ts-ignore
    const __VLS_24 = __VLS_asFunctionalComponent(__VLS_23, new __VLS_23({
        modelValue: (__VLS_ctx.dateRange),
        type: "daterange",
        valueFormat: "YYYY-MM-DD",
        startPlaceholder: "开始日期",
        endPlaceholder: "结束日期",
        ...{ style: {} },
    }));
    const __VLS_25 = __VLS_24({
        modelValue: (__VLS_ctx.dateRange),
        type: "daterange",
        valueFormat: "YYYY-MM-DD",
        startPlaceholder: "开始日期",
        endPlaceholder: "结束日期",
        ...{ style: {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_24));
    const __VLS_27 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_28 = __VLS_asFunctionalComponent(__VLS_27, new __VLS_27({
        ...{ 'onClick': {} },
        type: "primary",
    }));
    const __VLS_29 = __VLS_28({
        ...{ 'onClick': {} },
        type: "primary",
    }, ...__VLS_functionalComponentArgsRest(__VLS_28));
    let __VLS_31;
    let __VLS_32;
    let __VLS_33;
    const __VLS_34 = {
        onClick: (__VLS_ctx.handleSearch)
    };
    __VLS_30.slots.default;
    var __VLS_30;
    const __VLS_35 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_36 = __VLS_asFunctionalComponent(__VLS_35, new __VLS_35({
        ...{ 'onClick': {} },
    }));
    const __VLS_37 = __VLS_36({
        ...{ 'onClick': {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_36));
    let __VLS_39;
    let __VLS_40;
    let __VLS_41;
    const __VLS_42 = {
        onClick: (__VLS_ctx.handleReset)
    };
    __VLS_38.slots.default;
    var __VLS_38;
}
{
    const { toolbar: __VLS_thisSlot } = __VLS_2.slots;
    const __VLS_43 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_44 = __VLS_asFunctionalComponent(__VLS_43, new __VLS_43({
        ...{ 'onClick': {} },
    }));
    const __VLS_45 = __VLS_44({
        ...{ 'onClick': {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_44));
    let __VLS_47;
    let __VLS_48;
    let __VLS_49;
    const __VLS_50 = {
        onClick: (__VLS_ctx.handleExport)
    };
    __VLS_46.slots.default;
    var __VLS_46;
}
{
    const { order_no: __VLS_thisSlot } = __VLS_2.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_51 = {}.ElLink;
    /** @type {[typeof __VLS_components.ElLink, typeof __VLS_components.elLink, typeof __VLS_components.ElLink, typeof __VLS_components.elLink, ]} */ ;
    // @ts-ignore
    const __VLS_52 = __VLS_asFunctionalComponent(__VLS_51, new __VLS_51({
        ...{ 'onClick': {} },
        type: "primary",
    }));
    const __VLS_53 = __VLS_52({
        ...{ 'onClick': {} },
        type: "primary",
    }, ...__VLS_functionalComponentArgsRest(__VLS_52));
    let __VLS_55;
    let __VLS_56;
    let __VLS_57;
    const __VLS_58 = {
        onClick: (...[$event]) => {
            __VLS_ctx.router.push(`/order/detail/${row.id}`);
        }
    };
    __VLS_54.slots.default;
    (row.order_no);
    var __VLS_54;
}
{
    const { user: __VLS_thisSlot } = __VLS_2.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({});
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({});
    (row.user_nickname);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ style: {} },
    });
    (row.user_phone);
}
{
    const { items: __VLS_thisSlot } = __VLS_2.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    if (row.items?.[0]) {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "text-truncate" },
        });
        (row.items[0].product_title);
        if (row.items.length > 1) {
            __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
                ...{ style: {} },
            });
            (row.items.length);
        }
    }
}
{
    const { amount: __VLS_thisSlot } = __VLS_2.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ style: {} },
    });
    (__VLS_ctx.formatAmount(row.pay_amount));
}
{
    const { status: __VLS_thisSlot } = __VLS_2.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_59 = {}.ElTag;
    /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
    // @ts-ignore
    const __VLS_60 = __VLS_asFunctionalComponent(__VLS_59, new __VLS_59({
        type: (__VLS_ctx.orderStatusMap[row.status]?.type || ''),
        size: "small",
    }));
    const __VLS_61 = __VLS_60({
        type: (__VLS_ctx.orderStatusMap[row.status]?.type || ''),
        size: "small",
    }, ...__VLS_functionalComponentArgsRest(__VLS_60));
    __VLS_62.slots.default;
    (__VLS_ctx.orderStatusMap[row.status]?.label || row.status);
    var __VLS_62;
}
{
    const { actions: __VLS_thisSlot } = __VLS_2.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_63 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_64 = __VLS_asFunctionalComponent(__VLS_63, new __VLS_63({
        ...{ 'onClick': {} },
        text: true,
        type: "primary",
        size: "small",
    }));
    const __VLS_65 = __VLS_64({
        ...{ 'onClick': {} },
        text: true,
        type: "primary",
        size: "small",
    }, ...__VLS_functionalComponentArgsRest(__VLS_64));
    let __VLS_67;
    let __VLS_68;
    let __VLS_69;
    const __VLS_70 = {
        onClick: (...[$event]) => {
            __VLS_ctx.router.push(`/order/detail/${row.id}`);
        }
    };
    __VLS_66.slots.default;
    var __VLS_66;
    if (row.status === 'paid') {
        const __VLS_71 = {}.ElButton;
        /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
        // @ts-ignore
        const __VLS_72 = __VLS_asFunctionalComponent(__VLS_71, new __VLS_71({
            ...{ 'onClick': {} },
            text: true,
            type: "success",
            size: "small",
        }));
        const __VLS_73 = __VLS_72({
            ...{ 'onClick': {} },
            text: true,
            type: "success",
            size: "small",
        }, ...__VLS_functionalComponentArgsRest(__VLS_72));
        let __VLS_75;
        let __VLS_76;
        let __VLS_77;
        const __VLS_78 = {
            onClick: (...[$event]) => {
                if (!(row.status === 'paid'))
                    return;
                __VLS_ctx.openShipDialog(row.id);
            }
        };
        __VLS_74.slots.default;
        var __VLS_74;
    }
    if (row.status === 'pending_payment' || row.status === 'paid') {
        const __VLS_79 = {}.ElButton;
        /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
        // @ts-ignore
        const __VLS_80 = __VLS_asFunctionalComponent(__VLS_79, new __VLS_79({
            ...{ 'onClick': {} },
            text: true,
            type: "danger",
            size: "small",
        }));
        const __VLS_81 = __VLS_80({
            ...{ 'onClick': {} },
            text: true,
            type: "danger",
            size: "small",
        }, ...__VLS_functionalComponentArgsRest(__VLS_80));
        let __VLS_83;
        let __VLS_84;
        let __VLS_85;
        const __VLS_86 = {
            onClick: (...[$event]) => {
                if (!(row.status === 'pending_payment' || row.status === 'paid'))
                    return;
                __VLS_ctx.handleCancel(row);
            }
        };
        __VLS_82.slots.default;
        var __VLS_82;
    }
}
var __VLS_2;
const __VLS_87 = {}.ElDialog;
/** @type {[typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, ]} */ ;
// @ts-ignore
const __VLS_88 = __VLS_asFunctionalComponent(__VLS_87, new __VLS_87({
    modelValue: (__VLS_ctx.shipDialogVisible),
    title: "手动发货",
    width: "400px",
}));
const __VLS_89 = __VLS_88({
    modelValue: (__VLS_ctx.shipDialogVisible),
    title: "手动发货",
    width: "400px",
}, ...__VLS_functionalComponentArgsRest(__VLS_88));
__VLS_90.slots.default;
const __VLS_91 = {}.ElForm;
/** @type {[typeof __VLS_components.ElForm, typeof __VLS_components.elForm, typeof __VLS_components.ElForm, typeof __VLS_components.elForm, ]} */ ;
// @ts-ignore
const __VLS_92 = __VLS_asFunctionalComponent(__VLS_91, new __VLS_91({
    model: (__VLS_ctx.shipForm),
    labelWidth: "80px",
}));
const __VLS_93 = __VLS_92({
    model: (__VLS_ctx.shipForm),
    labelWidth: "80px",
}, ...__VLS_functionalComponentArgsRest(__VLS_92));
__VLS_94.slots.default;
const __VLS_95 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_96 = __VLS_asFunctionalComponent(__VLS_95, new __VLS_95({
    label: "快递公司",
}));
const __VLS_97 = __VLS_96({
    label: "快递公司",
}, ...__VLS_functionalComponentArgsRest(__VLS_96));
__VLS_98.slots.default;
const __VLS_99 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_100 = __VLS_asFunctionalComponent(__VLS_99, new __VLS_99({
    modelValue: (__VLS_ctx.shipForm.carrier_code),
    placeholder: "如：顺丰",
}));
const __VLS_101 = __VLS_100({
    modelValue: (__VLS_ctx.shipForm.carrier_code),
    placeholder: "如：顺丰",
}, ...__VLS_functionalComponentArgsRest(__VLS_100));
var __VLS_98;
const __VLS_103 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_104 = __VLS_asFunctionalComponent(__VLS_103, new __VLS_103({
    label: "运单号",
}));
const __VLS_105 = __VLS_104({
    label: "运单号",
}, ...__VLS_functionalComponentArgsRest(__VLS_104));
__VLS_106.slots.default;
const __VLS_107 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_108 = __VLS_asFunctionalComponent(__VLS_107, new __VLS_107({
    modelValue: (__VLS_ctx.shipForm.tracking_no),
    placeholder: "请输入快递单号",
}));
const __VLS_109 = __VLS_108({
    modelValue: (__VLS_ctx.shipForm.tracking_no),
    placeholder: "请输入快递单号",
}, ...__VLS_functionalComponentArgsRest(__VLS_108));
var __VLS_106;
var __VLS_94;
{
    const { footer: __VLS_thisSlot } = __VLS_90.slots;
    const __VLS_111 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_112 = __VLS_asFunctionalComponent(__VLS_111, new __VLS_111({
        ...{ 'onClick': {} },
    }));
    const __VLS_113 = __VLS_112({
        ...{ 'onClick': {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_112));
    let __VLS_115;
    let __VLS_116;
    let __VLS_117;
    const __VLS_118 = {
        onClick: (...[$event]) => {
            __VLS_ctx.shipDialogVisible = false;
        }
    };
    __VLS_114.slots.default;
    var __VLS_114;
    const __VLS_119 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_120 = __VLS_asFunctionalComponent(__VLS_119, new __VLS_119({
        ...{ 'onClick': {} },
        type: "primary",
        loading: (__VLS_ctx.shipLoading),
    }));
    const __VLS_121 = __VLS_120({
        ...{ 'onClick': {} },
        type: "primary",
        loading: (__VLS_ctx.shipLoading),
    }, ...__VLS_functionalComponentArgsRest(__VLS_120));
    let __VLS_123;
    let __VLS_124;
    let __VLS_125;
    const __VLS_126 = {
        onClick: (__VLS_ctx.handleShip)
    };
    __VLS_122.slots.default;
    var __VLS_122;
}
var __VLS_90;
/** @type {__VLS_StyleScopedClasses['page-card']} */ ;
/** @type {__VLS_StyleScopedClasses['text-truncate']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            ProTable: ProTable,
            formatAmount: formatAmount,
            orderStatusMap: orderStatusMap,
            router: router,
            searchForm: searchForm,
            dateRange: dateRange,
            list: list,
            total: total,
            page: page,
            pageSize: pageSize,
            loading: loading,
            fetch: fetch,
            columns: columns,
            shipDialogVisible: shipDialogVisible,
            shipForm: shipForm,
            shipLoading: shipLoading,
            openShipDialog: openShipDialog,
            handleShip: handleShip,
            handleCancel: handleCancel,
            handleExport: handleExport,
            handleSearch: handleSearch,
            handleReset: handleReset,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
