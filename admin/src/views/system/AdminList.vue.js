import { ref, onMounted } from 'vue';
import { ElMessage } from 'element-plus';
import ProTable from '@/components/ProTable/index.vue';
import { useTable } from '@/composables/useTable';
import { getAdminList, createAdmin, updateAdmin, disableAdmin, enableAdmin, resetAdminPwd, getRoleList, } from '@/api/account';
import { formatTime } from '@/utils/format';
const searchForm = ref({ username: '', status: '' });
const roles = ref([]);
const { list, total, page, pageSize, loading, fetch } = useTable((params) => getAdminList({ ...params, ...searchForm.value }));
const columns = [
    { label: '用户名', prop: 'username', width: 130 },
    { label: '真实姓名', prop: 'real_name', width: 120 },
    { label: '角色', slot: 'roles', minWidth: 150 },
    { label: '状态', slot: 'status', width: 80, align: 'center' },
    { label: '创建时间', prop: 'created_at', width: 150, formatter: (r) => formatTime(r.created_at) },
];
const dialogVisible = ref(false);
const isEdit = ref(false);
const saving = ref(false);
const formRef = ref();
const pwdFormRef = ref();
const form = ref({
    id: '',
    username: '',
    real_name: '',
    password: '',
    role_ids: [],
    status: 1,
});
const pwdDialog = ref({ visible: false, id: '', new_password: '' });
const pwdLoading = ref(false);
function validateStrongPassword(_, value, callback) {
    if (!value) {
        callback(new Error('请输入密码'));
        return;
    }
    if (value.length < 12) {
        callback(new Error('密码至少12位'));
        return;
    }
    const hasLetter = /[A-Za-z]/.test(value);
    const hasDigit = /\d/.test(value);
    const hasSpecial = /[^A-Za-z0-9]/.test(value);
    if (!hasLetter || !hasDigit || !hasSpecial) {
        callback(new Error('密码须包含字母、数字和特殊符号'));
        return;
    }
    callback();
}
const passwordRules = [
    { required: true, validator: validateStrongPassword, trigger: 'blur' },
];
function isAdminActive(status) {
    return status === 'active';
}
function resolveRoleIds(roleCodes) {
    if (!Array.isArray(roleCodes) || roleCodes.length === 0) {
        return [];
    }
    return roles.value
        .filter((role) => roleCodes.includes(role.code))
        .map((role) => role.id);
}
function openCreate() {
    form.value = { id: '', username: '', real_name: '', password: '', role_ids: [], status: 1 };
    isEdit.value = false;
    dialogVisible.value = true;
}
function openEdit(row) {
    form.value = {
        id: row.id,
        username: row.username,
        real_name: row.real_name || '',
        password: '',
        role_ids: resolveRoleIds(row.roles),
        status: row.status,
    };
    isEdit.value = true;
    dialogVisible.value = true;
}
async function handleSave() {
    await formRef.value?.validate();
    saving.value = true;
    try {
        const data = { ...form.value };
        if (isEdit.value) {
            await updateAdmin(form.value.id, data);
        }
        else {
            await createAdmin(data);
        }
        ElMessage.success('保存成功');
        dialogVisible.value = false;
        fetch(searchForm.value);
    }
    finally {
        saving.value = false;
    }
}
async function toggleStatus(row) {
    if (isAdminActive(row.status)) {
        await disableAdmin(row.id);
        ElMessage.success('已禁用');
    }
    else {
        await enableAdmin(row.id);
        ElMessage.success('已启用');
    }
    fetch(searchForm.value);
}
function openResetPwd(row) {
    pwdDialog.value = { visible: true, id: row.id, new_password: '' };
}
async function handleResetPwd() {
    await pwdFormRef.value?.validate();
    pwdLoading.value = true;
    try {
        await resetAdminPwd(pwdDialog.value.id, { new_password: pwdDialog.value.new_password });
        ElMessage.success('密码已重置');
        pwdDialog.value.visible = false;
    }
    finally {
        pwdLoading.value = false;
    }
}
onMounted(async () => {
    roles.value = await getRoleList().catch(() => []);
    fetch(searchForm.value);
});
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "page-card" },
});
/** @type {[typeof ProTable, typeof ProTable, ]} */ ;
// @ts-ignore
const __VLS_0 = __VLS_asFunctionalComponent(ProTable, new ProTable({
    ...{ 'onRefresh': {} },
    columns: (__VLS_ctx.columns),
    data: (__VLS_ctx.list),
    total: (__VLS_ctx.total),
    loading: (__VLS_ctx.loading),
    page: (__VLS_ctx.page),
    pageSize: (__VLS_ctx.pageSize),
}));
const __VLS_1 = __VLS_0({
    ...{ 'onRefresh': {} },
    columns: (__VLS_ctx.columns),
    data: (__VLS_ctx.list),
    total: (__VLS_ctx.total),
    loading: (__VLS_ctx.loading),
    page: (__VLS_ctx.page),
    pageSize: (__VLS_ctx.pageSize),
}, ...__VLS_functionalComponentArgsRest(__VLS_0));
let __VLS_3;
let __VLS_4;
let __VLS_5;
const __VLS_6 = {
    onRefresh: (...[$event]) => {
        __VLS_ctx.fetch(__VLS_ctx.searchForm);
    }
};
__VLS_2.slots.default;
{
    const { search: __VLS_thisSlot } = __VLS_2.slots;
    const __VLS_7 = {}.ElInput;
    /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
    // @ts-ignore
    const __VLS_8 = __VLS_asFunctionalComponent(__VLS_7, new __VLS_7({
        modelValue: (__VLS_ctx.searchForm.username),
        placeholder: "用户名",
        clearable: true,
        ...{ style: {} },
    }));
    const __VLS_9 = __VLS_8({
        modelValue: (__VLS_ctx.searchForm.username),
        placeholder: "用户名",
        clearable: true,
        ...{ style: {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_8));
    const __VLS_11 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_12 = __VLS_asFunctionalComponent(__VLS_11, new __VLS_11({
        ...{ 'onClick': {} },
        type: "primary",
    }));
    const __VLS_13 = __VLS_12({
        ...{ 'onClick': {} },
        type: "primary",
    }, ...__VLS_functionalComponentArgsRest(__VLS_12));
    let __VLS_15;
    let __VLS_16;
    let __VLS_17;
    const __VLS_18 = {
        onClick: (...[$event]) => {
            __VLS_ctx.fetch(__VLS_ctx.searchForm);
        }
    };
    __VLS_14.slots.default;
    var __VLS_14;
}
{
    const { toolbar: __VLS_thisSlot } = __VLS_2.slots;
    const __VLS_19 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_20 = __VLS_asFunctionalComponent(__VLS_19, new __VLS_19({
        ...{ 'onClick': {} },
        type: "primary",
    }));
    const __VLS_21 = __VLS_20({
        ...{ 'onClick': {} },
        type: "primary",
    }, ...__VLS_functionalComponentArgsRest(__VLS_20));
    let __VLS_23;
    let __VLS_24;
    let __VLS_25;
    const __VLS_26 = {
        onClick: (__VLS_ctx.openCreate)
    };
    __VLS_asFunctionalDirective(__VLS_directives.vPermission)(null, { ...__VLS_directiveBindingRestFields, value: ('system.admin.edit') }, null, null);
    __VLS_22.slots.default;
    var __VLS_22;
}
{
    const { roles: __VLS_thisSlot } = __VLS_2.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    for (const [r] of __VLS_getVForSourceType((row.roles || []))) {
        const __VLS_27 = {}.ElTag;
        /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
        // @ts-ignore
        const __VLS_28 = __VLS_asFunctionalComponent(__VLS_27, new __VLS_27({
            key: (r),
            size: "small",
            ...{ style: {} },
        }));
        const __VLS_29 = __VLS_28({
            key: (r),
            size: "small",
            ...{ style: {} },
        }, ...__VLS_functionalComponentArgsRest(__VLS_28));
        __VLS_30.slots.default;
        (r);
        var __VLS_30;
    }
}
{
    const { status: __VLS_thisSlot } = __VLS_2.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_31 = {}.ElTag;
    /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
    // @ts-ignore
    const __VLS_32 = __VLS_asFunctionalComponent(__VLS_31, new __VLS_31({
        type: (__VLS_ctx.isAdminActive(row.status) ? 'success' : 'danger'),
        size: "small",
    }));
    const __VLS_33 = __VLS_32({
        type: (__VLS_ctx.isAdminActive(row.status) ? 'success' : 'danger'),
        size: "small",
    }, ...__VLS_functionalComponentArgsRest(__VLS_32));
    __VLS_34.slots.default;
    (__VLS_ctx.isAdminActive(row.status) ? '正常' : '禁用');
    var __VLS_34;
}
{
    const { actions: __VLS_thisSlot } = __VLS_2.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_35 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_36 = __VLS_asFunctionalComponent(__VLS_35, new __VLS_35({
        ...{ 'onClick': {} },
        text: true,
        type: "primary",
        size: "small",
    }));
    const __VLS_37 = __VLS_36({
        ...{ 'onClick': {} },
        text: true,
        type: "primary",
        size: "small",
    }, ...__VLS_functionalComponentArgsRest(__VLS_36));
    let __VLS_39;
    let __VLS_40;
    let __VLS_41;
    const __VLS_42 = {
        onClick: (...[$event]) => {
            __VLS_ctx.openEdit(row);
        }
    };
    __VLS_38.slots.default;
    var __VLS_38;
    const __VLS_43 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_44 = __VLS_asFunctionalComponent(__VLS_43, new __VLS_43({
        ...{ 'onClick': {} },
        text: true,
        type: (__VLS_ctx.isAdminActive(row.status) ? 'danger' : 'success'),
        size: "small",
    }));
    const __VLS_45 = __VLS_44({
        ...{ 'onClick': {} },
        text: true,
        type: (__VLS_ctx.isAdminActive(row.status) ? 'danger' : 'success'),
        size: "small",
    }, ...__VLS_functionalComponentArgsRest(__VLS_44));
    let __VLS_47;
    let __VLS_48;
    let __VLS_49;
    const __VLS_50 = {
        onClick: (...[$event]) => {
            __VLS_ctx.toggleStatus(row);
        }
    };
    __VLS_46.slots.default;
    (__VLS_ctx.isAdminActive(row.status) ? '禁用' : '启用');
    var __VLS_46;
    const __VLS_51 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_52 = __VLS_asFunctionalComponent(__VLS_51, new __VLS_51({
        ...{ 'onClick': {} },
        text: true,
        size: "small",
    }));
    const __VLS_53 = __VLS_52({
        ...{ 'onClick': {} },
        text: true,
        size: "small",
    }, ...__VLS_functionalComponentArgsRest(__VLS_52));
    let __VLS_55;
    let __VLS_56;
    let __VLS_57;
    const __VLS_58 = {
        onClick: (...[$event]) => {
            __VLS_ctx.openResetPwd(row);
        }
    };
    __VLS_54.slots.default;
    var __VLS_54;
}
var __VLS_2;
const __VLS_59 = {}.ElDialog;
/** @type {[typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, ]} */ ;
// @ts-ignore
const __VLS_60 = __VLS_asFunctionalComponent(__VLS_59, new __VLS_59({
    modelValue: (__VLS_ctx.dialogVisible),
    title: (__VLS_ctx.isEdit ? '编辑员工' : '新建员工'),
    width: "480px",
}));
const __VLS_61 = __VLS_60({
    modelValue: (__VLS_ctx.dialogVisible),
    title: (__VLS_ctx.isEdit ? '编辑员工' : '新建员工'),
    width: "480px",
}, ...__VLS_functionalComponentArgsRest(__VLS_60));
__VLS_62.slots.default;
const __VLS_63 = {}.ElForm;
/** @type {[typeof __VLS_components.ElForm, typeof __VLS_components.elForm, typeof __VLS_components.ElForm, typeof __VLS_components.elForm, ]} */ ;
// @ts-ignore
const __VLS_64 = __VLS_asFunctionalComponent(__VLS_63, new __VLS_63({
    ref: "formRef",
    model: (__VLS_ctx.form),
    labelWidth: "90px",
}));
const __VLS_65 = __VLS_64({
    ref: "formRef",
    model: (__VLS_ctx.form),
    labelWidth: "90px",
}, ...__VLS_functionalComponentArgsRest(__VLS_64));
/** @type {typeof __VLS_ctx.formRef} */ ;
var __VLS_67 = {};
__VLS_66.slots.default;
const __VLS_69 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_70 = __VLS_asFunctionalComponent(__VLS_69, new __VLS_69({
    label: "用户名",
    prop: "username",
    rules: ([{ required: true }]),
}));
const __VLS_71 = __VLS_70({
    label: "用户名",
    prop: "username",
    rules: ([{ required: true }]),
}, ...__VLS_functionalComponentArgsRest(__VLS_70));
__VLS_72.slots.default;
const __VLS_73 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_74 = __VLS_asFunctionalComponent(__VLS_73, new __VLS_73({
    modelValue: (__VLS_ctx.form.username),
    disabled: (__VLS_ctx.isEdit),
}));
const __VLS_75 = __VLS_74({
    modelValue: (__VLS_ctx.form.username),
    disabled: (__VLS_ctx.isEdit),
}, ...__VLS_functionalComponentArgsRest(__VLS_74));
var __VLS_72;
const __VLS_77 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_78 = __VLS_asFunctionalComponent(__VLS_77, new __VLS_77({
    label: "真实姓名",
    prop: "real_name",
}));
const __VLS_79 = __VLS_78({
    label: "真实姓名",
    prop: "real_name",
}, ...__VLS_functionalComponentArgsRest(__VLS_78));
__VLS_80.slots.default;
const __VLS_81 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_82 = __VLS_asFunctionalComponent(__VLS_81, new __VLS_81({
    modelValue: (__VLS_ctx.form.real_name),
}));
const __VLS_83 = __VLS_82({
    modelValue: (__VLS_ctx.form.real_name),
}, ...__VLS_functionalComponentArgsRest(__VLS_82));
var __VLS_80;
if (!__VLS_ctx.isEdit) {
    const __VLS_85 = {}.ElFormItem;
    /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
    // @ts-ignore
    const __VLS_86 = __VLS_asFunctionalComponent(__VLS_85, new __VLS_85({
        label: "密码",
        prop: "password",
        rules: (__VLS_ctx.passwordRules),
    }));
    const __VLS_87 = __VLS_86({
        label: "密码",
        prop: "password",
        rules: (__VLS_ctx.passwordRules),
    }, ...__VLS_functionalComponentArgsRest(__VLS_86));
    __VLS_88.slots.default;
    const __VLS_89 = {}.ElInput;
    /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
    // @ts-ignore
    const __VLS_90 = __VLS_asFunctionalComponent(__VLS_89, new __VLS_89({
        modelValue: (__VLS_ctx.form.password),
        type: "password",
        showPassword: true,
        placeholder: "至少12位，含字母、数字、特殊符号",
    }));
    const __VLS_91 = __VLS_90({
        modelValue: (__VLS_ctx.form.password),
        type: "password",
        showPassword: true,
        placeholder: "至少12位，含字母、数字、特殊符号",
    }, ...__VLS_functionalComponentArgsRest(__VLS_90));
    var __VLS_88;
}
const __VLS_93 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_94 = __VLS_asFunctionalComponent(__VLS_93, new __VLS_93({
    label: "角色",
}));
const __VLS_95 = __VLS_94({
    label: "角色",
}, ...__VLS_functionalComponentArgsRest(__VLS_94));
__VLS_96.slots.default;
const __VLS_97 = {}.ElCheckboxGroup;
/** @type {[typeof __VLS_components.ElCheckboxGroup, typeof __VLS_components.elCheckboxGroup, typeof __VLS_components.ElCheckboxGroup, typeof __VLS_components.elCheckboxGroup, ]} */ ;
// @ts-ignore
const __VLS_98 = __VLS_asFunctionalComponent(__VLS_97, new __VLS_97({
    modelValue: (__VLS_ctx.form.role_ids),
}));
const __VLS_99 = __VLS_98({
    modelValue: (__VLS_ctx.form.role_ids),
}, ...__VLS_functionalComponentArgsRest(__VLS_98));
__VLS_100.slots.default;
for (const [role] of __VLS_getVForSourceType((__VLS_ctx.roles))) {
    const __VLS_101 = {}.ElCheckbox;
    /** @type {[typeof __VLS_components.ElCheckbox, typeof __VLS_components.elCheckbox, typeof __VLS_components.ElCheckbox, typeof __VLS_components.elCheckbox, ]} */ ;
    // @ts-ignore
    const __VLS_102 = __VLS_asFunctionalComponent(__VLS_101, new __VLS_101({
        key: (role.id),
        value: (role.id),
    }));
    const __VLS_103 = __VLS_102({
        key: (role.id),
        value: (role.id),
    }, ...__VLS_functionalComponentArgsRest(__VLS_102));
    __VLS_104.slots.default;
    (role.name);
    var __VLS_104;
}
var __VLS_100;
var __VLS_96;
var __VLS_66;
{
    const { footer: __VLS_thisSlot } = __VLS_62.slots;
    const __VLS_105 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_106 = __VLS_asFunctionalComponent(__VLS_105, new __VLS_105({
        ...{ 'onClick': {} },
    }));
    const __VLS_107 = __VLS_106({
        ...{ 'onClick': {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_106));
    let __VLS_109;
    let __VLS_110;
    let __VLS_111;
    const __VLS_112 = {
        onClick: (...[$event]) => {
            __VLS_ctx.dialogVisible = false;
        }
    };
    __VLS_108.slots.default;
    var __VLS_108;
    const __VLS_113 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_114 = __VLS_asFunctionalComponent(__VLS_113, new __VLS_113({
        ...{ 'onClick': {} },
        type: "primary",
        loading: (__VLS_ctx.saving),
    }));
    const __VLS_115 = __VLS_114({
        ...{ 'onClick': {} },
        type: "primary",
        loading: (__VLS_ctx.saving),
    }, ...__VLS_functionalComponentArgsRest(__VLS_114));
    let __VLS_117;
    let __VLS_118;
    let __VLS_119;
    const __VLS_120 = {
        onClick: (__VLS_ctx.handleSave)
    };
    __VLS_116.slots.default;
    var __VLS_116;
}
var __VLS_62;
const __VLS_121 = {}.ElDialog;
/** @type {[typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, ]} */ ;
// @ts-ignore
const __VLS_122 = __VLS_asFunctionalComponent(__VLS_121, new __VLS_121({
    modelValue: (__VLS_ctx.pwdDialog.visible),
    title: "重置密码",
    width: "360px",
}));
const __VLS_123 = __VLS_122({
    modelValue: (__VLS_ctx.pwdDialog.visible),
    title: "重置密码",
    width: "360px",
}, ...__VLS_functionalComponentArgsRest(__VLS_122));
__VLS_124.slots.default;
const __VLS_125 = {}.ElForm;
/** @type {[typeof __VLS_components.ElForm, typeof __VLS_components.elForm, typeof __VLS_components.ElForm, typeof __VLS_components.elForm, ]} */ ;
// @ts-ignore
const __VLS_126 = __VLS_asFunctionalComponent(__VLS_125, new __VLS_125({
    ref: "pwdFormRef",
    model: (__VLS_ctx.pwdDialog),
    labelWidth: "0",
}));
const __VLS_127 = __VLS_126({
    ref: "pwdFormRef",
    model: (__VLS_ctx.pwdDialog),
    labelWidth: "0",
}, ...__VLS_functionalComponentArgsRest(__VLS_126));
/** @type {typeof __VLS_ctx.pwdFormRef} */ ;
var __VLS_129 = {};
__VLS_128.slots.default;
const __VLS_131 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_132 = __VLS_asFunctionalComponent(__VLS_131, new __VLS_131({
    prop: "new_password",
    rules: (__VLS_ctx.passwordRules),
}));
const __VLS_133 = __VLS_132({
    prop: "new_password",
    rules: (__VLS_ctx.passwordRules),
}, ...__VLS_functionalComponentArgsRest(__VLS_132));
__VLS_134.slots.default;
const __VLS_135 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_136 = __VLS_asFunctionalComponent(__VLS_135, new __VLS_135({
    modelValue: (__VLS_ctx.pwdDialog.new_password),
    type: "password",
    showPassword: true,
    placeholder: "至少12位，含字母、数字、特殊符号",
}));
const __VLS_137 = __VLS_136({
    modelValue: (__VLS_ctx.pwdDialog.new_password),
    type: "password",
    showPassword: true,
    placeholder: "至少12位，含字母、数字、特殊符号",
}, ...__VLS_functionalComponentArgsRest(__VLS_136));
var __VLS_134;
var __VLS_128;
{
    const { footer: __VLS_thisSlot } = __VLS_124.slots;
    const __VLS_139 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_140 = __VLS_asFunctionalComponent(__VLS_139, new __VLS_139({
        ...{ 'onClick': {} },
    }));
    const __VLS_141 = __VLS_140({
        ...{ 'onClick': {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_140));
    let __VLS_143;
    let __VLS_144;
    let __VLS_145;
    const __VLS_146 = {
        onClick: (...[$event]) => {
            __VLS_ctx.pwdDialog.visible = false;
        }
    };
    __VLS_142.slots.default;
    var __VLS_142;
    const __VLS_147 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_148 = __VLS_asFunctionalComponent(__VLS_147, new __VLS_147({
        ...{ 'onClick': {} },
        type: "primary",
        loading: (__VLS_ctx.pwdLoading),
    }));
    const __VLS_149 = __VLS_148({
        ...{ 'onClick': {} },
        type: "primary",
        loading: (__VLS_ctx.pwdLoading),
    }, ...__VLS_functionalComponentArgsRest(__VLS_148));
    let __VLS_151;
    let __VLS_152;
    let __VLS_153;
    const __VLS_154 = {
        onClick: (__VLS_ctx.handleResetPwd)
    };
    __VLS_150.slots.default;
    var __VLS_150;
}
var __VLS_124;
/** @type {__VLS_StyleScopedClasses['page-card']} */ ;
// @ts-ignore
var __VLS_68 = __VLS_67, __VLS_130 = __VLS_129;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            ProTable: ProTable,
            searchForm: searchForm,
            roles: roles,
            list: list,
            total: total,
            page: page,
            pageSize: pageSize,
            loading: loading,
            fetch: fetch,
            columns: columns,
            dialogVisible: dialogVisible,
            isEdit: isEdit,
            saving: saving,
            formRef: formRef,
            pwdFormRef: pwdFormRef,
            form: form,
            pwdDialog: pwdDialog,
            pwdLoading: pwdLoading,
            passwordRules: passwordRules,
            isAdminActive: isAdminActive,
            openCreate: openCreate,
            openEdit: openEdit,
            handleSave: handleSave,
            toggleStatus: toggleStatus,
            openResetPwd: openResetPwd,
            handleResetPwd: handleResetPwd,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
