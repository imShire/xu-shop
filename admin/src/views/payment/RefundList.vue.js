import { ref, onMounted } from 'vue';
import ProTable from '@/components/ProTable/index.vue';
import { useTable } from '@/composables/useTable';
import { getRefundList } from '@/api/payment';
import { formatAmount, formatTime } from '@/utils/format';
const searchForm = ref({ order_no: '', status: '' });
const { list, total, page, pageSize, loading, fetch } = useTable((params) => getRefundList({ ...params, ...searchForm.value }));
const columns = [
    { label: '订单号', prop: 'order_no', width: 180 },
    { label: '退款单号', prop: 'refund_no', width: 200 },
    { label: '渠道', prop: 'channel', width: 100 },
    { label: '退款金额', slot: 'amount', width: 110, align: 'right' },
    { label: '原因', prop: 'reason', minWidth: 150 },
    { label: '状态', slot: 'status', width: 90, align: 'center' },
    { label: '时间', prop: 'created_at', width: 150, formatter: (r) => formatTime(r.created_at) },
];
const statusMap = {
    pending: { label: '处理中', type: 'warning' },
    success: { label: '退款成功', type: 'success' },
    failed: { label: '退款失败', type: 'danger' },
};
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
    const __VLS_27 = {}.ElTag;
    /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
    // @ts-ignore
    const __VLS_28 = __VLS_asFunctionalComponent(__VLS_27, new __VLS_27({
        type: (__VLS_ctx.statusMap[row.status]?.type || ''),
        size: "small",
    }));
    const __VLS_29 = __VLS_28({
        type: (__VLS_ctx.statusMap[row.status]?.type || ''),
        size: "small",
    }, ...__VLS_functionalComponentArgsRest(__VLS_28));
    __VLS_30.slots.default;
    (__VLS_ctx.statusMap[row.status]?.label || row.status);
    var __VLS_30;
}
var __VLS_2;
/** @type {__VLS_StyleScopedClasses['page-card']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            ProTable: ProTable,
            formatAmount: formatAmount,
            searchForm: searchForm,
            list: list,
            total: total,
            page: page,
            pageSize: pageSize,
            loading: loading,
            fetch: fetch,
            columns: columns,
            statusMap: statusMap,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
