import { ref, watch } from 'vue';
import VChart from 'vue-echarts';
import { use } from 'echarts/core';
import { LineChart, BarChart } from 'echarts/charts';
import { CanvasRenderer } from 'echarts/renderers';
import { GridComponent, TooltipComponent, LegendComponent, TitleComponent, } from 'echarts/components';
import { getStatsOverview, getStatsTrend, getCategoryStats } from '@/api/stats';
import { formatAmount } from '@/utils/format';
import dayjs from 'dayjs';
use([CanvasRenderer, LineChart, BarChart, GridComponent, TooltipComponent, LegendComponent, TitleComponent]);
const dateRange = ref([
    dayjs().subtract(6, 'day').format('YYYY-MM-DD'),
    dayjs().format('YYYY-MM-DD'),
]);
const loading = ref(false);
const overview = ref({});
const trendOption = ref({});
const categoryOption = ref({});
async function loadData() {
    if (!dateRange.value)
        return;
    loading.value = true;
    try {
        const [from, to] = dateRange.value;
        const [ov, trend, cats] = await Promise.all([
            getStatsOverview({ from, to }),
            getStatsTrend({ from, to }),
            getCategoryStats({ from, to }),
        ]);
        overview.value = ov;
        trendOption.value = buildTrend(trend || []);
        categoryOption.value = buildCategory(cats || []);
    }
    finally {
        loading.value = false;
    }
}
function buildTrend(data) {
    return {
        tooltip: { trigger: 'axis' },
        legend: { data: ['订单数', '销售额(元)'] },
        grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
        xAxis: { type: 'category', data: data.map((d) => d.date) },
        yAxis: [{ type: 'value', name: '订单数' }, { type: 'value', name: '销售额' }],
        series: [
            {
                name: '订单数',
                type: 'line',
                smooth: true,
                data: data.map((d) => d.order_count),
                itemStyle: { color: '#f59e0b' },
                areaStyle: { opacity: 0.1, color: '#f59e0b' },
            },
            {
                name: '销售额(元)',
                type: 'line',
                yAxisIndex: 1,
                smooth: true,
                data: data.map((d) => (d.order_amount / 100).toFixed(2)),
                itemStyle: { color: '#3b82f6' },
            },
        ],
    };
}
function buildCategory(data) {
    return {
        tooltip: { trigger: 'axis', axisPointer: { type: 'shadow' } },
        grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
        xAxis: { type: 'category', data: data.map((d) => d.name) },
        yAxis: { type: 'value' },
        series: [
            {
                name: '销售额',
                type: 'bar',
                data: data.map((d) => (d.amount / 100).toFixed(2)),
                itemStyle: { color: '#f59e0b', borderRadius: [4, 4, 0, 0] },
            },
        ],
    };
}
watch(dateRange, loadData, { immediate: true });
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({});
__VLS_asFunctionalDirective(__VLS_directives.vLoading)(null, { ...__VLS_directiveBindingRestFields, value: (__VLS_ctx.loading) }, null, null);
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "page-card" },
    ...{ style: {} },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
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
const __VLS_12 = {}.ElRow;
/** @type {[typeof __VLS_components.ElRow, typeof __VLS_components.elRow, typeof __VLS_components.ElRow, typeof __VLS_components.elRow, ]} */ ;
// @ts-ignore
const __VLS_13 = __VLS_asFunctionalComponent(__VLS_12, new __VLS_12({
    gutter: (16),
    ...{ style: {} },
}));
const __VLS_14 = __VLS_13({
    gutter: (16),
    ...{ style: {} },
}, ...__VLS_functionalComponentArgsRest(__VLS_13));
__VLS_15.slots.default;
const __VLS_16 = {}.ElCol;
/** @type {[typeof __VLS_components.ElCol, typeof __VLS_components.elCol, typeof __VLS_components.ElCol, typeof __VLS_components.elCol, ]} */ ;
// @ts-ignore
const __VLS_17 = __VLS_asFunctionalComponent(__VLS_16, new __VLS_16({
    span: (6),
}));
const __VLS_18 = __VLS_17({
    span: (6),
}, ...__VLS_functionalComponentArgsRest(__VLS_17));
__VLS_19.slots.default;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "stats-card" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "stats-label" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "stats-value" },
});
(__VLS_ctx.overview.order_count || 0);
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "stats-sub" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
    ...{ class: ((__VLS_ctx.overview.order_count_ratio || 0) >= 0 ? 'up' : 'down') },
});
((__VLS_ctx.overview.order_count_ratio || 0) >= 0 ? '↑' : '↓');
(Math.abs(__VLS_ctx.overview.order_count_ratio || 0).toFixed(1));
var __VLS_19;
const __VLS_20 = {}.ElCol;
/** @type {[typeof __VLS_components.ElCol, typeof __VLS_components.elCol, typeof __VLS_components.ElCol, typeof __VLS_components.elCol, ]} */ ;
// @ts-ignore
const __VLS_21 = __VLS_asFunctionalComponent(__VLS_20, new __VLS_20({
    span: (6),
}));
const __VLS_22 = __VLS_21({
    span: (6),
}, ...__VLS_functionalComponentArgsRest(__VLS_21));
__VLS_23.slots.default;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "stats-card" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "stats-label" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "stats-value" },
    ...{ style: {} },
});
(__VLS_ctx.formatAmount(__VLS_ctx.overview.order_amount || 0));
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "stats-sub" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
    ...{ class: ((__VLS_ctx.overview.order_amount_ratio || 0) >= 0 ? 'up' : 'down') },
});
((__VLS_ctx.overview.order_amount_ratio || 0) >= 0 ? '↑' : '↓');
(Math.abs(__VLS_ctx.overview.order_amount_ratio || 0).toFixed(1));
var __VLS_23;
const __VLS_24 = {}.ElCol;
/** @type {[typeof __VLS_components.ElCol, typeof __VLS_components.elCol, typeof __VLS_components.ElCol, typeof __VLS_components.elCol, ]} */ ;
// @ts-ignore
const __VLS_25 = __VLS_asFunctionalComponent(__VLS_24, new __VLS_24({
    span: (6),
}));
const __VLS_26 = __VLS_25({
    span: (6),
}, ...__VLS_functionalComponentArgsRest(__VLS_25));
__VLS_27.slots.default;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "stats-card" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "stats-label" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "stats-value" },
    ...{ style: {} },
});
(__VLS_ctx.formatAmount(__VLS_ctx.overview.refund_amount || 0));
var __VLS_27;
const __VLS_28 = {}.ElCol;
/** @type {[typeof __VLS_components.ElCol, typeof __VLS_components.elCol, typeof __VLS_components.ElCol, typeof __VLS_components.elCol, ]} */ ;
// @ts-ignore
const __VLS_29 = __VLS_asFunctionalComponent(__VLS_28, new __VLS_28({
    span: (6),
}));
const __VLS_30 = __VLS_29({
    span: (6),
}, ...__VLS_functionalComponentArgsRest(__VLS_29));
__VLS_31.slots.default;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "stats-card" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "stats-label" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "stats-value" },
    ...{ style: {} },
});
(__VLS_ctx.formatAmount(__VLS_ctx.overview.net_income || 0));
var __VLS_31;
var __VLS_15;
const __VLS_32 = {}.ElRow;
/** @type {[typeof __VLS_components.ElRow, typeof __VLS_components.elRow, typeof __VLS_components.ElRow, typeof __VLS_components.elRow, ]} */ ;
// @ts-ignore
const __VLS_33 = __VLS_asFunctionalComponent(__VLS_32, new __VLS_32({
    gutter: (16),
}));
const __VLS_34 = __VLS_33({
    gutter: (16),
}, ...__VLS_functionalComponentArgsRest(__VLS_33));
__VLS_35.slots.default;
const __VLS_36 = {}.ElCol;
/** @type {[typeof __VLS_components.ElCol, typeof __VLS_components.elCol, typeof __VLS_components.ElCol, typeof __VLS_components.elCol, ]} */ ;
// @ts-ignore
const __VLS_37 = __VLS_asFunctionalComponent(__VLS_36, new __VLS_36({
    span: (14),
}));
const __VLS_38 = __VLS_37({
    span: (14),
}, ...__VLS_functionalComponentArgsRest(__VLS_37));
__VLS_39.slots.default;
const __VLS_40 = {}.ElCard;
/** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
// @ts-ignore
const __VLS_41 = __VLS_asFunctionalComponent(__VLS_40, new __VLS_40({}));
const __VLS_42 = __VLS_41({}, ...__VLS_functionalComponentArgsRest(__VLS_41));
__VLS_43.slots.default;
{
    const { header: __VLS_thisSlot } = __VLS_43.slots;
}
const __VLS_44 = {}.VChart;
/** @type {[typeof __VLS_components.VChart, ]} */ ;
// @ts-ignore
const __VLS_45 = __VLS_asFunctionalComponent(__VLS_44, new __VLS_44({
    option: (__VLS_ctx.trendOption),
    ...{ style: {} },
    autoresize: true,
}));
const __VLS_46 = __VLS_45({
    option: (__VLS_ctx.trendOption),
    ...{ style: {} },
    autoresize: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_45));
