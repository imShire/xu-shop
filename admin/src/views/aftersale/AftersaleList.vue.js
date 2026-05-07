import { ref, onMounted } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import ProTable from '@/components/ProTable/index.vue';
import { useTable } from '@/composables/useTable';
import { getAftersaleList, approveAftersale, rejectAftersale, directRefund } from '@/api/aftersale';
import { formatAmount, formatTime } from '@/utils/format';
import PriceInput from '@/components/PriceInput/index.vue';
const searchForm = ref({ order_no: '', status: '' });
const { list, total, page, pageSize, loading, fetch } = useTable((params) => getAftersaleList({ ...params, ...searchForm.value }));
const columns = [
    { label: '订单号', prop: 'order_no', width: 180 },
    { label: '用户', prop: 'user_nickname', width: 100 },
    { label: '类型', slot: 'type', width: 90, align: 'center' },
    { label: '原因', prop: 'reason', minWidth: 150 },
    { label: '退款金额', slot: 'amount', width: 110, align: 'right' },
    { label: '状态', slot: 'status', width: 90, align: 'center' },
    { label: '申请时间', prop: 'created_at', width: 150, formatter: (r) => formatTime(r.created_at) },
];
const rejectDialog = ref({ visible: false, id: '', reason: '' });
const rejectLoading = ref(false);
const refundDialog = ref({ visible: false, id: '', amount: 0 });
const refundLoading = ref(false);
const typeMap = {
    cancel: '取消订单',
    refund: '仅退款',
    return: '退货退款',
};
const statusMap = {
    pending: { label: '待审核', type: 'warning' },
    approved: { label: '已同意', type: 'success' },
    rejected: { label: '已拒绝', type: 'danger' },
    refunding: { label: '退款中', type: '' },
    refunded: { label: '已退款', type: 'success' },
};
async function handleApprove(row) {
    try {
        await ElMessageBox.confirm('确认同意退款申请？退款将自动发起', '提示', { type: 'warning' });
        await approveAftersale(row.order_id);
        ElMessage.success('已同意');
        fetch(searchForm.value);
    }
    catch (e) {
        if (e !== 'cancel')
            ElMessage.error(e?.message || '操作失败');
    }
}
function openReject(row) {
    rejectDialog.value = { visible: true, id: row.order_id, reason: '' };
}
async function handleReject() {
    rejectLoading.value = true;
    try {
        await ElMessageBox.confirm('确认拒绝该售后申请？', '提示', { type: 'warning' });
        await rejectAftersale(rejectDialog.value.id, rejectDialog.value.reason);
        ElMessage.success('已拒绝');
        rejectDialog.value.visible = false;
        fetch(searchForm.value);
    }
    catch (e) {
        if (e !== 'cancel')
            ElMessage.error(e?.message || '操作失败');
    }
    finally {
        rejectLoading.value = false;
    }
}
function openRefund(row) {
    refundDialog.value = { visible: true, id: row.order_id, amount: row.amount };
}
async function handleRefund() {
    refundLoading.value = true;
    try {
        await ElMessageBox.confirm('确认直接退款？此操作不可撤销', '退款确认', { type: 'warning' });
        await directRefund(refundDialog.value.id, refundDialog.value.amount);
        ElMessage.success('退款成功');
        refundDialog.value.visible = false;
        fetch(searchForm.value);
    }
    catch (e) {
        if (e !== 'cancel')
            ElMessage.error(e?.message || '操作失败');
    }
    finally {
        refundLoading.value = false;
    }
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
    const __VLS_11 = {}.ElSelect;
    /** @type {[typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, ]} */ ;
    // @ts-ignore
    const __VLS_12 = __VLS_asFunctionalComponent(__VLS_11, new __VLS_11({
        modelValue: (__VLS_ctx.searchForm.status),
        placeholder: "状态",
        clearable: true,
        ...{ style: {} },
    }));
    const __VLS_13 = __VLS_12({
        modelValue: (__VLS_ctx.searchForm.status),
        placeholder: "状态",
        clearable: true,
        ...{ style: {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_12));
    __VLS_14.slots.default;
    for (const [v, k] of __VLS_getVForSourceType((__VLS_ctx.statusMap))) {
        const __VLS_15 = {}.ElOption;
        /** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
        // @ts-ignore
        const __VLS_16 = __VLS_asFunctionalComponent(__VLS_15, new __VLS_15({
            key: (k),
            label: (v.label),
            value: (k),
        }));
        const __VLS_17 = __VLS_16({
            key: (k),
            label: (v.label),
            value: (k),
        }, ...__VLS_functionalComponentArgsRest(__VLS_16));
    }
    var __VLS_14;
    const __VLS_19 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_20 = __VLS_asFunctionalComponent(__VLS_19, new __VLS_19({
        ...{ 'onClick': {} },
        type: "primary",
    }));
    const __VLS_21 = __VLS_20({
        ...{ 'onClick': {} },
        type: "primary",
    }, ...__VLS_functionalComponentArgsRest(__VLS_20));
    let __VLS_23;
    let __VLS_24;
    let __VLS_25;
    const __VLS_26 = {
        onClick: (...[$event]) => {
            __VLS_ctx.fetch(__VLS_ctx.searchForm);
        }
    };
    __VLS_22.slots.default;
    var __VLS_22;
}
{
    const { type: __VLS_thisSlot } = __VLS_2.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_27 = {}.ElTag;
    /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
    // @ts-ignore
    const __VLS_28 = __VLS_asFunctionalComponent(__VLS_27, new __VLS_27({
        size: "small",
    }));
    const __VLS_29 = __VLS_28({
        size: "small",
    }, ...__VLS_functionalComponentArgsRest(__VLS_28));
    __VLS_30.slots.default;
    (__VLS_ctx.typeMap[row.type] || row.type);
    var __VLS_30;
}
{
    const { amount: __VLS_thisSlot } = __VLS_2.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ style: {} },
    });
    (__VLS_ctx.formatAmount(row.amount));
}
{
    const { status: __VLS_thisSlot } = __VLS_2.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_31 = {}.ElTag;
    /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
    // @ts-ignore
    const __VLS_32 = __VLS_asFunctionalComponent(__VLS_31, new __VLS_31({
        type: (__VLS_ctx.statusMap[row.status]?.type || ''),
        size: "small",
    }));
    const __VLS_33 = __VLS_32({
        type: (__VLS_ctx.statusMap[row.status]?.type || ''),
        size: "small",
    }, ...__VLS_functionalComponentArgsRest(__VLS_32));
    __VLS_34.slots.default;
    (__VLS_ctx.statusMap[row.status]?.label || row.status);
    var __VLS_34;
}
{
    const { actions: __VLS_thisSlot } = __VLS_2.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    if (row.status === 'pending') {
        const __VLS_35 = {}.ElButton;
        /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
        // @ts-ignore
        const __VLS_36 = __VLS_asFunctionalComponent(__VLS_35, new __VLS_35({
            ...{ 'onClick': {} },
            text: true,
            type: "success",
            size: "small",
        }));
        const __VLS_37 = __VLS_36({
            ...{ 'onClick': {} },
            text: true,
            type: "success",
            size: "small",
        }, ...__VLS_functionalComponentArgsRest(__VLS_36));
        let __VLS_39;
        let __VLS_40;
        let __VLS_41;
        const __VLS_42 = {
            onClick: (...[$event]) => {
                if (!(row.status === 'pending'))
                    return;
                __VLS_ctx.handleApprove(row);
            }
        };
        __VLS_38.slots.default;
        var __VLS_38;
        const __VLS_43 = {}.ElButton;
        /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
        // @ts-ignore
        const __VLS_44 = __VLS_asFunctionalComponent(__VLS_43, new __VLS_43({
            ...{ 'onClick': {} },
            text: true,
            type: "danger",
            size: "small",
        }));
        const __VLS_45 = __VLS_44({
            ...{ 'onClick': {} },
            text: true,
            type: "danger",
            size: "small",
        }, ...__VLS_functionalComponentArgsRest(__VLS_44));
        let __VLS_47;
        let __VLS_48;
        let __VLS_49;
        const __VLS_50 = {
            onClick: (...[$event]) => {
                if (!(row.status === 'pending'))
                    return;
                __VLS_ctx.openReject(row);
            }
        };
        __VLS_46.slots.default;
        var __VLS_46;
    }
    if (['approved', 'pending'].includes(row.status)) {
        const __VLS_51 = {}.ElButton;
        /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
        // @ts-ignore
        const __VLS_52 = __VLS_asFunctionalComponent(__VLS_51, new __VLS_51({
            ...{ 'onClick': {} },
            text: true,
            type: "warning",
            size: "small",
        }));
        const __VLS_53 = __VLS_52({
            ...{ 'onClick': {} },
            text: true,
            type: "warning",
            size: "small",
        }, ...__VLS_functionalComponentArgsRest(__VLS_52));
        let __VLS_55;
        let __VLS_56;
        let __VLS_57;
        const __VLS_58 = {
            onClick: (...[$event]) => {
                if (!(['approved', 'pending'].includes(row.status)))
                    return;
                __VLS_ctx.openRefund(row);
            }
        };
        __VLS_54.slots.default;
        var __VLS_54;
    }
}
var __VLS_2;
const __VLS_59 = {}.ElDialog;
/** @type {[typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, ]} */ ;
// @ts-ignore
const __VLS_60 = __VLS_asFunctionalComponent(__VLS_59, new __VLS_59({
    modelValue: (__VLS_ctx.rejectDialog.visible),
    title: "拒绝原因",
    width: "400px",
}));
const __VLS_61 = __VLS_60({
    modelValue: (__VLS_ctx.rejectDialog.visible),
    title: "拒绝原因",
    width: "400px",
}, ...__VLS_functionalComponentArgsRest(__VLS_60));
__VLS_62.slots.default;
const __VLS_63 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_64 = __VLS_asFunctionalComponent(__VLS_63, new __VLS_63({
    modelValue: (__VLS_ctx.rejectDialog.reason),
    type: "textarea",
    placeholder: "请输入拒绝原因",
    rows: (3),
}));
const __VLS_65 = __VLS_64({
    modelValue: (__VLS_ctx.rejectDialog.reason),
    type: "textarea",
    placeholder: "请输入拒绝原因",
    rows: (3),
}, ...__VLS_functionalComponentArgsRest(__VLS_64));
{
    const { footer: __VLS_thisSlot } = __VLS_62.slots;
    const __VLS_67 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_68 = __VLS_asFunctionalComponent(__VLS_67, new __VLS_67({
        ...{ 'onClick': {} },
    }));
    const __VLS_69 = __VLS_68({
        ...{ 'onClick': {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_68));
    let __VLS_71;
    let __VLS_72;
    let __VLS_73;
    const __VLS_74 = {
        onClick: (...[$event]) => {
            __VLS_ctx.rejectDialog.visible = false;
        }
    };
    __VLS_70.slots.default;
    var __VLS_70;
    const __VLS_75 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_76 = __VLS_asFunctionalComponent(__VLS_75, new __VLS_75({
        ...{ 'onClick': {} },
        type: "danger",
        loading: (__VLS_ctx.rejectLoading),
    }));
    const __VLS_77 = __VLS_76({
        ...{ 'onClick': {} },
        type: "danger",
        loading: (__VLS_ctx.rejectLoading),
    }, ...__VLS_functionalComponentArgsRest(__VLS_76));
    let __VLS_79;
    let __VLS_80;
    let __VLS_81;
    const __VLS_82 = {
        onClick: (__VLS_ctx.handleReject)
    };
    __VLS_78.slots.default;
    var __VLS_78;
}
var __VLS_62;
const __VLS_83 = {}.ElDialog;
/** @type {[typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, ]} */ ;
// @ts-ignore
const __VLS_84 = __VLS_asFunctionalComponent(__VLS_83, new __VLS_83({
    modelValue: (__VLS_ctx.refundDialog.visible),
    title: "直接退款",
    width: "400px",
}));
const __VLS_85 = __VLS_84({
    modelValue: (__VLS_ctx.refundDialog.visible),
    title: "直接退款",
    width: "400px",
}, ...__VLS_functionalComponentArgsRest(__VLS_84));
__VLS_86.slots.default;
const __VLS_87 = {}.ElForm;
/** @type {[typeof __VLS_components.ElForm, typeof __VLS_components.elForm, typeof __VLS_components.ElForm, typeof __VLS_components.elForm, ]} */ ;
// @ts-ignore
const __VLS_88 = __VLS_asFunctionalComponent(__VLS_87, new __VLS_87({
    labelWidth: "80px",
}));
const __VLS_89 = __VLS_88({
    labelWidth: "80px",
}, ...__VLS_functionalComponentArgsRest(__VLS_88));
__VLS_90.slots.default;
const __VLS_91 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_92 = __VLS_asFunctionalComponent(__VLS_91, new __VLS_91({
    label: "退款金额",
}));
const __VLS_93 = __VLS_92({
    label: "退款金额",
}, ...__VLS_functionalComponentArgsRest(__VLS_92));
__VLS_94.slots.default;
/** @type {[typeof PriceInput, ]} */ ;
// @ts-ignore
const __VLS_95 = __VLS_asFunctionalComponent(PriceInput, new PriceInput({
    modelValue: (__VLS_ctx.refundDialog.amount),
}));
const __VLS_96 = __VLS_95({
    modelValue: (__VLS_ctx.refundDialog.amount),
}, ...__VLS_functionalComponentArgsRest(__VLS_95));
var __VLS_94;
var __VLS_90;
{
    const { footer: __VLS_thisSlot } = __VLS_86.slots;
    const __VLS_98 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_99 = __VLS_asFunctionalComponent(__VLS_98, new __VLS_98({
        ...{ 'onClick': {} },
    }));
    const __VLS_100 = __VLS_99({
        ...{ 'onClick': {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_99));
    let __VLS_102;
    let __VLS_103;
    let __VLS_104;
    const __VLS_105 = {
        onClick: (...[$event]) => {
            __VLS_ctx.refundDialog.visible = false;
        }
    };
    __VLS_101.slots.default;
    var __VLS_101;
    const __VLS_106 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_107 = __VLS_asFunctionalComponent(__VLS_106, new __VLS_106({
        ...{ 'onClick': {} },
        type: "warning",
        loading: (__VLS_ctx.refundLoading),
    }));
    const __VLS_108 = __VLS_107({
        ...{ 'onClick': {} },
        type: "warning",
        loading: (__VLS_ctx.refundLoading),
    }, ...__VLS_functionalComponentArgsRest(__VLS_107));
    let __VLS_110;
    let __VLS_111;
    let __VLS_112;
    const __VLS_113 = {
        onClick: (__VLS_ctx.handleRefund)
    };
    __VLS_109.slots.default;
    var __VLS_109;
}
var __VLS_86;
/** @type {__VLS_StyleScopedClasses['page-card']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            ProTable: ProTable,
            formatAmount: formatAmount,
            PriceInput: PriceInput,
            searchForm: searchForm,
            list: list,
            total: total,
            page: page,
            pageSize: pageSize,
            loading: loading,
            fetch: fetch,
            columns: columns,
            rejectDialog: rejectDialog,
            rejectLoading: rejectLoading,
            refundDialog: refundDialog,
            refundLoading: refundLoading,
            typeMap: typeMap,
            statusMap: statusMap,
            handleApprove: handleApprove,
            openReject: openReject,
            handleReject: handleReject,
            openRefund: openRefund,
            handleRefund: handleRefund,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
