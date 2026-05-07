import { ref, onMounted } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import { useAuthStore } from '@/stores/auth';
import UploadImage from '@/components/UploadImage/index.vue';
import { getBanners, createBanner, updateBanner, deleteBanner, toggleBanner, sortBanners, } from '@/api/banner';
const auth = useAuthStore();
const canEdit = () => auth.isSuperAdmin || auth.perms.includes('banner.edit');
const loading = ref(false);
const banners = ref([]);
const dialogVisible = ref(false);
const isEdit = ref(false);
const saving = ref(false);
const formRef = ref();
const defaultForm = () => ({
    title: '',
    image_url: '',
    link_url: '',
    sort: 0,
});
const form = ref(defaultForm());
const previewVisible = ref(false);
const previewUrl = ref('');
async function loadData() {
    loading.value = true;
    try {
        const res = await getBanners();
        banners.value = Array.isArray(res) ? res : [];
    }
    finally {
        loading.value = false;
    }
}
function openCreate() {
    form.value = defaultForm();
    isEdit.value = false;
    dialogVisible.value = true;
}
function openEdit(row) {
    form.value = {
        id: row.id,
        title: row.title,
        image_url: row.image_url,
        link_url: row.link_url,
        sort: row.sort,
    };
    isEdit.value = true;
    dialogVisible.value = true;
}
async function handleSave() {
    await formRef.value?.validate();
    if (!form.value.image_url) {
        ElMessage.warning('请上传图片');
        return;
    }
    saving.value = true;
    try {
        const payload = {
            title: form.value.title,
            image_url: form.value.image_url,
            link_url: form.value.link_url,
            sort: form.value.sort ?? 0,
        };
        if (isEdit.value && form.value.id) {
            await updateBanner(form.value.id, payload);
        }
        else {
            await createBanner(payload);
        }
        ElMessage.success('保存成功');
        dialogVisible.value = false;
        await loadData();
    }
    finally {
        saving.value = false;
    }
}
async function handleToggle(row) {
    try {
        await toggleBanner(row.id);
        await loadData();
    }
    catch {
        // revert on error — reload will restore correct state
        await loadData();
    }
}
async function handleSortChange(row) {
    try {
        await sortBanners([{ id: row.id, sort: row.sort }]);
    }
    catch {
        ElMessage.error('排序更新失败');
        await loadData();
    }
}
async function handleDelete(row) {
    await ElMessageBox.confirm(`确认删除 Banner「${row.title || row.id}」？`, '提示', { type: 'warning' });
    await deleteBanner(row.id);
    ElMessage.success('删除成功');
    await loadData();
}
function handlePreview(url) {
    previewUrl.value = url;
    previewVisible.value = true;
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
if (__VLS_ctx.canEdit()) {
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
}
const __VLS_8 = {}.ElTable;
/** @type {[typeof __VLS_components.ElTable, typeof __VLS_components.elTable, typeof __VLS_components.ElTable, typeof __VLS_components.elTable, ]} */ ;
// @ts-ignore
const __VLS_9 = __VLS_asFunctionalComponent(__VLS_8, new __VLS_8({
    data: (__VLS_ctx.banners),
    border: true,
    ...{ style: {} },
}));
const __VLS_10 = __VLS_9({
    data: (__VLS_ctx.banners),
    border: true,
    ...{ style: {} },
}, ...__VLS_functionalComponentArgsRest(__VLS_9));
__VLS_asFunctionalDirective(__VLS_directives.vLoading)(null, { ...__VLS_directiveBindingRestFields, value: (__VLS_ctx.loading) }, null, null);
__VLS_11.slots.default;
const __VLS_12 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_13 = __VLS_asFunctionalComponent(__VLS_12, new __VLS_12({
    label: "图片预览",
    width: "100",
    align: "center",
}));
const __VLS_14 = __VLS_13({
    label: "图片预览",
    width: "100",
    align: "center",
}, ...__VLS_functionalComponentArgsRest(__VLS_13));
__VLS_15.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_15.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_16 = {}.ElImage;
    /** @type {[typeof __VLS_components.ElImage, typeof __VLS_components.elImage, ]} */ ;
    // @ts-ignore
    const __VLS_17 = __VLS_asFunctionalComponent(__VLS_16, new __VLS_16({
        ...{ 'onClick': {} },
        src: (row.image_url),
        ...{ style: {} },
        fit: "cover",
    }));
    const __VLS_18 = __VLS_17({
        ...{ 'onClick': {} },
        src: (row.image_url),
        ...{ style: {} },
        fit: "cover",
    }, ...__VLS_functionalComponentArgsRest(__VLS_17));
    let __VLS_20;
    let __VLS_21;
    let __VLS_22;
    const __VLS_23 = {
        onClick: (...[$event]) => {
            __VLS_ctx.handlePreview(row.image_url);
        }
    };
    var __VLS_19;
}
var __VLS_15;
const __VLS_24 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_25 = __VLS_asFunctionalComponent(__VLS_24, new __VLS_24({
    prop: "title",
    label: "标题",
    minWidth: "120",
    showOverflowTooltip: true,
}));
const __VLS_26 = __VLS_25({
    prop: "title",
    label: "标题",
    minWidth: "120",
    showOverflowTooltip: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_25));
