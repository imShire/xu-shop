import { ref, onMounted } from 'vue';
import { ElMessage } from 'element-plus';
import { getCarriers, updateCarrier } from '@/api/shipping';
const loading = ref(false);
const carriers = ref([]);
const saving = ref({});
async function loadData() {
    loading.value = true;
    try {
        carriers.value = await getCarriers();
    }
    finally {
        loading.value = false;
    }
}
async function toggleEnabled(row) {
    saving.value[row.code] = true;
    try {
        await updateCarrier(row.code, { enabled: row.enabled });
        ElMessage.success('更新成功');
    }
    catch {
        row.enabled = !row.enabled;
    }
    finally {
        delete saving.value[row.code];
    }
}
onMounted(loadData);
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "page-card" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ style: {} },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.h3, __VLS_intrinsicElements.h3)({
    ...{ style: {} },
});
const __VLS_0 = {}.ElTable;
/** @type {[typeof __VLS_components.ElTable, typeof __VLS_components.elTable, typeof __VLS_components.ElTable, typeof __VLS_components.elTable, ]} */ ;
// @ts-ignore
const __VLS_1 = __VLS_asFunctionalComponent(__VLS_0, new __VLS_0({
    data: (__VLS_ctx.carriers),
    border: true,
}));
const __VLS_2 = __VLS_1({
    data: (__VLS_ctx.carriers),
    border: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_1));
__VLS_asFunctionalDirective(__VLS_directives.vLoading)(null, { ...__VLS_directiveBindingRestFields, value: (__VLS_ctx.loading) }, null, null);
__VLS_3.slots.default;
const __VLS_4 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_5 = __VLS_asFunctionalComponent(__VLS_4, new __VLS_4({
    prop: "name",
    label: "快递名称",
    minWidth: "120",
}));
const __VLS_6 = __VLS_5({
    prop: "name",
    label: "快递名称",
    minWidth: "120",
}, ...__VLS_functionalComponentArgsRest(__VLS_5));
const __VLS_8 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_9 = __VLS_asFunctionalComponent(__VLS_8, new __VLS_8({
    prop: "code",
    label: "快递代码",
    width: "120",
}));
const __VLS_10 = __VLS_9({
    prop: "code",
    label: "快递代码",
    width: "120",
}, ...__VLS_functionalComponentArgsRest(__VLS_9));
const __VLS_12 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_13 = __VLS_asFunctionalComponent(__VLS_12, new __VLS_12({
    label: "启用",
    width: "100",
    align: "center",
}));
const __VLS_14 = __VLS_13({
    label: "启用",
    width: "100",
    align: "center",
}, ...__VLS_functionalComponentArgsRest(__VLS_13));
__VLS_15.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_15.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_16 = {}.ElSwitch;
    /** @type {[typeof __VLS_components.ElSwitch, typeof __VLS_components.elSwitch, ]} */ ;
    // @ts-ignore
    const __VLS_17 = __VLS_asFunctionalComponent(__VLS_16, new __VLS_16({
        ...{ 'onChange': {} },
        modelValue: (row.enabled),
        loading: (!!__VLS_ctx.saving[row.code]),
    }));
    const __VLS_18 = __VLS_17({
        ...{ 'onChange': {} },
        modelValue: (row.enabled),
        loading: (!!__VLS_ctx.saving[row.code]),
    }, ...__VLS_functionalComponentArgsRest(__VLS_17));
    let __VLS_20;
    let __VLS_21;
    let __VLS_22;
    const __VLS_23 = {
        onChange: (...[$event]) => {
            __VLS_ctx.toggleEnabled(row);
        }
    };
    var __VLS_19;
}
var __VLS_15;
var __VLS_3;
/** @type {__VLS_StyleScopedClasses['page-card']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            loading: loading,
            carriers: carriers,
            saving: saving,
            toggleEnabled: toggleEnabled,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
