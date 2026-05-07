import { ref, onMounted } from 'vue';
import ProTable from '@/components/ProTable/index.vue';
import { useTable } from '@/composables/useTable';
import { getNotificationList } from '@/api/notification';
import { formatTime } from '@/utils/format';
const searchForm = ref({ status: '' });
const { list, total, page, pageSize, loading, fetch } = useTable((params) => getNotificationList({ ...params, ...searchForm.value }));
const columns = [
    { label: '模板', prop: 'template_name', width: 150 },
    { label: '目标', prop: 'target', minWidth: 150 },
    { label: '状态', slot: 'status', width: 90, align: 'center' },
    { label: '发送时间', prop: 'created_at', width: 150, formatter: (r) => formatTime(r.created_at) },
];
const statusMap = {
    pending: { label: '待发送', type: 'warning' },
    success: { label: '发送成功', type: 'success' },
    failed: { label: '发送失败', type: 'danger' },
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
    const __VLS_7 = {}.ElSelect;
    /** @type {[typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, ]} */ ;
    // @ts-ignore
    const __VLS_8 = __VLS_asFunctionalComponent(__VLS_7, new __VLS_7({
        modelValue: (__VLS_ctx.searchForm.status),
        placeholder: "状态",
        clearable: true,
        ...{ style: {} },
    }));
    const __VLS_9 = __VLS_8({
        modelValue: (__VLS_ctx.searchForm.status),
        placeholder: "状态",
        clearable: true,
        ...{ style: {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_8));
    __VLS_10.slots.default;
    for (const [v, k] of __VLS_getVForSourceType((__VLS_ctx.statusMap))) {
        const __VLS_11 = {}.ElOption;
        /** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
        // @ts-ignore
        const __VLS_12 = __VLS_asFunctionalComponent(__VLS_11, new __VLS_11({
            key: (k),
            label: (v.label),
            value: (k),
        }));
        const __VLS_13 = __VLS_12({
            key: (k),
            label: (v.label),
            value: (k),
        }, ...__VLS_functionalComponentArgsRest(__VLS_12));
    }
    var __VLS_10;
    const __VLS_15 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_16 = __VLS_asFunctionalComponent(__VLS_15, new __VLS_15({
        ...{ 'onClick': {} },
        type: "primary",
    }));
    const __VLS_17 = __VLS_16({
        ...{ 'onClick': {} },
        type: "primary",
    }, ...__VLS_functionalComponentArgsRest(__VLS_16));
    let __VLS_19;
    let __VLS_20;
    let __VLS_21;
    const __VLS_22 = {
        onClick: (...[$event]) => {
            __VLS_ctx.fetch(__VLS_ctx.searchForm);
        }
    };
    __VLS_18.slots.default;
    var __VLS_18;
}
{
    const { status: __VLS_thisSlot } = __VLS_2.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_23 = {}.ElTag;
    /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
    // @ts-ignore
    const __VLS_24 = __VLS_asFunctionalComponent(__VLS_23, new __VLS_23({
        type: (__VLS_ctx.statusMap[row.status]?.type || ''),
        size: "small",
    }));
    const __VLS_25 = __VLS_24({
        type: (__VLS_ctx.statusMap[row.status]?.type || ''),
        size: "small",
    }, ...__VLS_functionalComponentArgsRest(__VLS_24));
    __VLS_26.slots.default;
    (__VLS_ctx.statusMap[row.status]?.label || row.status);
    var __VLS_26;
}
var __VLS_2;
/** @type {__VLS_StyleScopedClasses['page-card']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            ProTable: ProTable,
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
