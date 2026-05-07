import { ref, onMounted } from 'vue';
import VChart from 'vue-echarts';
import { use } from 'echarts/core';
import { LineChart } from 'echarts/charts';
import { CanvasRenderer } from 'echarts/renderers';
import { GridComponent, TooltipComponent, LegendComponent, TitleComponent, } from 'echarts/components';
import { getUserStats } from '@/api/stats';
import dayjs from 'dayjs';
use([CanvasRenderer, LineChart, GridComponent, TooltipComponent, LegendComponent, TitleComponent]);
const dateRange = ref([
    dayjs().subtract(29, 'day').format('YYYY-MM-DD'),
    dayjs().format('YYYY-MM-DD'),
]);
const loading = ref(false);
const stats = ref({});
const trendOption = ref({});
async function loadData() {
    if (!dateRange.value)
        return;
    loading.value = true;
    try {
        const [from, to] = dateRange.value;
        const res = await getUserStats({ from, to });
        stats.value = res || {};
        trendOption.value = buildTrend(res?.new_users_trend || []);
    }
    finally {
        loading.value = false;
    }
}
function buildTrend(data) {
    return {
        tooltip: { trigger: 'axis' },
        grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
        xAxis: { type: 'category', data: data.map((d) => d.date) },
        yAxis: { type: 'value', name: '新增用户数' },
        series: [
            {
                name: '新增用户',
                type: 'line',
                smooth: true,
                data: data.map((d) => d.count),
                itemStyle: { color: '#3b82f6' },
                areaStyle: { opacity: 0.1, color: '#3b82f6' },
            },
        ],
    };
}
onMounted(loadData);
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
    span: (8),
}));
const __VLS_18 = __VLS_17({
    span: (8),
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
(__VLS_ctx.stats.new_users || 0);
var __VLS_19;
const __VLS_20 = {}.ElCol;
/** @type {[typeof __VLS_components.ElCol, typeof __VLS_components.elCol, typeof __VLS_components.ElCol, typeof __VLS_components.elCol, ]} */ ;
// @ts-ignore
const __VLS_21 = __VLS_asFunctionalComponent(__VLS_20, new __VLS_20({
    span: (8),
}));
const __VLS_22 = __VLS_21({
    span: (8),
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
(__VLS_ctx.stats.active_users || 0);
var __VLS_23;
const __VLS_24 = {}.ElCol;
/** @type {[typeof __VLS_components.ElCol, typeof __VLS_components.elCol, typeof __VLS_components.ElCol, typeof __VLS_components.elCol, ]} */ ;
// @ts-ignore
const __VLS_25 = __VLS_asFunctionalComponent(__VLS_24, new __VLS_24({
    span: (8),
}));
const __VLS_26 = __VLS_25({
    span: (8),
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
(__VLS_ctx.stats.total_users || 0);
var __VLS_27;
var __VLS_15;
const __VLS_28 = {}.ElCard;
/** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
// @ts-ignore
const __VLS_29 = __VLS_asFunctionalComponent(__VLS_28, new __VLS_28({}));
const __VLS_30 = __VLS_29({}, ...__VLS_functionalComponentArgsRest(__VLS_29));
__VLS_31.slots.default;
{
    const { header: __VLS_thisSlot } = __VLS_31.slots;
}
const __VLS_32 = {}.VChart;
/** @type {[typeof __VLS_components.VChart, ]} */ ;
// @ts-ignore
const __VLS_33 = __VLS_asFunctionalComponent(__VLS_32, new __VLS_32({
    option: (__VLS_ctx.trendOption),
    ...{ style: {} },
    autoresize: true,
}));
const __VLS_34 = __VLS_33({
    option: (__VLS_ctx.trendOption),
    ...{ style: {} },
    autoresize: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_33));
var __VLS_31;
/** @type {__VLS_StyleScopedClasses['page-card']} */ ;
/** @type {__VLS_StyleScopedClasses['stats-card']} */ ;
/** @type {__VLS_StyleScopedClasses['stats-label']} */ ;
/** @type {__VLS_StyleScopedClasses['stats-value']} */ ;
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
            dateRange: dateRange,
            loading: loading,
            stats: stats,
            trendOption: trendOption,
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
