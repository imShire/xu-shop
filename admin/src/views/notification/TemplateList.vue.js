import { ref, onMounted } from 'vue';
import { ElMessage } from 'element-plus';
import { getTemplateList, updateTemplate, testSendNotification } from '@/api/notification';
import { formatTime } from '@/utils/format';
const loading = ref(false);
const templates = ref([]);
const editDialog = ref({ visible: false, code: '' });
const editForm = ref({ template_id_external: '', fields_text: '{}', enabled: true });
const editFormRef = ref();
const saving = ref(false);
const testDialog = ref({ visible: false, code: '', openid: '' });
const testLoading = ref(false);
function stringifyFields(fields) {
    return JSON.stringify(fields || {}, null, 2);
}
function parseFields(text) {
    const normalized = text.trim() || '{}';
    const parsed = JSON.parse(normalized);
    if (parsed === null || Array.isArray(parsed) || typeof parsed !== 'object') {
        throw new Error('字段映射必须是 JSON 对象');
    }
    return parsed;
}
async function loadData() {
    loading.value = true;
    try {
        templates.value = await getTemplateList();
    }
    finally {
        loading.value = false;
    }
}
function openEdit(row) {
    editDialog.value = { visible: true, code: row.code };
    editForm.value = {
        template_id_external: row.template_id_external || '',
        fields_text: stringifyFields(row.fields),
        enabled: row.enabled,
    };
}
async function handleSave() {
    await editFormRef.value?.validate();
    let fields;
    try {
        fields = parseFields(editForm.value.fields_text);
    }
    catch (error) {
        ElMessage.error(error?.message || '字段映射格式错误');
        return;
    }
    saving.value = true;
    try {
        await updateTemplate(editDialog.value.code, {
            template_id_external: editForm.value.template_id_external.trim(),
            fields,
            enabled: editForm.value.enabled,
        });
        ElMessage.success('保存成功');
        editDialog.value.visible = false;
        await loadData();
    }
    catch (e) {
        ElMessage.error(e?.message || '操作失败');
    }
    finally {
        saving.value = false;
    }
}
function openTest(row) {
    testDialog.value = { visible: true, code: row.code, openid: '' };
}
async function handleTest() {
    if (!testDialog.value.openid)
        return ElMessage.warning('请输入测试目标');
    testLoading.value = true;
    try {
        await testSendNotification(testDialog.value.code, { openid: testDialog.value.openid });
        ElMessage.success('测试消息已发送');
        testDialog.value.visible = false;
    }
    catch (e) {
        ElMessage.error(e?.message || '操作失败');
    }
    finally {
        testLoading.value = false;
    }
}
async function toggleEnabled(row, nextEnabled) {
    const previousEnabled = row.enabled;
    row.enabled = nextEnabled;
    try {
        await updateTemplate(row.code, { enabled: row.enabled });
        ElMessage.success(row.enabled ? '已启用' : '已禁用');
    }
    catch (e) {
        row.enabled = previousEnabled;
        ElMessage.error(e?.message || '操作失败');
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
    data: (__VLS_ctx.templates),
    border: true,
    stripe: true,
}));
const __VLS_2 = __VLS_1({
    data: (__VLS_ctx.templates),
    border: true,
    stripe: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_1));
