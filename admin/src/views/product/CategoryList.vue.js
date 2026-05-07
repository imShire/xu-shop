import { ref, onMounted, computed } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import { getCategoryList, createCategory, updateCategory, deleteCategory } from '@/api/product';
import UploadImage from '@/components/UploadImage/index.vue';
const loading = ref(false);
const categories = ref([]);
const dialogVisible = ref(false);
const dialogTitle = ref('新建分类');
const formRef = ref();
const saving = ref(false);
function createDefaultForm() {
    return {
        id: '',
        name: '',
        parent_id: undefined,
        sort: 0,
        icon: '',
        status: 'enabled',
    };
}
function normalizeParentId(value) {
    if (typeof value === 'string' && value !== '' && value !== '0')
        return value;
    return '0';
}
const form = ref(createDefaultForm());
async function loadData() {
    loading.value = true;
    try {
        categories.value = await getCategoryList();
    }
    finally {
        loading.value = false;
    }
}
function openCreate() {
    form.value = createDefaultForm();
    dialogTitle.value = '新建分类';
    dialogVisible.value = true;
}
function openEdit(row) {
    const parentId = normalizeParentId(row.parent_id);
    form.value = {
        id: row.id,
        name: row.name,
        parent_id: parentId !== '0' ? parentId : undefined,
        sort: row.sort || 0,
        icon: row.icon || '',
        status: row.status || 'enabled',
    };
    dialogTitle.value = '编辑分类';
    dialogVisible.value = true;
}
async function handleSave() {
    await formRef.value?.validate();
    saving.value = true;
    try {
        const payload = {
            ...form.value,
            parent_id: normalizeParentId(form.value.parent_id),
        };
        if (form.value.id) {
            await updateCategory(form.value.id, payload);
        }
        else {
            await createCategory(payload);
        }
        ElMessage.success('保存成功');
        dialogVisible.value = false;
        await loadData();
    }
    finally {
        saving.value = false;
    }
}
async function handleDelete(row) {
    await ElMessageBox.confirm(`确认删除分类「${row.name}」？`, '提示', { type: 'warning' });
    await deleteCategory(row.id);
    ElMessage.success('删除成功');
    await loadData();
}
onMounted(loadData);
// 只有顶级分类（parent_id 为 '0' / null / undefined）才能被选为父级
const rootCategories = computed(() => categories.value.filter((c) => !c.parent_id || c.parent_id === '0'));
// id → name 映射，用于列表中显示父级名称
const categoryNameMap = computed(() => {
    const map = {};
    categories.value.forEach((c) => {
        map[c.id] = c.name;
    });
    return map;
});
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
const __VLS_8 = {}.ElTable;
/** @type {[typeof __VLS_components.ElTable, typeof __VLS_components.elTable, typeof __VLS_components.ElTable, typeof __VLS_components.elTable, ]} */ ;
// @ts-ignore
const __VLS_9 = __VLS_asFunctionalComponent(__VLS_8, new __VLS_8({
    data: (__VLS_ctx.categories),
    rowKey: "id",
    treeProps: ({ children: 'children', hasChildren: 'hasChildren' }),
    border: true,
}));
const __VLS_10 = __VLS_9({
    data: (__VLS_ctx.categories),
    rowKey: "id",
    treeProps: ({ children: 'children', hasChildren: 'hasChildren' }),
    border: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_9));
