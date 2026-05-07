import { computed } from 'vue';
const props = defineProps();
const emit = defineEmits();
const yuan = computed({
    get: () => (props.modelValue / 100).toFixed(2),
    set: (val) => {
        const num = parseFloat(val);
        emit('update:modelValue', isNaN(num) ? 0 : Math.round(num * 100));
    },
});
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
const __VLS_0 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_1 = __VLS_asFunctionalComponent(__VLS_0, new __VLS_0({
    modelValue: (__VLS_ctx.yuan),
    placeholder: (__VLS_ctx.placeholder || '请输入金额'),
    disabled: (__VLS_ctx.disabled),
    type: "number",
    step: "0.01",
    min: "0",
}));
const __VLS_2 = __VLS_1({
    modelValue: (__VLS_ctx.yuan),
    placeholder: (__VLS_ctx.placeholder || '请输入金额'),
    disabled: (__VLS_ctx.disabled),
    type: "number",
    step: "0.01",
    min: "0",
}, ...__VLS_functionalComponentArgsRest(__VLS_1));
var __VLS_4 = {};
__VLS_3.slots.default;
{
    const { prefix: __VLS_thisSlot } = __VLS_3.slots;
}
{
    const { suffix: __VLS_thisSlot } = __VLS_3.slots;
}
var __VLS_3;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            yuan: yuan,
        };
    },
    __typeEmits: {},
    __typeProps: {},
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
    __typeEmits: {},
    __typeProps: {},
});
; /* PartiallyEnd: #4569/main.vue */
