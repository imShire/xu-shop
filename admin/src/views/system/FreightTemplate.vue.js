import { ref, onMounted } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import { getFreightTemplates, createFreightTemplate, updateFreightTemplate, deleteFreightTemplate, } from '@/api/order';
import PriceInput from '@/components/PriceInput/index.vue';
const loading = ref(false);
const templates = ref([]);
const dialogVisible = ref(false);
const isEdit = ref(false);
const saving = ref(false);
const formRef = ref();
const form = ref({
    id: '',
    name: '',
    is_default: false,
    free_threshold: 0,
    base_fee: 0,
    rules: [],
});
async function loadData() {
    loading.value = true;
    try {
        templates.value = await getFreightTemplates();
    }
    finally {
        loading.value = false;
    }
}
function openCreate() {
    form.value = { id: '', name: '', is_default: false, free_threshold: 0, base_fee: 0, rules: [] };
    isEdit.value = false;
    dialogVisible.value = true;
}
function openEdit(row) {
    Object.assign(form.value, { ...row, rules: JSON.parse(JSON.stringify(row.rules || [])) });
    isEdit.value = true;
    dialogVisible.value = true;
}
async function handleSave() {
    await formRef.value?.validate();
    saving.value = true;
    try {
        if (isEdit.value) {
            await updateFreightTemplate(form.value.id, form.value);
        }
        else {
            await createFreightTemplate(form.value);
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
    try {
        await ElMessageBox.confirm('确认删除该运费模板？', '提示', { type: 'warning' });
        await deleteFreightTemplate(row.id);
        ElMessage.success('删除成功');
        await loadData();
    }
    catch (e) {
        if (e !== 'cancel')
            ElMessage.error(e?.message || '删除失败');
    }
}
function addRule() {
    form.value.rules.push({ regions: [], extra_fee: 0 });
}
function removeRule(idx) {
    form.value.rules.splice(idx, 1);
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
const __VLS_8 = {}.ElTable;
/** @type {[typeof __VLS_components.ElTable, typeof __VLS_components.elTable, typeof __VLS_components.ElTable, typeof __VLS_components.elTable, ]} */ ;
// @ts-ignore
const __VLS_9 = __VLS_asFunctionalComponent(__VLS_8, new __VLS_8({
    data: (__VLS_ctx.templates),
    border: true,
    stripe: true,
}));
const __VLS_10 = __VLS_9({
    data: (__VLS_ctx.templates),
    border: true,
    stripe: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_9));
__VLS_asFunctionalDirective(__VLS_directives.vLoading)(null, { ...__VLS_directiveBindingRestFields, value: (__VLS_ctx.loading) }, null, null);
__VLS_11.slots.default;
const __VLS_12 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_13 = __VLS_asFunctionalComponent(__VLS_12, new __VLS_12({
    prop: "name",
    label: "模板名称",
    minWidth: "150",
}));
const __VLS_14 = __VLS_13({
    prop: "name",
    label: "模板名称",
    minWidth: "150",
}, ...__VLS_functionalComponentArgsRest(__VLS_13));
const __VLS_16 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_17 = __VLS_asFunctionalComponent(__VLS_16, new __VLS_16({
    label: "基础运费",
    width: "120",
    align: "right",
}));
const __VLS_18 = __VLS_17({
    label: "基础运费",
    width: "120",
    align: "right",
}, ...__VLS_functionalComponentArgsRest(__VLS_17));
__VLS_19.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_19.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    ((row.base_fee / 100).toFixed(2));
}
var __VLS_19;
const __VLS_20 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_21 = __VLS_asFunctionalComponent(__VLS_20, new __VLS_20({
    label: "免运费门槛",
    width: "130",
    align: "right",
}));
const __VLS_22 = __VLS_21({
    label: "免运费门槛",
    width: "130",
    align: "right",
}, ...__VLS_functionalComponentArgsRest(__VLS_21));
__VLS_23.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_23.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    (row.free_threshold ? `¥${(row.free_threshold / 100).toFixed(2)}` : '-');
}
var __VLS_23;
const __VLS_24 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_25 = __VLS_asFunctionalComponent(__VLS_24, new __VLS_24({
    label: "默认",
    width: "80",
    align: "center",
}));
const __VLS_26 = __VLS_25({
    label: "默认",
    width: "80",
    align: "center",
}, ...__VLS_functionalComponentArgsRest(__VLS_25));
__VLS_27.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_27.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    if (row.is_default) {
        const __VLS_28 = {}.ElTag;
        /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
        // @ts-ignore
        const __VLS_29 = __VLS_asFunctionalComponent(__VLS_28, new __VLS_28({
            type: "success",
            size: "small",
        }));
        const __VLS_30 = __VLS_29({
            type: "success",
            size: "small",
        }, ...__VLS_functionalComponentArgsRest(__VLS_29));
        __VLS_31.slots.default;
        var __VLS_31;
    }
}
var __VLS_27;
const __VLS_32 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_33 = __VLS_asFunctionalComponent(__VLS_32, new __VLS_32({
    label: "操作",
    width: "140",
    fixed: "right",
}));
const __VLS_34 = __VLS_33({
    label: "操作",
    width: "140",
    fixed: "right",
}, ...__VLS_functionalComponentArgsRest(__VLS_33));
__VLS_35.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_35.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_36 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_37 = __VLS_asFunctionalComponent(__VLS_36, new __VLS_36({
        ...{ 'onClick': {} },
        text: true,
        type: "primary",
        size: "small",
    }));
    const __VLS_38 = __VLS_37({
        ...{ 'onClick': {} },
        text: true,
        type: "primary",
        size: "small",
    }, ...__VLS_functionalComponentArgsRest(__VLS_37));
    let __VLS_40;
    let __VLS_41;
    let __VLS_42;
    const __VLS_43 = {
        onClick: (...[$event]) => {
            __VLS_ctx.openEdit(row);
        }
    };
    __VLS_39.slots.default;
    var __VLS_39;
    const __VLS_44 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_45 = __VLS_asFunctionalComponent(__VLS_44, new __VLS_44({
        ...{ 'onClick': {} },
        text: true,
        type: "danger",
        size: "small",
    }));
    const __VLS_46 = __VLS_45({
        ...{ 'onClick': {} },
        text: true,
        type: "danger",
        size: "small",
    }, ...__VLS_functionalComponentArgsRest(__VLS_45));
    let __VLS_48;
    let __VLS_49;
    let __VLS_50;
    const __VLS_51 = {
        onClick: (...[$event]) => {
            __VLS_ctx.handleDelete(row);
        }
    };
    __VLS_47.slots.default;
    var __VLS_47;
}
var __VLS_35;
var __VLS_11;
const __VLS_52 = {}.ElDialog;
/** @type {[typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, ]} */ ;
// @ts-ignore
const __VLS_53 = __VLS_asFunctionalComponent(__VLS_52, new __VLS_52({
    modelValue: (__VLS_ctx.dialogVisible),
    title: (__VLS_ctx.isEdit ? '编辑模板' : '新建模板'),
    width: "560px",
}));
const __VLS_54 = __VLS_53({
    modelValue: (__VLS_ctx.dialogVisible),
    title: (__VLS_ctx.isEdit ? '编辑模板' : '新建模板'),
    width: "560px",
}, ...__VLS_functionalComponentArgsRest(__VLS_53));
__VLS_55.slots.default;
const __VLS_56 = {}.ElForm;
/** @type {[typeof __VLS_components.ElForm, typeof __VLS_components.elForm, typeof __VLS_components.ElForm, typeof __VLS_components.elForm, ]} */ ;
// @ts-ignore
const __VLS_57 = __VLS_asFunctionalComponent(__VLS_56, new __VLS_56({
    ref: "formRef",
    model: (__VLS_ctx.form),
    labelWidth: "100px",
}));
const __VLS_58 = __VLS_57({
    ref: "formRef",
    model: (__VLS_ctx.form),
    labelWidth: "100px",
}, ...__VLS_functionalComponentArgsRest(__VLS_57));
/** @type {typeof __VLS_ctx.formRef} */ ;
var __VLS_60 = {};
__VLS_59.slots.default;
const __VLS_62 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_63 = __VLS_asFunctionalComponent(__VLS_62, new __VLS_62({
    label: "模板名称",
    prop: "name",
    rules: ([{ required: true }]),
}));
const __VLS_64 = __VLS_63({
    label: "模板名称",
    prop: "name",
    rules: ([{ required: true }]),
}, ...__VLS_functionalComponentArgsRest(__VLS_63));
__VLS_65.slots.default;
const __VLS_66 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_67 = __VLS_asFunctionalComponent(__VLS_66, new __VLS_66({
    modelValue: (__VLS_ctx.form.name),
}));
const __VLS_68 = __VLS_67({
    modelValue: (__VLS_ctx.form.name),
}, ...__VLS_functionalComponentArgsRest(__VLS_67));
var __VLS_65;
const __VLS_70 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_71 = __VLS_asFunctionalComponent(__VLS_70, new __VLS_70({
    label: "基础运费",
}));
const __VLS_72 = __VLS_71({
    label: "基础运费",
}, ...__VLS_functionalComponentArgsRest(__VLS_71));
__VLS_73.slots.default;
/** @type {[typeof PriceInput, ]} */ ;
// @ts-ignore
const __VLS_74 = __VLS_asFunctionalComponent(PriceInput, new PriceInput({
    modelValue: (__VLS_ctx.form.base_fee),
}));
const __VLS_75 = __VLS_74({
    modelValue: (__VLS_ctx.form.base_fee),
}, ...__VLS_functionalComponentArgsRest(__VLS_74));
var __VLS_73;
const __VLS_77 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_78 = __VLS_asFunctionalComponent(__VLS_77, new __VLS_77({
    label: "免运费门槛",
}));
const __VLS_79 = __VLS_78({
    label: "免运费门槛",
}, ...__VLS_functionalComponentArgsRest(__VLS_78));
__VLS_80.slots.default;
/** @type {[typeof PriceInput, ]} */ ;
// @ts-ignore
const __VLS_81 = __VLS_asFunctionalComponent(PriceInput, new PriceInput({
    modelValue: (__VLS_ctx.form.free_threshold),
    placeholder: "0 表示不免运费",
}));
const __VLS_82 = __VLS_81({
    modelValue: (__VLS_ctx.form.free_threshold),
    placeholder: "0 表示不免运费",
}, ...__VLS_functionalComponentArgsRest(__VLS_81));
var __VLS_80;
const __VLS_84 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_85 = __VLS_asFunctionalComponent(__VLS_84, new __VLS_84({
    label: "设为默认",
}));
const __VLS_86 = __VLS_85({
    label: "设为默认",
}, ...__VLS_functionalComponentArgsRest(__VLS_85));
__VLS_87.slots.default;
const __VLS_88 = {}.ElSwitch;
/** @type {[typeof __VLS_components.ElSwitch, typeof __VLS_components.elSwitch, ]} */ ;
// @ts-ignore
const __VLS_89 = __VLS_asFunctionalComponent(__VLS_88, new __VLS_88({
    modelValue: (__VLS_ctx.form.is_default),
}));
const __VLS_90 = __VLS_89({
    modelValue: (__VLS_ctx.form.is_default),
}, ...__VLS_functionalComponentArgsRest(__VLS_89));
var __VLS_87;
const __VLS_92 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_93 = __VLS_asFunctionalComponent(__VLS_92, new __VLS_92({
    label: "偏远地区",
}));
const __VLS_94 = __VLS_93({
    label: "偏远地区",
}, ...__VLS_functionalComponentArgsRest(__VLS_93));
__VLS_95.slots.default;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ style: {} },
});
for (const [rule, idx] of __VLS_getVForSourceType((__VLS_ctx.form.rules))) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        key: (idx),
        ...{ style: {} },
    });
    const __VLS_96 = {}.ElSelect;
    /** @type {[typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, ]} */ ;
    // @ts-ignore
    const __VLS_97 = __VLS_asFunctionalComponent(__VLS_96, new __VLS_96({
        modelValue: (rule.regions),
        multiple: true,
        filterable: true,
        allowCreate: true,
        placeholder: "输入省份后回车",
        ...{ style: {} },
    }));
    const __VLS_98 = __VLS_97({
        modelValue: (rule.regions),
        multiple: true,
        filterable: true,
        allowCreate: true,
        placeholder: "输入省份后回车",
        ...{ style: {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_97));
    /** @type {[typeof PriceInput, ]} */ ;
    // @ts-ignore
    const __VLS_100 = __VLS_asFunctionalComponent(PriceInput, new PriceInput({
        modelValue: (rule.extra_fee),
        placeholder: "附加运费",
        ...{ style: {} },
    }));
    const __VLS_101 = __VLS_100({
        modelValue: (rule.extra_fee),
        placeholder: "附加运费",
        ...{ style: {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_100));
    const __VLS_103 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_104 = __VLS_asFunctionalComponent(__VLS_103, new __VLS_103({
        ...{ 'onClick': {} },
        text: true,
        type: "danger",
    }));
    const __VLS_105 = __VLS_104({
        ...{ 'onClick': {} },
        text: true,
        type: "danger",
    }, ...__VLS_functionalComponentArgsRest(__VLS_104));
    let __VLS_107;
    let __VLS_108;
    let __VLS_109;
    const __VLS_110 = {
        onClick: (...[$event]) => {
            __VLS_ctx.removeRule(idx);
        }
    };
    __VLS_106.slots.default;
    var __VLS_106;
}
const __VLS_111 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_112 = __VLS_asFunctionalComponent(__VLS_111, new __VLS_111({
    ...{ 'onClick': {} },
    size: "small",
}));
const __VLS_113 = __VLS_112({
    ...{ 'onClick': {} },
    size: "small",
}, ...__VLS_functionalComponentArgsRest(__VLS_112));
let __VLS_115;
let __VLS_116;
let __VLS_117;
const __VLS_118 = {
    onClick: (__VLS_ctx.addRule)
};
__VLS_114.slots.default;
var __VLS_114;
var __VLS_95;
var __VLS_59;
{
    const { footer: __VLS_thisSlot } = __VLS_55.slots;
    const __VLS_119 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_120 = __VLS_asFunctionalComponent(__VLS_119, new __VLS_119({
        ...{ 'onClick': {} },
    }));
    const __VLS_121 = __VLS_120({
        ...{ 'onClick': {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_120));
    let __VLS_123;
    let __VLS_124;
    let __VLS_125;
    const __VLS_126 = {
        onClick: (...[$event]) => {
            __VLS_ctx.dialogVisible = false;
        }
    };
    __VLS_122.slots.default;
    var __VLS_122;
    const __VLS_127 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_128 = __VLS_asFunctionalComponent(__VLS_127, new __VLS_127({
        ...{ 'onClick': {} },
        type: "primary",
        loading: (__VLS_ctx.saving),
    }));
    const __VLS_129 = __VLS_128({
        ...{ 'onClick': {} },
        type: "primary",
        loading: (__VLS_ctx.saving),
    }, ...__VLS_functionalComponentArgsRest(__VLS_128));
    let __VLS_131;
    let __VLS_132;
    let __VLS_133;
    const __VLS_134 = {
        onClick: (__VLS_ctx.handleSave)
    };
    __VLS_130.slots.default;
    var __VLS_130;
}
var __VLS_55;
/** @type {__VLS_StyleScopedClasses['page-card']} */ ;
// @ts-ignore
var __VLS_61 = __VLS_60;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            PriceInput: PriceInput,
            loading: loading,
            templates: templates,
            dialogVisible: dialogVisible,
            isEdit: isEdit,
            saving: saving,
            formRef: formRef,
            form: form,
            openCreate: openCreate,
            openEdit: openEdit,
            handleSave: handleSave,
            handleDelete: handleDelete,
            addRule: addRule,
            removeRule: removeRule,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
