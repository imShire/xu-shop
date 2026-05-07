import { ref, onMounted } from 'vue';
import ProTable from '@/components/ProTable/index.vue';
import { useTable } from '@/composables/useTable';
import { getAuditLogs } from '@/api/system';
import { formatTime } from '@/utils/format';
const searchForm = ref({
    module: '',
    operator: '',
    start_date: '',
    end_date: '',
});
const dateRange = ref(null);
const { list, total, page, pageSize, loading, fetch } = useTable((params) => getAuditLogs({ ...params, ...searchForm.value }));
const columns = [
    { label: '模块', prop: 'module', width: 100 },
    { label: '操作', prop: 'action', width: 150 },
    { label: '操作人', prop: 'operator', width: 100 },
    { label: 'IP', prop: 'ip', width: 130 },
    { label: '详情', prop: 'detail', minWidth: 200 },
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
        modelValue: (__VLS_ctx.searchForm.module),
        placeholder: "模块",
        clearable: true,
        ...{ style: {} },
    }));
    const __VLS_9 = __VLS_8({
        modelValue: (__VLS_ctx.searchForm.module),
        placeholder: "模块",
        clearable: true,
        ...{ style: {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_8));
    const __VLS_11 = {}.ElInput;
    /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
    // @ts-ignore
    const __VLS_12 = __VLS_asFunctionalComponent(__VLS_11, new __VLS_11({
        modelValue: (__VLS_ctx.searchForm.operator),
        placeholder: "操作人",
        clearable: true,
        ...{ style: {} },
    }));
    const __VLS_13 = __VLS_12({
        modelValue: (__VLS_ctx.searchForm.operator),
        placeholder: "操作人",
        clearable: true,
        ...{ style: {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_12));
    const __VLS_15 = {}.ElDatePicker;
    /** @type {[typeof __VLS_components.ElDatePicker, typeof __VLS_components.elDatePicker, ]} */ ;
    // @ts-ignore
    const __VLS_16 = __VLS_asFunctionalComponent(__VLS_15, new __VLS_15({
        modelValue: (__VLS_ctx.dateRange),
        type: "daterange",
        valueFormat: "YYYY-MM-DD",
        startPlaceholder: "开始",
        endPlaceholder: "结束",
        ...{ style: {} },
    }));
    const __VLS_17 = __VLS_16({
        modelValue: (__VLS_ctx.dateRange),
        type: "daterange",
        valueFormat: "YYYY-MM-DD",
        startPlaceholder: "开始",
        endPlaceholder: "结束",
        ...{ style: {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_16));
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
        onClick: (__VLS_ctx.handleSearch)
    };
    __VLS_22.slots.default;
    var __VLS_22;
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
