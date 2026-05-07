import { ref, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import VChart from 'vue-echarts';
import { use } from 'echarts/core';
import { LineChart } from 'echarts/charts';
import { CanvasRenderer } from 'echarts/renderers';
import { GridComponent, TooltipComponent, LegendComponent, } from 'echarts/components';
import { getWorkbenchStats, getStatsTrend } from '@/api/stats';
import { getInventoryAlerts } from '@/api/inventory';
import { getOrderList } from '@/api/order';
import { formatAmount, formatTime, orderStatusMap } from '@/utils/format';
import dayjs from 'dayjs';
use([CanvasRenderer, LineChart, GridComponent, TooltipComponent, LegendComponent]);
const router = useRouter();
const statsLoading = ref(false);
const stats = ref({
    today_order_count: 0,
    today_sales: 0,
    pending_ship: 0,
    aftersale_pending: 0,
});
const alertCount = ref(0);
const recentOrders = ref([]);
const trendOption = ref({});
async function loadData() {
    statsLoading.value = true;
    try {
        const [overview, trend, alerts, orders] = await Promise.all([
            getWorkbenchStats().catch(() => null),
            getStatsTrend({
                from: dayjs().subtract(6, 'day').format('YYYY-MM-DD'),
                to: dayjs().format('YYYY-MM-DD'),
            }).catch(() => []),
            getInventoryAlerts({ page: 1, page_size: 1, status: 'unread' }).catch(() => ({ total: 0 })),
            getOrderList({ page: 1, page_size: 10 }).catch(() => ({ list: [] })),
        ]);
        if (overview) {
            stats.value = overview;
        }
        alertCount.value = alerts?.total || 0;
        recentOrders.value = orders?.list || [];
        const trendData = trend || [];
        trendOption.value = buildTrendChart(trendData);
    }
    finally {
        statsLoading.value = false;
    }
}
function buildTrendChart(data) {
    return {
        tooltip: { trigger: 'axis' },
        legend: { data: ['订单数', '销售额(元)'] },
        grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
        xAxis: {
            type: 'category',
            data: data.map((d) => d.date),
        },
        yAxis: [
            { type: 'value', name: '订单数' },
            { type: 'value', name: '销售额' },
        ],
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
onMounted(loadData);
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
// CSS variable injection 
// CSS variable injection end 
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({});
const __VLS_0 = {}.ElRow;
/** @type {[typeof __VLS_components.ElRow, typeof __VLS_components.elRow, typeof __VLS_components.ElRow, typeof __VLS_components.elRow, ]} */ ;
// @ts-ignore
const __VLS_1 = __VLS_asFunctionalComponent(__VLS_0, new __VLS_0({
    gutter: (16),
    ...{ style: {} },
}));
const __VLS_2 = __VLS_1({
    gutter: (16),
    ...{ style: {} },
}, ...__VLS_functionalComponentArgsRest(__VLS_1));
__VLS_3.slots.default;
const __VLS_4 = {}.ElCol;
/** @type {[typeof __VLS_components.ElCol, typeof __VLS_components.elCol, typeof __VLS_components.ElCol, typeof __VLS_components.elCol, ]} */ ;
// @ts-ignore
const __VLS_5 = __VLS_asFunctionalComponent(__VLS_4, new __VLS_4({
    span: (6),
}));
const __VLS_6 = __VLS_5({
    span: (6),
}, ...__VLS_functionalComponentArgsRest(__VLS_5));
__VLS_7.slots.default;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "stats-card" },
});
__VLS_asFunctionalDirective(__VLS_directives.vLoading)(null, { ...__VLS_directiveBindingRestFields, value: (__VLS_ctx.statsLoading) }, null, null);
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "stats-label" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "stats-value" },
});
(__VLS_ctx.stats.today_order_count);
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "stats-sub" },
});
var __VLS_7;
const __VLS_8 = {}.ElCol;
/** @type {[typeof __VLS_components.ElCol, typeof __VLS_components.elCol, typeof __VLS_components.ElCol, typeof __VLS_components.elCol, ]} */ ;
// @ts-ignore
const __VLS_9 = __VLS_asFunctionalComponent(__VLS_8, new __VLS_8({
    span: (6),
}));
const __VLS_10 = __VLS_9({
    span: (6),
}, ...__VLS_functionalComponentArgsRest(__VLS_9));
__VLS_11.slots.default;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "stats-card" },
});
__VLS_asFunctionalDirective(__VLS_directives.vLoading)(null, { ...__VLS_directiveBindingRestFields, value: (__VLS_ctx.statsLoading) }, null, null);
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "stats-label" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "stats-value" },
});
(__VLS_ctx.formatAmount(__VLS_ctx.stats.today_sales));
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "stats-sub" },
});
var __VLS_11;
const __VLS_12 = {}.ElCol;
/** @type {[typeof __VLS_components.ElCol, typeof __VLS_components.elCol, typeof __VLS_components.ElCol, typeof __VLS_components.elCol, ]} */ ;
// @ts-ignore
const __VLS_13 = __VLS_asFunctionalComponent(__VLS_12, new __VLS_12({
    span: (6),
}));
const __VLS_14 = __VLS_13({
    span: (6),
}, ...__VLS_functionalComponentArgsRest(__VLS_13));
__VLS_15.slots.default;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "stats-card" },
});
__VLS_asFunctionalDirective(__VLS_directives.vLoading)(null, { ...__VLS_directiveBindingRestFields, value: (__VLS_ctx.statsLoading) }, null, null);
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "stats-label" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "stats-value" },
    ...{ style: {} },
});
(__VLS_ctx.stats.pending_ship);
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "stats-sub" },
});
const __VLS_16 = {}.ElLink;
/** @type {[typeof __VLS_components.ElLink, typeof __VLS_components.elLink, typeof __VLS_components.ElLink, typeof __VLS_components.elLink, ]} */ ;
// @ts-ignore
const __VLS_17 = __VLS_asFunctionalComponent(__VLS_16, new __VLS_16({
    ...{ 'onClick': {} },
    type: "primary",
}));
const __VLS_18 = __VLS_17({
    ...{ 'onClick': {} },
    type: "primary",
}, ...__VLS_functionalComponentArgsRest(__VLS_17));
let __VLS_20;
let __VLS_21;
let __VLS_22;
const __VLS_23 = {
    onClick: (...[$event]) => {
        __VLS_ctx.router.push('/shipping/pending');
    }
};
__VLS_19.slots.default;
var __VLS_19;
var __VLS_15;
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
__VLS_asFunctionalDirective(__VLS_directives.vLoading)(null, { ...__VLS_directiveBindingRestFields, value: (__VLS_ctx.statsLoading) }, null, null);
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "stats-label" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "stats-value" },
    ...{ style: {} },
});
(__VLS_ctx.stats.aftersale_pending);
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "stats-sub" },
});
const __VLS_28 = {}.ElLink;
/** @type {[typeof __VLS_components.ElLink, typeof __VLS_components.elLink, typeof __VLS_components.ElLink, typeof __VLS_components.elLink, ]} */ ;
// @ts-ignore
const __VLS_29 = __VLS_asFunctionalComponent(__VLS_28, new __VLS_28({
    ...{ 'onClick': {} },
    type: "danger",
}));
const __VLS_30 = __VLS_29({
    ...{ 'onClick': {} },
    type: "danger",
}, ...__VLS_functionalComponentArgsRest(__VLS_29));
let __VLS_32;
let __VLS_33;
let __VLS_34;
const __VLS_35 = {
    onClick: (...[$event]) => {
        __VLS_ctx.router.push('/aftersale/list');
    }
};
__VLS_31.slots.default;
var __VLS_31;
var __VLS_27;
var __VLS_3;
if (__VLS_ctx.alertCount > 0) {
    const __VLS_36 = {}.ElAlert;
    /** @type {[typeof __VLS_components.ElAlert, typeof __VLS_components.elAlert, typeof __VLS_components.ElAlert, typeof __VLS_components.elAlert, ]} */ ;
    // @ts-ignore
    const __VLS_37 = __VLS_asFunctionalComponent(__VLS_36, new __VLS_36({
        title: (`有 ${__VLS_ctx.alertCount} 条库存预警待处理`),
        type: "warning",
        showIcon: true,
        ...{ style: {} },
    }));
    const __VLS_38 = __VLS_37({
        title: (`有 ${__VLS_ctx.alertCount} 条库存预警待处理`),
        type: "warning",
        showIcon: true,
        ...{ style: {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_37));
    __VLS_39.slots.default;
    {
        const { default: __VLS_thisSlot } = __VLS_39.slots;
        const __VLS_40 = {}.ElLink;
        /** @type {[typeof __VLS_components.ElLink, typeof __VLS_components.elLink, typeof __VLS_components.ElLink, typeof __VLS_components.elLink, ]} */ ;
        // @ts-ignore
        const __VLS_41 = __VLS_asFunctionalComponent(__VLS_40, new __VLS_40({
            ...{ 'onClick': {} },
        }));
        const __VLS_42 = __VLS_41({
            ...{ 'onClick': {} },
        }, ...__VLS_functionalComponentArgsRest(__VLS_41));
        let __VLS_44;
        let __VLS_45;
        let __VLS_46;
        const __VLS_47 = {
            onClick: (...[$event]) => {
                if (!(__VLS_ctx.alertCount > 0))
                    return;
                __VLS_ctx.router.push('/inventory/alerts');
            }
        };
        __VLS_43.slots.default;
        var __VLS_43;
    }
    var __VLS_39;
}
const __VLS_48 = {}.ElRow;
/** @type {[typeof __VLS_components.ElRow, typeof __VLS_components.elRow, typeof __VLS_components.ElRow, typeof __VLS_components.elRow, ]} */ ;
// @ts-ignore
const __VLS_49 = __VLS_asFunctionalComponent(__VLS_48, new __VLS_48({
    gutter: (16),
}));
const __VLS_50 = __VLS_49({
    gutter: (16),
}, ...__VLS_functionalComponentArgsRest(__VLS_49));
__VLS_51.slots.default;
const __VLS_52 = {}.ElCol;
/** @type {[typeof __VLS_components.ElCol, typeof __VLS_components.elCol, typeof __VLS_components.ElCol, typeof __VLS_components.elCol, ]} */ ;
// @ts-ignore
const __VLS_53 = __VLS_asFunctionalComponent(__VLS_52, new __VLS_52({
    span: (16),
}));
const __VLS_54 = __VLS_53({
    span: (16),
}, ...__VLS_functionalComponentArgsRest(__VLS_53));
__VLS_55.slots.default;
const __VLS_56 = {}.ElCard;
/** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
// @ts-ignore
const __VLS_57 = __VLS_asFunctionalComponent(__VLS_56, new __VLS_56({}));
const __VLS_58 = __VLS_57({}, ...__VLS_functionalComponentArgsRest(__VLS_57));
__VLS_59.slots.default;
{
    const { header: __VLS_thisSlot } = __VLS_59.slots;
}
const __VLS_60 = {}.VChart;
/** @type {[typeof __VLS_components.VChart, ]} */ ;
// @ts-ignore
const __VLS_61 = __VLS_asFunctionalComponent(__VLS_60, new __VLS_60({
    option: (__VLS_ctx.trendOption),
    ...{ style: {} },
    autoresize: true,
}));
const __VLS_62 = __VLS_61({
    option: (__VLS_ctx.trendOption),
    ...{ style: {} },
    autoresize: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_61));
var __VLS_59;
var __VLS_55;
const __VLS_64 = {}.ElCol;
/** @type {[typeof __VLS_components.ElCol, typeof __VLS_components.elCol, typeof __VLS_components.ElCol, typeof __VLS_components.elCol, ]} */ ;
// @ts-ignore
const __VLS_65 = __VLS_asFunctionalComponent(__VLS_64, new __VLS_64({
    span: (8),
}));
const __VLS_66 = __VLS_65({
    span: (8),
}, ...__VLS_functionalComponentArgsRest(__VLS_65));
__VLS_67.slots.default;
const __VLS_68 = {}.ElCard;
/** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
// @ts-ignore
const __VLS_69 = __VLS_asFunctionalComponent(__VLS_68, new __VLS_68({}));
const __VLS_70 = __VLS_69({}, ...__VLS_functionalComponentArgsRest(__VLS_69));
__VLS_71.slots.default;
{
    const { header: __VLS_thisSlot } = __VLS_71.slots;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ style: {} },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
    const __VLS_72 = {}.ElLink;
    /** @type {[typeof __VLS_components.ElLink, typeof __VLS_components.elLink, typeof __VLS_components.ElLink, typeof __VLS_components.elLink, ]} */ ;
    // @ts-ignore
    const __VLS_73 = __VLS_asFunctionalComponent(__VLS_72, new __VLS_72({
        ...{ 'onClick': {} },
    }));
    const __VLS_74 = __VLS_73({
        ...{ 'onClick': {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_73));
    let __VLS_76;
    let __VLS_77;
    let __VLS_78;
    const __VLS_79 = {
        onClick: (...[$event]) => {
            __VLS_ctx.router.push('/order/list');
        }
    };
    __VLS_75.slots.default;
    var __VLS_75;
}
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ style: {} },
});
for (const [order] of __VLS_getVForSourceType((__VLS_ctx.recentOrders))) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ onClick: (...[$event]) => {
                __VLS_ctx.router.push(`/order/detail/${order.id}`);
            } },
        key: (order.id),
        ...{ class: "order-item" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "order-no" },
    });
    (order.order_no);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "order-info" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
    (__VLS_ctx.formatAmount(order.pay_amount));
    const __VLS_80 = {}.ElTag;
    /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
    // @ts-ignore
    const __VLS_81 = __VLS_asFunctionalComponent(__VLS_80, new __VLS_80({
        type: (__VLS_ctx.orderStatusMap[order.status]?.type || ''),
        size: "small",
    }));
    const __VLS_82 = __VLS_81({
        type: (__VLS_ctx.orderStatusMap[order.status]?.type || ''),
        size: "small",
    }, ...__VLS_functionalComponentArgsRest(__VLS_81));
    __VLS_83.slots.default;
    (__VLS_ctx.orderStatusMap[order.status]?.label || order.status);
    var __VLS_83;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "order-time" },
    });
    (__VLS_ctx.formatTime(order.created_at));
}
if (!__VLS_ctx.recentOrders.length) {
    const __VLS_84 = {}.ElEmpty;
    /** @type {[typeof __VLS_components.ElEmpty, typeof __VLS_components.elEmpty, ]} */ ;
    // @ts-ignore
    const __VLS_85 = __VLS_asFunctionalComponent(__VLS_84, new __VLS_84({
        description: "暂无订单",
        imageSize: (60),
    }));
    const __VLS_86 = __VLS_85({
        description: "暂无订单",
        imageSize: (60),
    }, ...__VLS_functionalComponentArgsRest(__VLS_85));
}
var __VLS_71;
var __VLS_67;
var __VLS_51;
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
/** @type {__VLS_StyleScopedClasses['stats-sub']} */ ;
/** @type {__VLS_StyleScopedClasses['stats-card']} */ ;
/** @type {__VLS_StyleScopedClasses['stats-label']} */ ;
/** @type {__VLS_StyleScopedClasses['stats-value']} */ ;
/** @type {__VLS_StyleScopedClasses['stats-sub']} */ ;
/** @type {__VLS_StyleScopedClasses['order-item']} */ ;
/** @type {__VLS_StyleScopedClasses['order-no']} */ ;
/** @type {__VLS_StyleScopedClasses['order-info']} */ ;
/** @type {__VLS_StyleScopedClasses['order-time']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            VChart: VChart,
            formatAmount: formatAmount,
            formatTime: formatTime,
            orderStatusMap: orderStatusMap,
            router: router,
            statsLoading: statsLoading,
            stats: stats,
            alertCount: alertCount,
            recentOrders: recentOrders,
            trendOption: trendOption,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