__VLS_asFunctionalDirective(__VLS_directives.vLoading)(null, { ...__VLS_directiveBindingRestFields, value: (__VLS_ctx.loading) }, null, null);
__VLS_11.slots.default;
const __VLS_12 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_13 = __VLS_asFunctionalComponent(__VLS_12, new __VLS_12({
    prop: "id",
    label: "ID",
    width: "60",
}));
const __VLS_14 = __VLS_13({
    prop: "id",
    label: "ID",
    width: "60",
}, ...__VLS_functionalComponentArgsRest(__VLS_13));
const __VLS_16 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_17 = __VLS_asFunctionalComponent(__VLS_16, new __VLS_16({
    label: "图标",
    width: "70",
    align: "center",
}));
const __VLS_18 = __VLS_17({
    label: "图标",
    width: "70",
    align: "center",
}, ...__VLS_functionalComponentArgsRest(__VLS_17));
__VLS_19.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_19.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    if (row.icon) {
        const __VLS_20 = {}.ElImage;
        /** @type {[typeof __VLS_components.ElImage, typeof __VLS_components.elImage, ]} */ ;
        // @ts-ignore
        const __VLS_21 = __VLS_asFunctionalComponent(__VLS_20, new __VLS_20({
            src: (row.icon),
            ...{ style: {} },
            fit: "cover",
        }));
        const __VLS_22 = __VLS_21({
            src: (row.icon),
            ...{ style: {} },
            fit: "cover",
        }, ...__VLS_functionalComponentArgsRest(__VLS_21));
    }
    else {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
            ...{ style: {} },
        });
    }
}
var __VLS_19;
const __VLS_24 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_25 = __VLS_asFunctionalComponent(__VLS_24, new __VLS_24({
    prop: "name",
    label: "分类名称",
    minWidth: "150",
}));
const __VLS_26 = __VLS_25({
    prop: "name",
    label: "分类名称",
    minWidth: "150",
}, ...__VLS_functionalComponentArgsRest(__VLS_25));
const __VLS_28 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_29 = __VLS_asFunctionalComponent(__VLS_28, new __VLS_28({
    label: "父级分类",
    width: "140",
}));
const __VLS_30 = __VLS_29({
    label: "父级分类",
    width: "140",
}, ...__VLS_functionalComponentArgsRest(__VLS_29));
__VLS_31.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_31.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    if (row.parent_id && row.parent_id !== '0' && __VLS_ctx.categoryNameMap[row.parent_id]) {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
        (__VLS_ctx.categoryNameMap[row.parent_id]);
    }
    else {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
            ...{ style: {} },
        });
    }
}
var __VLS_31;
const __VLS_32 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_33 = __VLS_asFunctionalComponent(__VLS_32, new __VLS_32({
    prop: "sort",
    label: "排序",
    width: "80",
    align: "center",
}));
const __VLS_34 = __VLS_33({
    prop: "sort",
    label: "排序",
    width: "80",
    align: "center",
}, ...__VLS_functionalComponentArgsRest(__VLS_33));
const __VLS_36 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_37 = __VLS_asFunctionalComponent(__VLS_36, new __VLS_36({
    label: "操作",
    width: "140",
    fixed: "right",
}));
const __VLS_38 = __VLS_37({
    label: "操作",
    width: "140",
    fixed: "right",
}, ...__VLS_functionalComponentArgsRest(__VLS_37));
__VLS_39.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_39.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_40 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_41 = __VLS_asFunctionalComponent(__VLS_40, new __VLS_40({
        ...{ 'onClick': {} },
        text: true,
        type: "primary",
        size: "small",
    }));
    const __VLS_42 = __VLS_41({
        ...{ 'onClick': {} },
        text: true,
        type: "primary",
        size: "small",
    }, ...__VLS_functionalComponentArgsRest(__VLS_41));
    let __VLS_44;
    let __VLS_45;
    let __VLS_46;
    const __VLS_47 = {
        onClick: (...[$event]) => {
            __VLS_ctx.openEdit(row);
        }
    };
    __VLS_43.slots.default;
    var __VLS_43;
    const __VLS_48 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_49 = __VLS_asFunctionalComponent(__VLS_48, new __VLS_48({
        ...{ 'onClick': {} },
        text: true,
        type: "danger",
        size: "small",
    }));
    const __VLS_50 = __VLS_49({
        ...{ 'onClick': {} },
        text: true,
        type: "danger",
        size: "small",
    }, ...__VLS_functionalComponentArgsRest(__VLS_49));
    let __VLS_52;
    let __VLS_53;
    let __VLS_54;
    const __VLS_55 = {
        onClick: (...[$event]) => {
            __VLS_ctx.handleDelete(row);
        }
    };
    __VLS_51.slots.default;
    var __VLS_51;
}
var __VLS_39;
var __VLS_11;
const __VLS_56 = {}.ElDialog;
/** @type {[typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, ]} */ ;
// @ts-ignore
const __VLS_57 = __VLS_asFunctionalComponent(__VLS_56, new __VLS_56({
    modelValue: (__VLS_ctx.dialogVisible),
    title: (__VLS_ctx.dialogTitle),
    width: "420px",
    destroyOnClose: true,
}));
const __VLS_58 = __VLS_57({
    modelValue: (__VLS_ctx.dialogVisible),
    title: (__VLS_ctx.dialogTitle),
    width: "420px",
    destroyOnClose: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_57));