__VLS_asFunctionalDirective(__VLS_directives.vLoading)(null, { ...__VLS_directiveBindingRestFields, value: (__VLS_ctx.loading) }, null, null);
__VLS_3.slots.default;
const __VLS_4 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_5 = __VLS_asFunctionalComponent(__VLS_4, new __VLS_4({
    prop: "code",
    label: "模板编码",
    minWidth: "160",
}));
const __VLS_6 = __VLS_5({
    prop: "code",
    label: "模板编码",
    minWidth: "160",
}, ...__VLS_functionalComponentArgsRest(__VLS_5));
const __VLS_8 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_9 = __VLS_asFunctionalComponent(__VLS_8, new __VLS_8({
    prop: "channel",
    label: "渠道",
    width: "120",
}));
const __VLS_10 = __VLS_9({
    prop: "channel",
    label: "渠道",
    width: "120",
}, ...__VLS_functionalComponentArgsRest(__VLS_9));
const __VLS_12 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_13 = __VLS_asFunctionalComponent(__VLS_12, new __VLS_12({
    prop: "template_id_external",
    label: "外部模板ID",
    minWidth: "200",
}));
const __VLS_14 = __VLS_13({
    prop: "template_id_external",
    label: "外部模板ID",
    minWidth: "200",
}, ...__VLS_functionalComponentArgsRest(__VLS_13));
const __VLS_16 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_17 = __VLS_asFunctionalComponent(__VLS_16, new __VLS_16({
    label: "字段映射",
    minWidth: "220",
}));
const __VLS_18 = __VLS_17({
    label: "字段映射",
    minWidth: "220",
}, ...__VLS_functionalComponentArgsRest(__VLS_17));
__VLS_19.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_19.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.code, __VLS_intrinsicElements.code)({});
    (__VLS_ctx.stringifyFields(row.fields));
}
var __VLS_19;
const __VLS_20 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_21 = __VLS_asFunctionalComponent(__VLS_20, new __VLS_20({
    label: "启用",
    width: "80",
    align: "center",
}));
const __VLS_22 = __VLS_21({
    label: "启用",
    width: "80",
    align: "center",
}, ...__VLS_functionalComponentArgsRest(__VLS_21));
__VLS_23.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_23.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_24 = {}.ElSwitch;
    /** @type {[typeof __VLS_components.ElSwitch, typeof __VLS_components.elSwitch, ]} */ ;
    // @ts-ignore
    const __VLS_25 = __VLS_asFunctionalComponent(__VLS_24, new __VLS_24({
        ...{ 'onChange': {} },
        modelValue: (row.enabled),
    }));
    const __VLS_26 = __VLS_25({
        ...{ 'onChange': {} },
        modelValue: (row.enabled),
    }, ...__VLS_functionalComponentArgsRest(__VLS_25));
    let __VLS_28;
    let __VLS_29;
    let __VLS_30;
    const __VLS_31 = {
        onChange: ((value) => __VLS_ctx.toggleEnabled(row, value))
    };
    var __VLS_27;
}
var __VLS_23;
const __VLS_32 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_33 = __VLS_asFunctionalComponent(__VLS_32, new __VLS_32({
    label: "更新时间",
    width: "150",
    formatter: ((row) => __VLS_ctx.formatTime(row.updated_at)),
}));
const __VLS_34 = __VLS_33({
    label: "更新时间",
    width: "150",
    formatter: ((row) => __VLS_ctx.formatTime(row.updated_at)),
}, ...__VLS_functionalComponentArgsRest(__VLS_33));
const __VLS_36 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_37 = __VLS_asFunctionalComponent(__VLS_36, new __VLS_36({
    label: "操作",
    width: "150",
    fixed: "right",
}));
const __VLS_38 = __VLS_37({
    label: "操作",
    width: "150",
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
        size: "small",
    }));
    const __VLS_50 = __VLS_49({
        ...{ 'onClick': {} },
        text: true,
        size: "small",
    }, ...__VLS_functionalComponentArgsRest(__VLS_49));
    let __VLS_52;
    let __VLS_53;
    let __VLS_54;
    const __VLS_55 = {
        onClick: (...[$event]) => {
            __VLS_ctx.openTest(row);
        }
    };
    __VLS_51.slots.default;
    var __VLS_51;
}
var __VLS_39;
var __VLS_3;
const __VLS_56 = {}.ElDialog;
/** @type {[typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, ]} */ ;
// @ts-ignore
const __VLS_57 = __VLS_asFunctionalComponent(__VLS_56, new __VLS_56({
    modelValue: (__VLS_ctx.editDialog.visible),
    title: "编辑模板",
    width: "500px",
}));
const __VLS_58 = __VLS_57({
    modelValue: (__VLS_ctx.editDialog.visible),
    title: "编辑模板",
    width: "500px",
}, ...__VLS_functionalComponentArgsRest(__VLS_57));
__VLS_59.slots.default;
const __VLS_60 = {}.ElForm;
/** @type {[typeof __VLS_components.ElForm, typeof __VLS_components.elForm, typeof __VLS_components.ElForm, typeof __VLS_components.elForm, ]} */ ;
// @ts-ignore
const __VLS_61 = __VLS_asFunctionalComponent(__VLS_60, new __VLS_60({
    ref: "editFormRef",
    model: (__VLS_ctx.editForm),
    labelWidth: "100px",
}));
const __VLS_62 = __VLS_61({
    ref: "editFormRef",
    model: (__VLS_ctx.editForm),
    labelWidth: "100px",
}, ...__VLS_functionalComponentArgsRest(__VLS_61));
/** @type {typeof __VLS_ctx.editFormRef} */ ;
var __VLS_64 = {};
__VLS_63.slots.default;
const __VLS_66 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_67 = __VLS_asFunctionalComponent(__VLS_66, new __VLS_66({
    label: "外部模板ID",
}));
const __VLS_68 = __VLS_67({
    label: "外部模板ID",
}, ...__VLS_functionalComponentArgsRest(__VLS_67));
__VLS_69.slots.default;
const __VLS_70 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_71 = __VLS_asFunctionalComponent(__VLS_70, new __VLS_70({
    modelValue: (__VLS_ctx.editForm.template_id_external),
    placeholder: "对接平台的模板ID",
}));
const __VLS_72 = __VLS_71({
    modelValue: (__VLS_ctx.editForm.template_id_external),
    placeholder: "对接平台的模板ID",
}, ...__VLS_functionalComponentArgsRest(__VLS_71));
var __VLS_69;
const __VLS_74 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_75 = __VLS_asFunctionalComponent(__VLS_74, new __VLS_74({
    label: "字段映射",
    prop: "fields_text",
    rules: ([{ required: true, message: '请输入字段映射 JSON', trigger: 'blur' }]),
}));
const __VLS_76 = __VLS_75({
    label: "字段映射",
    prop: "fields_text",
    rules: ([{ required: true, message: '请输入字段映射 JSON', trigger: 'blur' }]),
}, ...__VLS_functionalComponentArgsRest(__VLS_75));
__VLS_77.slots.default;
const __VLS_78 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_79 = __VLS_asFunctionalComponent(__VLS_78, new __VLS_78({
    modelValue: (__VLS_ctx.editForm.fields_text),
    type: "textarea",
    rows: (8),
    placeholder: '{"thing1":"content"}',
}));
const __VLS_80 = __VLS_79({
    modelValue: (__VLS_ctx.editForm.fields_text),
    type: "textarea",
    rows: (8),
    placeholder: '{"thing1":"content"}',
}, ...__VLS_functionalComponentArgsRest(__VLS_79));
var __VLS_77;
const __VLS_82 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_83 = __VLS_asFunctionalComponent(__VLS_82, new __VLS_82({
    label: "启用",
}));
const __VLS_84 = __VLS_83({
    label: "启用",
}, ...__VLS_functionalComponentArgsRest(__VLS_83));
__VLS_85.slots.default;
const __VLS_86 = {}.ElSwitch;
/** @type {[typeof __VLS_components.ElSwitch, typeof __VLS_components.elSwitch, ]} */ ;
// @ts-ignore
const __VLS_87 = __VLS_asFunctionalComponent(__VLS_86, new __VLS_86({
    modelValue: (__VLS_ctx.editForm.enabled),
}));
const __VLS_88 = __VLS_87({
    modelValue: (__VLS_ctx.editForm.enabled),
}, ...__VLS_functionalComponentArgsRest(__VLS_87));
var __VLS_85;
var __VLS_63;
{
    const { footer: __VLS_thisSlot } = __VLS_59.slots;
    const __VLS_90 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_91 = __VLS_asFunctionalComponent(__VLS_90, new __VLS_90({
        ...{ 'onClick': {} },
    }));
    const __VLS_92 = __VLS_91({
        ...{ 'onClick': {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_91));
    let __VLS_94;
    let __VLS_95;
    let __VLS_96;
    const __VLS_97 = {
        onClick: (...[$event]) => {
            __VLS_ctx.editDialog.visible = false;
        }
    };
    __VLS_93.slots.default;
    var __VLS_93;
    const __VLS_98 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_99 = __VLS_asFunctionalComponent(__VLS_98, new __VLS_98({
        ...{ 'onClick': {} },
        type: "primary",
        loading: (__VLS_ctx.saving),
    }));
    const __VLS_100 = __VLS_99({
        ...{ 'onClick': {} },
        type: "primary",
        loading: (__VLS_ctx.saving),
    }, ...__VLS_functionalComponentArgsRest(__VLS_99));
    let __VLS_102;
    let __VLS_103;
    let __VLS_104;
    const __VLS_105 = {
        onClick: (__VLS_ctx.handleSave)
    };
    __VLS_101.slots.default;
    var __VLS_101;
}
var __VLS_59;
const __VLS_106 = {}.ElDialog;
/** @type {[typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, ]} */ ;
// @ts-ignore
const __VLS_107 = __VLS_asFunctionalComponent(__VLS_106, new __VLS_106({
    modelValue: (__VLS_ctx.testDialog.visible),
    title: "测试发送",
    width: "400px",
}));
const __VLS_108 = __VLS_107({
    modelValue: (__VLS_ctx.testDialog.visible),
    title: "测试发送",
    width: "400px",
}, ...__VLS_functionalComponentArgsRest(__VLS_107));
__VLS_109.slots.default;
const __VLS_110 = {}.ElForm;
/** @type {[typeof __VLS_components.ElForm, typeof __VLS_components.elForm, typeof __VLS_components.ElForm, typeof __VLS_components.elForm, ]} */ ;
// @ts-ignore
const __VLS_111 = __VLS_asFunctionalComponent(__VLS_110, new __VLS_110({
    labelWidth: "80px",
}));
const __VLS_112 = __VLS_111({
    labelWidth: "80px",
}, ...__VLS_functionalComponentArgsRest(__VLS_111));
__VLS_113.slots.default;
const __VLS_114 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_115 = __VLS_asFunctionalComponent(__VLS_114, new __VLS_114({
    label: "目标",
}));
const __VLS_116 = __VLS_115({
    label: "目标",
}, ...__VLS_functionalComponentArgsRest(__VLS_115));
__VLS_117.slots.default;
const __VLS_118 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_119 = __VLS_asFunctionalComponent(__VLS_118, new __VLS_118({
    modelValue: (__VLS_ctx.testDialog.openid),
    placeholder: "用户 openid",
}));
const __VLS_120 = __VLS_119({
    modelValue: (__VLS_ctx.testDialog.openid),
    placeholder: "用户 openid",
}, ...__VLS_functionalComponentArgsRest(__VLS_119));
var __VLS_117;
var __VLS_113;
{
    const { footer: __VLS_thisSlot } = __VLS_109.slots;
    const __VLS_122 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_123 = __VLS_asFunctionalComponent(__VLS_122, new __VLS_122({
        ...{ 'onClick': {} },
    }));
    const __VLS_124 = __VLS_123({
        ...{ 'onClick': {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_123));
    let __VLS_126;
    let __VLS_127;
    let __VLS_128;
    const __VLS_129 = {
        onClick: (...[$event]) => {
            __VLS_ctx.testDialog.visible = false;
        }
    };
    __VLS_125.slots.default;
    var __VLS_125;
    const __VLS_130 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_131 = __VLS_asFunctionalComponent(__VLS_130, new __VLS_130({
        ...{ 'onClick': {} },
        type: "primary",
        loading: (__VLS_ctx.testLoading),
    }));
    const __VLS_132 = __VLS_131({
        ...{ 'onClick': {} },
        type: "primary",
        loading: (__VLS_ctx.testLoading),
    }, ...__VLS_functionalComponentArgsRest(__VLS_131));
    let __VLS_134;
    let __VLS_135;
    let __VLS_136;
    const __VLS_137 = {
        onClick: (__VLS_ctx.handleTest)
    };
    __VLS_133.slots.default;
    var __VLS_133;
}
var __VLS_109;
/** @type {__VLS_StyleScopedClasses['page-card']} */ ;
// @ts-ignore
var __VLS_65 = __VLS_64;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            formatTime: formatTime,
            loading: loading,
            templates: templates,
            editDialog: editDialog,
            editForm: editForm,
            editFormRef: editFormRef,
            saving: saving,
            testDialog: testDialog,
            testLoading: testLoading,
            stringifyFields: stringifyFields,
            openEdit: openEdit,
            handleSave: handleSave,
            openTest: openTest,
            handleTest: handleTest,
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
