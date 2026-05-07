import { ref, onMounted } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import { getRoleList, createRole, updateRole, deleteRole, getPermissions } from '@/api/account';
const loading = ref(false);
const roles = ref([]);
const permissions = ref([]);
const permGroups = ref({});
const dialogVisible = ref(false);
const isEdit = ref(false);
const saving = ref(false);
const formRef = ref();
const form = ref({
    id: '',
    name: '',
    code: '',
    perm_codes: [],
});
async function loadData() {
    loading.value = true;
    try {
        const [roleRes, permRes] = await Promise.all([getRoleList(), getPermissions()]);
        roles.value = roleRes;
        permissions.value = permRes;
        const groups = {};
        permRes.forEach((p) => {
            if (!groups[p.group])
                groups[p.group] = [];
            groups[p.group].push(p);
        });
        permGroups.value = groups;
    }
    finally {
        loading.value = false;
    }
}
function openCreate() {
    form.value = { id: '', name: '', code: '', perm_codes: [] };
    isEdit.value = false;
    dialogVisible.value = true;
}
function openEdit(row) {
    const permCodes = (row.permissions || []).map((p) => p.code);
    form.value = { id: row.id, name: row.name, code: row.code, perm_codes: permCodes };
    isEdit.value = true;
    dialogVisible.value = true;
}
async function handleSave() {
    await formRef.value?.validate();
    saving.value = true;
    try {
        if (isEdit.value) {
            await updateRole(form.value.id, form.value);
        }
        else {
            await createRole(form.value);
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
    await ElMessageBox.confirm(`确认删除角色「${row.name}」？`, '提示', { type: 'warning' });
    await deleteRole(row.id);
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
const __VLS_8 = {}.ElTable;
/** @type {[typeof __VLS_components.ElTable, typeof __VLS_components.elTable, typeof __VLS_components.ElTable, typeof __VLS_components.elTable, ]} */ ;
// @ts-ignore
const __VLS_9 = __VLS_asFunctionalComponent(__VLS_8, new __VLS_8({
    data: (__VLS_ctx.roles),
    border: true,
    stripe: true,
}));
const __VLS_10 = __VLS_9({
    data: (__VLS_ctx.roles),
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
    label: "角色名称",
    width: "150",
}));
const __VLS_14 = __VLS_13({
    prop: "name",
    label: "角色名称",
    width: "150",
}, ...__VLS_functionalComponentArgsRest(__VLS_13));
const __VLS_16 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_17 = __VLS_asFunctionalComponent(__VLS_16, new __VLS_16({
    prop: "code",
    label: "角色代码",
    width: "150",
}));
const __VLS_18 = __VLS_17({
    prop: "code",
    label: "角色代码",
    width: "150",
}, ...__VLS_functionalComponentArgsRest(__VLS_17));
const __VLS_20 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_21 = __VLS_asFunctionalComponent(__VLS_20, new __VLS_20({
    label: "权限",
    minWidth: "200",
}));
const __VLS_22 = __VLS_21({
    label: "权限",
    minWidth: "200",
}, ...__VLS_functionalComponentArgsRest(__VLS_21));
__VLS_23.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_23.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    for (const [perm] of __VLS_getVForSourceType(((row.permissions || []).slice(0, 5)))) {
        const __VLS_24 = {}.ElTag;
        /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
        // @ts-ignore
        const __VLS_25 = __VLS_asFunctionalComponent(__VLS_24, new __VLS_24({
            key: (perm.code),
            size: "small",
            ...{ style: {} },
        }));
        const __VLS_26 = __VLS_25({
            key: (perm.code),
            size: "small",
            ...{ style: {} },
        }, ...__VLS_functionalComponentArgsRest(__VLS_25));
        __VLS_27.slots.default;
        (perm.name);
        var __VLS_27;
    }
    if ((row.permissions || []).length > 5) {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
            ...{ style: {} },
        });
        (row.permissions.length - 5);
    }
}
var __VLS_23;
const __VLS_28 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_29 = __VLS_asFunctionalComponent(__VLS_28, new __VLS_28({
    label: "操作",
    width: "140",
    fixed: "right",
}));
const __VLS_30 = __VLS_29({
    label: "操作",
    width: "140",
    fixed: "right",
}, ...__VLS_functionalComponentArgsRest(__VLS_29));
__VLS_31.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_31.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_32 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_33 = __VLS_asFunctionalComponent(__VLS_32, new __VLS_32({
        ...{ 'onClick': {} },
        text: true,
        type: "primary",
        size: "small",
    }));
    const __VLS_34 = __VLS_33({
        ...{ 'onClick': {} },
        text: true,
        type: "primary",
        size: "small",
    }, ...__VLS_functionalComponentArgsRest(__VLS_33));
    let __VLS_36;
    let __VLS_37;
    let __VLS_38;
    const __VLS_39 = {
        onClick: (...[$event]) => {
            __VLS_ctx.openEdit(row);
        }
    };
    __VLS_35.slots.default;
    var __VLS_35;
    const __VLS_40 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_41 = __VLS_asFunctionalComponent(__VLS_40, new __VLS_40({
        ...{ 'onClick': {} },
        text: true,
        type: "danger",
        size: "small",
    }));
    const __VLS_42 = __VLS_41({
        ...{ 'onClick': {} },
        text: true,
        type: "danger",
        size: "small",
    }, ...__VLS_functionalComponentArgsRest(__VLS_41));
    let __VLS_44;
    let __VLS_45;
    let __VLS_46;
    const __VLS_47 = {
        onClick: (...[$event]) => {
            __VLS_ctx.handleDelete(row);
        }
    };
    __VLS_43.slots.default;
    var __VLS_43;
}
var __VLS_31;
var __VLS_11;
const __VLS_48 = {}.ElDialog;
/** @type {[typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, ]} */ ;
// @ts-ignore
const __VLS_49 = __VLS_asFunctionalComponent(__VLS_48, new __VLS_48({
    modelValue: (__VLS_ctx.dialogVisible),
    title: (__VLS_ctx.isEdit ? '编辑角色' : '新建角色'),
    width: "600px",
}));
const __VLS_50 = __VLS_49({
    modelValue: (__VLS_ctx.dialogVisible),
    title: (__VLS_ctx.isEdit ? '编辑角色' : '新建角色'),
    width: "600px",
}, ...__VLS_functionalComponentArgsRest(__VLS_49));
__VLS_51.slots.default;
const __VLS_52 = {}.ElForm;
/** @type {[typeof __VLS_components.ElForm, typeof __VLS_components.elForm, typeof __VLS_components.ElForm, typeof __VLS_components.elForm, ]} */ ;
// @ts-ignore
const __VLS_53 = __VLS_asFunctionalComponent(__VLS_52, new __VLS_52({
    ref: "formRef",
    model: (__VLS_ctx.form),
    labelWidth: "80px",
}));
const __VLS_54 = __VLS_53({
    ref: "formRef",
    model: (__VLS_ctx.form),
    labelWidth: "80px",
}, ...__VLS_functionalComponentArgsRest(__VLS_53));
/** @type {typeof __VLS_ctx.formRef} */ ;
var __VLS_56 = {};
__VLS_55.slots.default;
const __VLS_58 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_59 = __VLS_asFunctionalComponent(__VLS_58, new __VLS_58({
    label: "角色名",
    prop: "name",
    rules: ([{ required: true }]),
}));
const __VLS_60 = __VLS_59({
    label: "角色名",
    prop: "name",
    rules: ([{ required: true }]),
}, ...__VLS_functionalComponentArgsRest(__VLS_59));
__VLS_61.slots.default;
const __VLS_62 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_63 = __VLS_asFunctionalComponent(__VLS_62, new __VLS_62({
    modelValue: (__VLS_ctx.form.name),
}));
const __VLS_64 = __VLS_63({
    modelValue: (__VLS_ctx.form.name),
}, ...__VLS_functionalComponentArgsRest(__VLS_63));
var __VLS_61;
const __VLS_66 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_67 = __VLS_asFunctionalComponent(__VLS_66, new __VLS_66({
    label: "代码",
    prop: "code",
    rules: ([{ required: true }]),
}));
const __VLS_68 = __VLS_67({
    label: "代码",
    prop: "code",
    rules: ([{ required: true }]),
}, ...__VLS_functionalComponentArgsRest(__VLS_67));
__VLS_69.slots.default;
const __VLS_70 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_71 = __VLS_asFunctionalComponent(__VLS_70, new __VLS_70({
    modelValue: (__VLS_ctx.form.code),
    placeholder: "如：admin",
}));
const __VLS_72 = __VLS_71({
    modelValue: (__VLS_ctx.form.code),
    placeholder: "如：admin",
}, ...__VLS_functionalComponentArgsRest(__VLS_71));
var __VLS_69;
const __VLS_74 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_75 = __VLS_asFunctionalComponent(__VLS_74, new __VLS_74({
    label: "权限",
}));
const __VLS_76 = __VLS_75({
    label: "权限",
}, ...__VLS_functionalComponentArgsRest(__VLS_75));
__VLS_77.slots.default;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ style: {} },
});
for (const [perms, group] of __VLS_getVForSourceType((__VLS_ctx.permGroups))) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        key: (group),
        ...{ style: {} },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ style: {} },
    });
    (group);
    const __VLS_78 = {}.ElCheckboxGroup;
    /** @type {[typeof __VLS_components.ElCheckboxGroup, typeof __VLS_components.elCheckboxGroup, typeof __VLS_components.ElCheckboxGroup, typeof __VLS_components.elCheckboxGroup, ]} */ ;
    // @ts-ignore
    const __VLS_79 = __VLS_asFunctionalComponent(__VLS_78, new __VLS_78({
        modelValue: (__VLS_ctx.form.perm_codes),
        ...{ style: {} },
    }));
    const __VLS_80 = __VLS_79({
        modelValue: (__VLS_ctx.form.perm_codes),
        ...{ style: {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_79));
    __VLS_81.slots.default;
    for (const [p] of __VLS_getVForSourceType((perms))) {
        const __VLS_82 = {}.ElCheckbox;
        /** @type {[typeof __VLS_components.ElCheckbox, typeof __VLS_components.elCheckbox, typeof __VLS_components.ElCheckbox, typeof __VLS_components.elCheckbox, ]} */ ;
        // @ts-ignore
        const __VLS_83 = __VLS_asFunctionalComponent(__VLS_82, new __VLS_82({
            key: (p.code),
            value: (p.code),
            ...{ style: {} },
        }));
        const __VLS_84 = __VLS_83({
            key: (p.code),
            value: (p.code),
            ...{ style: {} },
        }, ...__VLS_functionalComponentArgsRest(__VLS_83));
        __VLS_85.slots.default;
        (p.name);
        var __VLS_85;
    }
    var __VLS_81;
}
var __VLS_77;
var __VLS_55;
{
    const { footer: __VLS_thisSlot } = __VLS_51.slots;
    const __VLS_86 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_87 = __VLS_asFunctionalComponent(__VLS_86, new __VLS_86({
        ...{ 'onClick': {} },
    }));
    const __VLS_88 = __VLS_87({
        ...{ 'onClick': {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_87));
    let __VLS_90;
    let __VLS_91;
    let __VLS_92;
    const __VLS_93 = {
        onClick: (...[$event]) => {
            __VLS_ctx.dialogVisible = false;
        }
    };
    __VLS_89.slots.default;
    var __VLS_89;
    const __VLS_94 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_95 = __VLS_asFunctionalComponent(__VLS_94, new __VLS_94({
        ...{ 'onClick': {} },
        type: "primary",
        loading: (__VLS_ctx.saving),
    }));
    const __VLS_96 = __VLS_95({
        ...{ 'onClick': {} },
        type: "primary",
        loading: (__VLS_ctx.saving),
    }, ...__VLS_functionalComponentArgsRest(__VLS_95));
    let __VLS_98;
    let __VLS_99;
    let __VLS_100;
    const __VLS_101 = {
        onClick: (__VLS_ctx.handleSave)
    };
    __VLS_97.slots.default;
    var __VLS_97;
}
var __VLS_51;
/** @type {__VLS_StyleScopedClasses['page-card']} */ ;
// @ts-ignore
var __VLS_57 = __VLS_56;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            loading: loading,
            roles: roles,
            permGroups: permGroups,
            dialogVisible: dialogVisible,
            isEdit: isEdit,
            saving: saving,
            formRef: formRef,
            form: form,
            openCreate: openCreate,
            openEdit: openEdit,
            handleSave: handleSave,
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