__VLS_59.slots.default;
const __VLS_60 = {}.ElForm;
/** @type {[typeof __VLS_components.ElForm, typeof __VLS_components.elForm, typeof __VLS_components.ElForm, typeof __VLS_components.elForm, ]} */ ;
// @ts-ignore
const __VLS_61 = __VLS_asFunctionalComponent(__VLS_60, new __VLS_60({
    ref: "formRef",
    model: (__VLS_ctx.form),
    labelWidth: "80px",
}));
const __VLS_62 = __VLS_61({
    ref: "formRef",
    model: (__VLS_ctx.form),
    labelWidth: "80px",
}, ...__VLS_functionalComponentArgsRest(__VLS_61));
/** @type {typeof __VLS_ctx.formRef} */ ;
var __VLS_64 = {};
__VLS_63.slots.default;
const __VLS_66 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_67 = __VLS_asFunctionalComponent(__VLS_66, new __VLS_66({
    label: "分类名",
    prop: "name",
    rules: ([{ required: true }]),
}));
const __VLS_68 = __VLS_67({
    label: "分类名",
    prop: "name",
    rules: ([{ required: true }]),
}, ...__VLS_functionalComponentArgsRest(__VLS_67));
__VLS_69.slots.default;
const __VLS_70 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_71 = __VLS_asFunctionalComponent(__VLS_70, new __VLS_70({
    modelValue: (__VLS_ctx.form.name),
}));
const __VLS_72 = __VLS_71({
    modelValue: (__VLS_ctx.form.name),
}, ...__VLS_functionalComponentArgsRest(__VLS_71));
var __VLS_69;
const __VLS_74 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_75 = __VLS_asFunctionalComponent(__VLS_74, new __VLS_74({
    label: "图标",
}));
const __VLS_76 = __VLS_75({
    label: "图标",
}, ...__VLS_functionalComponentArgsRest(__VLS_75));
__VLS_77.slots.default;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({});
/** @type {[typeof UploadImage, ]} */ ;
// @ts-ignore
const __VLS_78 = __VLS_asFunctionalComponent(UploadImage, new UploadImage({
    modelValue: (__VLS_ctx.form.icon),
}));
const __VLS_79 = __VLS_78({
    modelValue: (__VLS_ctx.form.icon),
}, ...__VLS_functionalComponentArgsRest(__VLS_78));
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ style: {} },
});
var __VLS_77;
const __VLS_81 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_82 = __VLS_asFunctionalComponent(__VLS_81, new __VLS_81({
    label: "父分类",
}));
const __VLS_83 = __VLS_82({
    label: "父分类",
}, ...__VLS_functionalComponentArgsRest(__VLS_82));
__VLS_84.slots.default;
const __VLS_85 = {}.ElSelect;
/** @type {[typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, ]} */ ;
// @ts-ignore
const __VLS_86 = __VLS_asFunctionalComponent(__VLS_85, new __VLS_85({
    modelValue: (__VLS_ctx.form.parent_id),
    clearable: true,
    placeholder: "顶级分类",
}));
const __VLS_87 = __VLS_86({
    modelValue: (__VLS_ctx.form.parent_id),
    clearable: true,
    placeholder: "顶级分类",
}, ...__VLS_functionalComponentArgsRest(__VLS_86));
__VLS_88.slots.default;
for (const [cat] of __VLS_getVForSourceType((__VLS_ctx.rootCategories))) {
    const __VLS_89 = {}.ElOption;
    /** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
    // @ts-ignore
    const __VLS_90 = __VLS_asFunctionalComponent(__VLS_89, new __VLS_89({
        key: (cat.id),
        label: (cat.name),
        value: (cat.id),
    }));
    const __VLS_91 = __VLS_90({
        key: (cat.id),
        label: (cat.name),
        value: (cat.id),
    }, ...__VLS_functionalComponentArgsRest(__VLS_90));
}
var __VLS_88;
var __VLS_84;
const __VLS_93 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_94 = __VLS_asFunctionalComponent(__VLS_93, new __VLS_93({
    label: "排序",
}));
const __VLS_95 = __VLS_94({
    label: "排序",
}, ...__VLS_functionalComponentArgsRest(__VLS_94));
__VLS_96.slots.default;
const __VLS_97 = {}.ElInputNumber;
/** @type {[typeof __VLS_components.ElInputNumber, typeof __VLS_components.elInputNumber, ]} */ ;
// @ts-ignore
const __VLS_98 = __VLS_asFunctionalComponent(__VLS_97, new __VLS_97({
    modelValue: (__VLS_ctx.form.sort),
    min: (0),
}));
const __VLS_99 = __VLS_98({
    modelValue: (__VLS_ctx.form.sort),
    min: (0),
}, ...__VLS_functionalComponentArgsRest(__VLS_98));
var __VLS_96;
const __VLS_101 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_102 = __VLS_asFunctionalComponent(__VLS_101, new __VLS_101({
    label: "状态",
}));
const __VLS_103 = __VLS_102({
    label: "状态",
}, ...__VLS_functionalComponentArgsRest(__VLS_102));
__VLS_104.slots.default;
const __VLS_105 = {}.ElRadioGroup;
/** @type {[typeof __VLS_components.ElRadioGroup, typeof __VLS_components.elRadioGroup, typeof __VLS_components.ElRadioGroup, typeof __VLS_components.elRadioGroup, ]} */ ;
// @ts-ignore
const __VLS_106 = __VLS_asFunctionalComponent(__VLS_105, new __VLS_105({
    modelValue: (__VLS_ctx.form.status),
}));
const __VLS_107 = __VLS_106({
    modelValue: (__VLS_ctx.form.status),
}, ...__VLS_functionalComponentArgsRest(__VLS_106));
__VLS_108.slots.default;
const __VLS_109 = {}.ElRadio;
/** @type {[typeof __VLS_components.ElRadio, typeof __VLS_components.elRadio, typeof __VLS_components.ElRadio, typeof __VLS_components.elRadio, ]} */ ;
// @ts-ignore
const __VLS_110 = __VLS_asFunctionalComponent(__VLS_109, new __VLS_109({
    value: "enabled",
}));
const __VLS_111 = __VLS_110({
    value: "enabled",
}, ...__VLS_functionalComponentArgsRest(__VLS_110));
__VLS_112.slots.default;
var __VLS_112;
const __VLS_113 = {}.ElRadio;
/** @type {[typeof __VLS_components.ElRadio, typeof __VLS_components.elRadio, typeof __VLS_components.ElRadio, typeof __VLS_components.elRadio, ]} */ ;
// @ts-ignore
const __VLS_114 = __VLS_asFunctionalComponent(__VLS_113, new __VLS_113({
    value: "disabled",
}));
const __VLS_115 = __VLS_114({
    value: "disabled",
}, ...__VLS_functionalComponentArgsRest(__VLS_114));
__VLS_116.slots.default;
var __VLS_116;
var __VLS_108;
var __VLS_104;
var __VLS_63;
{
    const { footer: __VLS_thisSlot } = __VLS_59.slots;
    const __VLS_117 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_118 = __VLS_asFunctionalComponent(__VLS_117, new __VLS_117({
        ...{ 'onClick': {} },
    }));
    const __VLS_119 = __VLS_118({
        ...{ 'onClick': {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_118));
    let __VLS_121;
    let __VLS_122;
    let __VLS_123;
    const __VLS_124 = {
        onClick: (...[$event]) => {
            __VLS_ctx.dialogVisible = false;
        }
    };
    __VLS_120.slots.default;
    var __VLS_120;
    const __VLS_125 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_126 = __VLS_asFunctionalComponent(__VLS_125, new __VLS_125({
        ...{ 'onClick': {} },
        type: "primary",
        loading: (__VLS_ctx.saving),
    }));
    const __VLS_127 = __VLS_126({
        ...{ 'onClick': {} },
        type: "primary",
        loading: (__VLS_ctx.saving),
    }, ...__VLS_functionalComponentArgsRest(__VLS_126));
    let __VLS_129;
    let __VLS_130;
    let __VLS_131;
    const __VLS_132 = {
        onClick: (__VLS_ctx.handleSave)
    };
    __VLS_128.slots.default;
    var __VLS_128;
}
var __VLS_59;
/** @type {__VLS_StyleScopedClasses['page-card']} */ ;
// @ts-ignore
var __VLS_65 = __VLS_64;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            UploadImage: UploadImage,
            loading: loading,
            categories: categories,
            dialogVisible: dialogVisible,
            dialogTitle: dialogTitle,
            formRef: formRef,
            saving: saving,
            form: form,
            openCreate: openCreate,
            openEdit: openEdit,
            handleSave: handleSave,
            handleDelete: handleDelete,
            rootCategories: rootCategories,
            categoryNameMap: categoryNameMap,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