const __VLS_28 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_29 = __VLS_asFunctionalComponent(__VLS_28, new __VLS_28({
    prop: "link_url",
    label: "跳转链接",
    minWidth: "200",
}));
const __VLS_30 = __VLS_29({
    prop: "link_url",
    label: "跳转链接",
    minWidth: "200",
}, ...__VLS_functionalComponentArgsRest(__VLS_29));
__VLS_31.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_31.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ style: {} },
        title: (row.link_url),
    });
    (row.link_url || '—');
}
var __VLS_31;
const __VLS_32 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_33 = __VLS_asFunctionalComponent(__VLS_32, new __VLS_32({
    label: "排序",
    width: "100",
    align: "center",
}));
const __VLS_34 = __VLS_33({
    label: "排序",
    width: "100",
    align: "center",
}, ...__VLS_functionalComponentArgsRest(__VLS_33));
__VLS_35.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_35.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_36 = {}.ElInputNumber;
    /** @type {[typeof __VLS_components.ElInputNumber, typeof __VLS_components.elInputNumber, ]} */ ;
    // @ts-ignore
    const __VLS_37 = __VLS_asFunctionalComponent(__VLS_36, new __VLS_36({
        ...{ 'onChange': {} },
        modelValue: (row.sort),
        min: (0),
        max: (9999),
        controls: (false),
        size: "small",
        ...{ style: {} },
    }));
    const __VLS_38 = __VLS_37({
        ...{ 'onChange': {} },
        modelValue: (row.sort),
        min: (0),
        max: (9999),
        controls: (false),
        size: "small",
        ...{ style: {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_37));
    let __VLS_40;
    let __VLS_41;
    let __VLS_42;
    const __VLS_43 = {
        onChange: (...[$event]) => {
            __VLS_ctx.handleSortChange(row);
        }
    };
    var __VLS_39;
}
var __VLS_35;
const __VLS_44 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_45 = __VLS_asFunctionalComponent(__VLS_44, new __VLS_44({
    label: "状态",
    width: "90",
    align: "center",
}));
const __VLS_46 = __VLS_45({
    label: "状态",
    width: "90",
    align: "center",
}, ...__VLS_functionalComponentArgsRest(__VLS_45));
__VLS_47.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_47.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_48 = {}.ElSwitch;
    /** @type {[typeof __VLS_components.ElSwitch, typeof __VLS_components.elSwitch, ]} */ ;
    // @ts-ignore
    const __VLS_49 = __VLS_asFunctionalComponent(__VLS_48, new __VLS_48({
        ...{ 'onChange': {} },
        modelValue: (row.is_active),
        disabled: (!__VLS_ctx.canEdit()),
    }));
    const __VLS_50 = __VLS_49({
        ...{ 'onChange': {} },
        modelValue: (row.is_active),
        disabled: (!__VLS_ctx.canEdit()),
    }, ...__VLS_functionalComponentArgsRest(__VLS_49));
    let __VLS_52;
    let __VLS_53;
    let __VLS_54;
    const __VLS_55 = {
        onChange: (...[$event]) => {
            __VLS_ctx.handleToggle(row);
        }
    };
    var __VLS_51;
}
var __VLS_47;
if (__VLS_ctx.canEdit()) {
    const __VLS_56 = {}.ElTableColumn;
    /** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
    // @ts-ignore
    const __VLS_57 = __VLS_asFunctionalComponent(__VLS_56, new __VLS_56({
        label: "操作",
        width: "140",
        align: "center",
    }));
    const __VLS_58 = __VLS_57({
        label: "操作",
        width: "140",
        align: "center",
    }, ...__VLS_functionalComponentArgsRest(__VLS_57));
    __VLS_59.slots.default;
    {
        const { default: __VLS_thisSlot } = __VLS_59.slots;
        const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
        const __VLS_60 = {}.ElButton;
        /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
        // @ts-ignore
        const __VLS_61 = __VLS_asFunctionalComponent(__VLS_60, new __VLS_60({
            ...{ 'onClick': {} },
            text: true,
            size: "small",
            type: "primary",
        }));
        const __VLS_62 = __VLS_61({
            ...{ 'onClick': {} },
            text: true,
            size: "small",
            type: "primary",
        }, ...__VLS_functionalComponentArgsRest(__VLS_61));
        let __VLS_64;
        let __VLS_65;
        let __VLS_66;
        const __VLS_67 = {
            onClick: (...[$event]) => {
                if (!(__VLS_ctx.canEdit()))
                    return;
                __VLS_ctx.openEdit(row);
            }
        };
        __VLS_63.slots.default;
        var __VLS_63;
        const __VLS_68 = {}.ElButton;
        /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
        // @ts-ignore
        const __VLS_69 = __VLS_asFunctionalComponent(__VLS_68, new __VLS_68({
            ...{ 'onClick': {} },
            text: true,
            size: "small",
            type: "danger",
        }));
        const __VLS_70 = __VLS_69({
            ...{ 'onClick': {} },
            text: true,
            size: "small",
            type: "danger",
        }, ...__VLS_functionalComponentArgsRest(__VLS_69));
        let __VLS_72;
        let __VLS_73;
        let __VLS_74;
        const __VLS_75 = {
            onClick: (...[$event]) => {
                if (!(__VLS_ctx.canEdit()))
                    return;
                __VLS_ctx.handleDelete(row);
            }
        };
        __VLS_71.slots.default;
        var __VLS_71;
    }
    var __VLS_59;
}
var __VLS_11;
if (!__VLS_ctx.banners.length && !__VLS_ctx.loading) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ style: {} },
    });
    const __VLS_76 = {}.ElEmpty;
    /** @type {[typeof __VLS_components.ElEmpty, typeof __VLS_components.elEmpty, ]} */ ;
    // @ts-ignore
    const __VLS_77 = __VLS_asFunctionalComponent(__VLS_76, new __VLS_76({
        description: "暂无 Banner",
    }));
    const __VLS_78 = __VLS_77({
        description: "暂无 Banner",
    }, ...__VLS_functionalComponentArgsRest(__VLS_77));
}
const __VLS_80 = {}.ElDialog;
/** @type {[typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, ]} */ ;
// @ts-ignore
const __VLS_81 = __VLS_asFunctionalComponent(__VLS_80, new __VLS_80({
    modelValue: (__VLS_ctx.dialogVisible),
    title: (__VLS_ctx.isEdit ? '编辑 Banner' : '新增 Banner'),
    width: "480px",
    destroyOnClose: true,
}));
const __VLS_82 = __VLS_81({
    modelValue: (__VLS_ctx.dialogVisible),
    title: (__VLS_ctx.isEdit ? '编辑 Banner' : '新增 Banner'),
    width: "480px",
    destroyOnClose: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_81));
