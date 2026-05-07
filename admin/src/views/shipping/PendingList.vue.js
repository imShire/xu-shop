import { ref, onMounted, onUnmounted, computed } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import ProTable from '@/components/ProTable/index.vue';
import { useTable } from '@/composables/useTable';
import { getOrderList } from '@/api/order';
import { shipOrder, batchShipOrders, getBatchShipStatus } from '@/api/shipping';
import { formatAmount, formatTime } from '@/utils/format';
const searchForm = ref({ order_no: '', phone: '' });
const selectedRows = ref([]);
const { list, total, page, pageSize, loading, fetch } = useTable((params) => getOrderList({ ...params, status: 'paid', ...searchForm.value }));
const columns = [
    { label: '订单号', prop: 'order_no', width: 180 },
    { label: '用户', slot: 'user', width: 120 },
    { label: '商品', slot: 'items', minWidth: 160 },
    { label: '实付', slot: 'amount', width: 100, align: 'right' },
    { label: '付款时间', prop: 'paid_at', width: 150, formatter: (r) => formatTime(r.paid_at) },
];
const singleShip = ref({ visible: false, orderId: '', carrier_code: '', tracking_no: '' });
const batchShip = ref({ visible: false, carrier_code: '', tracking_no: '' });
const shipLoading = ref(false);
// Progress dialog state
const progress = ref({
    visible: false,
    total: 0,
    done: 0,
    failed: 0,
    errors: [],
    pdf_url: '',
});
let pollTimer = null;
const progressPercent = computed(() => {
    if (!progress.value.total)
        return 0;
    return Math.round(((progress.value.done + progress.value.failed) / progress.value.total) * 100);
});
const progressDone = computed(() => progress.value.done + progress.value.failed === progress.value.total && progress.value.total > 0);
function clearPollTimer() {
    if (pollTimer !== null) {
        clearInterval(pollTimer);
        pollTimer = null;
    }
}
function startPolling(taskId) {
    clearPollTimer();
    pollTimer = setInterval(async () => {
        try {
            const res = await getBatchShipStatus(taskId);
            const data = res?.data ?? res;
            progress.value.total = data.total ?? 0;
            progress.value.done = data.done ?? 0;
            progress.value.failed = data.failed ?? 0;
            progress.value.errors = data.errors ?? [];
            progress.value.pdf_url = data.pdf_url ?? '';
            if (progress.value.done + progress.value.failed >= progress.value.total && progress.value.total > 0) {
                clearPollTimer();
                fetch(searchForm.value);
            }
        }
        catch {
            // silently retry on poll error
        }
    }, 2000);
}
function closeProgressDialog() {
    clearPollTimer();
    progress.value.visible = false;
}
function openSingleShip(row) {
    singleShip.value = { visible: true, orderId: row.id, carrier_code: '', tracking_no: '' };
}
async function handleSingleShip() {
    shipLoading.value = true;
    try {
        await ElMessageBox.confirm('确认发货？发货后无法撤回', '发货确认', { type: 'warning' });
        await shipOrder(singleShip.value.orderId, {
            carrier_code: singleShip.value.carrier_code,
            tracking_no: singleShip.value.tracking_no,
        });
        ElMessage.success('发货成功');
        singleShip.value.visible = false;
        fetch(searchForm.value);
    }
    catch (e) {
        if (e !== 'cancel')
            ElMessage.error(e?.message || '操作失败');
    }
    finally {
        shipLoading.value = false;
    }
}
function openBatchShip() {
    if (!selectedRows.value.length)
        return ElMessage.warning('请先选择订单');
    batchShip.value = { visible: true, carrier_code: '', tracking_no: '' };
}
async function handleBatchShip() {
    shipLoading.value = true;
    try {
        await ElMessageBox.confirm(`确认批量发货 ${selectedRows.value.length} 个订单？`, '批量发货', { type: 'warning' });
        const res = await batchShipOrders(selectedRows.value.map((r) => ({
            order_id: r.id,
            carrier_code: batchShip.value.carrier_code,
            tracking_no: batchShip.value.tracking_no,
        })));
        const taskId = res?.data?.task_id ?? res?.task_id;
        batchShip.value.visible = false;
        // Reset and open progress dialog
        progress.value = {
            visible: true,
            total: selectedRows.value.length,
            done: 0,
            failed: 0,
            errors: [],
            pdf_url: '',
        };
        startPolling(taskId);
    }
    catch (e) {
        if (e !== 'cancel')
            ElMessage.error(e?.message || '操作失败');
    }
    finally {
        shipLoading.value = false;
    }
}
onMounted(() => fetch(searchForm.value));
onUnmounted(() => clearPollTimer());
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
    ...{ 'onSelectionChange': {} },
    columns: (__VLS_ctx.columns),
    data: (__VLS_ctx.list),
    total: (__VLS_ctx.total),
    loading: (__VLS_ctx.loading),
    page: (__VLS_ctx.page),
    pageSize: (__VLS_ctx.pageSize),
    selection: true,
}));
const __VLS_1 = __VLS_0({
    ...{ 'onRefresh': {} },
    ...{ 'onSelectionChange': {} },
    columns: (__VLS_ctx.columns),
    data: (__VLS_ctx.list),
    total: (__VLS_ctx.total),
    loading: (__VLS_ctx.loading),
    page: (__VLS_ctx.page),
    pageSize: (__VLS_ctx.pageSize),
    selection: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_0));
