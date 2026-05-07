import { ref, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { ElMessage } from 'element-plus';
import { useAuthStore } from '@/stores/auth';
import { getCaptcha } from '@/api/account';
const router = useRouter();
const authStore = useAuthStore();
const formRef = ref();
const loading = ref(false);
const captchaId = ref('');
const captchaB64 = ref('');
const captchaError = ref(false);
const signalCards = [
    { label: '待发货订单', value: '23', detail: '高峰时段 3 分钟内响应' },
    { label: '库存预警', value: '05', detail: '补货队列已同步' },
    { label: '私域转化', value: '+18%', detail: '近 24 小时渠道增长' },
];
const operationRows = [
    { module: '订单', metric: '1,284', trend: '+12.6%', tone: 'up' },
    { module: '客户', metric: '486', trend: '+08.4%', tone: 'up' },
    { module: '库存', metric: '32', trend: '-02.1%', tone: 'down' },
    { module: '投放', metric: '74', trend: '+05.7%', tone: 'up' },
];
const checkpoints = [
    '验证码动态刷新，降低撞库风险',
    '角色权限与路由守卫统一校验',
    '登录成功后直达工作台',
];
const form = ref({
    username: '',
    password: '',
    captcha_code: '',
});
const rules = {
    username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
    password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
    captcha_code: [{ required: true, message: '请输入验证码', trigger: 'blur' }],
};
async function loadCaptcha() {
    captchaError.value = false;
    try {
        const res = await getCaptcha();
        captchaId.value = res.captcha_id;
        captchaB64.value = res.captcha_b64;
    }
    catch {
        captchaB64.value = '';
        captchaError.value = true;
    }
}
async function handleLogin() {
    await formRef.value?.validate();
    loading.value = true;
    try {
        await authStore.login({
            ...form.value,
            captcha_id: captchaId.value,
        });
        ElMessage.success('登录成功');
        router.push('/workbench');
    }
    catch {
        await loadCaptcha();
        form.value.captcha_code = '';
    }
    finally {
        loading.value = false;
    }
}
onMounted(loadCaptcha);
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
/** @type {__VLS_StyleScopedClasses['login-shell__noise']} */ ;
/** @type {__VLS_StyleScopedClasses['login-shell__beam']} */ ;
/** @type {__VLS_StyleScopedClasses['market-stage']} */ ;
/** @type {__VLS_StyleScopedClasses['stage-logo__tile']} */ ;
/** @type {__VLS_StyleScopedClasses['stage-logo__core']} */ ;
/** @type {__VLS_StyleScopedClasses['market-stage__intro']} */ ;
/** @type {__VLS_StyleScopedClasses['signal-card']} */ ;
/** @type {__VLS_StyleScopedClasses['signal-card']} */ ;
/** @type {__VLS_StyleScopedClasses['ops-board__ticker']} */ ;
/** @type {__VLS_StyleScopedClasses['ticker-head']} */ ;
/** @type {__VLS_StyleScopedClasses['ticker-row']} */ ;
/** @type {__VLS_StyleScopedClasses['rail-card']} */ ;
/** @type {__VLS_StyleScopedClasses['auth-panel']} */ ;
/** @type {__VLS_StyleScopedClasses['el-input__inner']} */ ;
/** @type {__VLS_StyleScopedClasses['el-input__wrapper']} */ ;
/** @type {__VLS_StyleScopedClasses['el-form-item']} */ ;
/** @type {__VLS_StyleScopedClasses['el-input__wrapper']} */ ;
/** @type {__VLS_StyleScopedClasses['captcha-panel']} */ ;
/** @type {__VLS_StyleScopedClasses['captcha-panel--fallback']} */ ;
/** @type {__VLS_StyleScopedClasses['footer-note']} */ ;
/** @type {__VLS_StyleScopedClasses['el-icon']} */ ;
/** @type {__VLS_StyleScopedClasses['login-grid']} */ ;
/** @type {__VLS_StyleScopedClasses['auth-panel']} */ ;
/** @type {__VLS_StyleScopedClasses['market-stage']} */ ;
/** @type {__VLS_StyleScopedClasses['signal-strip']} */ ;
/** @type {__VLS_StyleScopedClasses['ops-board__grid']} */ ;
/** @type {__VLS_StyleScopedClasses['login-grid']} */ ;
/** @type {__VLS_StyleScopedClasses['market-stage']} */ ;
/** @type {__VLS_StyleScopedClasses['auth-panel']} */ ;
/** @type {__VLS_StyleScopedClasses['market-stage__masthead']} */ ;
/** @type {__VLS_StyleScopedClasses['market-stage__intro']} */ ;
/** @type {__VLS_StyleScopedClasses['market-stage__intro']} */ ;
/** @type {__VLS_StyleScopedClasses['captcha-row']} */ ;
/** @type {__VLS_StyleScopedClasses['captcha-panel']} */ ;
/** @type {__VLS_StyleScopedClasses['captcha-panel--fallback']} */ ;
/** @type {__VLS_StyleScopedClasses['login-grid']} */ ;
/** @type {__VLS_StyleScopedClasses['market-stage']} */ ;
/** @type {__VLS_StyleScopedClasses['auth-panel']} */ ;
/** @type {__VLS_StyleScopedClasses['market-stage']} */ ;
/** @type {__VLS_StyleScopedClasses['market-stage__intro']} */ ;
/** @type {__VLS_StyleScopedClasses['market-stage__lead']} */ ;
/** @type {__VLS_StyleScopedClasses['signal-card']} */ ;
/** @type {__VLS_StyleScopedClasses['market-stage__lead']} */ ;
/** @type {__VLS_StyleScopedClasses['signal-card']} */ ;
/** @type {__VLS_StyleScopedClasses['rail-card']} */ ;
/** @type {__VLS_StyleScopedClasses['auth-panel__header']} */ ;
/** @type {__VLS_StyleScopedClasses['auth-panel__header']} */ ;
// CSS variable injection 
// CSS variable injection end 
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "login-shell" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div)({
    ...{ class: "login-shell__noise" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div)({
    ...{ class: "login-shell__beam login-shell__beam--left" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div)({
    ...{ class: "login-shell__beam login-shell__beam--right" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "login-grid" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.section, __VLS_intrinsicElements.section)({
    ...{ class: "market-stage" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "market-stage__masthead" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "stage-logo" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
    ...{ class: "stage-logo__mark" },
    'aria-hidden': "true",
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
    ...{ class: "stage-logo__glyph" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.i)({
    ...{ class: "stage-logo__tile stage-logo__tile--top" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.i)({
    ...{ class: "stage-logo__tile stage-logo__tile--right" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.i)({
    ...{ class: "stage-logo__tile stage-logo__tile--bottom" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.i)({
    ...{ class: "stage-logo__tile stage-logo__tile--left" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.i)({
    ...{ class: "stage-logo__core" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "stage-status" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.span)({
    ...{ class: "stage-status__dot" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "market-stage__intro" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({
    ...{ class: "market-stage__eyebrow" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.h1, __VLS_intrinsicElements.h1)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({
    ...{ class: "market-stage__lead" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "signal-strip" },
});
for (const [card] of __VLS_getVForSourceType((__VLS_ctx.signalCards))) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.article, __VLS_intrinsicElements.article)({
        key: (card.label),
        ...{ class: "signal-card" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
    (card.label);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.strong, __VLS_intrinsicElements.strong)({});
    (card.value);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.small, __VLS_intrinsicElements.small)({});
    (card.detail);
}
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "ops-board" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "ops-board__header" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "ops-board__tag" },
});
const __VLS_0 = {}.ElIcon;
/** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
// @ts-ignore
const __VLS_1 = __VLS_asFunctionalComponent(__VLS_0, new __VLS_0({}));
const __VLS_2 = __VLS_1({}, ...__VLS_functionalComponentArgsRest(__VLS_1));
__VLS_3.slots.default;
const __VLS_4 = {}.DataLine;
/** @type {[typeof __VLS_components.DataLine, ]} */ ;
// @ts-ignore
const __VLS_5 = __VLS_asFunctionalComponent(__VLS_4, new __VLS_4({}));
const __VLS_6 = __VLS_5({}, ...__VLS_functionalComponentArgsRest(__VLS_5));
var __VLS_3;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "ops-board__grid" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "ops-board__ticker" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "ticker-head" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
for (const [row] of __VLS_getVForSourceType((__VLS_ctx.operationRows))) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        key: (row.module),
        ...{ class: "ticker-row" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ class: "ticker-row__module" },
    });
    (row.module);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.strong, __VLS_intrinsicElements.strong)({});
    (row.metric);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ class: (['ticker-row__trend', `ticker-row__trend--${row.tone}`]) },
    });
    (row.trend);
}
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "ops-board__rail" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "rail-card" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.strong, __VLS_intrinsicElements.strong)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "rail-card" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.strong, __VLS_intrinsicElements.strong)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.section, __VLS_intrinsicElements.section)({
    ...{ class: "auth-panel" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "auth-panel__header" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "auth-panel__badge" },
});
const __VLS_8 = {}.ElIcon;
/** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
// @ts-ignore
const __VLS_9 = __VLS_asFunctionalComponent(__VLS_8, new __VLS_8({}));
const __VLS_10 = __VLS_9({}, ...__VLS_functionalComponentArgsRest(__VLS_9));
__VLS_11.slots.default;
const __VLS_12 = {}.Key;
/** @type {[typeof __VLS_components.Key, ]} */ ;
// @ts-ignore
const __VLS_13 = __VLS_asFunctionalComponent(__VLS_12, new __VLS_12({}));
const __VLS_14 = __VLS_13({}, ...__VLS_functionalComponentArgsRest(__VLS_13));
var __VLS_11;
__VLS_asFunctionalElement(__VLS_intrinsicElements.h2, __VLS_intrinsicElements.h2)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({});
const __VLS_16 = {}.ElForm;
/** @type {[typeof __VLS_components.ElForm, typeof __VLS_components.elForm, typeof __VLS_components.ElForm, typeof __VLS_components.elForm, ]} */ ;
// @ts-ignore
const __VLS_17 = __VLS_asFunctionalComponent(__VLS_16, new __VLS_16({
    ...{ 'onKeyup': {} },
    ref: "formRef",
    ...{ class: "auth-form" },
    model: (__VLS_ctx.form),
    rules: (__VLS_ctx.rules),
    labelPosition: "top",
    size: "large",
}));
const __VLS_18 = __VLS_17({
    ...{ 'onKeyup': {} },
    ref: "formRef",
    ...{ class: "auth-form" },
    model: (__VLS_ctx.form),
    rules: (__VLS_ctx.rules),
    labelPosition: "top",
    size: "large",
}, ...__VLS_functionalComponentArgsRest(__VLS_17));
let __VLS_20;
let __VLS_21;
let __VLS_22;
const __VLS_23 = {
    onKeyup: (__VLS_ctx.handleLogin)
};
/** @type {typeof __VLS_ctx.formRef} */ ;
var __VLS_24 = {};
__VLS_19.slots.default;
const __VLS_26 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_27 = __VLS_asFunctionalComponent(__VLS_26, new __VLS_26({
    label: "用户名",
    prop: "username",
}));
const __VLS_28 = __VLS_27({
    label: "用户名",
    prop: "username",
}, ...__VLS_functionalComponentArgsRest(__VLS_27));
__VLS_29.slots.default;
const __VLS_30 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_31 = __VLS_asFunctionalComponent(__VLS_30, new __VLS_30({
    modelValue: (__VLS_ctx.form.username),
    placeholder: "请输入用户名",
    autocomplete: "username",
}));
const __VLS_32 = __VLS_31({
    modelValue: (__VLS_ctx.form.username),
    placeholder: "请输入用户名",
    autocomplete: "username",
}, ...__VLS_functionalComponentArgsRest(__VLS_31));
__VLS_33.slots.default;
{
    const { prefix: __VLS_thisSlot } = __VLS_33.slots;
    const __VLS_34 = {}.ElIcon;
    /** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
    // @ts-ignore
    const __VLS_35 = __VLS_asFunctionalComponent(__VLS_34, new __VLS_34({}));
    const __VLS_36 = __VLS_35({}, ...__VLS_functionalComponentArgsRest(__VLS_35));
    __VLS_37.slots.default;
    const __VLS_38 = {}.User;
    /** @type {[typeof __VLS_components.User, ]} */ ;
    // @ts-ignore
    const __VLS_39 = __VLS_asFunctionalComponent(__VLS_38, new __VLS_38({}));
    const __VLS_40 = __VLS_39({}, ...__VLS_functionalComponentArgsRest(__VLS_39));
    var __VLS_37;
}
var __VLS_33;
var __VLS_29;
const __VLS_42 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_43 = __VLS_asFunctionalComponent(__VLS_42, new __VLS_42({
    label: "密码",
    prop: "password",
}));
const __VLS_44 = __VLS_43({
    label: "密码",
    prop: "password",
}, ...__VLS_functionalComponentArgsRest(__VLS_43));
__VLS_45.slots.default;
const __VLS_46 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_47 = __VLS_asFunctionalComponent(__VLS_46, new __VLS_46({
    modelValue: (__VLS_ctx.form.password),
    type: "password",
    placeholder: "请输入密码",
    showPassword: true,
    autocomplete: "current-password",
}));
const __VLS_48 = __VLS_47({
    modelValue: (__VLS_ctx.form.password),
    type: "password",
    placeholder: "请输入密码",
    showPassword: true,
    autocomplete: "current-password",
}, ...__VLS_functionalComponentArgsRest(__VLS_47));
__VLS_49.slots.default;
{
    const { prefix: __VLS_thisSlot } = __VLS_49.slots;
    const __VLS_50 = {}.ElIcon;
    /** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
    // @ts-ignore
    const __VLS_51 = __VLS_asFunctionalComponent(__VLS_50, new __VLS_50({}));
    const __VLS_52 = __VLS_51({}, ...__VLS_functionalComponentArgsRest(__VLS_51));
    __VLS_53.slots.default;
    const __VLS_54 = {}.Key;
    /** @type {[typeof __VLS_components.Key, ]} */ ;
    // @ts-ignore
    const __VLS_55 = __VLS_asFunctionalComponent(__VLS_54, new __VLS_54({}));
    const __VLS_56 = __VLS_55({}, ...__VLS_functionalComponentArgsRest(__VLS_55));
    var __VLS_53;
}
var __VLS_49;
var __VLS_45;
const __VLS_58 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_59 = __VLS_asFunctionalComponent(__VLS_58, new __VLS_58({
    label: "验证码",
    prop: "captcha_code",
}));
const __VLS_60 = __VLS_59({
    label: "验证码",
    prop: "captcha_code",
}, ...__VLS_functionalComponentArgsRest(__VLS_59));
__VLS_61.slots.default;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "captcha-row" },
});
const __VLS_62 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_63 = __VLS_asFunctionalComponent(__VLS_62, new __VLS_62({
    modelValue: (__VLS_ctx.form.captcha_code),
    placeholder: "输入图形验证码",
}));
const __VLS_64 = __VLS_63({
    modelValue: (__VLS_ctx.form.captcha_code),
    placeholder: "输入图形验证码",
}, ...__VLS_functionalComponentArgsRest(__VLS_63));
__VLS_65.slots.default;
{
    const { prefix: __VLS_thisSlot } = __VLS_65.slots;
    const __VLS_66 = {}.ElIcon;
    /** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
    // @ts-ignore
    const __VLS_67 = __VLS_asFunctionalComponent(__VLS_66, new __VLS_66({}));
    const __VLS_68 = __VLS_67({}, ...__VLS_functionalComponentArgsRest(__VLS_67));
    __VLS_69.slots.default;
    const __VLS_70 = {}.Connection;
    /** @type {[typeof __VLS_components.Connection, ]} */ ;
    // @ts-ignore
    const __VLS_71 = __VLS_asFunctionalComponent(__VLS_70, new __VLS_70({}));
    const __VLS_72 = __VLS_71({}, ...__VLS_functionalComponentArgsRest(__VLS_71));
    var __VLS_69;
}
var __VLS_65;
if (__VLS_ctx.captchaB64 && !__VLS_ctx.captchaError) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.button, __VLS_intrinsicElements.button)({
        ...{ onClick: (__VLS_ctx.loadCaptcha) },
        type: "button",
        ...{ class: "captcha-panel" },
        title: "点击刷新验证码",
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.img, __VLS_intrinsicElements.img)({
        ...{ onError: (...[$event]) => {
                if (!(__VLS_ctx.captchaB64 && !__VLS_ctx.captchaError))
                    return;
                __VLS_ctx.captchaError = true;
            } },
        src: (__VLS_ctx.captchaB64),
        alt: "验证码",
    });
}
else {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.button, __VLS_intrinsicElements.button)({
        ...{ onClick: (__VLS_ctx.loadCaptcha) },
        type: "button",
        ...{ class: "captcha-panel captcha-panel--fallback" },
    });
    const __VLS_74 = {}.ElIcon;
    /** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
    // @ts-ignore
    const __VLS_75 = __VLS_asFunctionalComponent(__VLS_74, new __VLS_74({}));
    const __VLS_76 = __VLS_75({}, ...__VLS_functionalComponentArgsRest(__VLS_75));
    __VLS_77.slots.default;
    const __VLS_78 = {}.Warning;
    /** @type {[typeof __VLS_components.Warning, ]} */ ;
    // @ts-ignore
    const __VLS_79 = __VLS_asFunctionalComponent(__VLS_78, new __VLS_78({}));
    const __VLS_80 = __VLS_79({}, ...__VLS_functionalComponentArgsRest(__VLS_79));
    var __VLS_77;
    (__VLS_ctx.captchaError ? '获取失败，点击重试' : '刷新验证码');
}
var __VLS_61;
const __VLS_82 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_83 = __VLS_asFunctionalComponent(__VLS_82, new __VLS_82({
    ...{ 'onClick': {} },
    type: "primary",
    ...{ class: "submit-button" },
    loading: (__VLS_ctx.loading),
}));
const __VLS_84 = __VLS_83({
    ...{ 'onClick': {} },
    type: "primary",
    ...{ class: "submit-button" },
    loading: (__VLS_ctx.loading),
}, ...__VLS_functionalComponentArgsRest(__VLS_83));
let __VLS_86;
let __VLS_87;
let __VLS_88;
const __VLS_89 = {
    onClick: (__VLS_ctx.handleLogin)
};
__VLS_85.slots.default;
const __VLS_90 = {}.ElIcon;
/** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
// @ts-ignore
const __VLS_91 = __VLS_asFunctionalComponent(__VLS_90, new __VLS_90({}));
const __VLS_92 = __VLS_91({}, ...__VLS_functionalComponentArgsRest(__VLS_91));
__VLS_93.slots.default;
const __VLS_94 = {}.Right;
/** @type {[typeof __VLS_components.Right, ]} */ ;
// @ts-ignore
const __VLS_95 = __VLS_asFunctionalComponent(__VLS_94, new __VLS_94({}));
const __VLS_96 = __VLS_95({}, ...__VLS_functionalComponentArgsRest(__VLS_95));
var __VLS_93;
var __VLS_85;
var __VLS_19;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "auth-panel__footer" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "footer-note" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
    ...{ class: "footer-note__label" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.ul, __VLS_intrinsicElements.ul)({});
for (const [item] of __VLS_getVForSourceType((__VLS_ctx.checkpoints))) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.li, __VLS_intrinsicElements.li)({
        key: (item),
    });
    const __VLS_98 = {}.ElIcon;
    /** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
    // @ts-ignore
    const __VLS_99 = __VLS_asFunctionalComponent(__VLS_98, new __VLS_98({}));
    const __VLS_100 = __VLS_99({}, ...__VLS_functionalComponentArgsRest(__VLS_99));
    __VLS_101.slots.default;
    const __VLS_102 = {}.CircleCheckFilled;
    /** @type {[typeof __VLS_components.CircleCheckFilled, ]} */ ;
    // @ts-ignore
    const __VLS_103 = __VLS_asFunctionalComponent(__VLS_102, new __VLS_102({}));
    const __VLS_104 = __VLS_103({}, ...__VLS_functionalComponentArgsRest(__VLS_103));
    var __VLS_101;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
    (item);
}
/** @type {__VLS_StyleScopedClasses['login-shell']} */ ;
/** @type {__VLS_StyleScopedClasses['login-shell__noise']} */ ;
/** @type {__VLS_StyleScopedClasses['login-shell__beam']} */ ;
/** @type {__VLS_StyleScopedClasses['login-shell__beam--left']} */ ;
/** @type {__VLS_StyleScopedClasses['login-shell__beam']} */ ;
/** @type {__VLS_StyleScopedClasses['login-shell__beam--right']} */ ;
/** @type {__VLS_StyleScopedClasses['login-grid']} */ ;
/** @type {__VLS_StyleScopedClasses['market-stage']} */ ;
/** @type {__VLS_StyleScopedClasses['market-stage__masthead']} */ ;
/** @type {__VLS_StyleScopedClasses['stage-logo']} */ ;
/** @type {__VLS_StyleScopedClasses['stage-logo__mark']} */ ;
/** @type {__VLS_StyleScopedClasses['stage-logo__glyph']} */ ;
/** @type {__VLS_StyleScopedClasses['stage-logo__tile']} */ ;
/** @type {__VLS_StyleScopedClasses['stage-logo__tile--top']} */ ;
/** @type {__VLS_StyleScopedClasses['stage-logo__tile']} */ ;
/** @type {__VLS_StyleScopedClasses['stage-logo__tile--right']} */ ;
/** @type {__VLS_StyleScopedClasses['stage-logo__tile']} */ ;
/** @type {__VLS_StyleScopedClasses['stage-logo__tile--bottom']} */ ;
/** @type {__VLS_StyleScopedClasses['stage-logo__tile']} */ ;
/** @type {__VLS_StyleScopedClasses['stage-logo__tile--left']} */ ;
/** @type {__VLS_StyleScopedClasses['stage-logo__core']} */ ;
/** @type {__VLS_StyleScopedClasses['stage-status']} */ ;
/** @type {__VLS_StyleScopedClasses['stage-status__dot']} */ ;
/** @type {__VLS_StyleScopedClasses['market-stage__intro']} */ ;
/** @type {__VLS_StyleScopedClasses['market-stage__eyebrow']} */ ;
/** @type {__VLS_StyleScopedClasses['market-stage__lead']} */ ;
/** @type {__VLS_StyleScopedClasses['signal-strip']} */ ;
/** @type {__VLS_StyleScopedClasses['signal-card']} */ ;
/** @type {__VLS_StyleScopedClasses['ops-board']} */ ;
/** @type {__VLS_StyleScopedClasses['ops-board__header']} */ ;
/** @type {__VLS_StyleScopedClasses['ops-board__tag']} */ ;
/** @type {__VLS_StyleScopedClasses['ops-board__grid']} */ ;
/** @type {__VLS_StyleScopedClasses['ops-board__ticker']} */ ;
/** @type {__VLS_StyleScopedClasses['ticker-head']} */ ;
/** @type {__VLS_StyleScopedClasses['ticker-row']} */ ;
/** @type {__VLS_StyleScopedClasses['ticker-row__module']} */ ;
/** @type {__VLS_StyleScopedClasses['ops-board__rail']} */ ;
/** @type {__VLS_StyleScopedClasses['rail-card']} */ ;
/** @type {__VLS_StyleScopedClasses['rail-card']} */ ;
/** @type {__VLS_StyleScopedClasses['auth-panel']} */ ;
/** @type {__VLS_StyleScopedClasses['auth-panel__header']} */ ;
/** @type {__VLS_StyleScopedClasses['auth-panel__badge']} */ ;
/** @type {__VLS_StyleScopedClasses['auth-form']} */ ;
/** @type {__VLS_StyleScopedClasses['captcha-row']} */ ;
/** @type {__VLS_StyleScopedClasses['captcha-panel']} */ ;
/** @type {__VLS_StyleScopedClasses['captcha-panel']} */ ;
/** @type {__VLS_StyleScopedClasses['captcha-panel--fallback']} */ ;
/** @type {__VLS_StyleScopedClasses['submit-button']} */ ;
/** @type {__VLS_StyleScopedClasses['auth-panel__footer']} */ ;
/** @type {__VLS_StyleScopedClasses['footer-note']} */ ;
/** @type {__VLS_StyleScopedClasses['footer-note__label']} */ ;
// @ts-ignore
var __VLS_25 = __VLS_24;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            formRef: formRef,
            loading: loading,
            captchaB64: captchaB64,
            captchaError: captchaError,
            signalCards: signalCards,
            operationRows: operationRows,
            checkpoints: checkpoints,
            form: form,
            rules: rules,
            loadCaptcha: loadCaptcha,
            handleLogin: handleLogin,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