__VLS_83.slots.default;
const __VLS_84 = {}.ElForm;
/** @type {[typeof __VLS_components.ElForm, typeof __VLS_components.elForm, typeof __VLS_components.ElForm, typeof __VLS_components.elForm, ]} */ ;
// @ts-ignore
const __VLS_85 = __VLS_asFunctionalComponent(__VLS_84, new __VLS_84({
    ref: "formRef",
    model: (__VLS_ctx.form),
    labelWidth: "80px",
}));
const __VLS_86 = __VLS_85({
    ref: "formRef",
    model: (__VLS_ctx.form),
    labelWidth: "80px",
}, ...__VLS_functionalComponentArgsRest(__VLS_85));
/** @type {typeof __VLS_ctx.formRef} */ ;
var __VLS_88 = {};
__VLS_87.slots.default;
const __VLS_90 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_91 = __VLS_asFunctionalComponent(__VLS_90, new __VLS_90({
    label: "图片",
    prop: "image_url",
}));
const __VLS_92 = __VLS_91({
    label: "图片",
    prop: "image_url",
}, ...__VLS_functionalComponentArgsRest(__VLS_91));
__VLS_93.slots.default;
/** @type {[typeof UploadImage, ]} */ ;
// @ts-ignore
const __VLS_94 = __VLS_asFunctionalComponent(UploadImage, new UploadImage({
    modelValue: (__VLS_ctx.form.image_url),
}));
const __VLS_95 = __VLS_94({
    modelValue: (__VLS_ctx.form.image_url),
}, ...__VLS_functionalComponentArgsRest(__VLS_94));
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ style: {} },
});
var __VLS_93;
const __VLS_97 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_98 = __VLS_asFunctionalComponent(__VLS_97, new __VLS_97({
    label: "标题",
    prop: "title",
}));
const __VLS_99 = __VLS_98({
    label: "标题",
    prop: "title",
}, ...__VLS_functionalComponentArgsRest(__VLS_98));
__VLS_100.slots.default;
const __VLS_101 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_102 = __VLS_asFunctionalComponent(__VLS_101, new __VLS_101({
    modelValue: (__VLS_ctx.form.title),
    placeholder: "可选",
    maxlength: "128",
    showWordLimit: true,
}));
const __VLS_103 = __VLS_102({
    modelValue: (__VLS_ctx.form.title),
    placeholder: "可选",
    maxlength: "128",
    showWordLimit: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_102));
