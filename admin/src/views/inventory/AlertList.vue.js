import { ref, onMounted } from 'vue';
import { ElMessage } from 'element-plus';
import ProTable from '@/components/ProTable/index.vue';
import { useTable } from '@/composables/useTable';
import { getInventoryAlerts, markAlertRead, markAllAlertsRead } from '@/api/inventory';
import { formatTime } from '@/utils/format';
const searchForm = ref({ status: '' });
const { list, total, page, pageSize, loading, fetch } = useTable((params) => getInventoryAlerts({ ...params, ...searchForm.value }));
const columns = [
    { label: '商品', prop: 'product_title', minWidth: 160 },
    { label: 'SKU', slot: 'sku', width: 160 },
    { label: '当前库存', prop: 'stock', width: 100, align: 'center' },
    { label: '预警阈值', prop: 'threshold', width: 100, align: 'center' },
    { label: '状态', slot: 'status', width: 90, align: 'center' },
    { label: '时间', prop: 'created_at', width: 150, formatter: (r) => formatTime(r.created_at) },
];
async function handleMarkRead(row) {
    try {
        await markAlertRead(row.id);
        ElMessage.success('已标记已读');
        fetch(searchForm.value);
    }
    catch (e) {
        ElMessage.error(e?.message || '操作失败');
    }
}
async function handleMarkAllRead() {
    try {
        await markAllAlertsRead();
        ElMessage.success('全部已读');
        fetch(searchForm.value);
    }
    catch (e) {
        ElMessage.error(e?.message || '操作失败');
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
    const __VLS_11 = {}.ElOption;
    /** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
    // @ts-ignore
    const __VLS_12 = __VLS_asFunctionalComponent(__VLS_11, new __VLS_11({
        label: "未读",
        value: "unread",
    }));
    const __VLS_13 = __VLS_12({
        label: "未读",
        value: "unread",
    }, ...__VLS_functionalComponentArgsRest(__VLS_12));
    const __VLS_15 = {}.ElOption;
    /** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
    // @ts-ignore
    const __VLS_16 = __VLS_asFunctionalComponent(__VLS_15, new __VLS_15({
        label: "已读",
        value: "read",
    }));
    const __VLS_17 = __VLS_16({
        label: "已读",
        value: "read",
    }, ...__VLS_functionalComponentArgsRest(__VLS_16));
    var __VLS_10;
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
    const { toolbar: __VLS_thisSlot } = __VLS_2.slots;
    const __VLS_27 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_28 = __VLS_asFunctionalComponent(__VLS_27, new __VLS_27({
        ...{ 'onClick': {} },
    }));
    const __VLS_29 = __VLS_28({
        ...{ 'onClick': {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_28));
    let __VLS_31;
    let __VLS_32;
    let __VLS_33;
    const __VLS_34 = {
        onClick: (__VLS_ctx.handleMarkAllRead)
    };
    __VLS_30.slots.default;
    var __VLS_30;
}
{
    const { sku: __VLS_thisSlot } = __VLS_2.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({});
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ style: {} },
    });
    (row.sku_code);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ style: {} },
    });
    (row.spec_values?.join(' / '));
}
{
    const { status: __VLS_thisSlot } = __VLS_2.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_35 = {}.ElTag;
    /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
    // @ts-ignore
    const __VLS_36 = __VLS_asFunctionalComponent(__VLS_35, new __VLS_35({
        type: (row.status === 'unread' ? 'danger' : 'info'),
        size: "small",
    }));
    const __VLS_37 = __VLS_36({
        type: (row.status === 'unread' ? 'danger' : 'info'),
        size: "small",
    }, ...__VLS_functionalComponentArgsRest(__VLS_36));
    __VLS_38.slots.default;
    (row.status === 'unread' ? '未读' : '已读');
    var __VLS_38;
}
{
    const { actions: __VLS_thisSlot } = __VLS_2.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    if (row.status === 'unread') {
        const __VLS_39 = {}.ElButton;
        /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
        // @ts-ignore
        const __VLS_40 = __VLS_asFunctionalComponent(__VLS_39, new __VLS_39({
            ...{ 'onClick': {} },
            text: true,
            type: "primary",
            size: "small",
        }));
        const __VLS_41 = __VLS_40({
            ...{ 'onClick': {} },
            text: true,
            type: "primary",
            size: "small",
        }, ...__VLS_functionalComponentArgsRest(__VLS_40));
        let __VLS_43;
        let __VLS_44;
        let __VLS_45;
        const __VLS_46 = {
            onClick: (...[$event]) => {
                if (!(row.status === 'unread'))
                    return;
                __VLS_ctx.handleMarkRead(row);
            }
        };
        __VLS_42.slots.default;
        var __VLS_42;
    }
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
            handleMarkRead: handleMarkRead,
            handleMarkAllRead: handleMarkAllRead,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
