import { ref } from 'vue';
import { ElMessage } from 'element-plus';
import { Plus, Delete } from '@element-plus/icons-vue';
import { uploadImage } from '@/api/product';
const props = withDefaults(defineProps(), {
    multiple: false,
    limit: 9,
    accept: 'image/*',
});
const emit = defineEmits();
const uploading = ref(false);
const previewVisible = ref(false);
const previewUrl = ref('');
function getList() {
    if (!props.modelValue)
        return [];
    if (Array.isArray(props.modelValue))
        return [...props.modelValue];
    return props.modelValue ? [props.modelValue] : [];
}
async function handleChange(e) {
    const input = e.target;
    if (!input.files?.length)
        return;
    const files = Array.from(input.files);
    uploading.value = true;
    try {
        const urls = [];
        for (const file of files) {
            const res = await uploadImage(file);
            urls.push(res.url);
        }
        if (props.multiple) {
            const list = [...getList(), ...urls].slice(0, props.limit);
            emit('update:modelValue', list);
        }
        else {
            emit('update:modelValue', urls[0]);
        }
    }
    catch {
        ElMessage.error('上传失败');
    }
    finally {
        uploading.value = false;
        input.value = '';
    }
}
function removeImage(idx) {
    if (props.multiple) {
        const list = getList();
        list.splice(idx, 1);
        emit('update:modelValue', list);
    }
    else {
        emit('update:modelValue', '');
    }
}
function preview(url) {
    previewUrl.value = url;
    previewVisible.value = true;
}
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_withDefaultsArg = (function (t) { return t; })({
    multiple: false,
    limit: 9,
    accept: 'image/*',
});
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
/** @type {__VLS_StyleScopedClasses['image-actions']} */ ;
// CSS variable injection 
// CSS variable injection end 
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "upload-image" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "image-list" },
});
for (const [url, idx] of __VLS_getVForSourceType((__VLS_ctx.getList()))) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        key: (idx),
        ...{ class: "image-item" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.img)({
        ...{ onClick: (...[$event]) => {
                __VLS_ctx.preview(url);
            } },
        src: (url),
        ...{ class: "preview-img" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "image-actions" },
    });
    const __VLS_0 = {}.ElIcon;
    /** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
    // @ts-ignore
    const __VLS_1 = __VLS_asFunctionalComponent(__VLS_0, new __VLS_0({
        ...{ 'onClick': {} },
        ...{ class: "action-icon" },
    }));
    const __VLS_2 = __VLS_1({
        ...{ 'onClick': {} },
        ...{ class: "action-icon" },
    }, ...__VLS_functionalComponentArgsRest(__VLS_1));
    let __VLS_4;
    let __VLS_5;
    let __VLS_6;
    const __VLS_7 = {
        onClick: (...[$event]) => {
            __VLS_ctx.preview(url);
        }
    };
    __VLS_3.slots.default;
    const __VLS_8 = {}.ZoomIn;
    /** @type {[typeof __VLS_components.ZoomIn, ]} */ ;
    // @ts-ignore
    const __VLS_9 = __VLS_asFunctionalComponent(__VLS_8, new __VLS_8({}));
    const __VLS_10 = __VLS_9({}, ...__VLS_functionalComponentArgsRest(__VLS_9));
    var __VLS_3;
    const __VLS_12 = {}.ElIcon;
    /** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
    // @ts-ignore
    const __VLS_13 = __VLS_asFunctionalComponent(__VLS_12, new __VLS_12({
        ...{ 'onClick': {} },
        ...{ class: "action-icon" },
    }));
    const __VLS_14 = __VLS_13({
        ...{ 'onClick': {} },
        ...{ class: "action-icon" },
    }, ...__VLS_functionalComponentArgsRest(__VLS_13));
    let __VLS_16;
    let __VLS_17;
    let __VLS_18;
    const __VLS_19 = {
        onClick: (...[$event]) => {
            __VLS_ctx.removeImage(idx);
        }
    };
    __VLS_15.slots.default;
    const __VLS_20 = {}.Delete;
    /** @type {[typeof __VLS_components.Delete, ]} */ ;
    // @ts-ignore
    const __VLS_21 = __VLS_asFunctionalComponent(__VLS_20, new __VLS_20({}));
    const __VLS_22 = __VLS_21({}, ...__VLS_functionalComponentArgsRest(__VLS_21));
    var __VLS_15;
}
if (__VLS_ctx.multiple ? __VLS_ctx.getList().length < __VLS_ctx.limit : !__VLS_ctx.modelValue) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.label, __VLS_intrinsicElements.label)({
        ...{ class: "upload-trigger" },
        ...{ class: ({ uploading: __VLS_ctx.uploading }) },
    });
    if (!__VLS_ctx.uploading) {
        const __VLS_24 = {}.ElIcon;
        /** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
        // @ts-ignore
        const __VLS_25 = __VLS_asFunctionalComponent(__VLS_24, new __VLS_24({
            size: "24",
        }));
        const __VLS_26 = __VLS_25({
            size: "24",
        }, ...__VLS_functionalComponentArgsRest(__VLS_25));
        __VLS_27.slots.default;
        const __VLS_28 = {}.Plus;
        /** @type {[typeof __VLS_components.Plus, ]} */ ;
        // @ts-ignore
        const __VLS_29 = __VLS_asFunctionalComponent(__VLS_28, new __VLS_28({}));
        const __VLS_30 = __VLS_29({}, ...__VLS_functionalComponentArgsRest(__VLS_29));
        var __VLS_27;
    }
    else {
        const __VLS_32 = {}.ElIcon;
        /** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
        // @ts-ignore
        const __VLS_33 = __VLS_asFunctionalComponent(__VLS_32, new __VLS_32({
            ...{ class: "rotating" },
            size: "24",
        }));
        const __VLS_34 = __VLS_33({
            ...{ class: "rotating" },
            size: "24",
        }, ...__VLS_functionalComponentArgsRest(__VLS_33));
        __VLS_35.slots.default;
        const __VLS_36 = {}.Loading;
        /** @type {[typeof __VLS_components.Loading, ]} */ ;
        // @ts-ignore
        const __VLS_37 = __VLS_asFunctionalComponent(__VLS_36, new __VLS_36({}));
        const __VLS_38 = __VLS_37({}, ...__VLS_functionalComponentArgsRest(__VLS_37));
        var __VLS_35;
    }
    __VLS_asFunctionalElement(__VLS_intrinsicElements.input)({
        ...{ onChange: (__VLS_ctx.handleChange) },
        type: "file",
        accept: (__VLS_ctx.accept),
        multiple: (__VLS_ctx.multiple),
        ...{ style: {} },
    });
}
const __VLS_40 = {}.ElDialog;
/** @type {[typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, ]} */ ;
// @ts-ignore
const __VLS_41 = __VLS_asFunctionalComponent(__VLS_40, new __VLS_40({
    modelValue: (__VLS_ctx.previewVisible),
    title: "图片预览",
    width: "600px",
}));
const __VLS_42 = __VLS_41({
    modelValue: (__VLS_ctx.previewVisible),
    title: "图片预览",
    width: "600px",
}, ...__VLS_functionalComponentArgsRest(__VLS_41));
__VLS_43.slots.default;
__VLS_asFunctionalElement(__VLS_intrinsicElements.img)({
    src: (__VLS_ctx.previewUrl),
    ...{ style: {} },
});
var __VLS_43;
/** @type {__VLS_StyleScopedClasses['upload-image']} */ ;
/** @type {__VLS_StyleScopedClasses['image-list']} */ ;
/** @type {__VLS_StyleScopedClasses['image-item']} */ ;
/** @type {__VLS_StyleScopedClasses['preview-img']} */ ;
/** @type {__VLS_StyleScopedClasses['image-actions']} */ ;
/** @type {__VLS_StyleScopedClasses['action-icon']} */ ;
/** @type {__VLS_StyleScopedClasses['action-icon']} */ ;
/** @type {__VLS_StyleScopedClasses['upload-trigger']} */ ;
/** @type {__VLS_StyleScopedClasses['rotating']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            Plus: Plus,
            Delete: Delete,
            uploading: uploading,
            previewVisible: previewVisible,
            previewUrl: previewUrl,
            getList: getList,
            handleChange: handleChange,
            removeImage: removeImage,
            preview: preview,
        };
    },
    __typeEmits: {},
    __typeProps: {},
    props: {},
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
    __typeEmits: {},
    __typeProps: {},
    props: {},
});
; /* PartiallyEnd: #4569/main.vue */
