import { ref, reactive, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { ElMessage, ElMessageBox } from 'element-plus';
import ProTable from '@/components/ProTable/index.vue';
import { useTable } from '@/composables/useTable';
import { getUserList, createUser, disableUser, enableUser } from '@/api/user';
import { formatTime } from '@/utils/format';
const router = useRouter();
const searchForm = ref({ phone: '', nickname: '', status: '' });
const { list, total, page, pageSize, loading, fetch } = useTable((params) => getUserList({ ...params, ...searchForm.value }));
const columns = [
    { label: '用户', slot: 'user', minWidth: 160 },
    { label: '手机号', prop: 'phone', width: 130 },
    { label: '状态', slot: 'status', width: 90, align: 'center' },
    { label: '注册时间', prop: 'created_at', width: 150, formatter: (r) => formatTime(r.created_at) },
    { label: '最近登录', prop: 'last_login_at', width: 150, formatter: (r) => formatTime(r.last_login_at) },
];
async function toggleStatus(row) {
    const action = row.status === 'active' ? '禁用' : '启用';
    await ElMessageBox.confirm(`确认${action}该用户？`, '提示', { type: 'warning' });
    if (row.status === 'active') {
        await disableUser(row.id);
    }
    else {
        await enableUser(row.id);
    }
    ElMessage.success(`${action}成功`);
    fetch(searchForm.value);
}
// 新建用户弹窗
const dialogVisible = ref(false);
const formRef = ref();
const createForm = reactive({
    phone: '',
    password: '',
    nickname: '',
});
const createLoading = ref(false);
const phoneRule = [
    { required: true, message: '请输入手机号', trigger: 'blur' },
    { pattern: /^1[3-9]\d{9}$/, message: '请输入正确的手机号', trigger: 'blur' },
];
const passwordRule = [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, message: '密码至少6位', trigger: 'blur' },
];
function openCreateDialog() {
    createForm.phone = '';
    createForm.password = '';
    createForm.nickname = '';
    dialogVisible.value = true;
}
async function handleCreate() {
    await formRef.value.validate();
    createLoading.value = true;
    try {
        await createUser({
            phone: createForm.phone,
            password: createForm.password,
            nickname: createForm.nickname || undefined,
        });
        ElMessage.success('创建成功');
        dialogVisible.value = false;
        fetch(searchForm.value);
    }
    catch (e) {
        ElMessage.error(e?.message || '创建失败');
    }
    finally {
        createLoading.value = false;
    }
}
onMounted(() => fetch(searchForm.value));
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
    const { toolbar: __VLS_thisSlot } = __VLS_2.slots;
    const __VLS_7 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_8 = __VLS_asFunctionalComponent(__VLS_7, new __VLS_7({
        ...{ 'onClick': {} },
        type: "primary",
    }));
    const __VLS_9 = __VLS_8({
        ...{ 'onClick': {} },
        type: "primary",
    }, ...__VLS_functionalComponentArgsRest(__VLS_8));
    let __VLS_11;
    let __VLS_12;
    let __VLS_13;
    const __VLS_14 = {
        onClick: (__VLS_ctx.openCreateDialog)
    };
    __VLS_10.slots.default;
    var __VLS_10;
}
{
    const { search: __VLS_thisSlot } = __VLS_2.slots;
    const __VLS_15 = {}.ElInput;
    /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
    // @ts-ignore
    const __VLS_16 = __VLS_asFunctionalComponent(__VLS_15, new __VLS_15({
        modelValue: (__VLS_ctx.searchForm.phone),
        placeholder: "手机号",
        clearable: true,
        ...{ style: {} },
    }));
    const __VLS_17 = __VLS_16({
        modelValue: (__VLS_ctx.searchForm.phone),
        placeholder: "手机号",
        clearable: true,
        ...{ style: {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_16));
    const __VLS_19 = {}.ElInput;
    /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
    // @ts-ignore
    const __VLS_20 = __VLS_asFunctionalComponent(__VLS_19, new __VLS_19({
        modelValue: (__VLS_ctx.searchForm.nickname),
        placeholder: "昵称",
        clearable: true,
        ...{ style: {} },
    }));
    const __VLS_21 = __VLS_20({
        modelValue: (__VLS_ctx.searchForm.nickname),
        placeholder: "昵称",
        clearable: true,
        ...{ style: {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_20));
    const __VLS_23 = {}.ElSelect;
    /** @type {[typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, ]} */ ;
    // @ts-ignore
    const __VLS_24 = __VLS_asFunctionalComponent(__VLS_23, new __VLS_23({
        modelValue: (__VLS_ctx.searchForm.status),
        placeholder: "状态",
        clearable: true,
        ...{ style: {} },
    }));
    const __VLS_25 = __VLS_24({
        modelValue: (__VLS_ctx.searchForm.status),
        placeholder: "状态",
        clearable: true,
        ...{ style: {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_24));
    __VLS_26.slots.default;
    const __VLS_27 = {}.ElOption;
    /** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
    // @ts-ignore
    const __VLS_28 = __VLS_asFunctionalComponent(__VLS_27, new __VLS_27({
        label: "正常",
        value: "active",
    }));
    const __VLS_29 = __VLS_28({
        label: "正常",
        value: "active",
    }, ...__VLS_functionalComponentArgsRest(__VLS_28));
    const __VLS_31 = {}.ElOption;
    /** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
    // @ts-ignore
    const __VLS_32 = __VLS_asFunctionalComponent(__VLS_31, new __VLS_31({
        label: "禁用",
        value: "disabled",
    }));
    const __VLS_33 = __VLS_32({
        label: "禁用",
        value: "disabled",
    }, ...__VLS_functionalComponentArgsRest(__VLS_32));
    var __VLS_26;
    const __VLS_35 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_36 = __VLS_asFunctionalComponent(__VLS_35, new __VLS_35({
        ...{ 'onClick': {} },
        type: "primary",
    }));
    const __VLS_37 = __VLS_36({
        ...{ 'onClick': {} },
        type: "primary",
    }, ...__VLS_functionalComponentArgsRest(__VLS_36));
    let __VLS_39;
    let __VLS_40;
    let __VLS_41;
    const __VLS_42 = {
        onClick: (...[$event]) => {
            __VLS_ctx.fetch(__VLS_ctx.searchForm);
        }
    };
    __VLS_38.slots.default;
    var __VLS_38;
}
{
    const { user: __VLS_thisSlot } = __VLS_2.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ style: {} },
    });
    const __VLS_43 = {}.ElAvatar;
    /** @type {[typeof __VLS_components.ElAvatar, typeof __VLS_components.elAvatar, typeof __VLS_components.ElAvatar, typeof __VLS_components.elAvatar, ]} */ ;
    // @ts-ignore
    const __VLS_44 = __VLS_asFunctionalComponent(__VLS_43, new __VLS_43({
        size: (32),
        src: (row.avatar),
    }));
    const __VLS_45 = __VLS_44({
        size: (32),
        src: (row.avatar),
    }, ...__VLS_functionalComponentArgsRest(__VLS_44));
    __VLS_46.slots.default;
    (row.nickname?.charAt(0));
    var __VLS_46;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
    (row.nickname || '未知');
}
{
    const { status: __VLS_thisSlot } = __VLS_2.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_47 = {}.ElTag;
    /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
    // @ts-ignore
    const __VLS_48 = __VLS_asFunctionalComponent(__VLS_47, new __VLS_47({
        type: (row.status === 'active' ? 'success' : 'danger'),
        size: "small",
    }));
    const __VLS_49 = __VLS_48({
        type: (row.status === 'active' ? 'success' : 'danger'),
        size: "small",
    }, ...__VLS_functionalComponentArgsRest(__VLS_48));
    __VLS_50.slots.default;
    (row.status === 'active' ? '正常' : '禁用');
    var __VLS_50;
}
{
    const { actions: __VLS_thisSlot } = __VLS_2.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_51 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_52 = __VLS_asFunctionalComponent(__VLS_51, new __VLS_51({
        ...{ 'onClick': {} },
        text: true,
        type: "primary",
        size: "small",
    }));
    const __VLS_53 = __VLS_52({
        ...{ 'onClick': {} },
        text: true,
        type: "primary",
        size: "small",
    }, ...__VLS_functionalComponentArgsRest(__VLS_52));
    let __VLS_55;
    let __VLS_56;
    let __VLS_57;
    const __VLS_58 = {
        onClick: (...[$event]) => {
            __VLS_ctx.router.push(`/user/${row.id}`);
        }
    };
    __VLS_54.slots.default;
    var __VLS_54;
    const __VLS_59 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_60 = __VLS_asFunctionalComponent(__VLS_59, new __VLS_59({
        ...{ 'onClick': {} },
        text: true,
        type: (row.status === 'active' ? 'danger' : 'success'),
        size: "small",
    }));
    const __VLS_61 = __VLS_60({
        ...{ 'onClick': {} },
        text: true,
        type: (row.status === 'active' ? 'danger' : 'success'),
        size: "small",
    }, ...__VLS_functionalComponentArgsRest(__VLS_60));
    let __VLS_63;
    let __VLS_64;
    let __VLS_65;
    const __VLS_66 = {
        onClick: (...[$event]) => {
            __VLS_ctx.toggleStatus(row);
        }
    };
    __VLS_62.slots.default;
    (row.status === 'active' ? '禁用' : '启用');
    var __VLS_62;
}
var __VLS_2;
const __VLS_67 = {}.ElDialog;
/** @type {[typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, ]} */ ;
// @ts-ignore
const __VLS_68 = __VLS_asFunctionalComponent(__VLS_67, new __VLS_67({
    modelValue: (__VLS_ctx.dialogVisible),
    title: "新建用户",
    width: "420px",
    closeOnClickModal: (false),
}));
const __VLS_69 = __VLS_68({
    modelValue: (__VLS_ctx.dialogVisible),
    title: "新建用户",
    width: "420px",
    closeOnClickModal: (false),
}, ...__VLS_functionalComponentArgsRest(__VLS_68));
__VLS_70.slots.default;
const __VLS_71 = {}.ElForm;
/** @type {[typeof __VLS_components.ElForm, typeof __VLS_components.elForm, typeof __VLS_components.ElForm, typeof __VLS_components.elForm, ]} */ ;
// @ts-ignore
const __VLS_72 = __VLS_asFunctionalComponent(__VLS_71, new __VLS_71({
    ref: "formRef",
    model: (__VLS_ctx.createForm),
    labelWidth: "80px",
}));
const __VLS_73 = __VLS_72({
    ref: "formRef",
    model: (__VLS_ctx.createForm),
    labelWidth: "80px",
}, ...__VLS_functionalComponentArgsRest(__VLS_72));
/** @type {typeof __VLS_ctx.formRef} */ ;
var __VLS_75 = {};
__VLS_74.slots.default;
const __VLS_77 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_78 = __VLS_asFunctionalComponent(__VLS_77, new __VLS_77({
    label: "手机号",
    prop: "phone",
    rules: (__VLS_ctx.phoneRule),
}));
const __VLS_79 = __VLS_78({
    label: "手机号",
    prop: "phone",
    rules: (__VLS_ctx.phoneRule),
}, ...__VLS_functionalComponentArgsRest(__VLS_78));
__VLS_80.slots.default;
const __VLS_81 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_82 = __VLS_asFunctionalComponent(__VLS_81, new __VLS_81({
    modelValue: (__VLS_ctx.createForm.phone),
    placeholder: "请输入手机号",
    maxlength: "11",
}));
const __VLS_83 = __VLS_82({
    modelValue: (__VLS_ctx.createForm.phone),
    placeholder: "请输入手机号",
    maxlength: "11",
}, ...__VLS_functionalComponentArgsRest(__VLS_82));
var __VLS_80;
const __VLS_85 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_86 = __VLS_asFunctionalComponent(__VLS_85, new __VLS_85({
    label: "密码",
    prop: "password",
    rules: (__VLS_ctx.passwordRule),
}));
const __VLS_87 = __VLS_86({
    label: "密码",
    prop: "password",
    rules: (__VLS_ctx.passwordRule),
}, ...__VLS_functionalComponentArgsRest(__VLS_86));
__VLS_88.slots.default;
const __VLS_89 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_90 = __VLS_asFunctionalComponent(__VLS_89, new __VLS_89({
    modelValue: (__VLS_ctx.createForm.password),
    type: "password",
    placeholder: "请输入密码（至少6位）",
    showPassword: true,
}));
const __VLS_91 = __VLS_90({
    modelValue: (__VLS_ctx.createForm.password),
    type: "password",
    placeholder: "请输入密码（至少6位）",
    showPassword: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_90));
var __VLS_88;
const __VLS_93 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_94 = __VLS_asFunctionalComponent(__VLS_93, new __VLS_93({
    label: "昵称",
    prop: "nickname",
}));
const __VLS_95 = __VLS_94({
    label: "昵称",
    prop: "nickname",
}, ...__VLS_functionalComponentArgsRest(__VLS_94));
__VLS_96.slots.default;
const __VLS_97 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_98 = __VLS_asFunctionalComponent(__VLS_97, new __VLS_97({
    modelValue: (__VLS_ctx.createForm.nickname),
    placeholder: "选填，默认使用手机后4位",
}));
const __VLS_99 = __VLS_98({
    modelValue: (__VLS_ctx.createForm.nickname),
    placeholder: "选填，默认使用手机后4位",
}, ...__VLS_functionalComponentArgsRest(__VLS_98));
var __VLS_96;
var __VLS_74;
{
    const { footer: __VLS_thisSlot } = __VLS_70.slots;
    const __VLS_101 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_102 = __VLS_asFunctionalComponent(__VLS_101, new __VLS_101({
        ...{ 'onClick': {} },
    }));
    const __VLS_103 = __VLS_102({
        ...{ 'onClick': {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_102));
    let __VLS_105;
    let __VLS_106;
    let __VLS_107;
    const __VLS_108 = {
        onClick: (...[$event]) => {
            __VLS_ctx.dialogVisible = false;
        }
    };
    __VLS_104.slots.default;
    var __VLS_104;
    const __VLS_109 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_110 = __VLS_asFunctionalComponent(__VLS_109, new __VLS_109({
        ...{ 'onClick': {} },
        type: "primary",
        loading: (__VLS_ctx.createLoading),
    }));
    const __VLS_111 = __VLS_110({
        ...{ 'onClick': {} },
        type: "primary",
        loading: (__VLS_ctx.createLoading),
    }, ...__VLS_functionalComponentArgsRest(__VLS_110));
    let __VLS_113;
    let __VLS_114;
    let __VLS_115;
    const __VLS_116 = {
        onClick: (__VLS_ctx.handleCreate)
    };
    __VLS_112.slots.default;
    var __VLS_112;
}
var __VLS_70;
/** @type {__VLS_StyleScopedClasses['page-card']} */ ;
// @ts-ignore
var __VLS_76 = __VLS_75;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            ProTable: ProTable,
            router: router,
            searchForm: searchForm,
            list: list,
            total: total,
            page: page,
            pageSize: pageSize,
            loading: loading,
            fetch: fetch,
            columns: columns,
            toggleStatus: toggleStatus,
            dialogVisible: dialogVisible,
            formRef: formRef,
            createForm: createForm,
            createLoading: createLoading,
            phoneRule: phoneRule,
            passwordRule: passwordRule,
            openCreateDialog: openCreateDialog,
            handleCreate: handleCreate,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
