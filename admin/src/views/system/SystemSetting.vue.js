import { computed, reactive, ref } from 'vue';
import { ElMessage } from 'element-plus';
import { getSettings, updateSettings } from '@/api/system';
import { useAuthStore } from '@/stores/auth';
const auth = useAuthStore();
const canEdit = computed(() => auth.isSuperAdmin || auth.perms.includes('system.setting.edit'));
const tabs = [
    { name: 'basic', label: '基本设置' },
    { name: 'wxpay', label: '微信支付' },
    { name: 'wxlogin', label: '微信登录' },
    { name: 'qywx', label: '企业微信' },
    { name: 'kdniao', label: '快递鸟' },
    { name: 'sms', label: '短信' },
    { name: 'security', label: '安全设置' },
];
// Secret fields per group — when backend returns "", it means "value is set but masked"
const secretFields = {
    basic: [],
    wxpay: ['api_v3_key', 'cert_pem', 'key_pem'],
    wxlogin: ['mp_app_secret'],
    qywx: ['app_secret', 'robot_webhook'],
    kdniao: ['api_key'],
    sms: ['secret_key'],
    security: [],
};
// Textarea fields (large content)
const textareaFields = {
    basic: [],
    wxpay: ['cert_pem', 'key_pem'],
    wxlogin: [],
    qywx: [],
    kdniao: [],
    sms: [],
    security: [],
};
// Forms per group
const forms = reactive({
    basic: { shop_name: '', shop_logo: '', contact_phone: '', icp_no: '' },
    wxpay: { mch_id: '', app_id: '', api_v3_key: '', serial_no: '', cert_pem: '', key_pem: '' },
    wxlogin: { mp_app_id: '', mp_app_secret: '' },
    qywx: { corp_id: '', agent_id: '', app_secret: '', robot_webhook: '' },
    kdniao: { e_business_id: '', api_key: '', env: 'sandbox' },
    sms: { provider: 'aliyun', access_key: '', secret_key: '', sign_name: '', verify_template: '' },
    security: { session_hours: '24', max_login_attempts: '5', admin_pw_min_len: '12' },
});
// Track which groups have been loaded and which are currently being saved
const loaded = reactive({
    basic: false, wxpay: false, wxlogin: false, qywx: false, kdniao: false, sms: false, security: false,
});
const loadingGroup = ref(null);
const savingGroup = ref(null);
// Track which secret fields already have a value on the server (mask display)
const secretSet = reactive({
    basic: {},
    wxpay: { api_v3_key: false, cert_pem: false, key_pem: false },
    wxlogin: { mp_app_secret: false },
    qywx: { app_secret: false, robot_webhook: false },
    kdniao: { api_key: false },
    sms: { secret_key: false },
    security: {},
});
const activeTab = ref('basic');
async function loadGroup(group) {
    if (loaded[group])
        return;
    loadingGroup.value = group;
    try {
        const data = await getSettings(group);
        const secrets = secretFields[group];
        const newForm = { ...forms[group] };
        for (const [k, v] of Object.entries(data)) {
            if (secrets.includes(k)) {
                // Backend returns "" for masked secrets
                secretSet[group][k] = true; // assume it's set if backend ever had a value
                newForm[k] = ''; // always keep local field empty so user must retype to change
            }
            else {
                newForm[k] = v;
            }
        }
        Object.assign(forms[group], newForm);
        loaded[group] = true;
    }
    catch {
        ElMessage.error(`加载 ${group} 配置失败`);
    }
    finally {
        loadingGroup.value = null;
    }
}
function handleTabChange(name) {
    activeTab.value = name;
    loadGroup(name);
}
async function handleSave(group) {
    savingGroup.value = group;
    try {
        const payload = {};
        const secrets = secretFields[group];
        for (const [k, v] of Object.entries(forms[group])) {
            // Skip empty secret fields — leave backend value unchanged
            if (secrets.includes(k) && v === '')
                continue;
            payload[k] = v;
        }
        await updateSettings(group, payload);
        ElMessage.success('保存成功');
        // Reload to sync
        loaded[group] = false;
        await loadGroup(group);
    }
    catch {
        ElMessage.error('保存失败');
    }
    finally {
        savingGroup.value = null;
    }
}
// Initial load for default tab
loadGroup('basic');
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
// CSS variable injection 
// CSS variable injection end 
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "page-card" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "page-head" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.h3, __VLS_intrinsicElements.h3)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({});
if (!__VLS_ctx.canEdit) {
    const __VLS_0 = {}.ElAlert;
    /** @type {[typeof __VLS_components.ElAlert, typeof __VLS_components.elAlert, ]} */ ;
    // @ts-ignore
    const __VLS_1 = __VLS_asFunctionalComponent(__VLS_0, new __VLS_0({
        title: "当前账号只有查看权限，不能修改系统设置。",
        type: "warning",
        showIcon: true,
        closable: (false),
        ...{ style: {} },
    }));
    const __VLS_2 = __VLS_1({
        title: "当前账号只有查看权限，不能修改系统设置。",
        type: "warning",
        showIcon: true,
        closable: (false),
        ...{ style: {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_1));
}
const __VLS_4 = {}.ElTabs;
/** @type {[typeof __VLS_components.ElTabs, typeof __VLS_components.elTabs, typeof __VLS_components.ElTabs, typeof __VLS_components.elTabs, ]} */ ;
// @ts-ignore
const __VLS_5 = __VLS_asFunctionalComponent(__VLS_4, new __VLS_4({
    ...{ 'onTabChange': {} },
    modelValue: (__VLS_ctx.activeTab),
}));
const __VLS_6 = __VLS_5({
    ...{ 'onTabChange': {} },
    modelValue: (__VLS_ctx.activeTab),
}, ...__VLS_functionalComponentArgsRest(__VLS_5));
let __VLS_8;
let __VLS_9;
let __VLS_10;
const __VLS_11 = {
    onTabChange: (__VLS_ctx.handleTabChange)
};
__VLS_7.slots.default;
for (const [tab] of __VLS_getVForSourceType((__VLS_ctx.tabs))) {
    const __VLS_12 = {}.ElTabPane;
    /** @type {[typeof __VLS_components.ElTabPane, typeof __VLS_components.elTabPane, typeof __VLS_components.ElTabPane, typeof __VLS_components.elTabPane, ]} */ ;
    // @ts-ignore
    const __VLS_13 = __VLS_asFunctionalComponent(__VLS_12, new __VLS_12({
        key: (tab.name),
        label: (tab.label),
        name: (tab.name),
    }));
    const __VLS_14 = __VLS_13({
        key: (tab.name),
        label: (tab.label),
        name: (tab.name),
    }, ...__VLS_functionalComponentArgsRest(__VLS_13));
    __VLS_15.slots.default;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "tab-content" },
    });
    __VLS_asFunctionalDirective(__VLS_directives.vLoading)(null, { ...__VLS_directiveBindingRestFields, value: (__VLS_ctx.loadingGroup === tab.name) }, null, null);
    if (tab.name === 'basic') {
        const __VLS_16 = {}.ElForm;
        /** @type {[typeof __VLS_components.ElForm, typeof __VLS_components.elForm, typeof __VLS_components.ElForm, typeof __VLS_components.elForm, ]} */ ;
        // @ts-ignore
        const __VLS_17 = __VLS_asFunctionalComponent(__VLS_16, new __VLS_16({
            labelWidth: "160px",
            ...{ class: "setting-form" },
        }));
        const __VLS_18 = __VLS_17({
            labelWidth: "160px",
            ...{ class: "setting-form" },
        }, ...__VLS_functionalComponentArgsRest(__VLS_17));
        __VLS_19.slots.default;
        const __VLS_20 = {}.ElFormItem;
        /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
        // @ts-ignore
        const __VLS_21 = __VLS_asFunctionalComponent(__VLS_20, new __VLS_20({
            label: "店铺名称",
        }));
        const __VLS_22 = __VLS_21({
            label: "店铺名称",
        }, ...__VLS_functionalComponentArgsRest(__VLS_21));
        __VLS_23.slots.default;
        const __VLS_24 = {}.ElInput;
        /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
        // @ts-ignore
        const __VLS_25 = __VLS_asFunctionalComponent(__VLS_24, new __VLS_24({
            modelValue: (__VLS_ctx.forms.basic.shop_name),
            placeholder: "请输入店铺名称",
            disabled: (!__VLS_ctx.canEdit),
        }));
        const __VLS_26 = __VLS_25({
            modelValue: (__VLS_ctx.forms.basic.shop_name),
            placeholder: "请输入店铺名称",
            disabled: (!__VLS_ctx.canEdit),
        }, ...__VLS_functionalComponentArgsRest(__VLS_25));
        var __VLS_23;
        const __VLS_28 = {}.ElFormItem;
        /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
        // @ts-ignore
        const __VLS_29 = __VLS_asFunctionalComponent(__VLS_28, new __VLS_28({
            label: "店铺 Logo URL",
        }));
        const __VLS_30 = __VLS_29({
            label: "店铺 Logo URL",
        }, ...__VLS_functionalComponentArgsRest(__VLS_29));
        __VLS_31.slots.default;
        const __VLS_32 = {}.ElInput;
        /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
        // @ts-ignore
        const __VLS_33 = __VLS_asFunctionalComponent(__VLS_32, new __VLS_32({
            modelValue: (__VLS_ctx.forms.basic.shop_logo),
            placeholder: "请输入图片地址",
            disabled: (!__VLS_ctx.canEdit),
        }));
        const __VLS_34 = __VLS_33({
            modelValue: (__VLS_ctx.forms.basic.shop_logo),
            placeholder: "请输入图片地址",
            disabled: (!__VLS_ctx.canEdit),
        }, ...__VLS_functionalComponentArgsRest(__VLS_33));
        var __VLS_31;
        const __VLS_36 = {}.ElFormItem;
        /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
        // @ts-ignore
        const __VLS_37 = __VLS_asFunctionalComponent(__VLS_36, new __VLS_36({
            label: "联系电话",
        }));
        const __VLS_38 = __VLS_37({
            label: "联系电话",
        }, ...__VLS_functionalComponentArgsRest(__VLS_37));
        __VLS_39.slots.default;
        const __VLS_40 = {}.ElInput;
        /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
        // @ts-ignore
        const __VLS_41 = __VLS_asFunctionalComponent(__VLS_40, new __VLS_40({
            modelValue: (__VLS_ctx.forms.basic.contact_phone),
            placeholder: "请输入联系电话",
            disabled: (!__VLS_ctx.canEdit),
        }));
        const __VLS_42 = __VLS_41({
            modelValue: (__VLS_ctx.forms.basic.contact_phone),
            placeholder: "请输入联系电话",
            disabled: (!__VLS_ctx.canEdit),
        }, ...__VLS_functionalComponentArgsRest(__VLS_41));
        var __VLS_39;
        const __VLS_44 = {}.ElFormItem;
        /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
        // @ts-ignore
        const __VLS_45 = __VLS_asFunctionalComponent(__VLS_44, new __VLS_44({
            label: "ICP 备案号",
        }));
        const __VLS_46 = __VLS_45({
            label: "ICP 备案号",
        }, ...__VLS_functionalComponentArgsRest(__VLS_45));
        __VLS_47.slots.default;
        const __VLS_48 = {}.ElInput;
        /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
        // @ts-ignore
        const __VLS_49 = __VLS_asFunctionalComponent(__VLS_48, new __VLS_48({
            modelValue: (__VLS_ctx.forms.basic.icp_no),
            placeholder: "例如 粤ICP备XXXXXXXX号",
            disabled: (!__VLS_ctx.canEdit),
        }));
        const __VLS_50 = __VLS_49({
            modelValue: (__VLS_ctx.forms.basic.icp_no),
            placeholder: "例如 粤ICP备XXXXXXXX号",
            disabled: (!__VLS_ctx.canEdit),
        }, ...__VLS_functionalComponentArgsRest(__VLS_49));
        var __VLS_47;
        if (__VLS_ctx.canEdit) {
            const __VLS_52 = {}.ElFormItem;
            /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
            // @ts-ignore
            const __VLS_53 = __VLS_asFunctionalComponent(__VLS_52, new __VLS_52({}));
            const __VLS_54 = __VLS_53({}, ...__VLS_functionalComponentArgsRest(__VLS_53));
            __VLS_55.slots.default;
            const __VLS_56 = {}.ElButton;
            /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
            // @ts-ignore
            const __VLS_57 = __VLS_asFunctionalComponent(__VLS_56, new __VLS_56({
                ...{ 'onClick': {} },
                type: "primary",
                loading: (__VLS_ctx.savingGroup === 'basic'),
            }));
            const __VLS_58 = __VLS_57({
                ...{ 'onClick': {} },
                type: "primary",
                loading: (__VLS_ctx.savingGroup === 'basic'),
            }, ...__VLS_functionalComponentArgsRest(__VLS_57));
            let __VLS_60;
            let __VLS_61;
            let __VLS_62;
            const __VLS_63 = {
                onClick: (...[$event]) => {
                    if (!(tab.name === 'basic'))
                        return;
                    if (!(__VLS_ctx.canEdit))
                        return;
                    __VLS_ctx.handleSave('basic');
                }
            };
            __VLS_59.slots.default;
            var __VLS_59;
            var __VLS_55;
        }
        var __VLS_19;
    }
    else if (tab.name === 'wxpay') {
        const __VLS_64 = {}.ElForm;
        /** @type {[typeof __VLS_components.ElForm, typeof __VLS_components.elForm, typeof __VLS_components.ElForm, typeof __VLS_components.elForm, ]} */ ;
        // @ts-ignore
        const __VLS_65 = __VLS_asFunctionalComponent(__VLS_64, new __VLS_64({
            labelWidth: "160px",
            ...{ class: "setting-form" },
        }));
        const __VLS_66 = __VLS_65({
            labelWidth: "160px",
            ...{ class: "setting-form" },
        }, ...__VLS_functionalComponentArgsRest(__VLS_65));
        __VLS_67.slots.default;
        const __VLS_68 = {}.ElFormItem;
        /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
        // @ts-ignore
        const __VLS_69 = __VLS_asFunctionalComponent(__VLS_68, new __VLS_68({
            label: "商户号 (mch_id)",
        }));
        const __VLS_70 = __VLS_69({
            label: "商户号 (mch_id)",
        }, ...__VLS_functionalComponentArgsRest(__VLS_69));
        __VLS_71.slots.default;
        const __VLS_72 = {}.ElInput;
        /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
        // @ts-ignore
        const __VLS_73 = __VLS_asFunctionalComponent(__VLS_72, new __VLS_72({
            modelValue: (__VLS_ctx.forms.wxpay.mch_id),
            placeholder: "请输入商户号",
            disabled: (!__VLS_ctx.canEdit),
        }));
        const __VLS_74 = __VLS_73({
            modelValue: (__VLS_ctx.forms.wxpay.mch_id),
            placeholder: "请输入商户号",
            disabled: (!__VLS_ctx.canEdit),
        }, ...__VLS_functionalComponentArgsRest(__VLS_73));
        var __VLS_71;
        const __VLS_76 = {}.ElFormItem;
        /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
        // @ts-ignore
        const __VLS_77 = __VLS_asFunctionalComponent(__VLS_76, new __VLS_76({
            label: "AppID",
        }));
        const __VLS_78 = __VLS_77({
            label: "AppID",
        }, ...__VLS_functionalComponentArgsRest(__VLS_77));
        __VLS_79.slots.default;
        const __VLS_80 = {}.ElInput;
        /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
        // @ts-ignore
        const __VLS_81 = __VLS_asFunctionalComponent(__VLS_80, new __VLS_80({
            modelValue: (__VLS_ctx.forms.wxpay.app_id),
            placeholder: "微信支付关联 AppID",
            disabled: (!__VLS_ctx.canEdit),
        }));
        const __VLS_82 = __VLS_81({
            modelValue: (__VLS_ctx.forms.wxpay.app_id),
            placeholder: "微信支付关联 AppID",
            disabled: (!__VLS_ctx.canEdit),
        }, ...__VLS_functionalComponentArgsRest(__VLS_81));
        var __VLS_79;
        const __VLS_84 = {}.ElFormItem;
        /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
        // @ts-ignore
        const __VLS_85 = __VLS_asFunctionalComponent(__VLS_84, new __VLS_84({
            label: "APIv3 密钥",
        }));
        const __VLS_86 = __VLS_85({
            label: "APIv3 密钥",
        }, ...__VLS_functionalComponentArgsRest(__VLS_85));
        __VLS_87.slots.default;
        const __VLS_88 = {}.ElInput;
        /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
        // @ts-ignore
        const __VLS_89 = __VLS_asFunctionalComponent(__VLS_88, new __VLS_88({
            modelValue: (__VLS_ctx.forms.wxpay.api_v3_key),
            placeholder: (__VLS_ctx.secretSet.wxpay.api_v3_key ? '●●●●●● 已配置，留空则保持不变' : '请输入 APIv3 密钥'),
            disabled: (!__VLS_ctx.canEdit),
            showPassword: true,
        }));
        const __VLS_90 = __VLS_89({
            modelValue: (__VLS_ctx.forms.wxpay.api_v3_key),
            placeholder: (__VLS_ctx.secretSet.wxpay.api_v3_key ? '●●●●●● 已配置，留空则保持不变' : '请输入 APIv3 密钥'),
            disabled: (!__VLS_ctx.canEdit),
            showPassword: true,
        }, ...__VLS_functionalComponentArgsRest(__VLS_89));
        var __VLS_87;
        const __VLS_92 = {}.ElFormItem;
        /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
        // @ts-ignore
        const __VLS_93 = __VLS_asFunctionalComponent(__VLS_92, new __VLS_92({
            label: "证书序列号",
        }));
        const __VLS_94 = __VLS_93({
            label: "证书序列号",
        }, ...__VLS_functionalComponentArgsRest(__VLS_93));
        __VLS_95.slots.default;
        const __VLS_96 = {}.ElInput;
        /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
        // @ts-ignore
        const __VLS_97 = __VLS_asFunctionalComponent(__VLS_96, new __VLS_96({
            modelValue: (__VLS_ctx.forms.wxpay.serial_no),
            placeholder: "请输入证书序列号",
            disabled: (!__VLS_ctx.canEdit),
        }));
        const __VLS_98 = __VLS_97({
            modelValue: (__VLS_ctx.forms.wxpay.serial_no),
            placeholder: "请输入证书序列号",
            disabled: (!__VLS_ctx.canEdit),
        }, ...__VLS_functionalComponentArgsRest(__VLS_97));
        var __VLS_95;
        const __VLS_100 = {}.ElFormItem;
        /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
        // @ts-ignore
        const __VLS_101 = __VLS_asFunctionalComponent(__VLS_100, new __VLS_100({
            label: "cert.pem",
        }));
        const __VLS_102 = __VLS_101({
            label: "cert.pem",
        }, ...__VLS_functionalComponentArgsRest(__VLS_101));
        __VLS_103.slots.default;
        const __VLS_104 = {}.ElInput;
        /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
        // @ts-ignore
        const __VLS_105 = __VLS_asFunctionalComponent(__VLS_104, new __VLS_104({
            modelValue: (__VLS_ctx.forms.wxpay.cert_pem),
            type: "textarea",
            rows: (4),
            placeholder: (__VLS_ctx.secretSet.wxpay.cert_pem ? '●●●●●● 已配置，留空则保持不变' : '请粘贴 apiclient_cert.pem 内容'),
            disabled: (!__VLS_ctx.canEdit),
        }));
        const __VLS_106 = __VLS_105({
            modelValue: (__VLS_ctx.forms.wxpay.cert_pem),
            type: "textarea",
            rows: (4),
            placeholder: (__VLS_ctx.secretSet.wxpay.cert_pem ? '●●●●●● 已配置，留空则保持不变' : '请粘贴 apiclient_cert.pem 内容'),
            disabled: (!__VLS_ctx.canEdit),
        }, ...__VLS_functionalComponentArgsRest(__VLS_105));
        var __VLS_103;
        const __VLS_108 = {}.ElFormItem;
        /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
        // @ts-ignore
        const __VLS_109 = __VLS_asFunctionalComponent(__VLS_108, new __VLS_108({
            label: "key.pem",
        }));
        const __VLS_110 = __VLS_109({
            label: "key.pem",
        }, ...__VLS_functionalComponentArgsRest(__VLS_109));
        __VLS_111.slots.default;
        const __VLS_112 = {}.ElInput;
        /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
        // @ts-ignore
        const __VLS_113 = __VLS_asFunctionalComponent(__VLS_112, new __VLS_112({
            modelValue: (__VLS_ctx.forms.wxpay.key_pem),
            type: "textarea",
            rows: (4),
            placeholder: (__VLS_ctx.secretSet.wxpay.key_pem ? '●●●●●● 已配置，留空则保持不变' : '请粘贴 apiclient_key.pem 内容'),
            disabled: (!__VLS_ctx.canEdit),
        }));
        const __VLS_114 = __VLS_113({
            modelValue: (__VLS_ctx.forms.wxpay.key_pem),
            type: "textarea",
            rows: (4),
            placeholder: (__VLS_ctx.secretSet.wxpay.key_pem ? '●●●●●● 已配置，留空则保持不变' : '请粘贴 apiclient_key.pem 内容'),
            disabled: (!__VLS_ctx.canEdit),
        }, ...__VLS_functionalComponentArgsRest(__VLS_113));
        var __VLS_111;
        if (__VLS_ctx.canEdit) {
            const __VLS_116 = {}.ElFormItem;
            /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
            // @ts-ignore
            const __VLS_117 = __VLS_asFunctionalComponent(__VLS_116, new __VLS_116({}));
            const __VLS_118 = __VLS_117({}, ...__VLS_functionalComponentArgsRest(__VLS_117));
            __VLS_119.slots.default;
            const __VLS_120 = {}.ElButton;
            /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
            // @ts-ignore
            const __VLS_121 = __VLS_asFunctionalComponent(__VLS_120, new __VLS_120({
                ...{ 'onClick': {} },
                type: "primary",
                loading: (__VLS_ctx.savingGroup === 'wxpay'),
            }));
            const __VLS_122 = __VLS_121({
                ...{ 'onClick': {} },
                type: "primary",
                loading: (__VLS_ctx.savingGroup === 'wxpay'),
            }, ...__VLS_functionalComponentArgsRest(__VLS_121));
            let __VLS_124;
            let __VLS_125;
            let __VLS_126;
            const __VLS_127 = {
                onClick: (...[$event]) => {
                    if (!!(tab.name === 'basic'))
                        return;
                    if (!(tab.name === 'wxpay'))
                        return;
                    if (!(__VLS_ctx.canEdit))
                        return;
                    __VLS_ctx.handleSave('wxpay');
                }
            };
            __VLS_123.slots.default;
            var __VLS_123;
            var __VLS_119;
        }
        var __VLS_67;
    }
    else if (tab.name === 'wxlogin') {
        const __VLS_128 = {}.ElForm;
        /** @type {[typeof __VLS_components.ElForm, typeof __VLS_components.elForm, typeof __VLS_components.ElForm, typeof __VLS_components.elForm, ]} */ ;
        // @ts-ignore
        const __VLS_129 = __VLS_asFunctionalComponent(__VLS_128, new __VLS_128({
            labelWidth: "160px",
            ...{ class: "setting-form" },
        }));
        const __VLS_130 = __VLS_129({
            labelWidth: "160px",
            ...{ class: "setting-form" },
        }, ...__VLS_functionalComponentArgsRest(__VLS_129));
        __VLS_131.slots.default;
        const __VLS_132 = {}.ElFormItem;
        /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
        // @ts-ignore
        const __VLS_133 = __VLS_asFunctionalComponent(__VLS_132, new __VLS_132({
            label: "小程序 AppID",
        }));
        const __VLS_134 = __VLS_133({
            label: "小程序 AppID",
        }, ...__VLS_functionalComponentArgsRest(__VLS_133));
        __VLS_135.slots.default;
        const __VLS_136 = {}.ElInput;
        /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
        // @ts-ignore
        const __VLS_137 = __VLS_asFunctionalComponent(__VLS_136, new __VLS_136({
            modelValue: (__VLS_ctx.forms.wxlogin.mp_app_id),
            placeholder: "微信小程序 AppID",
            disabled: (!__VLS_ctx.canEdit),
        }));
        const __VLS_138 = __VLS_137({
            modelValue: (__VLS_ctx.forms.wxlogin.mp_app_id),
            placeholder: "微信小程序 AppID",
            disabled: (!__VLS_ctx.canEdit),
        }, ...__VLS_functionalComponentArgsRest(__VLS_137));
        var __VLS_135;
        const __VLS_140 = {}.ElFormItem;
        /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
        // @ts-ignore
        const __VLS_141 = __VLS_asFunctionalComponent(__VLS_140, new __VLS_140({
            label: "小程序 AppSecret",
        }));
        const __VLS_142 = __VLS_141({
            label: "小程序 AppSecret",
        }, ...__VLS_functionalComponentArgsRest(__VLS_141));
        __VLS_143.slots.default;
        const __VLS_144 = {}.ElInput;
        /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
        // @ts-ignore
        const __VLS_145 = __VLS_asFunctionalComponent(__VLS_144, new __VLS_144({
            modelValue: (__VLS_ctx.forms.wxlogin.mp_app_secret),
            placeholder: (__VLS_ctx.secretSet.wxlogin.mp_app_secret ? '●●●●●● 已配置，留空则保持不变' : '请输入 AppSecret'),
            disabled: (!__VLS_ctx.canEdit),
            showPassword: true,
        }));
        const __VLS_146 = __VLS_145({
            modelValue: (__VLS_ctx.forms.wxlogin.mp_app_secret),
            placeholder: (__VLS_ctx.secretSet.wxlogin.mp_app_secret ? '●●●●●● 已配置，留空则保持不变' : '请输入 AppSecret'),
            disabled: (!__VLS_ctx.canEdit),
            showPassword: true,
        }, ...__VLS_functionalComponentArgsRest(__VLS_145));
        var __VLS_143;
        if (__VLS_ctx.canEdit) {
            const __VLS_148 = {}.ElFormItem;
            /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
            // @ts-ignore
            const __VLS_149 = __VLS_asFunctionalComponent(__VLS_148, new __VLS_148({}));
            const __VLS_150 = __VLS_149({}, ...__VLS_functionalComponentArgsRest(__VLS_149));
            __VLS_151.slots.default;
            const __VLS_152 = {}.ElButton;
            /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
            // @ts-ignore
            const __VLS_153 = __VLS_asFunctionalComponent(__VLS_152, new __VLS_152({
                ...{ 'onClick': {} },
                type: "primary",
                loading: (__VLS_ctx.savingGroup === 'wxlogin'),
            }));
            const __VLS_154 = __VLS_153({
                ...{ 'onClick': {} },
                type: "primary",
                loading: (__VLS_ctx.savingGroup === 'wxlogin'),
            }, ...__VLS_functionalComponentArgsRest(__VLS_153));
            let __VLS_156;
            let __VLS_157;
            let __VLS_158;
            const __VLS_159 = {
                onClick: (...[$event]) => {
                    if (!!(tab.name === 'basic'))
                        return;
                    if (!!(tab.name === 'wxpay'))
                        return;
                    if (!(tab.name === 'wxlogin'))
                        return;
                    if (!(__VLS_ctx.canEdit))
                        return;
                    __VLS_ctx.handleSave('wxlogin');
                }
            };
            __VLS_155.slots.default;
            var __VLS_155;
            var __VLS_151;
        }
        var __VLS_131;
    }
    else if (tab.name === 'qywx') {
        const __VLS_160 = {}.ElForm;
        /** @type {[typeof __VLS_components.ElForm, typeof __VLS_components.elForm, typeof __VLS_components.ElForm, typeof __VLS_components.elForm, ]} */ ;
        // @ts-ignore
        const __VLS_161 = __VLS_asFunctionalComponent(__VLS_160, new __VLS_160({
            labelWidth: "160px",
            ...{ class: "setting-form" },
        }));
        const __VLS_162 = __VLS_161({
            labelWidth: "160px",
            ...{ class: "setting-form" },
        }, ...__VLS_functionalComponentArgsRest(__VLS_161));
        __VLS_163.slots.default;
        const __VLS_164 = {}.ElFormItem;
        /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
        // @ts-ignore
        const __VLS_165 = __VLS_asFunctionalComponent(__VLS_164, new __VLS_164({
            label: "企业 ID (corp_id)",
        }));
        const __VLS_166 = __VLS_165({
            label: "企业 ID (corp_id)",
        }, ...__VLS_functionalComponentArgsRest(__VLS_165));
        __VLS_167.slots.default;
        const __VLS_168 = {}.ElInput;
        /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
        // @ts-ignore
        const __VLS_169 = __VLS_asFunctionalComponent(__VLS_168, new __VLS_168({
            modelValue: (__VLS_ctx.forms.qywx.corp_id),
            placeholder: "企业微信 corpid",
            disabled: (!__VLS_ctx.canEdit),
        }));
        const __VLS_170 = __VLS_169({
            modelValue: (__VLS_ctx.forms.qywx.corp_id),
            placeholder: "企业微信 corpid",
            disabled: (!__VLS_ctx.canEdit),
        }, ...__VLS_functionalComponentArgsRest(__VLS_169));
        var __VLS_167;
        const __VLS_172 = {}.ElFormItem;
        /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
        // @ts-ignore
        const __VLS_173 = __VLS_asFunctionalComponent(__VLS_172, new __VLS_172({
            label: "应用 ID (agent_id)",
        }));
        const __VLS_174 = __VLS_173({
            label: "应用 ID (agent_id)",
        }, ...__VLS_functionalComponentArgsRest(__VLS_173));
        __VLS_175.slots.default;
        const __VLS_176 = {}.ElInput;
        /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
        // @ts-ignore
        const __VLS_177 = __VLS_asFunctionalComponent(__VLS_176, new __VLS_176({
            modelValue: (__VLS_ctx.forms.qywx.agent_id),
            placeholder: "企业微信应用 agentid",
            disabled: (!__VLS_ctx.canEdit),
        }));
        const __VLS_178 = __VLS_177({
            modelValue: (__VLS_ctx.forms.qywx.agent_id),
            placeholder: "企业微信应用 agentid",
            disabled: (!__VLS_ctx.canEdit),
        }, ...__VLS_functionalComponentArgsRest(__VLS_177));
        var __VLS_175;
        const __VLS_180 = {}.ElFormItem;
        /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
        // @ts-ignore
        const __VLS_181 = __VLS_asFunctionalComponent(__VLS_180, new __VLS_180({
            label: "应用 Secret",
        }));
        const __VLS_182 = __VLS_181({
            label: "应用 Secret",
        }, ...__VLS_functionalComponentArgsRest(__VLS_181));
        __VLS_183.slots.default;
        const __VLS_184 = {}.ElInput;
        /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
        // @ts-ignore
        const __VLS_185 = __VLS_asFunctionalComponent(__VLS_184, new __VLS_184({
            modelValue: (__VLS_ctx.forms.qywx.app_secret),
            placeholder: (__VLS_ctx.secretSet.qywx.app_secret ? '●●●●●● 已配置，留空则保持不变' : '请输入应用 Secret'),
            disabled: (!__VLS_ctx.canEdit),
            showPassword: true,
        }));
        const __VLS_186 = __VLS_185({
            modelValue: (__VLS_ctx.forms.qywx.app_secret),
            placeholder: (__VLS_ctx.secretSet.qywx.app_secret ? '●●●●●● 已配置，留空则保持不变' : '请输入应用 Secret'),
            disabled: (!__VLS_ctx.canEdit),
            showPassword: true,
        }, ...__VLS_functionalComponentArgsRest(__VLS_185));
        var __VLS_183;
        const __VLS_188 = {}.ElFormItem;
        /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
        // @ts-ignore
        const __VLS_189 = __VLS_asFunctionalComponent(__VLS_188, new __VLS_188({
            label: "机器人 Webhook",
        }));
        const __VLS_190 = __VLS_189({
            label: "机器人 Webhook",
        }, ...__VLS_functionalComponentArgsRest(__VLS_189));
        __VLS_191.slots.default;
        const __VLS_192 = {}.ElInput;
        /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
        // @ts-ignore
        const __VLS_193 = __VLS_asFunctionalComponent(__VLS_192, new __VLS_192({
            modelValue: (__VLS_ctx.forms.qywx.robot_webhook),
            placeholder: (__VLS_ctx.secretSet.qywx.robot_webhook ? '●●●●●● 已配置，留空则保持不变' : 'https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=...'),
            disabled: (!__VLS_ctx.canEdit),
            showPassword: true,
        }));
        const __VLS_194 = __VLS_193({
            modelValue: (__VLS_ctx.forms.qywx.robot_webhook),
            placeholder: (__VLS_ctx.secretSet.qywx.robot_webhook ? '●●●●●● 已配置，留空则保持不变' : 'https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=...'),
            disabled: (!__VLS_ctx.canEdit),
            showPassword: true,
        }, ...__VLS_functionalComponentArgsRest(__VLS_193));
        var __VLS_191;
        if (__VLS_ctx.canEdit) {
            const __VLS_196 = {}.ElFormItem;
            /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
            // @ts-ignore
            const __VLS_197 = __VLS_asFunctionalComponent(__VLS_196, new __VLS_196({}));
            const __VLS_198 = __VLS_197({}, ...__VLS_functionalComponentArgsRest(__VLS_197));
            __VLS_199.slots.default;
            const __VLS_200 = {}.ElButton;
            /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
            // @ts-ignore
            const __VLS_201 = __VLS_asFunctionalComponent(__VLS_200, new __VLS_200({
                ...{ 'onClick': {} },
                type: "primary",
                loading: (__VLS_ctx.savingGroup === 'qywx'),
            }));
            const __VLS_202 = __VLS_201({
                ...{ 'onClick': {} },
                type: "primary",
                loading: (__VLS_ctx.savingGroup === 'qywx'),
            }, ...__VLS_functionalComponentArgsRest(__VLS_201));
            let __VLS_204;
            let __VLS_205;
            let __VLS_206;
            const __VLS_207 = {
                onClick: (...[$event]) => {
                    if (!!(tab.name === 'basic'))
                        return;
                    if (!!(tab.name === 'wxpay'))
                        return;
                    if (!!(tab.name === 'wxlogin'))
                        return;
                    if (!(tab.name === 'qywx'))
                        return;
                    if (!(__VLS_ctx.canEdit))
                        return;
                    __VLS_ctx.handleSave('qywx');
                }
            };
            __VLS_203.slots.default;
            var __VLS_203;
            var __VLS_199;
        }
        var __VLS_163;
    }
    else if (tab.name === 'kdniao') {
        const __VLS_208 = {}.ElForm;
        /** @type {[typeof __VLS_components.ElForm, typeof __VLS_components.elForm, typeof __VLS_components.ElForm, typeof __VLS_components.elForm, ]} */ ;
        // @ts-ignore
        const __VLS_209 = __VLS_asFunctionalComponent(__VLS_208, new __VLS_208({
            labelWidth: "160px",
            ...{ class: "setting-form" },
        }));
        const __VLS_210 = __VLS_209({
            labelWidth: "160px",
            ...{ class: "setting-form" },
        }, ...__VLS_functionalComponentArgsRest(__VLS_209));
        __VLS_211.slots.default;
        const __VLS_212 = {}.ElFormItem;
        /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
        // @ts-ignore
        const __VLS_213 = __VLS_asFunctionalComponent(__VLS_212, new __VLS_212({
            label: "电商 ID",
        }));
        const __VLS_214 = __VLS_213({
            label: "电商 ID",
        }, ...__VLS_functionalComponentArgsRest(__VLS_213));
        __VLS_215.slots.default;
        const __VLS_216 = {}.ElInput;
        /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
        // @ts-ignore
        const __VLS_217 = __VLS_asFunctionalComponent(__VLS_216, new __VLS_216({
            modelValue: (__VLS_ctx.forms.kdniao.e_business_id),
            placeholder: "快递鸟电商 EBusinessID",
            disabled: (!__VLS_ctx.canEdit),
        }));
        const __VLS_218 = __VLS_217({
            modelValue: (__VLS_ctx.forms.kdniao.e_business_id),
            placeholder: "快递鸟电商 EBusinessID",
            disabled: (!__VLS_ctx.canEdit),
        }, ...__VLS_functionalComponentArgsRest(__VLS_217));
        var __VLS_215;
        const __VLS_220 = {}.ElFormItem;
        /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
        // @ts-ignore
        const __VLS_221 = __VLS_asFunctionalComponent(__VLS_220, new __VLS_220({
            label: "API Key",
        }));
        const __VLS_222 = __VLS_221({
            label: "API Key",
        }, ...__VLS_functionalComponentArgsRest(__VLS_221));
        __VLS_223.slots.default;
        const __VLS_224 = {}.ElInput;
        /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
        // @ts-ignore
        const __VLS_225 = __VLS_asFunctionalComponent(__VLS_224, new __VLS_224({
            modelValue: (__VLS_ctx.forms.kdniao.api_key),
            placeholder: (__VLS_ctx.secretSet.kdniao.api_key ? '●●●●●● 已配置，留空则保持不变' : '请输入 API Key'),
            disabled: (!__VLS_ctx.canEdit),
            showPassword: true,
        }));
        const __VLS_226 = __VLS_225({
            modelValue: (__VLS_ctx.forms.kdniao.api_key),
            placeholder: (__VLS_ctx.secretSet.kdniao.api_key ? '●●●●●● 已配置，留空则保持不变' : '请输入 API Key'),
            disabled: (!__VLS_ctx.canEdit),
            showPassword: true,
        }, ...__VLS_functionalComponentArgsRest(__VLS_225));
        var __VLS_223;
        const __VLS_228 = {}.ElFormItem;
        /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
        // @ts-ignore
        const __VLS_229 = __VLS_asFunctionalComponent(__VLS_228, new __VLS_228({
            label: "环境",
        }));
        const __VLS_230 = __VLS_229({
            label: "环境",
        }, ...__VLS_functionalComponentArgsRest(__VLS_229));
        __VLS_231.slots.default;
        const __VLS_232 = {}.ElSelect;
        /** @type {[typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, ]} */ ;
        // @ts-ignore
        const __VLS_233 = __VLS_asFunctionalComponent(__VLS_232, new __VLS_232({
            modelValue: (__VLS_ctx.forms.kdniao.env),
            ...{ style: {} },
            disabled: (!__VLS_ctx.canEdit),
        }));
        const __VLS_234 = __VLS_233({
            modelValue: (__VLS_ctx.forms.kdniao.env),
            ...{ style: {} },
            disabled: (!__VLS_ctx.canEdit),
        }, ...__VLS_functionalComponentArgsRest(__VLS_233));
        __VLS_235.slots.default;
        const __VLS_236 = {}.ElOption;
        /** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
        // @ts-ignore
        const __VLS_237 = __VLS_asFunctionalComponent(__VLS_236, new __VLS_236({
            label: "沙箱环境",
            value: "sandbox",
        }));
        const __VLS_238 = __VLS_237({
            label: "沙箱环境",
            value: "sandbox",
        }, ...__VLS_functionalComponentArgsRest(__VLS_237));
        const __VLS_240 = {}.ElOption;
        /** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
        // @ts-ignore
        const __VLS_241 = __VLS_asFunctionalComponent(__VLS_240, new __VLS_240({
            label: "生产环境",
            value: "production",
        }));
        const __VLS_242 = __VLS_241({
            label: "生产环境",
            value: "production",
        }, ...__VLS_functionalComponentArgsRest(__VLS_241));
        var __VLS_235;
        var __VLS_231;
        if (__VLS_ctx.canEdit) {
            const __VLS_244 = {}.ElFormItem;
            /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
            // @ts-ignore
            const __VLS_245 = __VLS_asFunctionalComponent(__VLS_244, new __VLS_244({}));
            const __VLS_246 = __VLS_245({}, ...__VLS_functionalComponentArgsRest(__VLS_245));
            __VLS_247.slots.default;
            const __VLS_248 = {}.ElButton;
            /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
            // @ts-ignore
            const __VLS_249 = __VLS_asFunctionalComponent(__VLS_248, new __VLS_248({
                ...{ 'onClick': {} },
                type: "primary",
                loading: (__VLS_ctx.savingGroup === 'kdniao'),
            }));
            const __VLS_250 = __VLS_249({
                ...{ 'onClick': {} },
                type: "primary",
                loading: (__VLS_ctx.savingGroup === 'kdniao'),
            }, ...__VLS_functionalComponentArgsRest(__VLS_249));
            let __VLS_252;
            let __VLS_253;
            let __VLS_254;
            const __VLS_255 = {
                onClick: (...[$event]) => {
                    if (!!(tab.name === 'basic'))
                        return;
                    if (!!(tab.name === 'wxpay'))
                        return;
                    if (!!(tab.name === 'wxlogin'))
                        return;
                    if (!!(tab.name === 'qywx'))
                        return;
                    if (!(tab.name === 'kdniao'))
                        return;
                    if (!(__VLS_ctx.canEdit))
                        return;
                    __VLS_ctx.handleSave('kdniao');
                }
            };
            __VLS_251.slots.default;
            var __VLS_251;
            var __VLS_247;
        }
        var __VLS_211;
    }
    else if (tab.name === 'sms') {
        const __VLS_256 = {}.ElForm;
        /** @type {[typeof __VLS_components.ElForm, typeof __VLS_components.elForm, typeof __VLS_components.ElForm, typeof __VLS_components.elForm, ]} */ ;
        // @ts-ignore
        const __VLS_257 = __VLS_asFunctionalComponent(__VLS_256, new __VLS_256({
            labelWidth: "160px",
            ...{ class: "setting-form" },
        }));
        const __VLS_258 = __VLS_257({
            labelWidth: "160px",
            ...{ class: "setting-form" },
        }, ...__VLS_functionalComponentArgsRest(__VLS_257));
        __VLS_259.slots.default;
        const __VLS_260 = {}.ElFormItem;
        /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
        // @ts-ignore
        const __VLS_261 = __VLS_asFunctionalComponent(__VLS_260, new __VLS_260({
            label: "短信服务商",
        }));
        const __VLS_262 = __VLS_261({
            label: "短信服务商",
        }, ...__VLS_functionalComponentArgsRest(__VLS_261));
        __VLS_263.slots.default;
        const __VLS_264 = {}.ElSelect;
        /** @type {[typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, ]} */ ;
        // @ts-ignore
        const __VLS_265 = __VLS_asFunctionalComponent(__VLS_264, new __VLS_264({
            modelValue: (__VLS_ctx.forms.sms.provider),
            ...{ style: {} },
            disabled: (!__VLS_ctx.canEdit),
        }));
        const __VLS_266 = __VLS_265({
            modelValue: (__VLS_ctx.forms.sms.provider),
            ...{ style: {} },
            disabled: (!__VLS_ctx.canEdit),
        }, ...__VLS_functionalComponentArgsRest(__VLS_265));
        __VLS_267.slots.default;
        const __VLS_268 = {}.ElOption;
        /** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
        // @ts-ignore
        const __VLS_269 = __VLS_asFunctionalComponent(__VLS_268, new __VLS_268({
            label: "阿里云",
            value: "aliyun",
        }));
        const __VLS_270 = __VLS_269({
            label: "阿里云",
            value: "aliyun",
        }, ...__VLS_functionalComponentArgsRest(__VLS_269));
        const __VLS_272 = {}.ElOption;
        /** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
        // @ts-ignore
        const __VLS_273 = __VLS_asFunctionalComponent(__VLS_272, new __VLS_272({
            label: "腾讯云",
            value: "tencent",
        }));
        const __VLS_274 = __VLS_273({
            label: "腾讯云",
            value: "tencent",
        }, ...__VLS_functionalComponentArgsRest(__VLS_273));
        var __VLS_267;
        var __VLS_263;
        const __VLS_276 = {}.ElFormItem;
        /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
        // @ts-ignore
        const __VLS_277 = __VLS_asFunctionalComponent(__VLS_276, new __VLS_276({
            label: "Access Key",
        }));
        const __VLS_278 = __VLS_277({
            label: "Access Key",
        }, ...__VLS_functionalComponentArgsRest(__VLS_277));
        __VLS_279.slots.default;
        const __VLS_280 = {}.ElInput;
        /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
        // @ts-ignore
        const __VLS_281 = __VLS_asFunctionalComponent(__VLS_280, new __VLS_280({
            modelValue: (__VLS_ctx.forms.sms.access_key),
            placeholder: "请输入 Access Key ID",
            disabled: (!__VLS_ctx.canEdit),
        }));
        const __VLS_282 = __VLS_281({
            modelValue: (__VLS_ctx.forms.sms.access_key),
            placeholder: "请输入 Access Key ID",
            disabled: (!__VLS_ctx.canEdit),
        }, ...__VLS_functionalComponentArgsRest(__VLS_281));
        var __VLS_279;
        const __VLS_284 = {}.ElFormItem;
        /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
        // @ts-ignore
        const __VLS_285 = __VLS_asFunctionalComponent(__VLS_284, new __VLS_284({
            label: "Secret Key",
        }));
        const __VLS_286 = __VLS_285({
            label: "Secret Key",
        }, ...__VLS_functionalComponentArgsRest(__VLS_285));
        __VLS_287.slots.default;
        const __VLS_288 = {}.ElInput;
        /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
        // @ts-ignore
        const __VLS_289 = __VLS_asFunctionalComponent(__VLS_288, new __VLS_288({
            modelValue: (__VLS_ctx.forms.sms.secret_key),
            placeholder: (__VLS_ctx.secretSet.sms.secret_key ? '●●●●●● 已配置，留空则保持不变' : '请输入 Secret Key'),
            disabled: (!__VLS_ctx.canEdit),
            showPassword: true,
        }));
        const __VLS_290 = __VLS_289({
            modelValue: (__VLS_ctx.forms.sms.secret_key),
            placeholder: (__VLS_ctx.secretSet.sms.secret_key ? '●●●●●● 已配置，留空则保持不变' : '请输入 Secret Key'),
            disabled: (!__VLS_ctx.canEdit),
            showPassword: true,
        }, ...__VLS_functionalComponentArgsRest(__VLS_289));
        var __VLS_287;
        const __VLS_292 = {}.ElFormItem;
        /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
        // @ts-ignore
        const __VLS_293 = __VLS_asFunctionalComponent(__VLS_292, new __VLS_292({
            label: "短信签名",
        }));
        const __VLS_294 = __VLS_293({
            label: "短信签名",
        }, ...__VLS_functionalComponentArgsRest(__VLS_293));
        __VLS_295.slots.default;
        const __VLS_296 = {}.ElInput;
        /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
        // @ts-ignore
        const __VLS_297 = __VLS_asFunctionalComponent(__VLS_296, new __VLS_296({
            modelValue: (__VLS_ctx.forms.sms.sign_name),
            placeholder: "例如 【xu-shop】",
            disabled: (!__VLS_ctx.canEdit),
        }));
        const __VLS_298 = __VLS_297({
            modelValue: (__VLS_ctx.forms.sms.sign_name),
            placeholder: "例如 【xu-shop】",
            disabled: (!__VLS_ctx.canEdit),
        }, ...__VLS_functionalComponentArgsRest(__VLS_297));
        var __VLS_295;
        const __VLS_300 = {}.ElFormItem;
        /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
        // @ts-ignore
        const __VLS_301 = __VLS_asFunctionalComponent(__VLS_300, new __VLS_300({
            label: "验证码模板 ID",
        }));
        const __VLS_302 = __VLS_301({
            label: "验证码模板 ID",
        }, ...__VLS_functionalComponentArgsRest(__VLS_301));
        __VLS_303.slots.default;
        const __VLS_304 = {}.ElInput;
        /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
        // @ts-ignore
        const __VLS_305 = __VLS_asFunctionalComponent(__VLS_304, new __VLS_304({
            modelValue: (__VLS_ctx.forms.sms.verify_template),
            placeholder: "短信验证码模板 ID",
            disabled: (!__VLS_ctx.canEdit),
        }));
        const __VLS_306 = __VLS_305({
            modelValue: (__VLS_ctx.forms.sms.verify_template),
            placeholder: "短信验证码模板 ID",
            disabled: (!__VLS_ctx.canEdit),
        }, ...__VLS_functionalComponentArgsRest(__VLS_305));
        var __VLS_303;
        if (__VLS_ctx.canEdit) {
            const __VLS_308 = {}.ElFormItem;
            /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
            // @ts-ignore
            const __VLS_309 = __VLS_asFunctionalComponent(__VLS_308, new __VLS_308({}));
            const __VLS_310 = __VLS_309({}, ...__VLS_functionalComponentArgsRest(__VLS_309));
            __VLS_311.slots.default;
            const __VLS_312 = {}.ElButton;
            /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
            // @ts-ignore
            const __VLS_313 = __VLS_asFunctionalComponent(__VLS_312, new __VLS_312({
                ...{ 'onClick': {} },
                type: "primary",
                loading: (__VLS_ctx.savingGroup === 'sms'),
            }));
            const __VLS_314 = __VLS_313({
                ...{ 'onClick': {} },
                type: "primary",
                loading: (__VLS_ctx.savingGroup === 'sms'),
            }, ...__VLS_functionalComponentArgsRest(__VLS_313));
            let __VLS_316;
            let __VLS_317;
            let __VLS_318;
            const __VLS_319 = {
                onClick: (...[$event]) => {
                    if (!!(tab.name === 'basic'))
                        return;
                    if (!!(tab.name === 'wxpay'))
                        return;
                    if (!!(tab.name === 'wxlogin'))
                        return;
                    if (!!(tab.name === 'qywx'))
                        return;
                    if (!!(tab.name === 'kdniao'))
                        return;
                    if (!(tab.name === 'sms'))
                        return;
                    if (!(__VLS_ctx.canEdit))
                        return;
                    __VLS_ctx.handleSave('sms');
                }
            };
            __VLS_315.slots.default;
            var __VLS_315;
            var __VLS_311;
        }
        var __VLS_259;
    }
    else if (tab.name === 'security') {
        const __VLS_320 = {}.ElForm;
        /** @type {[typeof __VLS_components.ElForm, typeof __VLS_components.elForm, typeof __VLS_components.ElForm, typeof __VLS_components.elForm, ]} */ ;
        // @ts-ignore
        const __VLS_321 = __VLS_asFunctionalComponent(__VLS_320, new __VLS_320({
            labelWidth: "160px",
            ...{ class: "setting-form" },
        }));
        const __VLS_322 = __VLS_321({
            labelWidth: "160px",
            ...{ class: "setting-form" },
        }, ...__VLS_functionalComponentArgsRest(__VLS_321));
        __VLS_323.slots.default;
        const __VLS_324 = {}.ElFormItem;
        /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
        // @ts-ignore
        const __VLS_325 = __VLS_asFunctionalComponent(__VLS_324, new __VLS_324({
            label: "会话时长（小时）",
        }));
        const __VLS_326 = __VLS_325({
            label: "会话时长（小时）",
        }, ...__VLS_functionalComponentArgsRest(__VLS_325));
        __VLS_327.slots.default;
        const __VLS_328 = {}.ElInputNumber;
        /** @type {[typeof __VLS_components.ElInputNumber, typeof __VLS_components.elInputNumber, ]} */ ;
        // @ts-ignore
        const __VLS_329 = __VLS_asFunctionalComponent(__VLS_328, new __VLS_328({
            modelValue: (__VLS_ctx.forms.security.session_hours),
            modelModifiers: { number: true, },
            min: (1),
            max: (720),
            disabled: (!__VLS_ctx.canEdit),
        }));
        const __VLS_330 = __VLS_329({
            modelValue: (__VLS_ctx.forms.security.session_hours),
            modelModifiers: { number: true, },
            min: (1),
            max: (720),
            disabled: (!__VLS_ctx.canEdit),
        }, ...__VLS_functionalComponentArgsRest(__VLS_329));
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
            ...{ class: "field-tip" },
        });
        var __VLS_327;
        const __VLS_332 = {}.ElFormItem;
        /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
        // @ts-ignore
        const __VLS_333 = __VLS_asFunctionalComponent(__VLS_332, new __VLS_332({
            label: "最大登录失败次数",
        }));
        const __VLS_334 = __VLS_333({
            label: "最大登录失败次数",
        }, ...__VLS_functionalComponentArgsRest(__VLS_333));
        __VLS_335.slots.default;
        const __VLS_336 = {}.ElInputNumber;
        /** @type {[typeof __VLS_components.ElInputNumber, typeof __VLS_components.elInputNumber, ]} */ ;
        // @ts-ignore
        const __VLS_337 = __VLS_asFunctionalComponent(__VLS_336, new __VLS_336({
            modelValue: (__VLS_ctx.forms.security.max_login_attempts),
            modelModifiers: { number: true, },
            min: (1),
            max: (100),
            disabled: (!__VLS_ctx.canEdit),
        }));
        const __VLS_338 = __VLS_337({
            modelValue: (__VLS_ctx.forms.security.max_login_attempts),
            modelModifiers: { number: true, },
            min: (1),
            max: (100),
            disabled: (!__VLS_ctx.canEdit),
        }, ...__VLS_functionalComponentArgsRest(__VLS_337));
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
            ...{ class: "field-tip" },
        });
        var __VLS_335;
        const __VLS_340 = {}.ElFormItem;
        /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
        // @ts-ignore
        const __VLS_341 = __VLS_asFunctionalComponent(__VLS_340, new __VLS_340({
            label: "密码最短长度",
        }));
        const __VLS_342 = __VLS_341({
            label: "密码最短长度",
        }, ...__VLS_functionalComponentArgsRest(__VLS_341));
        __VLS_343.slots.default;
        const __VLS_344 = {}.ElInputNumber;
        /** @type {[typeof __VLS_components.ElInputNumber, typeof __VLS_components.elInputNumber, ]} */ ;
        // @ts-ignore
        const __VLS_345 = __VLS_asFunctionalComponent(__VLS_344, new __VLS_344({
            modelValue: (__VLS_ctx.forms.security.admin_pw_min_len),
            modelModifiers: { number: true, },
            min: (8),
            max: (64),
            disabled: (!__VLS_ctx.canEdit),
        }));
        const __VLS_346 = __VLS_345({
            modelValue: (__VLS_ctx.forms.security.admin_pw_min_len),
            modelModifiers: { number: true, },
            min: (8),
            max: (64),
            disabled: (!__VLS_ctx.canEdit),
        }, ...__VLS_functionalComponentArgsRest(__VLS_345));
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
            ...{ class: "field-tip" },
        });
        var __VLS_343;
        if (__VLS_ctx.canEdit) {
            const __VLS_348 = {}.ElFormItem;
            /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
            // @ts-ignore
            const __VLS_349 = __VLS_asFunctionalComponent(__VLS_348, new __VLS_348({}));
            const __VLS_350 = __VLS_349({}, ...__VLS_functionalComponentArgsRest(__VLS_349));
            __VLS_351.slots.default;
            const __VLS_352 = {}.ElButton;
            /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
            // @ts-ignore
            const __VLS_353 = __VLS_asFunctionalComponent(__VLS_352, new __VLS_352({
                ...{ 'onClick': {} },
                type: "primary",
                loading: (__VLS_ctx.savingGroup === 'security'),
            }));
            const __VLS_354 = __VLS_353({
                ...{ 'onClick': {} },
                type: "primary",
                loading: (__VLS_ctx.savingGroup === 'security'),
            }, ...__VLS_functionalComponentArgsRest(__VLS_353));
            let __VLS_356;
            let __VLS_357;
            let __VLS_358;
            const __VLS_359 = {
                onClick: (...[$event]) => {
                    if (!!(tab.name === 'basic'))
                        return;
                    if (!!(tab.name === 'wxpay'))
                        return;
                    if (!!(tab.name === 'wxlogin'))
                        return;
                    if (!!(tab.name === 'qywx'))
                        return;
                    if (!!(tab.name === 'kdniao'))
                        return;
                    if (!!(tab.name === 'sms'))
                        return;
                    if (!(tab.name === 'security'))
                        return;
                    if (!(__VLS_ctx.canEdit))
                        return;
                    __VLS_ctx.handleSave('security');
                }
            };
            __VLS_355.slots.default;
            var __VLS_355;
            var __VLS_351;
        }
        var __VLS_323;
    }
    var __VLS_15;
}
var __VLS_7;
/** @type {__VLS_StyleScopedClasses['page-card']} */ ;
/** @type {__VLS_StyleScopedClasses['page-head']} */ ;
/** @type {__VLS_StyleScopedClasses['tab-content']} */ ;
/** @type {__VLS_StyleScopedClasses['setting-form']} */ ;
/** @type {__VLS_StyleScopedClasses['setting-form']} */ ;
/** @type {__VLS_StyleScopedClasses['setting-form']} */ ;
/** @type {__VLS_StyleScopedClasses['setting-form']} */ ;
/** @type {__VLS_StyleScopedClasses['setting-form']} */ ;
/** @type {__VLS_StyleScopedClasses['setting-form']} */ ;
/** @type {__VLS_StyleScopedClasses['setting-form']} */ ;
/** @type {__VLS_StyleScopedClasses['field-tip']} */ ;
/** @type {__VLS_StyleScopedClasses['field-tip']} */ ;
/** @type {__VLS_StyleScopedClasses['field-tip']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            canEdit: canEdit,
            tabs: tabs,
            forms: forms,
            loadingGroup: loadingGroup,
            savingGroup: savingGroup,
            secretSet: secretSet,
            activeTab: activeTab,
            handleTabChange: handleTabChange,
            handleSave: handleSave,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
