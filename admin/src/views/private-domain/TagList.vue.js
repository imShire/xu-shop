import { ref, onMounted } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import { getTagList, createTag, updateTag, deleteTag } from '@/api/private-domain';
const loading = ref(false);
const tags = ref([]);
const dialogVisible = ref(false);
const isEdit = ref(false);
const saving = ref(false);
const formRef = ref();
const form = ref({ id: '', name: '' });
const colors = ['#f59e0b', '#3b82f6', '#22c55e', '#ef4444', '#8b5cf6', '#ec4899', '#06b6d4'];
async function loadData() {
    loading.value = true;
    try {
        const res = await getTagList();
        tags.value = Array.isArray(res) ? res : [];
    }
    finally {
        loading.value = false;
    }
}
function openCreate() {
    form.value = { id: '', name: '' };
    isEdit.value = false;
    dialogVisible.value = true;
}
function openEdit(row) {
    form.value = { id: row.id, name: row.name };
    isEdit.value = true;
    dialogVisible.value = true;
}
async function handleSave() {
    await formRef.value?.validate();
    saving.value = true;
    try {
        if (isEdit.value) {
            await updateTag(form.value.id, form.value);
        }
        else {
            await createTag(form.value);
        }
        ElMessage.success('保存成功');
        dialogVisible.value = false;
        await loadData();
    }
    finally {
        saving.value = false;
    }
}
function getTagColor(id) {
    const seed = Number(id) || 0;
    return colors[Math.abs(seed) % colors.length];
}
async function handleDelete(row) {
    await ElMessageBox.confirm(`确认删除标签「${row.name}」？`, '提示', { type: 'warning' });
    await deleteTag(row.id);
    ElMessage.success('删除成功');
    await loadData();
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
const __VLS_0 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_1 = __VLS_asFunctionalComponent(__VLS_0, new __VLS_0({
    ...{ 'onClick': {} },
    type: "primary",
}));
const __VLS_2 = __VLS_1({
    ...{ 'onClick': {} },
    type: "primary",
}, ...__VLS_functionalComponentArgsRest(__VLS_1));
let __VLS_4;
let __VLS_5;
let __VLS_6;
const __VLS_7 = {
    onClick: (__VLS_ctx.openCreate)
};
__VLS_3.slots.default;
var __VLS_3;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ style: {} },
});
__VLS_asFunctionalDirective(__VLS_directives.vLoading)(null, { ...__VLS_directiveBindingRestFields, value: (__VLS_ctx.loading) }, null, null);
for (const [tag] of __VLS_getVForSourceType((__VLS_ctx.tags))) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        key: (tag.id),
        ...{ style: {} },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div)({
        ...{ style: ({ width: '10px', height: '10px', borderRadius: '50%', background: __VLS_ctx.getTagColor(tag.id), flexShrink: 0 }) },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ style: {} },
    });
    (tag.name);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ style: {} },
    });
    (tag.source);
    const __VLS_8 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_9 = __VLS_asFunctionalComponent(__VLS_8, new __VLS_8({
        ...{ 'onClick': {} },
        text: true,
        size: "small",
        type: "primary",
    }));
    const __VLS_10 = __VLS_9({
        ...{ 'onClick': {} },
        text: true,
        size: "small",
        type: "primary",
    }, ...__VLS_functionalComponentArgsRest(__VLS_9));
    let __VLS_12;
    let __VLS_13;
    let __VLS_14;
    const __VLS_15 = {
        onClick: (...[$event]) => {
            __VLS_ctx.openEdit(tag);
        }
    };
    __VLS_11.slots.default;
    var __VLS_11;
    const __VLS_16 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_17 = __VLS_asFunctionalComponent(__VLS_16, new __VLS_16({
        ...{ 'onClick': {} },
        text: true,
        size: "small",
        type: "danger",
    }));
    const __VLS_18 = __VLS_17({
        ...{ 'onClick': {} },
        text: true,
        size: "small",
        type: "danger",
    }, ...__VLS_functionalComponentArgsRest(__VLS_17));
    let __VLS_20;
    let __VLS_21;
    let __VLS_22;
    const __VLS_23 = {
        onClick: (...[$event]) => {
            __VLS_ctx.handleDelete(tag);
        }
    };
    __VLS_19.slots.default;
    var __VLS_19;
}
if (!__VLS_ctx.tags.length && !__VLS_ctx.loading) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ style: {} },
    });
    const __VLS_24 = {}.ElEmpty;
    /** @type {[typeof __VLS_components.ElEmpty, typeof __VLS_components.elEmpty, ]} */ ;
    // @ts-ignore
    const __VLS_25 = __VLS_asFunctionalComponent(__VLS_24, new __VLS_24({
        description: "暂无标签",
    }));
    const __VLS_26 = __VLS_25({
        description: "暂无标签",
    }, ...__VLS_functionalComponentArgsRest(__VLS_25));
}
const __VLS_28 = {}.ElDialog;
/** @type {[typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, ]} */ ;
// @ts-ignore
const __VLS_29 = __VLS_asFunctionalComponent(__VLS_28, new __VLS_28({
    modelValue: (__VLS_ctx.dialogVisible),
    title: (__VLS_ctx.isEdit ? '编辑标签' : '新建标签'),
    width: "380px",
}));
const __VLS_30 = __VLS_29({
    modelValue: (__VLS_ctx.dialogVisible),
    title: (__VLS_ctx.isEdit ? '编辑标签' : '新建标签'),
    width: "380px",
}, ...__VLS_functionalComponentArgsRest(__VLS_29));
__VLS_31.slots.default;
const __VLS_32 = {}.ElForm;
/** @type {[typeof __VLS_components.ElForm, typeof __VLS_components.elForm, typeof __VLS_components.ElForm, typeof __VLS_components.elForm, ]} */ ;
// @ts-ignore
const __VLS_33 = __VLS_asFunctionalComponent(__VLS_32, new __VLS_32({
    ref: "formRef",
    model: (__VLS_ctx.form),
    labelWidth: "70px",
}));
const __VLS_34 = __VLS_33({
    ref: "formRef",
    model: (__VLS_ctx.form),
    labelWidth: "70px",
}, ...__VLS_functionalComponentArgsRest(__VLS_33));
/** @type {typeof __VLS_ctx.formRef} */ ;
var __VLS_36 = {};
__VLS_35.slots.default;
const __VLS_38 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_39 = __VLS_asFunctionalComponent(__VLS_38, new __VLS_38({
    label: "标签名",
    prop: "name",
    rules: ([{ required: true }]),
}));
const __VLS_40 = __VLS_39({
    label: "标签名",
    prop: "name",
    rules: ([{ required: true }]),
}, ...__VLS_functionalComponentArgsRest(__VLS_39));
__VLS_41.slots.default;
const __VLS_42 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_43 = __VLS_asFunctionalComponent(__VLS_42, new __VLS_42({
    modelValue: (__VLS_ctx.form.name),
}));
const __VLS_44 = __VLS_43({
    modelValue: (__VLS_ctx.form.name),
}, ...__VLS_functionalComponentArgsRest(__VLS_43));
var __VLS_41;
var __VLS_35;
{
    const { footer: __VLS_thisSlot } = __VLS_31.slots;
    const __VLS_46 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_47 = __VLS_asFunctionalComponent(__VLS_46, new __VLS_46({
        ...{ 'onClick': {} },
    }));
    const __VLS_48 = __VLS_47({
        ...{ 'onClick': {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_47));
    let __VLS_50;
    let __VLS_51;
    let __VLS_52;
    const __VLS_53 = {
        onClick: (...[$event]) => {
            __VLS_ctx.dialogVisible = false;
        }
    };
    __VLS_49.slots.default;
    var __VLS_49;
    const __VLS_54 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_55 = __VLS_asFunctionalComponent(__VLS_54, new __VLS_54({
        ...{ 'onClick': {} },
        type: "primary",
        loading: (__VLS_ctx.saving),
    }));
    const __VLS_56 = __VLS_55({
        ...{ 'onClick': {} },
        type: "primary",
        loading: (__VLS_ctx.saving),
    }, ...__VLS_functionalComponentArgsRest(__VLS_55));
    let __VLS_58;
    let __VLS_59;
    let __VLS_60;
    const __VLS_61 = {
        onClick: (__VLS_ctx.handleSave)
    };
    __VLS_57.slots.default;
    var __VLS_57;
}
var __VLS_31;
/** @type {__VLS_StyleScopedClasses['page-card']} */ ;
// @ts-ignore
var __VLS_37 = __VLS_36;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            loading: loading,
            tags: tags,
            dialogVisible: dialogVisible,
            isEdit: isEdit,
            saving: saving,
            formRef: formRef,
            form: form,
            openCreate: openCreate,
            openEdit: openEdit,
            handleSave: handleSave,
            getTagColor: getTagColor,
            handleDelete: handleDelete,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