let __VLS_3;
let __VLS_4;
let __VLS_5;
const __VLS_6 = {
    onRefresh: (...[$event]) => {
        __VLS_ctx.fetch(__VLS_ctx.searchForm);
    }
};
const __VLS_7 = {
    onSelectionChange: ((rows) => (__VLS_ctx.selectedRows = rows))
};
__VLS_2.slots.default;
{
    const { search: __VLS_thisSlot } = __VLS_2.slots;
    const __VLS_8 = {}.ElInput;
    /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
    // @ts-ignore
    const __VLS_9 = __VLS_asFunctionalComponent(__VLS_8, new __VLS_8({
        modelValue: (__VLS_ctx.searchForm.order_no),
        placeholder: "订单号",
        clearable: true,
        ...{ style: {} },
    }));
    const __VLS_10 = __VLS_9({
        modelValue: (__VLS_ctx.searchForm.order_no),
        placeholder: "订单号",
        clearable: true,
        ...{ style: {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_9));
    const __VLS_12 = {}.ElInput;
    /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
    // @ts-ignore
    const __VLS_13 = __VLS_asFunctionalComponent(__VLS_12, new __VLS_12({
        modelValue: (__VLS_ctx.searchForm.phone),
        placeholder: "手机号",
        clearable: true,
        ...{ style: {} },
    }));
    const __VLS_14 = __VLS_13({
        modelValue: (__VLS_ctx.searchForm.phone),
        placeholder: "手机号",
        clearable: true,
        ...{ style: {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_13));
    const __VLS_16 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_17 = __VLS_asFunctionalComponent(__VLS_16, new __VLS_16({
        ...{ 'onClick': {} },
        type: "primary",
    }));
    const __VLS_18 = __VLS_17({
        ...{ 'onClick': {} },
        type: "primary",
    }, ...__VLS_functionalComponentArgsRest(__VLS_17));
    let __VLS_20;
    let __VLS_21;
    let __VLS_22;
    const __VLS_23 = {
        onClick: (...[$event]) => {
            __VLS_ctx.fetch(__VLS_ctx.searchForm);
        }
    };
    __VLS_19.slots.default;
    var __VLS_19;
}
{
    const { toolbar: __VLS_thisSlot } = __VLS_2.slots;
    const __VLS_24 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_25 = __VLS_asFunctionalComponent(__VLS_24, new __VLS_24({
        ...{ 'onClick': {} },
        type: "primary",
    }));
    const __VLS_26 = __VLS_25({
        ...{ 'onClick': {} },
        type: "primary",
    }, ...__VLS_functionalComponentArgsRest(__VLS_25));
    let __VLS_28;
    let __VLS_29;
    let __VLS_30;
    const __VLS_31 = {
        onClick: (__VLS_ctx.openBatchShip)
    };
    __VLS_27.slots.default;
    var __VLS_27;
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
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "text-truncate" },
    });
    (row.items?.[0]?.product_title);
    if (row.items?.length > 1) {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
            ...{ style: {} },
        });
        (row.items.length);
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
    const { actions: __VLS_thisSlot } = __VLS_2.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_32 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_33 = __VLS_asFunctionalComponent(__VLS_32, new __VLS_32({
        ...{ 'onClick': {} },
        type: "primary",
        size: "small",
    }));
    const __VLS_34 = __VLS_33({
        ...{ 'onClick': {} },
        type: "primary",
        size: "small",
    }, ...__VLS_functionalComponentArgsRest(__VLS_33));
    let __VLS_36;
    let __VLS_37;
    let __VLS_38;
    const __VLS_39 = {
        onClick: (...[$event]) => {
            __VLS_ctx.openSingleShip(row);
        }
    };
    __VLS_35.slots.default;
    var __VLS_35;
}
var __VLS_2;
const __VLS_40 = {}.ElDialog;
/** @type {[typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, ]} */ ;
// @ts-ignore
const __VLS_41 = __VLS_asFunctionalComponent(__VLS_40, new __VLS_40({
    modelValue: (__VLS_ctx.singleShip.visible),
    title: "快速发货",
    width: "400px",
}));
const __VLS_42 = __VLS_41({
    modelValue: (__VLS_ctx.singleShip.visible),
    title: "快速发货",
    width: "400px",
}, ...__VLS_functionalComponentArgsRest(__VLS_41));
__VLS_43.slots.default;
const __VLS_44 = {}.ElForm;
/** @type {[typeof __VLS_components.ElForm, typeof __VLS_components.elForm, typeof __VLS_components.ElForm, typeof __VLS_components.elForm, ]} */ ;
// @ts-ignore
const __VLS_45 = __VLS_asFunctionalComponent(__VLS_44, new __VLS_44({
    labelWidth: "80px",
}));
const __VLS_46 = __VLS_45({
    labelWidth: "80px",
}, ...__VLS_functionalComponentArgsRest(__VLS_45));
__VLS_47.slots.default;
const __VLS_48 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_49 = __VLS_asFunctionalComponent(__VLS_48, new __VLS_48({
    label: "快递公司",
}));
const __VLS_50 = __VLS_49({
    label: "快递公司",
}, ...__VLS_functionalComponentArgsRest(__VLS_49));
__VLS_51.slots.default;
const __VLS_52 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_53 = __VLS_asFunctionalComponent(__VLS_52, new __VLS_52({
    modelValue: (__VLS_ctx.singleShip.carrier_code),
    placeholder: "如：顺丰",
}));
const __VLS_54 = __VLS_53({
    modelValue: (__VLS_ctx.singleShip.carrier_code),
    placeholder: "如：顺丰",
}, ...__VLS_functionalComponentArgsRest(__VLS_53));
var __VLS_51;
const __VLS_56 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_57 = __VLS_asFunctionalComponent(__VLS_56, new __VLS_56({
    label: "运单号",
}));
const __VLS_58 = __VLS_57({
    label: "运单号",
}, ...__VLS_functionalComponentArgsRest(__VLS_57));
__VLS_59.slots.default;
const __VLS_60 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_61 = __VLS_asFunctionalComponent(__VLS_60, new __VLS_60({
    modelValue: (__VLS_ctx.singleShip.tracking_no),
}));
const __VLS_62 = __VLS_61({
    modelValue: (__VLS_ctx.singleShip.tracking_no),
}, ...__VLS_functionalComponentArgsRest(__VLS_61));
var __VLS_59;
var __VLS_47;
{
    const { footer: __VLS_thisSlot } = __VLS_43.slots;
    const __VLS_64 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_65 = __VLS_asFunctionalComponent(__VLS_64, new __VLS_64({
        ...{ 'onClick': {} },
    }));
    const __VLS_66 = __VLS_65({
        ...{ 'onClick': {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_65));
    let __VLS_68;
    let __VLS_69;
    let __VLS_70;
    const __VLS_71 = {
        onClick: (...[$event]) => {
            __VLS_ctx.singleShip.visible = false;
        }
    };
    __VLS_67.slots.default;
    var __VLS_67;
    const __VLS_72 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_73 = __VLS_asFunctionalComponent(__VLS_72, new __VLS_72({
        ...{ 'onClick': {} },
        type: "primary",
        loading: (__VLS_ctx.shipLoading),
    }));
    const __VLS_74 = __VLS_73({
        ...{ 'onClick': {} },
        type: "primary",
        loading: (__VLS_ctx.shipLoading),
    }, ...__VLS_functionalComponentArgsRest(__VLS_73));
    let __VLS_76;
    let __VLS_77;
    let __VLS_78;
    const __VLS_79 = {
        onClick: (__VLS_ctx.handleSingleShip)
    };
    __VLS_75.slots.default;
    var __VLS_75;
}
var __VLS_43;
const __VLS_80 = {}.ElDialog;
/** @type {[typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, ]} */ ;
// @ts-ignore
const __VLS_81 = __VLS_asFunctionalComponent(__VLS_80, new __VLS_80({
    modelValue: (__VLS_ctx.batchShip.visible),
    title: (`批量发货（已选 ${__VLS_ctx.selectedRows.length} 单）`),
    width: "420px",
}));
const __VLS_82 = __VLS_81({
    modelValue: (__VLS_ctx.batchShip.visible),
    title: (`批量发货（已选 ${__VLS_ctx.selectedRows.length} 单）`),
    width: "420px",
}, ...__VLS_functionalComponentArgsRest(__VLS_81));
__VLS_83.slots.default;
const __VLS_84 = {}.ElForm;
/** @type {[typeof __VLS_components.ElForm, typeof __VLS_components.elForm, typeof __VLS_components.ElForm, typeof __VLS_components.elForm, ]} */ ;
// @ts-ignore
const __VLS_85 = __VLS_asFunctionalComponent(__VLS_84, new __VLS_84({
    labelWidth: "80px",
}));
const __VLS_86 = __VLS_85({
    labelWidth: "80px",
}, ...__VLS_functionalComponentArgsRest(__VLS_85));
__VLS_87.slots.default;
const __VLS_88 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_89 = __VLS_asFunctionalComponent(__VLS_88, new __VLS_88({
    label: "快递公司",
}));
const __VLS_90 = __VLS_89({
    label: "快递公司",
}, ...__VLS_functionalComponentArgsRest(__VLS_89));
__VLS_91.slots.default;
const __VLS_92 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_93 = __VLS_asFunctionalComponent(__VLS_92, new __VLS_92({
    modelValue: (__VLS_ctx.batchShip.carrier_code),
    placeholder: "统一快递公司",
}));
const __VLS_94 = __VLS_93({
    modelValue: (__VLS_ctx.batchShip.carrier_code),
    placeholder: "统一快递公司",
}, ...__VLS_functionalComponentArgsRest(__VLS_93));
var __VLS_91;
const __VLS_96 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_97 = __VLS_asFunctionalComponent(__VLS_96, new __VLS_96({
    label: "运单号",
}));
const __VLS_98 = __VLS_97({
    label: "运单号",
}, ...__VLS_functionalComponentArgsRest(__VLS_97));
__VLS_99.slots.default;
const __VLS_100 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_101 = __VLS_asFunctionalComponent(__VLS_100, new __VLS_100({
    modelValue: (__VLS_ctx.batchShip.tracking_no),
    placeholder: "同一运单号或留空",
}));
const __VLS_102 = __VLS_101({
    modelValue: (__VLS_ctx.batchShip.tracking_no),
    placeholder: "同一运单号或留空",
}, ...__VLS_functionalComponentArgsRest(__VLS_101));
var __VLS_99;
var __VLS_87;
{
    const { footer: __VLS_thisSlot } = __VLS_83.slots;
    const __VLS_104 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_105 = __VLS_asFunctionalComponent(__VLS_104, new __VLS_104({
        ...{ 'onClick': {} },
    }));
    const __VLS_106 = __VLS_105({
        ...{ 'onClick': {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_105));
    let __VLS_108;
    let __VLS_109;
    let __VLS_110;
    const __VLS_111 = {
        onClick: (...[$event]) => {
            __VLS_ctx.batchShip.visible = false;
        }
    };
    __VLS_107.slots.default;
    var __VLS_107;
    const __VLS_112 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_113 = __VLS_asFunctionalComponent(__VLS_112, new __VLS_112({
        ...{ 'onClick': {} },
        type: "primary",
        loading: (__VLS_ctx.shipLoading),
    }));
    const __VLS_114 = __VLS_113({
        ...{ 'onClick': {} },
        type: "primary",
        loading: (__VLS_ctx.shipLoading),
    }, ...__VLS_functionalComponentArgsRest(__VLS_113));
    let __VLS_116;
    let __VLS_117;
    let __VLS_118;
    const __VLS_119 = {
        onClick: (__VLS_ctx.handleBatchShip)
    };
    __VLS_115.slots.default;
    var __VLS_115;
}
var __VLS_83;
const __VLS_120 = {}.ElDialog;
/** @type {[typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, ]} */ ;
// @ts-ignore
const __VLS_121 = __VLS_asFunctionalComponent(__VLS_120, new __VLS_120({
    ...{ 'onClosed': {} },
    modelValue: (__VLS_ctx.progress.visible),
    title: "批量发货进度",
    width: "480px",
    closeOnClickModal: (false),
    closeOnPressEscape: (false),
    showClose: (__VLS_ctx.progressDone),
}));
const __VLS_122 = __VLS_121({
    ...{ 'onClosed': {} },
    modelValue: (__VLS_ctx.progress.visible),
    title: "批量发货进度",
    width: "480px",
    closeOnClickModal: (false),
    closeOnPressEscape: (false),
    showClose: (__VLS_ctx.progressDone),
}, ...__VLS_functionalComponentArgsRest(__VLS_121));
let __VLS_124;
let __VLS_125;
let __VLS_126;
const __VLS_127 = {
    onClosed: (__VLS_ctx.clearPollTimer)
};
__VLS_123.slots.default;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ style: {} },
});
const __VLS_128 = {}.ElProgress;
/** @type {[typeof __VLS_components.ElProgress, typeof __VLS_components.elProgress, ]} */ ;
// @ts-ignore
const __VLS_129 = __VLS_asFunctionalComponent(__VLS_128, new __VLS_128({
    percentage: (__VLS_ctx.progressPercent),
    status: (__VLS_ctx.progressDone ? (__VLS_ctx.progress.failed > 0 ? 'warning' : 'success') : undefined),
    striped: (!__VLS_ctx.progressDone),
    stripedFlow: (!__VLS_ctx.progressDone),
    duration: (10),
}));
const __VLS_130 = __VLS_129({
    percentage: (__VLS_ctx.progressPercent),
    status: (__VLS_ctx.progressDone ? (__VLS_ctx.progress.failed > 0 ? 'warning' : 'success') : undefined),
    striped: (!__VLS_ctx.progressDone),
    stripedFlow: (!__VLS_ctx.progressDone),
    duration: (10),
}, ...__VLS_functionalComponentArgsRest(__VLS_129));
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ style: {} },
});
if (!__VLS_ctx.progressDone) {
    (__VLS_ctx.progress.done + __VLS_ctx.progress.failed);
    (__VLS_ctx.progress.total);
}
else {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ style: {} },
    });
    (__VLS_ctx.progress.done);
    if (__VLS_ctx.progress.failed > 0) {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
            ...{ style: {} },
        });
        (__VLS_ctx.progress.failed);
    }
}
if (__VLS_ctx.progressDone && __VLS_ctx.progress.errors.length > 0) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ style: {} },
    });
    const __VLS_132 = {}.ElScrollbar;
    /** @type {[typeof __VLS_components.ElScrollbar, typeof __VLS_components.elScrollbar, typeof __VLS_components.ElScrollbar, typeof __VLS_components.elScrollbar, ]} */ ;
    // @ts-ignore
    const __VLS_133 = __VLS_asFunctionalComponent(__VLS_132, new __VLS_132({
        ...{ style: {} },
    }));
    const __VLS_134 = __VLS_133({
        ...{ style: {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_133));
    __VLS_135.slots.default;
    for (const [err, idx] of __VLS_getVForSourceType((__VLS_ctx.progress.errors))) {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            key: (idx),
            ...{ style: {} },
        });
        (err);
    }
    var __VLS_135;
}
if (__VLS_ctx.progressDone && __VLS_ctx.progress.pdf_url) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ style: {} },
    });
    const __VLS_136 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_137 = __VLS_asFunctionalComponent(__VLS_136, new __VLS_136({
        type: "primary",
        link: true,
        href: (__VLS_ctx.progress.pdf_url),
        target: "_blank",
        tag: "a",
    }));
    const __VLS_138 = __VLS_137({
        type: "primary",
        link: true,
        href: (__VLS_ctx.progress.pdf_url),
        target: "_blank",
        tag: "a",
    }, ...__VLS_functionalComponentArgsRest(__VLS_137));
    __VLS_139.slots.default;
    var __VLS_139;
}
{
    const { footer: __VLS_thisSlot } = __VLS_123.slots;
    const __VLS_140 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_141 = __VLS_asFunctionalComponent(__VLS_140, new __VLS_140({
        ...{ 'onClick': {} },
        disabled: (!__VLS_ctx.progressDone),
    }));
    const __VLS_142 = __VLS_141({
        ...{ 'onClick': {} },
        disabled: (!__VLS_ctx.progressDone),
    }, ...__VLS_functionalComponentArgsRest(__VLS_141));
    let __VLS_144;
    let __VLS_145;
    let __VLS_146;
    const __VLS_147 = {
        onClick: (__VLS_ctx.closeProgressDialog)
    };
    __VLS_143.slots.default;
    var __VLS_143;
}
var __VLS_123;
/** @type {__VLS_StyleScopedClasses['page-card']} */ ;
/** @type {__VLS_StyleScopedClasses['text-truncate']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            ProTable: ProTable,
            formatAmount: formatAmount,
            searchForm: searchForm,
            selectedRows: selectedRows,
            list: list,
            total: total,
            page: page,
            pageSize: pageSize,
            loading: loading,
            fetch: fetch,
            columns: columns,
            singleShip: singleShip,
            batchShip: batchShip,
            shipLoading: shipLoading,
            progress: progress,
            progressPercent: progressPercent,
            progressDone: progressDone,
            clearPollTimer: clearPollTimer,
            closeProgressDialog: closeProgressDialog,
            openSingleShip: openSingleShip,
            handleSingleShip: handleSingleShip,
            openBatchShip: openBatchShip,
            handleBatchShip: handleBatchShip,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
