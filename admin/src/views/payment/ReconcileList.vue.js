import { ref, onMounted } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import ProTable from '@/components/ProTable/index.vue';
import { useTable } from '@/composables/useTable';
import { getReconcileList, resolveReconcile } from '@/api/payment';
import { formatAmount, formatDate } from '@/utils/format';
const searchForm = ref({ date: '', channel: '' });
const { list, total, page, pageSize, loading, fetch } = useTable((params) => getReconcileList({ ...params, ...searchForm.value }));
const columns = [
    { label: '日期', prop: 'date', width: 120, formatter: (r) => formatDate(r.date) },
    { label: '渠道', prop: 'channel', width: 100 },
    { label: '交易笔数', prop: 'order_count', width: 100, align: 'right' },
    { label: '交易金额', slot: 'total_amount', width: 120, align: 'right' },
    { label: '退款金额', slot: 'refund_amount', width: 120, align: 'right' },
    { label: '净收入', slot: 'net_amount', width: 120, align: 'right' },
    { label: '状态', slot: 'status', width: 90, align: 'center' },
];
async function handleResolve(row) {
    try {
        await ElMessageBox.confirm('确认标记该差异已处理？', '提示', { type: 'warning' });
        await resolveReconcile(row.id);
        ElMessage.success('已处理');
        fetch(searchForm.value);
    }
    catch (e) {
        if (e !== 'cancel')
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
    const __VLS_7 = {}.ElDatePicker;
    /** @type {[typeof __VLS_components.ElDatePicker, typeof __VLS_components.elDatePicker, ]} */ ;
    // @ts-ignore
    const __VLS_8 = __VLS_asFunctionalComponent(__VLS_7, new __VLS_7({
        modelValue: (__VLS_ctx.searchForm.date),
        type: "date",
        valueFormat: "YYYY-MM-DD",
        placeholder: "选择日期",
        ...{ style: {} },
    }));
    const __VLS_9 = __VLS_8({
        modelValue: (__VLS_ctx.searchForm.date),
        type: "date",
        valueFormat: "YYYY-MM-DD",
        placeholder: "选择日期",
        ...{ style: {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_8));
    const __VLS_11 = {}.ElSelect;
    /** @type {[typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, ]} */ ;
    // @ts-ignore
    const __VLS_12 = __VLS_asFunctionalComponent(__VLS_11, new __VLS_11({
        modelValue: (__VLS_ctx.searchForm.channel),
        placeholder: "渠道",
        clearable: true,
        ...{ style: {} },
    }));
    const __VLS_13 = __VLS_12({
        modelValue: (__VLS_ctx.searchForm.channel),
        placeholder: "渠道",
        clearable: true,
        ...{ style: {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_12));
    __VLS_14.slots.default;
    const __VLS_15 = {}.ElOption;
    /** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
    // @ts-ignore
    const __VLS_16 = __VLS_asFunctionalComponent(__VLS_15, new __VLS_15({
        label: "微信",
        value: "wechat",
    }));
    const __VLS_17 = __VLS_16({
        label: "微信",
        value: "wechat",
    }, ...__VLS_functionalComponentArgsRest(__VLS_16));
    const __VLS_19 = {}.ElOption;
    /** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
    // @ts-ignore
    const __VLS_20 = __VLS_asFunctionalComponent(__VLS_19, new __VLS_19({
        label: "支付宝",
        value: "alipay",
    }));
    const __VLS_21 = __VLS_20({
        label: "支付宝",
        value: "alipay",
    }, ...__VLS_functionalComponentArgsRest(__VLS_20));
    var __VLS_14;
    const __VLS_23 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_24 = __VLS_asFunctionalComponent(__VLS_23, new __VLS_23({
        ...{ 'onClick': {} },
        type: "primary",
    }));
    const __VLS_25 = __VLS_24({
        ...{ 'onClick': {} },
        type: "primary",
    }, ...__VLS_functionalComponentArgsRest(__VLS_24));
    let __VLS_27;
    let __VLS_28;
    let __VLS_29;
    const __VLS_30 = {
        onClick: (...[$event]) => {
            __VLS_ctx.fetch(__VLS_ctx.searchForm);
        }
    };
    __VLS_26.slots.default;
    var __VLS_26;
}
{
    const { total_amount: __VLS_thisSlot } = __VLS_2.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    (__VLS_ctx.formatAmount(row.total_amount));
}
{
    const { refund_amount: __VLS_thisSlot } = __VLS_2.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ style: {} },
    });
    (__VLS_ctx.formatAmount(row.refund_amount));
}
{
    const { net_amount: __VLS_thisSlot } = __VLS_2.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ style: {} },
    });
    (__VLS_ctx.formatAmount(row.net_amount));
}
{
    const { status: __VLS_thisSlot } = __VLS_2.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_31 = {}.ElTag;
    /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
    // @ts-ignore
    const __VLS_32 = __VLS_asFunctionalComponent(__VLS_31, new __VLS_31({
        type: (row.status === 'balanced' ? 'success' : 'warning'),
        size: "small",
    }));
    const __VLS_33 = __VLS_32({
        type: (row.status === 'balanced' ? 'success' : 'warning'),
        size: "small",
    }, ...__VLS_functionalComponentArgsRest(__VLS_32));
    __VLS_34.slots.default;
    (row.status === 'balanced' ? '已平账' : '待处理');
    var __VLS_34;
}
{
    const { actions: __VLS_thisSlot } = __VLS_2.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    if (row.status !== 'balanced') {
        const __VLS_35 = {}.ElButton;
        /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
        // @ts-ignore
        const __VLS_36 = __VLS_asFunctionalComponent(__VLS_35, new __VLS_35({
            ...{ 'onClick': {} },
            text: true,
            type: "primary",
            size: "small",
        }));
        const __VLS_37 = __VLS_36({
            ...{ 'onClick': {} },
            text: true,
            type: "primary",
            size: "small",
        }, ...__VLS_functionalComponentArgsRest(__VLS_36));
        let __VLS_39;
        let __VLS_40;
        let __VLS_41;
        const __VLS_42 = {
            onClick: (...[$event]) => {
                if (!(row.status !== 'balanced'))
                    return;
                __VLS_ctx.handleResolve(row);
            }
        };
        __VLS_38.slots.default;
        var __VLS_38;
    }
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
            handleResolve: handleResolve,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
