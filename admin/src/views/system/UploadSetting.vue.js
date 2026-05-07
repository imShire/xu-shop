import { computed, onMounted, reactive, ref } from 'vue';
import { ElMessage } from 'element-plus';
import { getUploadSettings, probeUploadSettings, testUploadSettings, updateUploadSettings } from '@/api/system';
import { useAuthStore } from '@/stores/auth';
const auth = useAuthStore();
const loading = ref(false);
const saving = ref(false);
const testing = ref(false);
const probing = ref(false);
const formRef = ref();
const probeInputRef = ref();
const probeResult = ref('');
const form = reactive({
    driver: 'local',
    public_base_url: 'http://localhost:8080/uploads',
    max_size_mb: 10,
    local_dir: 'uploads',
    s3_vendor: 'generic',
    s3_endpoint: '',
    s3_region: '',
    s3_bucket: '',
    s3_prefix: '',
    s3_force_path_style: false,
    s3_access_key: '',
    s3_secret_key: '',
    s3_access_key_set: false,
    s3_secret_key_set: false,
});
const canEdit = computed(() => auth.isSuperAdmin || auth.perms.includes('system.upload.edit'));
const rules = {
    driver: [{ required: true, message: '请选择上传驱动', trigger: 'change' }],
    public_base_url: [{ required: true, message: '请填写上传访问地址', trigger: 'blur' }],
    max_size_mb: [{ required: true, message: '请填写上传大小限制', trigger: 'change' }],
    local_dir: [{
            validator: (_rule, value, callback) => {
                if (form.driver === 'local' && !value)
                    return callback(new Error('请填写本地上传目录'));
                callback();
            },
            trigger: 'blur',
        }],
    s3_endpoint: [{
            validator: (_rule, value, callback) => {
                if (form.driver === 's3' && !value)
                    return callback(new Error('请填写 S3 Endpoint'));
                callback();
            },
            trigger: 'blur',
        }],
    s3_bucket: [{
            validator: (_rule, value, callback) => {
                if (form.driver === 's3' && !value)
                    return callback(new Error('请填写 S3 Bucket'));
                callback();
            },
            trigger: 'blur',
        }],
    s3_access_key: [{
            validator: (_rule, value, callback) => {
                if (form.driver === 's3' && !value && !form.s3_access_key_set)
                    return callback(new Error('请填写 S3 Access Key'));
                callback();
            },
            trigger: 'blur',
        }],
    s3_secret_key: [{
            validator: (_rule, value, callback) => {
                if (form.driver === 's3' && !value && !form.s3_secret_key_set)
                    return callback(new Error('请填写 S3 Secret Key'));
                callback();
            },
            trigger: 'blur',
        }],
};
async function loadData() {
    loading.value = true;
    try {
        const res = await getUploadSettings();
        Object.assign(form, res, { s3_access_key: '', s3_secret_key: '' });
    }
    finally {
        loading.value = false;
    }
}
async function handleSave() {
    await formRef.value?.validate();
    saving.value = true;
    try {
        await updateUploadSettings(toPayload());
        ElMessage.success('保存成功');
        await loadData();
    }
    finally {
        saving.value = false;
    }
}
async function handleTest() {
    await formRef.value?.validate();
    testing.value = true;
    try {
        await testUploadSettings(toPayload());
        ElMessage.success('连接测试成功');
    }
    finally {
        testing.value = false;
    }
}
async function handleProbeChange(event) {
    const input = event.target;
    const file = input.files?.[0];
    if (!file)
        return;
    probing.value = true;
    probeResult.value = '';
    try {
        const res = await probeUploadSettings(file);
        probeResult.value = res.url;
        ElMessage.success('测试上传成功');
    }
    finally {
        probing.value = false;
        input.value = '';
    }
}
function openProbePicker() {
    probeInputRef.value?.click();
}
function toPayload() {
    return {
        driver: form.driver,
        public_base_url: form.public_base_url,
        max_size_mb: Number(form.max_size_mb),
        local_dir: form.local_dir,
        s3_vendor: form.s3_vendor,
        s3_endpoint: form.s3_endpoint,
        s3_region: form.s3_region,
        s3_bucket: form.s3_bucket,
        s3_access_key: form.s3_access_key,
        s3_secret_key: form.s3_secret_key,
        s3_prefix: form.s3_prefix,
        s3_force_path_style: form.s3_force_path_style,
    };
}
onMounted(loadData);
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
// CSS variable injection 
// CSS variable injection end 
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "page-card" },
});
__VLS_asFunctionalDirective(__VLS_directives.vLoading)(null, { ...__VLS_directiveBindingRestFields, value: (__VLS_ctx.loading) }, null, null);
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "page-head" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.h3, __VLS_intrinsicElements.h3)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "page-actions" },
});
const __VLS_0 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_1 = __VLS_asFunctionalComponent(__VLS_0, new __VLS_0({
    ...{ 'onClick': {} },
    loading: (__VLS_ctx.testing),
    disabled: (!__VLS_ctx.canEdit),
}));
const __VLS_2 = __VLS_1({
    ...{ 'onClick': {} },
    loading: (__VLS_ctx.testing),
    disabled: (!__VLS_ctx.canEdit),
}, ...__VLS_functionalComponentArgsRest(__VLS_1));
let __VLS_4;
let __VLS_5;
let __VLS_6;
const __VLS_7 = {
    onClick: (__VLS_ctx.handleTest)
};
__VLS_3.slots.default;
var __VLS_3;
const __VLS_8 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_9 = __VLS_asFunctionalComponent(__VLS_8, new __VLS_8({
    ...{ 'onClick': {} },
    type: "primary",
    loading: (__VLS_ctx.saving),
    disabled: (!__VLS_ctx.canEdit),
}));
const __VLS_10 = __VLS_9({
    ...{ 'onClick': {} },
    type: "primary",
    loading: (__VLS_ctx.saving),
    disabled: (!__VLS_ctx.canEdit),
}, ...__VLS_functionalComponentArgsRest(__VLS_9));
let __VLS_12;
let __VLS_13;
let __VLS_14;
const __VLS_15 = {
    onClick: (__VLS_ctx.handleSave)
};
__VLS_11.slots.default;
var __VLS_11;
if (!__VLS_ctx.canEdit) {
    const __VLS_16 = {}.ElAlert;
    /** @type {[typeof __VLS_components.ElAlert, typeof __VLS_components.elAlert, ]} */ ;
    // @ts-ignore
    const __VLS_17 = __VLS_asFunctionalComponent(__VLS_16, new __VLS_16({
        title: "当前账号只有查看权限，不能修改上传设置。",
        type: "warning",
        showIcon: true,
        closable: (false),
        ...{ style: {} },
    }));
    const __VLS_18 = __VLS_17({
        title: "当前账号只有查看权限，不能修改上传设置。",
        type: "warning",
        showIcon: true,
        closable: (false),
        ...{ style: {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_17));
}
const __VLS_20 = {}.ElForm;
/** @type {[typeof __VLS_components.ElForm, typeof __VLS_components.elForm, typeof __VLS_components.ElForm, typeof __VLS_components.elForm, ]} */ ;
// @ts-ignore
const __VLS_21 = __VLS_asFunctionalComponent(__VLS_20, new __VLS_20({
    ref: "formRef",
    model: (__VLS_ctx.form),
    rules: (__VLS_ctx.rules),
    labelWidth: "140px",
    ...{ class: "upload-form" },
}));
const __VLS_22 = __VLS_21({
    ref: "formRef",
    model: (__VLS_ctx.form),
    rules: (__VLS_ctx.rules),
    labelWidth: "140px",
    ...{ class: "upload-form" },
}, ...__VLS_functionalComponentArgsRest(__VLS_21));
/** @type {typeof __VLS_ctx.formRef} */ ;
var __VLS_24 = {};
__VLS_23.slots.default;
const __VLS_26 = {}.ElCard;
/** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
// @ts-ignore
const __VLS_27 = __VLS_asFunctionalComponent(__VLS_26, new __VLS_26({
    shadow: "never",
    ...{ class: "setting-card" },
}));
const __VLS_28 = __VLS_27({
    shadow: "never",
    ...{ class: "setting-card" },
}, ...__VLS_functionalComponentArgsRest(__VLS_27));
__VLS_29.slots.default;
{
    const { header: __VLS_thisSlot } = __VLS_29.slots;
}
const __VLS_30 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_31 = __VLS_asFunctionalComponent(__VLS_30, new __VLS_30({
    label: "上传驱动",
    prop: "driver",
}));
const __VLS_32 = __VLS_31({
    label: "上传驱动",
    prop: "driver",
}, ...__VLS_functionalComponentArgsRest(__VLS_31));
__VLS_33.slots.default;
const __VLS_34 = {}.ElRadioGroup;
/** @type {[typeof __VLS_components.ElRadioGroup, typeof __VLS_components.elRadioGroup, typeof __VLS_components.ElRadioGroup, typeof __VLS_components.elRadioGroup, ]} */ ;
// @ts-ignore
const __VLS_35 = __VLS_asFunctionalComponent(__VLS_34, new __VLS_34({
    modelValue: (__VLS_ctx.form.driver),
}));
const __VLS_36 = __VLS_35({
    modelValue: (__VLS_ctx.form.driver),
}, ...__VLS_functionalComponentArgsRest(__VLS_35));
__VLS_37.slots.default;
const __VLS_38 = {}.ElRadio;
/** @type {[typeof __VLS_components.ElRadio, typeof __VLS_components.elRadio, typeof __VLS_components.ElRadio, typeof __VLS_components.elRadio, ]} */ ;
// @ts-ignore
const __VLS_39 = __VLS_asFunctionalComponent(__VLS_38, new __VLS_38({
    value: "local",
}));
const __VLS_40 = __VLS_39({
    value: "local",
}, ...__VLS_functionalComponentArgsRest(__VLS_39));
__VLS_41.slots.default;
var __VLS_41;
const __VLS_42 = {}.ElRadio;
/** @type {[typeof __VLS_components.ElRadio, typeof __VLS_components.elRadio, typeof __VLS_components.ElRadio, typeof __VLS_components.elRadio, ]} */ ;
// @ts-ignore
const __VLS_43 = __VLS_asFunctionalComponent(__VLS_42, new __VLS_42({
    value: "s3",
}));
const __VLS_44 = __VLS_43({
    value: "s3",
}, ...__VLS_functionalComponentArgsRest(__VLS_43));
__VLS_45.slots.default;
var __VLS_45;
var __VLS_37;
var __VLS_33;
const __VLS_46 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_47 = __VLS_asFunctionalComponent(__VLS_46, new __VLS_46({
    label: "访问前缀",
    prop: "public_base_url",
}));
const __VLS_48 = __VLS_47({
    label: "访问前缀",
    prop: "public_base_url",
}, ...__VLS_functionalComponentArgsRest(__VLS_47));
__VLS_49.slots.default;
const __VLS_50 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_51 = __VLS_asFunctionalComponent(__VLS_50, new __VLS_50({
    modelValue: (__VLS_ctx.form.public_base_url),
    placeholder: "例如 http://localhost:8080/uploads 或 https://cdn.example.com",
}));
const __VLS_52 = __VLS_51({
    modelValue: (__VLS_ctx.form.public_base_url),
    placeholder: "例如 http://localhost:8080/uploads 或 https://cdn.example.com",
}, ...__VLS_functionalComponentArgsRest(__VLS_51));
var __VLS_49;
const __VLS_54 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_55 = __VLS_asFunctionalComponent(__VLS_54, new __VLS_54({
    label: "大小限制",
    prop: "max_size_mb",
}));
const __VLS_56 = __VLS_55({
    label: "大小限制",
    prop: "max_size_mb",
}, ...__VLS_functionalComponentArgsRest(__VLS_55));
__VLS_57.slots.default;
const __VLS_58 = {}.ElInputNumber;
/** @type {[typeof __VLS_components.ElInputNumber, typeof __VLS_components.elInputNumber, ]} */ ;
// @ts-ignore
const __VLS_59 = __VLS_asFunctionalComponent(__VLS_58, new __VLS_58({
    modelValue: (__VLS_ctx.form.max_size_mb),
    min: (1),
    max: (50),
}));
const __VLS_60 = __VLS_59({
    modelValue: (__VLS_ctx.form.max_size_mb),
    min: (1),
    max: (50),
}, ...__VLS_functionalComponentArgsRest(__VLS_59));
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
    ...{ class: "field-tip" },
});
var __VLS_57;
var __VLS_29;
if (__VLS_ctx.form.driver === 'local') {
    const __VLS_62 = {}.ElCard;
    /** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
    // @ts-ignore
    const __VLS_63 = __VLS_asFunctionalComponent(__VLS_62, new __VLS_62({
        shadow: "never",
        ...{ class: "setting-card" },
    }));
    const __VLS_64 = __VLS_63({
        shadow: "never",
        ...{ class: "setting-card" },
    }, ...__VLS_functionalComponentArgsRest(__VLS_63));
    __VLS_65.slots.default;
    {
        const { header: __VLS_thisSlot } = __VLS_65.slots;
    }
    const __VLS_66 = {}.ElFormItem;
    /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
    // @ts-ignore
    const __VLS_67 = __VLS_asFunctionalComponent(__VLS_66, new __VLS_66({
        label: "保存目录",
        prop: "local_dir",
    }));
    const __VLS_68 = __VLS_67({
        label: "保存目录",
        prop: "local_dir",
    }, ...__VLS_functionalComponentArgsRest(__VLS_67));
    __VLS_69.slots.default;
    const __VLS_70 = {}.ElInput;
    /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
    // @ts-ignore
    const __VLS_71 = __VLS_asFunctionalComponent(__VLS_70, new __VLS_70({
        modelValue: (__VLS_ctx.form.local_dir),
        placeholder: "例如 uploads",
    }));
    const __VLS_72 = __VLS_71({
        modelValue: (__VLS_ctx.form.local_dir),
        placeholder: "例如 uploads",
    }, ...__VLS_functionalComponentArgsRest(__VLS_71));
    var __VLS_69;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "hint-block" },
    });
    var __VLS_65;
}
else {
    const __VLS_74 = {}.ElCard;
    /** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
    // @ts-ignore
    const __VLS_75 = __VLS_asFunctionalComponent(__VLS_74, new __VLS_74({
        shadow: "never",
        ...{ class: "setting-card" },
    }));
    const __VLS_76 = __VLS_75({
        shadow: "never",
        ...{ class: "setting-card" },
    }, ...__VLS_functionalComponentArgsRest(__VLS_75));
    __VLS_77.slots.default;
    {
        const { header: __VLS_thisSlot } = __VLS_77.slots;
    }
    const __VLS_78 = {}.ElFormItem;
    /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
    // @ts-ignore
    const __VLS_79 = __VLS_asFunctionalComponent(__VLS_78, new __VLS_78({
        label: "存储厂商",
    }));
    const __VLS_80 = __VLS_79({
        label: "存储厂商",
    }, ...__VLS_functionalComponentArgsRest(__VLS_79));
    __VLS_81.slots.default;
    const __VLS_82 = {}.ElSelect;
    /** @type {[typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, ]} */ ;
    // @ts-ignore
    const __VLS_83 = __VLS_asFunctionalComponent(__VLS_82, new __VLS_82({
        modelValue: (__VLS_ctx.form.s3_vendor),
        ...{ style: {} },
    }));
    const __VLS_84 = __VLS_83({
        modelValue: (__VLS_ctx.form.s3_vendor),
        ...{ style: {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_83));
    __VLS_85.slots.default;
    const __VLS_86 = {}.ElOption;
    /** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
    // @ts-ignore
    const __VLS_87 = __VLS_asFunctionalComponent(__VLS_86, new __VLS_86({
        label: "通用 S3",
        value: "generic",
    }));
    const __VLS_88 = __VLS_87({
        label: "通用 S3",
        value: "generic",
    }, ...__VLS_functionalComponentArgsRest(__VLS_87));
    const __VLS_90 = {}.ElOption;
    /** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
    // @ts-ignore
    const __VLS_91 = __VLS_asFunctionalComponent(__VLS_90, new __VLS_90({
        label: "七牛 Kodo S3",
        value: "qiniu",
    }));
    const __VLS_92 = __VLS_91({
        label: "七牛 Kodo S3",
        value: "qiniu",
    }, ...__VLS_functionalComponentArgsRest(__VLS_91));
    var __VLS_85;
    var __VLS_81;
    const __VLS_94 = {}.ElFormItem;
    /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
    // @ts-ignore
    const __VLS_95 = __VLS_asFunctionalComponent(__VLS_94, new __VLS_94({
        label: "Endpoint",
        prop: "s3_endpoint",
    }));
    const __VLS_96 = __VLS_95({
        label: "Endpoint",
        prop: "s3_endpoint",
    }, ...__VLS_functionalComponentArgsRest(__VLS_95));
    __VLS_97.slots.default;
    const __VLS_98 = {}.ElInput;
    /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
    // @ts-ignore
    const __VLS_99 = __VLS_asFunctionalComponent(__VLS_98, new __VLS_98({
        modelValue: (__VLS_ctx.form.s3_endpoint),
        placeholder: "例如 https://s3-cn-east-1.qiniucs.com",
    }));
    const __VLS_100 = __VLS_99({
        modelValue: (__VLS_ctx.form.s3_endpoint),
        placeholder: "例如 https://s3-cn-east-1.qiniucs.com",
    }, ...__VLS_functionalComponentArgsRest(__VLS_99));
    var __VLS_97;
    const __VLS_102 = {}.ElFormItem;
    /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
    // @ts-ignore
    const __VLS_103 = __VLS_asFunctionalComponent(__VLS_102, new __VLS_102({
        label: "Region",
    }));
    const __VLS_104 = __VLS_103({
        label: "Region",
    }, ...__VLS_functionalComponentArgsRest(__VLS_103));
    __VLS_105.slots.default;
    const __VLS_106 = {}.ElInput;
    /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
    // @ts-ignore
    const __VLS_107 = __VLS_asFunctionalComponent(__VLS_106, new __VLS_106({
        modelValue: (__VLS_ctx.form.s3_region),
        placeholder: "可选，例如 z0 / us-east-1",
    }));
    const __VLS_108 = __VLS_107({
        modelValue: (__VLS_ctx.form.s3_region),
        placeholder: "可选，例如 z0 / us-east-1",
    }, ...__VLS_functionalComponentArgsRest(__VLS_107));
    var __VLS_105;
    const __VLS_110 = {}.ElFormItem;
    /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
    // @ts-ignore
    const __VLS_111 = __VLS_asFunctionalComponent(__VLS_110, new __VLS_110({
        label: "Bucket",
        prop: "s3_bucket",
    }));
    const __VLS_112 = __VLS_111({
        label: "Bucket",
        prop: "s3_bucket",
    }, ...__VLS_functionalComponentArgsRest(__VLS_111));
    __VLS_113.slots.default;
    const __VLS_114 = {}.ElInput;
    /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
    // @ts-ignore
    const __VLS_115 = __VLS_asFunctionalComponent(__VLS_114, new __VLS_114({
        modelValue: (__VLS_ctx.form.s3_bucket),
    }));
    const __VLS_116 = __VLS_115({
        modelValue: (__VLS_ctx.form.s3_bucket),
    }, ...__VLS_functionalComponentArgsRest(__VLS_115));
    var __VLS_113;
    const __VLS_118 = {}.ElFormItem;
    /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
    // @ts-ignore
    const __VLS_119 = __VLS_asFunctionalComponent(__VLS_118, new __VLS_118({
        label: "Access Key",
        prop: "s3_access_key",
    }));
    const __VLS_120 = __VLS_119({
        label: "Access Key",
        prop: "s3_access_key",
    }, ...__VLS_functionalComponentArgsRest(__VLS_119));
    __VLS_121.slots.default;
    const __VLS_122 = {}.ElInput;
    /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
    // @ts-ignore
    const __VLS_123 = __VLS_asFunctionalComponent(__VLS_122, new __VLS_122({
        modelValue: (__VLS_ctx.form.s3_access_key),
        placeholder: (__VLS_ctx.form.s3_access_key_set ? '已配置，留空则保持不变' : '请输入 Access Key'),
    }));
    const __VLS_124 = __VLS_123({
        modelValue: (__VLS_ctx.form.s3_access_key),
        placeholder: (__VLS_ctx.form.s3_access_key_set ? '已配置，留空则保持不变' : '请输入 Access Key'),
    }, ...__VLS_functionalComponentArgsRest(__VLS_123));
    var __VLS_121;
    const __VLS_126 = {}.ElFormItem;
    /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
    // @ts-ignore
    const __VLS_127 = __VLS_asFunctionalComponent(__VLS_126, new __VLS_126({
        label: "Secret Key",
        prop: "s3_secret_key",
    }));
    const __VLS_128 = __VLS_127({
        label: "Secret Key",
        prop: "s3_secret_key",
    }, ...__VLS_functionalComponentArgsRest(__VLS_127));
    __VLS_129.slots.default;
    const __VLS_130 = {}.ElInput;
    /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
    // @ts-ignore
    const __VLS_131 = __VLS_asFunctionalComponent(__VLS_130, new __VLS_130({
        modelValue: (__VLS_ctx.form.s3_secret_key),
        type: "password",
        showPassword: true,
        placeholder: (__VLS_ctx.form.s3_secret_key_set ? '已配置，留空则保持不变' : '请输入 Secret Key'),
    }));
    const __VLS_132 = __VLS_131({
        modelValue: (__VLS_ctx.form.s3_secret_key),
        type: "password",
        showPassword: true,
        placeholder: (__VLS_ctx.form.s3_secret_key_set ? '已配置，留空则保持不变' : '请输入 Secret Key'),
    }, ...__VLS_functionalComponentArgsRest(__VLS_131));
    var __VLS_129;
    const __VLS_134 = {}.ElFormItem;
    /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
    // @ts-ignore
    const __VLS_135 = __VLS_asFunctionalComponent(__VLS_134, new __VLS_134({
        label: "对象前缀",
    }));
    const __VLS_136 = __VLS_135({
        label: "对象前缀",
    }, ...__VLS_functionalComponentArgsRest(__VLS_135));
    __VLS_137.slots.default;
    const __VLS_138 = {}.ElInput;
    /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
    // @ts-ignore
    const __VLS_139 = __VLS_asFunctionalComponent(__VLS_138, new __VLS_138({
        modelValue: (__VLS_ctx.form.s3_prefix),
        placeholder: "可选，例如 media",
    }));
    const __VLS_140 = __VLS_139({
        modelValue: (__VLS_ctx.form.s3_prefix),
        placeholder: "可选，例如 media",
    }, ...__VLS_functionalComponentArgsRest(__VLS_139));
    var __VLS_137;
    const __VLS_142 = {}.ElFormItem;
    /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
    // @ts-ignore
    const __VLS_143 = __VLS_asFunctionalComponent(__VLS_142, new __VLS_142({
        label: "Path Style",
    }));
    const __VLS_144 = __VLS_143({
        label: "Path Style",
    }, ...__VLS_functionalComponentArgsRest(__VLS_143));
    __VLS_145.slots.default;
    const __VLS_146 = {}.ElSwitch;
    /** @type {[typeof __VLS_components.ElSwitch, typeof __VLS_components.elSwitch, ]} */ ;
    // @ts-ignore
    const __VLS_147 = __VLS_asFunctionalComponent(__VLS_146, new __VLS_146({
        modelValue: (__VLS_ctx.form.s3_force_path_style),
    }));
    const __VLS_148 = __VLS_147({
        modelValue: (__VLS_ctx.form.s3_force_path_style),
    }, ...__VLS_functionalComponentArgsRest(__VLS_147));
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ class: "field-tip" },
    });
    var __VLS_145;
    var __VLS_77;
}
const __VLS_150 = {}.ElCard;
/** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
// @ts-ignore
const __VLS_151 = __VLS_asFunctionalComponent(__VLS_150, new __VLS_150({
    shadow: "never",
    ...{ class: "setting-card" },
}));
const __VLS_152 = __VLS_151({
    shadow: "never",
    ...{ class: "setting-card" },
}, ...__VLS_functionalComponentArgsRest(__VLS_151));
__VLS_153.slots.default;
{
    const { header: __VLS_thisSlot } = __VLS_153.slots;
}
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "probe-panel" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "probe-copy" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "probe-actions" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.input)({
    ...{ onChange: (__VLS_ctx.handleProbeChange) },
    ref: "probeInputRef",
    type: "file",
    accept: "image/*",
    ...{ style: {} },
});
/** @type {typeof __VLS_ctx.probeInputRef} */ ;
const __VLS_154 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_155 = __VLS_asFunctionalComponent(__VLS_154, new __VLS_154({
    ...{ 'onClick': {} },
    loading: (__VLS_ctx.probing),
    disabled: (!__VLS_ctx.canEdit),
}));
const __VLS_156 = __VLS_155({
    ...{ 'onClick': {} },
    loading: (__VLS_ctx.probing),
    disabled: (!__VLS_ctx.canEdit),
}, ...__VLS_functionalComponentArgsRest(__VLS_155));
let __VLS_158;
let __VLS_159;
let __VLS_160;
const __VLS_161 = {
    onClick: (__VLS_ctx.openProbePicker)
};
__VLS_157.slots.default;
var __VLS_157;
if (__VLS_ctx.probeResult) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "probe-result" },
    });
    const __VLS_162 = {}.ElInput;
    /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
    // @ts-ignore
    const __VLS_163 = __VLS_asFunctionalComponent(__VLS_162, new __VLS_162({
        modelValue: (__VLS_ctx.probeResult),
        readonly: true,
    }));
    const __VLS_164 = __VLS_163({
        modelValue: (__VLS_ctx.probeResult),
        readonly: true,
    }, ...__VLS_functionalComponentArgsRest(__VLS_163));
    __VLS_asFunctionalElement(__VLS_intrinsicElements.a, __VLS_intrinsicElements.a)({
        href: (__VLS_ctx.probeResult),
        target: "_blank",
        rel: "noreferrer",
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.img)({
        src: (__VLS_ctx.probeResult),
        alt: "probe preview",
        ...{ class: "probe-preview" },
    });
}
var __VLS_153;
var __VLS_23;
/** @type {__VLS_StyleScopedClasses['page-card']} */ ;
/** @type {__VLS_StyleScopedClasses['page-head']} */ ;
/** @type {__VLS_StyleScopedClasses['page-actions']} */ ;
/** @type {__VLS_StyleScopedClasses['upload-form']} */ ;
/** @type {__VLS_StyleScopedClasses['setting-card']} */ ;
/** @type {__VLS_StyleScopedClasses['field-tip']} */ ;
/** @type {__VLS_StyleScopedClasses['setting-card']} */ ;
/** @type {__VLS_StyleScopedClasses['hint-block']} */ ;
/** @type {__VLS_StyleScopedClasses['setting-card']} */ ;
/** @type {__VLS_StyleScopedClasses['field-tip']} */ ;
/** @type {__VLS_StyleScopedClasses['setting-card']} */ ;
/** @type {__VLS_StyleScopedClasses['probe-panel']} */ ;
/** @type {__VLS_StyleScopedClasses['probe-copy']} */ ;
/** @type {__VLS_StyleScopedClasses['probe-actions']} */ ;
/** @type {__VLS_StyleScopedClasses['probe-result']} */ ;
/** @type {__VLS_StyleScopedClasses['probe-preview']} */ ;
// @ts-ignore
var __VLS_25 = __VLS_24;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            loading: loading,
            saving: saving,
            testing: testing,
            probing: probing,
            formRef: formRef,
            probeInputRef: probeInputRef,
            probeResult: probeResult,
            form: form,
            canEdit: canEdit,
            rules: rules,
            handleSave: handleSave,
            handleTest: handleTest,
            handleProbeChange: handleProbeChange,
            openProbePicker: openProbePicker,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
