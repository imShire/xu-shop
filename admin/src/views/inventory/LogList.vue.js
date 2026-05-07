import { ref, onMounted } from 'vue';
import ProTable from '@/components/ProTable/index.vue';
import { useTable } from '@/composables/useTable';
import { getInventoryLogs } from '@/api/inventory';
import { formatTime } from '@/utils/format';
const searchForm = ref({
    sku_code: '',
    change_type: '',
    start_date: '',
    end_date: '',
});
const dateRange = ref(null);
const { list, total, page, pageSize, loading, fetch } = useTable((params) => getInventoryLogs({ ...params, ...searchForm.value }));
const columns = [
    { label: '商品', prop: 'product_title', minWidth: 150 },
    { label: 'SKU码', prop: 'sku_code', width: 120 },
    { label: '操作类型', slot: 'type', width: 100, align: 'center' },
    { label: '变动数量', slot: 'change', width: 100, align: 'center' },
    { label: '变动前', prop: 'before_qty', width: 80, align: 'center' },
    { label: '变动后', prop: 'after_qty', width: 80, align: 'center' },
    { label: '备注', prop: 'remark', minWidth: 120 },
    { label: '操作人', prop: 'operator', width: 90 },
    { label: '时间', prop: 'created_at', width: 150, formatter: (r) => formatTime(r.created_at) },
];
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
        modelValue: (__VLS_ctx.searchForm.sku_code),
        placeholder: "SKU码",
        clearable: true,
        ...{ style: {} },
    }));
    const __VLS_9 = __VLS_8({
        modelValue: (__VLS_ctx.searchForm.sku_code),
        placeholder: "SKU码",
        clearable: true,
        ...{ style: {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_8));
    const __VLS_11 = {}.ElSelect;
    /** @type {[typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, ]} */ ;
    // @ts-ignore
    const __VLS_12 = __VLS_asFunctionalComponent(__VLS_11, new __VLS_11({
        modelValue: (__VLS_ctx.searchForm.change_type),
        placeholder: "操作类型",
        clearable: true,
        ...{ style: {} },
    }));
    const __VLS_13 = __VLS_12({
        modelValue: (__VLS_ctx.searchForm.change_type),
        placeholder: "操作类型",
        clearable: true,
        ...{ style: {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_12));
    __VLS_14.slots.default;
    const __VLS_15 = {}.ElOption;
    /** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
    // @ts-ignore
    const __VLS_16 = __VLS_asFunctionalComponent(__VLS_15, new __VLS_15({
        label: "入库",
        value: "in",
    }));
    const __VLS_17 = __VLS_16({
        label: "入库",
        value: "in",
    }, ...__VLS_functionalComponentArgsRest(__VLS_16));
    const __VLS_19 = {}.ElOption;
    /** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
    // @ts-ignore
    const __VLS_20 = __VLS_asFunctionalComponent(__VLS_19, new __VLS_19({
        label: "出库",
        value: "out",
    }));
    const __VLS_21 = __VLS_20({
        label: "出库",
        value: "out",
    }, ...__VLS_functionalComponentArgsRest(__VLS_20));
    const __VLS_23 = {}.ElOption;
    /** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
    // @ts-ignore
    const __VLS_24 = __VLS_asFunctionalComponent(__VLS_23, new __VLS_23({
        label: "手动调整",
        value: "adjust",
    }));
    const __VLS_25 = __VLS_24({
        label: "手动调整",
        value: "adjust",
    }, ...__VLS_functionalComponentArgsRest(__VLS_24));
    const __VLS_27 = {}.ElOption;
    /** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
    // @ts-ignore
    const __VLS_28 = __VLS_asFunctionalComponent(__VLS_27, new __VLS_27({
        label: "订单出库",
        value: "order",
    }));
    const __VLS_29 = __VLS_28({
        label: "订单出库",
        value: "order",
    }, ...__VLS_functionalComponentArgsRest(__VLS_28));
    const __VLS_31 = {}.ElOption;
    /** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
    // @ts-ignore
    const __VLS_32 = __VLS_asFunctionalComponent(__VLS_31, new __VLS_31({
        label: "退款入库",
        value: "refund",
    }));
    const __VLS_33 = __VLS_32({
        label: "退款入库",
        value: "refund",
    }, ...__VLS_functionalComponentArgsRest(__VLS_32));
    var __VLS_14;
    const __VLS_35 = {}.ElDatePicker;
    /** @type {[typeof __VLS_components.ElDatePicker, typeof __VLS_components.elDatePicker, ]} */ ;
    // @ts-ignore
    const __VLS_36 = __VLS_asFunctionalComponent(__VLS_35, new __VLS_35({
        modelValue: (__VLS_ctx.dateRange),
        type: "daterange",
        valueFormat: "YYYY-MM-DD",
        startPlaceholder: "开始日期",
        endPlaceholder: "结束日期",
        ...{ style: {} },
    }));
    const __VLS_37 = __VLS_36({
        modelValue: (__VLS_ctx.dateRange),
        type: "daterange",
        valueFormat: "YYYY-MM-DD",
        startPlaceholder: "开始日期",
        endPlaceholder: "结束日期",
        ...{ style: {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_36));
    const __VLS_39 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_40 = __VLS_asFunctionalComponent(__VLS_39, new __VLS_39({
        ...{ 'onClick': {} },
        type: "primary",
    }));
    const __VLS_41 = __VLS_40({
        ...{ 'onClick': {} },
        type: "primary",
    }, ...__VLS_functionalComponentArgsRest(__VLS_40));
    let __VLS_43;
    let __VLS_44;
    let __VLS_45;
    const __VLS_46 = {
        onClick: (__VLS_ctx.handleSearch)
    };
    __VLS_42.slots.default;
    var __VLS_42;
}
{
    const { type: __VLS_thisSlot } = __VLS_2.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_47 = {}.ElTag;
    /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
    // @ts-ignore
    const __VLS_48 = __VLS_asFunctionalComponent(__VLS_47, new __VLS_47({
        size: "small",
    }));
    const __VLS_49 = __VLS_48({
        size: "small",
    }, ...__VLS_functionalComponentArgsRest(__VLS_48));
    __VLS_50.slots.default;
    (row.change_type);
    var __VLS_50;
}
{
    const { change: __VLS_thisSlot } = __VLS_2.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ style: ({ color: row.change_qty > 0 ? '#22c55e' : '#ef4444', fontWeight: 600 }) },
    });
    (row.change_qty > 0 ? '+' : '');
    (row.change_qty);
}
var __VLS_2;
/** @type {__VLS_StyleScopedClasses['page-card']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            ProTable: ProTable,
            searchForm: searchForm,
            dateRange: dateRange,
            list: list,
            total: total,
            page: page,
            pageSize: pageSize,
            loading: loading,
            fetch: fetch,
            columns: columns,
            handleSearch: handleSearch,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