var __VLS_43;
var __VLS_39;
const __VLS_48 = {}.ElCol;
/** @type {[typeof __VLS_components.ElCol, typeof __VLS_components.elCol, typeof __VLS_components.ElCol, typeof __VLS_components.elCol, ]} */ ;
// @ts-ignore
const __VLS_49 = __VLS_asFunctionalComponent(__VLS_48, new __VLS_48({
    span: (10),
}));
const __VLS_50 = __VLS_49({
    span: (10),
}, ...__VLS_functionalComponentArgsRest(__VLS_49));
__VLS_51.slots.default;
const __VLS_52 = {}.ElCard;
/** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
// @ts-ignore
const __VLS_53 = __VLS_asFunctionalComponent(__VLS_52, new __VLS_52({}));
const __VLS_54 = __VLS_53({}, ...__VLS_functionalComponentArgsRest(__VLS_53));
__VLS_55.slots.default;
{
    const { header: __VLS_thisSlot } = __VLS_55.slots;
}
const __VLS_56 = {}.VChart;
/** @type {[typeof __VLS_components.VChart, ]} */ ;
// @ts-ignore
const __VLS_57 = __VLS_asFunctionalComponent(__VLS_56, new __VLS_56({
    option: (__VLS_ctx.categoryOption),
    ...{ style: {} },
    autoresize: true,
}));
const __VLS_58 = __VLS_57({
    option: (__VLS_ctx.categoryOption),
    ...{ style: {} },
    autoresize: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_57));