var __VLS_100;
const __VLS_105 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_106 = __VLS_asFunctionalComponent(__VLS_105, new __VLS_105({
    label: "跳转链接",
    prop: "link_url",
}));
const __VLS_107 = __VLS_106({
    label: "跳转链接",
    prop: "link_url",
}, ...__VLS_functionalComponentArgsRest(__VLS_106));
__VLS_108.slots.default;
const __VLS_109 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_110 = __VLS_asFunctionalComponent(__VLS_109, new __VLS_109({
    modelValue: (__VLS_ctx.form.link_url),
    placeholder: "可选，如 /pages/product/detail?id=1",
}));
const __VLS_111 = __VLS_110({
    modelValue: (__VLS_ctx.form.link_url),
    placeholder: "可选，如 /pages/product/detail?id=1",
}, ...__VLS_functionalComponentArgsRest(__VLS_110));
var __VLS_108;
const __VLS_113 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_114 = __VLS_asFunctionalComponent(__VLS_113, new __VLS_113({
    label: "排序",
    prop: "sort",
}));
const __VLS_115 = __VLS_114({
    label: "排序",
    prop: "sort",
}, ...__VLS_functionalComponentArgsRest(__VLS_114));
__VLS_116.slots.default;
const __VLS_117 = {}.ElInputNumber;
/** @type {[typeof __VLS_components.ElInputNumber, typeof __VLS_components.elInputNumber, ]} */ ;
// @ts-ignore
const __VLS_118 = __VLS_asFunctionalComponent(__VLS_117, new __VLS_117({
    modelValue: (__VLS_ctx.form.sort),
    min: (0),
    max: (9999),
}));
const __VLS_119 = __VLS_118({
    modelValue: (__VLS_ctx.form.sort),
    min: (0),
    max: (9999),
}, ...__VLS_functionalComponentArgsRest(__VLS_118));
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
    ...{ style: {} },
});
var __VLS_116;
var __VLS_87;
{
    const { footer: __VLS_thisSlot } = __VLS_83.slots;
    const __VLS_121 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_122 = __VLS_asFunctionalComponent(__VLS_121, new __VLS_121({
        ...{ 'onClick': {} },
    }));
    const __VLS_123 = __VLS_122({
        ...{ 'onClick': {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_122));
    let __VLS_125;
    let __VLS_126;
    let __VLS_127;
    const __VLS_128 = {
        onClick: (...[$event]) => {
            __VLS_ctx.dialogVisible = false;
        }
    };
    __VLS_124.slots.default;
    var __VLS_124;
    const __VLS_129 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_130 = __VLS_asFunctionalComponent(__VLS_129, new __VLS_129({
        ...{ 'onClick': {} },
        type: "primary",
        loading: (__VLS_ctx.saving),
    }));
    const __VLS_131 = __VLS_130({
        ...{ 'onClick': {} },
        type: "primary",
        loading: (__VLS_ctx.saving),
    }, ...__VLS_functionalComponentArgsRest(__VLS_130));
    let __VLS_133;
    let __VLS_134;
    let __VLS_135;
    const __VLS_136 = {
        onClick: (__VLS_ctx.handleSave)
    };
    __VLS_132.slots.default;
    var __VLS_132;
}
var __VLS_83;
const __VLS_137 = {}.ElDialog;
/** @type {[typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, ]} */ ;
// @ts-ignore
const __VLS_138 = __VLS_asFunctionalComponent(__VLS_137, new __VLS_137({
    modelValue: (__VLS_ctx.previewVisible),
    title: "图片预览",
    width: "600px",
}));
const __VLS_139 = __VLS_138({
    modelValue: (__VLS_ctx.previewVisible),
    title: "图片预览",
    width: "600px",
}, ...__VLS_functionalComponentArgsRest(__VLS_138));
__VLS_140.slots.default;
__VLS_asFunctionalElement(__VLS_intrinsicElements.img)({
    src: (__VLS_ctx.previewUrl),
    ...{ style: {} },
});
var __VLS_140;
/** @type {__VLS_StyleScopedClasses['page-card']} */ ;
// @ts-ignore
var __VLS_89 = __VLS_88;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            UploadImage: UploadImage,
            canEdit: canEdit,
            loading: loading,
            banners: banners,
            dialogVisible: dialogVisible,
            isEdit: isEdit,
            saving: saving,
            formRef: formRef,
            form: form,
            previewVisible: previewVisible,
            previewUrl: previewUrl,
            openCreate: openCreate,
            openEdit: openEdit,
            handleSave: handleSave,
            handleToggle: handleToggle,
            handleSortChange: handleSortChange,
            handleDelete: handleDelete,
            handlePreview: handlePreview,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
