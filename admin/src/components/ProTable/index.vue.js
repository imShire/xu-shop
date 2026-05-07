import { ref, computed } from 'vue';
import { RefreshRight, Setting } from '@element-plus/icons-vue';
const props = withDefaults(defineProps(), {
    total: 0,
    loading: false,
    rowKey: 'id',
    selection: false,
    pageSizes: () => [10, 20, 50, 100],
});
const page = defineModel('page', { default: 1 });
const pageSize = defineModel('pageSize', { default: 20 });
const emit = defineEmits();
const hiddenCols = ref(new Set());
const settingVisible = ref(false);
const visibleColumns = computed(() => props.columns.filter((col) => !col.hidden && !hiddenCols.value.has(col.label)));
function toggleCol(label) {
    if (hiddenCols.value.has(label)) {
        hiddenCols.value.delete(label);
    }
    else {
        hiddenCols.value.add(label);
    }
}
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_withDefaultsArg = (function (t) { return t; })({
    total: 0,
    loading: false,
    rowKey: 'id',
    selection: false,
    pageSizes: () => [10, 20, 50, 100],
});
const __VLS_defaults = {
    'page': 1,
    'pageSize': 20,
};
const __VLS_modelEmit = defineEmits();
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "pro-table" },
});
if (__VLS_ctx.$slots.search) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "search-bar" },
    });
    var __VLS_0 = {};
}
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "toolbar" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "toolbar-left" },
});
var __VLS_2 = {};
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "toolbar-right" },
    ...{ style: {} },
});
const __VLS_4 = {}.ElTooltip;
/** @type {[typeof __VLS_components.ElTooltip, typeof __VLS_components.elTooltip, typeof __VLS_components.ElTooltip, typeof __VLS_components.elTooltip, ]} */ ;
// @ts-ignore
const __VLS_5 = __VLS_asFunctionalComponent(__VLS_4, new __VLS_4({
    content: "刷新",
}));
const __VLS_6 = __VLS_5({
    content: "刷新",
}, ...__VLS_functionalComponentArgsRest(__VLS_5));
__VLS_7.slots.default;
const __VLS_8 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_9 = __VLS_asFunctionalComponent(__VLS_8, new __VLS_8({
    ...{ 'onClick': {} },
    icon: (__VLS_ctx.RefreshRight),
    circle: true,
    size: "small",
}));
const __VLS_10 = __VLS_9({
    ...{ 'onClick': {} },
    icon: (__VLS_ctx.RefreshRight),
    circle: true,
    size: "small",
}, ...__VLS_functionalComponentArgsRest(__VLS_9));
let __VLS_12;
let __VLS_13;
let __VLS_14;
const __VLS_15 = {
    onClick: (...[$event]) => {
        __VLS_ctx.emit('refresh');
    }
};
var __VLS_11;
var __VLS_7;
const __VLS_16 = {}.ElPopover;
/** @type {[typeof __VLS_components.ElPopover, typeof __VLS_components.elPopover, typeof __VLS_components.ElPopover, typeof __VLS_components.elPopover, ]} */ ;
// @ts-ignore
const __VLS_17 = __VLS_asFunctionalComponent(__VLS_16, new __VLS_16({
    visible: (__VLS_ctx.settingVisible),
    placement: "bottom-end",
    trigger: "click",
    width: "180",
}));
const __VLS_18 = __VLS_17({
    visible: (__VLS_ctx.settingVisible),
    placement: "bottom-end",
    trigger: "click",
    width: "180",
}, ...__VLS_functionalComponentArgsRest(__VLS_17));
__VLS_19.slots.default;
{
    const { reference: __VLS_thisSlot } = __VLS_19.slots;
    const __VLS_20 = {}.ElTooltip;
    /** @type {[typeof __VLS_components.ElTooltip, typeof __VLS_components.elTooltip, typeof __VLS_components.ElTooltip, typeof __VLS_components.elTooltip, ]} */ ;
    // @ts-ignore
    const __VLS_21 = __VLS_asFunctionalComponent(__VLS_20, new __VLS_20({
        content: "列设置",
    }));
    const __VLS_22 = __VLS_21({
        content: "列设置",
    }, ...__VLS_functionalComponentArgsRest(__VLS_21));
    __VLS_23.slots.default;
    const __VLS_24 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_25 = __VLS_asFunctionalComponent(__VLS_24, new __VLS_24({
        icon: (__VLS_ctx.Setting),
        circle: true,
        size: "small",
    }));
    const __VLS_26 = __VLS_25({
        icon: (__VLS_ctx.Setting),
        circle: true,
        size: "small",
    }, ...__VLS_functionalComponentArgsRest(__VLS_25));
    var __VLS_23;
}
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ style: {} },
});
for (const [col] of __VLS_getVForSourceType((__VLS_ctx.columns))) {
    const __VLS_28 = {}.ElCheckbox;
    /** @type {[typeof __VLS_components.ElCheckbox, typeof __VLS_components.elCheckbox, typeof __VLS_components.ElCheckbox, typeof __VLS_components.elCheckbox, ]} */ ;
    // @ts-ignore
    const __VLS_29 = __VLS_asFunctionalComponent(__VLS_28, new __VLS_28({
        ...{ 'onChange': {} },
        key: (col.label),
        modelValue: (!__VLS_ctx.hiddenCols.has(col.label)),
        label: (col.label),
    }));
    const __VLS_30 = __VLS_29({
        ...{ 'onChange': {} },
        key: (col.label),
        modelValue: (!__VLS_ctx.hiddenCols.has(col.label)),
        label: (col.label),
    }, ...__VLS_functionalComponentArgsRest(__VLS_29));
    let __VLS_32;
    let __VLS_33;
    let __VLS_34;
    const __VLS_35 = {
        onChange: (...[$event]) => {
            __VLS_ctx.toggleCol(col.label);
        }
    };
    __VLS_31.slots.default;
    (col.label);
    var __VLS_31;
}
var __VLS_19;
const __VLS_36 = {}.ElTable;
/** @type {[typeof __VLS_components.ElTable, typeof __VLS_components.elTable, typeof __VLS_components.ElTable, typeof __VLS_components.elTable, ]} */ ;
// @ts-ignore
const __VLS_37 = __VLS_asFunctionalComponent(__VLS_36, new __VLS_36({
    ...{ 'onSelectionChange': {} },
    data: (__VLS_ctx.data),
    rowKey: (__VLS_ctx.rowKey),
    border: true,
    stripe: true,
    ...{ style: {} },
}));
const __VLS_38 = __VLS_37({
    ...{ 'onSelectionChange': {} },
    data: (__VLS_ctx.data),
    rowKey: (__VLS_ctx.rowKey),
    border: true,
    stripe: true,
    ...{ style: {} },
}, ...__VLS_functionalComponentArgsRest(__VLS_37));
let __VLS_40;
let __VLS_41;
let __VLS_42;
const __VLS_43 = {
    onSelectionChange: ((rows) => __VLS_ctx.emit('selectionChange', rows))
};
__VLS_asFunctionalDirective(__VLS_directives.vLoading)(null, { ...__VLS_directiveBindingRestFields, value: (__VLS_ctx.loading) }, null, null);
__VLS_39.slots.default;
if (__VLS_ctx.selection) {
    const __VLS_44 = {}.ElTableColumn;
    /** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
    // @ts-ignore
    const __VLS_45 = __VLS_asFunctionalComponent(__VLS_44, new __VLS_44({
        type: "selection",
        width: "48",
        fixed: "left",
    }));
    const __VLS_46 = __VLS_45({
        type: "selection",
        width: "48",
        fixed: "left",
    }, ...__VLS_functionalComponentArgsRest(__VLS_45));
}
for (const [col] of __VLS_getVForSourceType((__VLS_ctx.visibleColumns))) {
    const __VLS_48 = {}.ElTableColumn;
    /** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
    // @ts-ignore
    const __VLS_49 = __VLS_asFunctionalComponent(__VLS_48, new __VLS_48({
        prop: (col.prop),
        label: (col.label),
        width: (col.width),
        minWidth: (col.minWidth),
        align: (col.align || 'left'),
        fixed: (col.fixed),
        showOverflowTooltip: true,
    }));
    const __VLS_50 = __VLS_49({
        prop: (col.prop),
        label: (col.label),
        width: (col.width),
        minWidth: (col.minWidth),
        align: (col.align || 'left'),
        fixed: (col.fixed),
        showOverflowTooltip: true,
    }, ...__VLS_functionalComponentArgsRest(__VLS_49));
    __VLS_51.slots.default;
    if (col.slot || col.formatter) {
        {
            const { default: __VLS_thisSlot } = __VLS_51.slots;
            const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
            if (col.slot) {
                var __VLS_52 = {
                    row: (row),
                };
                var __VLS_53 = __VLS_tryAsConstant(col.slot);
            }
            else if (col.formatter) {
                __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
                (col.formatter(row, col.prop ? row[col.prop] : undefined));
            }
        }
    }
    var __VLS_51;
}
if (__VLS_ctx.$slots.actions) {
    const __VLS_56 = {}.ElTableColumn;
    /** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
    // @ts-ignore
    const __VLS_57 = __VLS_asFunctionalComponent(__VLS_56, new __VLS_56({
        label: "操作",
        fixed: "right",
        width: "160",
    }));
    const __VLS_58 = __VLS_57({
        label: "操作",
        fixed: "right",
        width: "160",
    }, ...__VLS_functionalComponentArgsRest(__VLS_57));
    __VLS_59.slots.default;
    {
        const { default: __VLS_thisSlot } = __VLS_59.slots;
        const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
        var __VLS_60 = {
            row: (row),
        };
    }
    var __VLS_59;
}
var __VLS_39;
if (__VLS_ctx.total > 0) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "pagination-wrapper" },
    });
    const __VLS_62 = {}.ElPagination;
    /** @type {[typeof __VLS_components.ElPagination, typeof __VLS_components.elPagination, ]} */ ;
    // @ts-ignore
    const __VLS_63 = __VLS_asFunctionalComponent(__VLS_62, new __VLS_62({
        currentPage: (__VLS_ctx.page),
        pageSize: (__VLS_ctx.pageSize),
        total: (__VLS_ctx.total),
        pageSizes: (__VLS_ctx.pageSizes),
        layout: "total, sizes, prev, pager, next, jumper",
        background: true,
    }));
    const __VLS_64 = __VLS_63({
        currentPage: (__VLS_ctx.page),
        pageSize: (__VLS_ctx.pageSize),
        total: (__VLS_ctx.total),
        pageSizes: (__VLS_ctx.pageSizes),
        layout: "total, sizes, prev, pager, next, jumper",
        background: true,
    }, ...__VLS_functionalComponentArgsRest(__VLS_63));
}
/** @type {__VLS_StyleScopedClasses['pro-table']} */ ;
/** @type {__VLS_StyleScopedClasses['search-bar']} */ ;
/** @type {__VLS_StyleScopedClasses['toolbar']} */ ;
/** @type {__VLS_StyleScopedClasses['toolbar-left']} */ ;
/** @type {__VLS_StyleScopedClasses['toolbar-right']} */ ;
/** @type {__VLS_StyleScopedClasses['pagination-wrapper']} */ ;
// @ts-ignore
var __VLS_1 = __VLS_0, __VLS_3 = __VLS_2, __VLS_54 = __VLS_53, __VLS_55 = __VLS_52, __VLS_61 = __VLS_60;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            RefreshRight: RefreshRight,
            Setting: Setting,
            page: page,
            pageSize: pageSize,
            emit: emit,
            hiddenCols: hiddenCols,
            settingVisible: settingVisible,
            visibleColumns: visibleColumns,
            toggleCol: toggleCol,
        };
    },
    __typeEmits: {},
    __typeProps: {},
    props: {},
});
const __VLS_component = (await import('vue')).defineComponent({
    setup() {
        return {};
    },
    __typeEmits: {},
    __typeProps: {},
    props: {},
});
export default {};
; /* PartiallyEnd: #4569/main.vue */