var __VLS_55;
var __VLS_51;
var __VLS_35;
/** @type {__VLS_StyleScopedClasses['page-card']} */ ;
/** @type {__VLS_StyleScopedClasses['stats-card']} */ ;
/** @type {__VLS_StyleScopedClasses['stats-label']} */ ;
/** @type {__VLS_StyleScopedClasses['stats-value']} */ ;
/** @type {__VLS_StyleScopedClasses['stats-sub']} */ ;
/** @type {__VLS_StyleScopedClasses['stats-card']} */ ;
/** @type {__VLS_StyleScopedClasses['stats-label']} */ ;
/** @type {__VLS_StyleScopedClasses['stats-value']} */ ;
/** @type {__VLS_StyleScopedClasses['stats-sub']} */ ;
/** @type {__VLS_StyleScopedClasses['stats-card']} */ ;
/** @type {__VLS_StyleScopedClasses['stats-label']} */ ;
/** @type {__VLS_StyleScopedClasses['stats-value']} */ ;
/** @type {__VLS_StyleScopedClasses['stats-card']} */ ;
/** @type {__VLS_StyleScopedClasses['stats-label']} */ ;
/** @type {__VLS_StyleScopedClasses['stats-value']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            VChart: VChart,
            formatAmount: formatAmount,
            dateRange: dateRange,
            loading: loading,
            overview: overview,
            trendOption: trendOption,
            categoryOption: categoryOption,
            loadData: loadData,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
