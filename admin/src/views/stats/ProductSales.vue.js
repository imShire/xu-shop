import { ref, onMounted } from 'vue';
import { getProductRanking, exportStats } from '@/api/stats';
import { formatAmount } from '@/utils/format';
import dayjs from 'dayjs';
const dateRange = ref([
    dayjs().subtract(29, 'day').format('YYYY-MM-DD'),
    dayjs().format('YYYY-MM-DD'),
]);
const loading = ref(false);
const list = ref([]);
async function loadData() {
    if (!dateRange.value)
        return;
    loading.value = true;
    try {
        const [from, to] = dateRange.value;
        const res = await getProductRanking({ from, to, page_size: 50 });
        list.value = res?.list || [];
    }
    finally {
        loading.value = false;
    }
}
async function handleExport() {
    const [from, to] = dateRange.value || [];
    const blob = await exportStats({ type: 'products', from, to });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `product_sales_${Date.now()}.xlsx`;
    a.click();
    URL.revokeObjectURL(url);
}
onMounted(loadData);
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "page-card" },
    ...{ style: {} },
});
const __VLS_0 = {}.ElDatePicker;
/** @type {[typeof __VLS_components.ElDatePicker, typeof __VLS_components.elDatePicker, ]} */ ;
// @ts-ignore
const __VLS_1 = __VLS_asFunctionalComponent(__VLS_0, new __VLS_0({
    modelValue: (__VLS_ctx.dateRange),
    type: "daterange",
    valueFormat: "YYYY-MM-DD",
    startPlaceholder: "开始日期",
    endPlaceholder: "结束日期",
    ...{ style: {} },
}));
const __VLS_2 = __VLS_1({
    modelValue: (__VLS_ctx.dateRange),
    type: "daterange",
    valueFormat: "YYYY-MM-DD",
    startPlaceholder: "开始日期",
    endPlaceholder: "结束日期",
    ...{ style: {} },
}, ...__VLS_functionalComponentArgsRest(__VLS_1));
const __VLS_4 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_5 = __VLS_asFunctionalComponent(__VLS_4, new __VLS_4({
    ...{ 'onClick': {} },
    type: "primary",
}));
const __VLS_6 = __VLS_5({
    ...{ 'onClick': {} },
    type: "primary",
}, ...__VLS_functionalComponentArgsRest(__VLS_5));
let __VLS_8;
let __VLS_9;
let __VLS_10;
const __VLS_11 = {
    onClick: (__VLS_ctx.loadData)
};
__VLS_7.slots.default;
var __VLS_7;
const __VLS_12 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_13 = __VLS_asFunctionalComponent(__VLS_12, new __VLS_12({
    ...{ 'onClick': {} },
}));
const __VLS_14 = __VLS_13({
    ...{ 'onClick': {} },
}, ...__VLS_functionalComponentArgsRest(__VLS_13));
let __VLS_16;
let __VLS_17;
let __VLS_18;
const __VLS_19 = {
    onClick: (__VLS_ctx.handleExport)
};
__VLS_15.slots.default;
var __VLS_15;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "page-card" },
});
const __VLS_20 = {}.ElTable;
/** @type {[typeof __VLS_components.ElTable, typeof __VLS_components.elTable, typeof __VLS_components.ElTable, typeof __VLS_components.elTable, ]} */ ;
// @ts-ignore
const __VLS_21 = __VLS_asFunctionalComponent(__VLS_20, new __VLS_20({
    data: (__VLS_ctx.list),
    border: true,
    stripe: true,
}));
const __VLS_22 = __VLS_21({
    data: (__VLS_ctx.list),
    border: true,
    stripe: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_21));
__VLS_asFunctionalDirective(__VLS_directives.vLoading)(null, { ...__VLS_directiveBindingRestFields, value: (__VLS_ctx.loading) }, null, null);
__VLS_23.slots.default;
const __VLS_24 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_25 = __VLS_asFunctionalComponent(__VLS_24, new __VLS_24({
    type: "index",
    label: "排名",
    width: "60",
    align: "center",
}));
const __VLS_26 = __VLS_25({
    type: "index",
    label: "排名",
    width: "60",
    align: "center",
}, ...__VLS_functionalComponentArgsRest(__VLS_25));
const __VLS_28 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_29 = __VLS_asFunctionalComponent(__VLS_28, new __VLS_28({
    prop: "product_name",
    label: "商品名称",
    minWidth: "180",
}));
const __VLS_30 = __VLS_29({
    prop: "product_name",
    label: "商品名称",
    minWidth: "180",
}, ...__VLS_functionalComponentArgsRest(__VLS_29));
const __VLS_32 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_33 = __VLS_asFunctionalComponent(__VLS_32, new __VLS_32({
    label: "销售数量",
    prop: "qty",
    width: "100",
    align: "right",
}));
const __VLS_34 = __VLS_33({
    label: "销售数量",
    prop: "qty",
    width: "100",
    align: "right",
}, ...__VLS_functionalComponentArgsRest(__VLS_33));
const __VLS_36 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_37 = __VLS_asFunctionalComponent(__VLS_36, new __VLS_36({
    label: "销售金额",
    width: "120",
    align: "right",
}));
const __VLS_38 = __VLS_37({
    label: "销售金额",
    width: "120",
    align: "right",
}, ...__VLS_functionalComponentArgsRest(__VLS_37));
__VLS_39.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_39.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ style: {} },
    });
    (__VLS_ctx.formatAmount(row.amount_cents));
}
var __VLS_39;
var __VLS_23;
/** @type {__VLS_StyleScopedClasses['page-card']} */ ;
/** @type {__VLS_StyleScopedClasses['page-card']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            formatAmount: formatAmount,
            dateRange: dateRange,
            loading: loading,
            list: list,
            loadData: loadData,
            handleExport: handleExport,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
