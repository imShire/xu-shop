import { ref, onMounted, computed, watch, shallowRef, onBeforeUnmount } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { ElMessage } from 'element-plus';
import { Editor, Toolbar } from '@wangeditor/editor-for-vue';
import '@wangeditor/editor/dist/css/style.css';
import UploadImage from '@/components/UploadImage/index.vue';
import PriceInput from '@/components/PriceInput/index.vue';
import { getProduct, createProduct, updateProduct, getCategoryList, getFreightTemplates, uploadImage, } from '@/api/product';
const route = useRoute();
const router = useRouter();
const productId = computed(() => (route.params.id ? String(route.params.id) : null));
const isEdit = computed(() => !!productId.value);
const formRef = ref();
const saving = ref(false);
const categories = ref([]);
const freightTemplates = ref([]);
// wangEditor
const editorRef = shallowRef();
const editorConfig = {
    MENU_CONF: {
        uploadImage: {
            async customUpload(file, insertFn) {
                const res = await uploadImage(file);
                insertFn(res.url);
            },
        },
    },
};
onBeforeUnmount(() => editorRef.value?.destroy());
const UNIT_OPTIONS = ['件', '个', '套', 'kg', '份', '盒'];
const form = ref({
    title: '',
    subtitle: '',
    category_id: undefined,
    unit: '件',
    tags: [],
    sort: 0,
    status: 'draft',
    main_image: '',
    images: [],
    video_url: '',
    specs: [],
    skus: [],
    is_virtual: false,
    freight_template_id: null,
    detail_html: '',
    virtual_sales: 0,
    on_sale_at: null,
});
// Cartesian product for SKU generation
function cartesian(specs) {
    if (!specs.length)
        return [[]];
    const [head, ...tail] = specs;
    const tailCombinations = cartesian(tail);
    return head.values.flatMap((v) => tailCombinations.map((combo) => [v, ...combo]));
}
function rebuildSkus() {
    if (!form.value.specs.length)
        return;
    const combinations = cartesian(form.value.specs);
    const existingMap = new Map(form.value.skus.map((s) => [s.spec_values.join('|'), s]));
    form.value.skus = combinations.map((combo) => {
        const key = combo.join('|');
        const existing = existingMap.get(key);
        return {
            id: existing?.id,
            spec_values: combo,
            price_cents: existing?.price_cents ?? 0,
            original_price_cents: existing?.original_price_cents ?? 0,
            stock: existing?.stock ?? 0,
            weight_g: existing?.weight_g ?? 0,
            sku_code: existing?.sku_code ?? '',
            barcode: existing?.barcode ?? '',
            image: existing?.image ?? '',
            status: existing?.status ?? 'active',
        };
    });
}
watch(() => form.value.specs, rebuildSkus, { deep: true });
function addSpec() {
    if (form.value.specs.length >= 3)
        return;
    form.value.specs.push({ name: '', values: [] });
}
function removeSpec(idx) {
    form.value.specs.splice(idx, 1);
    if (!form.value.specs.length) {
        form.value.skus = [];
    }
}
const specInputs = ref([]);
function addSpecValue(specIdx, val) {
    if (val && !form.value.specs[specIdx].values.includes(val)) {
        form.value.specs[specIdx].values.push(val);
        specInputs.value[specIdx] = '';
    }
}
function removeSpecValue(specIdx, valIdx) {
    form.value.specs[specIdx].values.splice(valIdx, 1);
}
// 无规格单 SKU 快捷模式
const singleSku = computed(() => {
    if (form.value.specs.length > 0)
        return null;
    return form.value.skus[0] ?? null;
});
watch(() => form.value.specs.length, (len) => {
    if (len === 0) {
        if (!form.value.skus.length) {
            form.value.skus = [
                {
                    spec_values: [],
                    price_cents: 0,
                    original_price_cents: 0,
                    stock: 0,
                    weight_g: 0,
                    sku_code: '',
                    barcode: '',
                    image: '',
                    status: 'active',
                },
            ];
        }
    }
}, { immediate: true });
const formRules = {
    title: [
        { required: true, message: '请输入商品标题', trigger: 'blur' },
        { max: 60, message: '标题最多 60 个字符', trigger: 'blur' },
    ],
    category_id: [{ required: true, message: '请选择商品分类', trigger: 'change' }],
    main_image: [{ required: true, message: '请上传商品主图', trigger: 'change' }],
};
async function handleSave(targetStatus) {
    await formRef.value?.validate();
    // Validate SKU prices
    const hasInvalidSku = form.value.skus.some((sku) => !(sku.price_cents > 0));
    if (hasInvalidSku) {
        ElMessage.warning('请为所有 SKU 设置售价');
        return;
    }
    saving.value = true;
    try {
        const payload = {
            ...form.value,
            skus: form.value.skus.map((sku) => ({
                ...sku,
                price_cents: Math.round(sku.price_cents),
                original_price_cents: Math.round(sku.original_price_cents),
            })),
        };
        if (targetStatus) {
            payload.status = targetStatus;
        }
        if (isEdit.value) {
            await updateProduct(productId.value, payload);
            ElMessage.success('保存成功');
            if (targetStatus)
                form.value.status = targetStatus;
        }
        else {
            await createProduct(payload);
            ElMessage.success('创建成功');
            router.push('/product/list');
        }
    }
    finally {
        saving.value = false;
    }
}
onMounted(async () => {
    ;
    [categories.value, freightTemplates.value] = await Promise.all([
        getCategoryList().catch(() => []),
        getFreightTemplates().catch(() => []),
    ]);
    if (isEdit.value) {
        const product = await getProduct(productId.value);
        Object.assign(form.value, {
            ...product,
            unit: product.unit || '件',
            sort: product.sort ?? 0,
            virtual_sales: product.virtual_sales ?? 0,
            is_virtual: product.is_virtual ?? false,
            freight_template_id: product.freight_template_id ?? null,
            on_sale_at: product.on_sale_at ?? null,
            detail_html: product.detail_html ?? '',
            video_url: product.video_url ?? '',
            skus: (product.skus || []).map((sku) => ({
                ...sku,
                price_cents: sku.price_cents,
                original_price_cents: sku.original_price_cents,
                weight_g: sku.weight_g ?? 0,
                barcode: sku.barcode ?? '',
                image: sku.image ?? '',
                status: sku.status ?? 'active',
            })),
        });
    }
    else {
        // Init single SKU for no-spec mode
        form.value.skus = [
            {
                spec_values: [],
                price_cents: 0,
                original_price_cents: 0,
                stock: 0,
                weight_g: 0,
                sku_code: '',
                barcode: '',
                image: '',
                status: 'active',
            },
        ];
    }
});
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "page-card" },
    ...{ style: {} },
});
const __VLS_0 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_1 = __VLS_asFunctionalComponent(__VLS_0, new __VLS_0({
    ...{ 'onClick': {} },
}));
const __VLS_2 = __VLS_1({
    ...{ 'onClick': {} },
}, ...__VLS_functionalComponentArgsRest(__VLS_1));
let __VLS_4;
let __VLS_5;
let __VLS_6;
const __VLS_7 = {
    onClick: (...[$event]) => {
        __VLS_ctx.router.back();
    }
};
__VLS_3.slots.default;
var __VLS_3;
__VLS_asFunctionalElement(__VLS_intrinsicElements.h3, __VLS_intrinsicElements.h3)({
    ...{ style: {} },
});
(__VLS_ctx.isEdit ? '编辑商品' : '新建商品');
const __VLS_8 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_9 = __VLS_asFunctionalComponent(__VLS_8, new __VLS_8({
    ...{ 'onClick': {} },
    loading: (__VLS_ctx.saving),
}));
const __VLS_10 = __VLS_9({
    ...{ 'onClick': {} },
    loading: (__VLS_ctx.saving),
}, ...__VLS_functionalComponentArgsRest(__VLS_9));
let __VLS_12;
let __VLS_13;
let __VLS_14;
const __VLS_15 = {
    onClick: (...[$event]) => {
        __VLS_ctx.handleSave();
    }
};
__VLS_11.slots.default;
var __VLS_11;
const __VLS_16 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_17 = __VLS_asFunctionalComponent(__VLS_16, new __VLS_16({
    ...{ 'onClick': {} },
    type: "primary",
    loading: (__VLS_ctx.saving),
}));
const __VLS_18 = __VLS_17({
    ...{ 'onClick': {} },
    type: "primary",
    loading: (__VLS_ctx.saving),
}, ...__VLS_functionalComponentArgsRest(__VLS_17));
let __VLS_20;
let __VLS_21;
let __VLS_22;
const __VLS_23 = {
    onClick: (...[$event]) => {
        __VLS_ctx.handleSave('onsale');
    }
};
__VLS_19.slots.default;
var __VLS_19;
const __VLS_24 = {}.ElForm;
/** @type {[typeof __VLS_components.ElForm, typeof __VLS_components.elForm, typeof __VLS_components.ElForm, typeof __VLS_components.elForm, ]} */ ;
// @ts-ignore
const __VLS_25 = __VLS_asFunctionalComponent(__VLS_24, new __VLS_24({
    ref: "formRef",
    model: (__VLS_ctx.form),
    rules: (__VLS_ctx.formRules),
    labelWidth: "110px",
}));
const __VLS_26 = __VLS_25({
    ref: "formRef",
    model: (__VLS_ctx.form),
    rules: (__VLS_ctx.formRules),
    labelWidth: "110px",
}, ...__VLS_functionalComponentArgsRest(__VLS_25));
/** @type {typeof __VLS_ctx.formRef} */ ;
var __VLS_28 = {};
__VLS_27.slots.default;
const __VLS_30 = {}.ElCard;
/** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
// @ts-ignore
const __VLS_31 = __VLS_asFunctionalComponent(__VLS_30, new __VLS_30({
    ...{ style: {} },
}));
const __VLS_32 = __VLS_31({
    ...{ style: {} },
}, ...__VLS_functionalComponentArgsRest(__VLS_31));
__VLS_33.slots.default;
{
    const { header: __VLS_thisSlot } = __VLS_33.slots;
}
const __VLS_34 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_35 = __VLS_asFunctionalComponent(__VLS_34, new __VLS_34({
    label: "商品标题",
    prop: "title",
}));
const __VLS_36 = __VLS_35({
    label: "商品标题",
    prop: "title",
}, ...__VLS_functionalComponentArgsRest(__VLS_35));
__VLS_37.slots.default;
const __VLS_38 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_39 = __VLS_asFunctionalComponent(__VLS_38, new __VLS_38({
    modelValue: (__VLS_ctx.form.title),
    placeholder: "请输入商品标题（最多 60 字）",
    maxlength: "60",
    showWordLimit: true,
}));
const __VLS_40 = __VLS_39({
    modelValue: (__VLS_ctx.form.title),
    placeholder: "请输入商品标题（最多 60 字）",
    maxlength: "60",
    showWordLimit: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_39));
var __VLS_37;
const __VLS_42 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_43 = __VLS_asFunctionalComponent(__VLS_42, new __VLS_42({
    label: "副标题",
    prop: "subtitle",
}));
const __VLS_44 = __VLS_43({
    label: "副标题",
    prop: "subtitle",
}, ...__VLS_functionalComponentArgsRest(__VLS_43));
__VLS_45.slots.default;
const __VLS_46 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_47 = __VLS_asFunctionalComponent(__VLS_46, new __VLS_46({
    modelValue: (__VLS_ctx.form.subtitle),
    placeholder: "请输入副标题",
}));
const __VLS_48 = __VLS_47({
    modelValue: (__VLS_ctx.form.subtitle),
    placeholder: "请输入副标题",
}, ...__VLS_functionalComponentArgsRest(__VLS_47));
var __VLS_45;
const __VLS_50 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_51 = __VLS_asFunctionalComponent(__VLS_50, new __VLS_50({
    label: "商品分类",
    prop: "category_id",
}));
const __VLS_52 = __VLS_51({
    label: "商品分类",
    prop: "category_id",
}, ...__VLS_functionalComponentArgsRest(__VLS_51));
__VLS_53.slots.default;
const __VLS_54 = {}.ElSelect;
/** @type {[typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, ]} */ ;
// @ts-ignore
const __VLS_55 = __VLS_asFunctionalComponent(__VLS_54, new __VLS_54({
    modelValue: (__VLS_ctx.form.category_id),
    placeholder: "请选择分类",
    filterable: true,
    ...{ style: {} },
}));
const __VLS_56 = __VLS_55({
    modelValue: (__VLS_ctx.form.category_id),
    placeholder: "请选择分类",
    filterable: true,
    ...{ style: {} },
}, ...__VLS_functionalComponentArgsRest(__VLS_55));
__VLS_57.slots.default;
for (const [cat] of __VLS_getVForSourceType((__VLS_ctx.categories))) {
    const __VLS_58 = {}.ElOption;
    /** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
    // @ts-ignore
    const __VLS_59 = __VLS_asFunctionalComponent(__VLS_58, new __VLS_58({
        key: (cat.id),
        label: (cat.name),
        value: (cat.id),
    }));
    const __VLS_60 = __VLS_59({
        key: (cat.id),
        label: (cat.name),
        value: (cat.id),
    }, ...__VLS_functionalComponentArgsRest(__VLS_59));
}
var __VLS_57;
var __VLS_53;
const __VLS_62 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_63 = __VLS_asFunctionalComponent(__VLS_62, new __VLS_62({
    label: "单位",
    prop: "unit",
}));
const __VLS_64 = __VLS_63({
    label: "单位",
    prop: "unit",
}, ...__VLS_functionalComponentArgsRest(__VLS_63));
__VLS_65.slots.default;
const __VLS_66 = {}.ElSelect;
/** @type {[typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, ]} */ ;
// @ts-ignore
const __VLS_67 = __VLS_asFunctionalComponent(__VLS_66, new __VLS_66({
    modelValue: (__VLS_ctx.form.unit),
    allowCreate: true,
    filterable: true,
    placeholder: "选择或输入单位",
    ...{ style: {} },
}));
const __VLS_68 = __VLS_67({
    modelValue: (__VLS_ctx.form.unit),
    allowCreate: true,
    filterable: true,
    placeholder: "选择或输入单位",
    ...{ style: {} },
}, ...__VLS_functionalComponentArgsRest(__VLS_67));
__VLS_69.slots.default;
for (const [u] of __VLS_getVForSourceType((__VLS_ctx.UNIT_OPTIONS))) {
    const __VLS_70 = {}.ElOption;
    /** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
    // @ts-ignore
    const __VLS_71 = __VLS_asFunctionalComponent(__VLS_70, new __VLS_70({
        key: (u),
        label: (u),
        value: (u),
    }));
    const __VLS_72 = __VLS_71({
        key: (u),
        label: (u),
        value: (u),
    }, ...__VLS_functionalComponentArgsRest(__VLS_71));
}
var __VLS_69;
var __VLS_65;
const __VLS_74 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_75 = __VLS_asFunctionalComponent(__VLS_74, new __VLS_74({
    label: "标签",
}));
const __VLS_76 = __VLS_75({
    label: "标签",
}, ...__VLS_functionalComponentArgsRest(__VLS_75));
__VLS_77.slots.default;
const __VLS_78 = {}.ElSelect;
/** @type {[typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, ]} */ ;
// @ts-ignore
const __VLS_79 = __VLS_asFunctionalComponent(__VLS_78, new __VLS_78({
    modelValue: (__VLS_ctx.form.tags),
    multiple: true,
    allowCreate: true,
    filterable: true,
    placeholder: "输入标签按回车添加",
    ...{ style: {} },
}));
const __VLS_80 = __VLS_79({
    modelValue: (__VLS_ctx.form.tags),
    multiple: true,
    allowCreate: true,
    filterable: true,
    placeholder: "输入标签按回车添加",
    ...{ style: {} },
}, ...__VLS_functionalComponentArgsRest(__VLS_79));
var __VLS_77;
const __VLS_82 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_83 = __VLS_asFunctionalComponent(__VLS_82, new __VLS_82({
    label: "排序",
    prop: "sort",
}));
const __VLS_84 = __VLS_83({
    label: "排序",
    prop: "sort",
}, ...__VLS_functionalComponentArgsRest(__VLS_83));
__VLS_85.slots.default;
const __VLS_86 = {}.ElInputNumber;
/** @type {[typeof __VLS_components.ElInputNumber, typeof __VLS_components.elInputNumber, ]} */ ;
// @ts-ignore
const __VLS_87 = __VLS_asFunctionalComponent(__VLS_86, new __VLS_86({
    modelValue: (__VLS_ctx.form.sort),
    min: (0),
    controlsPosition: "right",
}));
const __VLS_88 = __VLS_87({
    modelValue: (__VLS_ctx.form.sort),
    min: (0),
    controlsPosition: "right",
}, ...__VLS_functionalComponentArgsRest(__VLS_87));
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
    ...{ style: {} },
});
var __VLS_85;
var __VLS_33;
const __VLS_90 = {}.ElCard;
/** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
// @ts-ignore
const __VLS_91 = __VLS_asFunctionalComponent(__VLS_90, new __VLS_90({
    ...{ style: {} },
}));
const __VLS_92 = __VLS_91({
    ...{ style: {} },
}, ...__VLS_functionalComponentArgsRest(__VLS_91));
__VLS_93.slots.default;
{
    const { header: __VLS_thisSlot } = __VLS_93.slots;
}
const __VLS_94 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_95 = __VLS_asFunctionalComponent(__VLS_94, new __VLS_94({
    label: "主图",
    prop: "main_image",
}));
const __VLS_96 = __VLS_95({
    label: "主图",
    prop: "main_image",
}, ...__VLS_functionalComponentArgsRest(__VLS_95));
__VLS_97.slots.default;
/** @type {[typeof UploadImage, ]} */ ;
// @ts-ignore
const __VLS_98 = __VLS_asFunctionalComponent(UploadImage, new UploadImage({
    modelValue: (__VLS_ctx.form.main_image),
}));
const __VLS_99 = __VLS_98({
    modelValue: (__VLS_ctx.form.main_image),
}, ...__VLS_functionalComponentArgsRest(__VLS_98));
var __VLS_97;
const __VLS_101 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_102 = __VLS_asFunctionalComponent(__VLS_101, new __VLS_101({
    label: "轮播图",
}));
const __VLS_103 = __VLS_102({
    label: "轮播图",
}, ...__VLS_functionalComponentArgsRest(__VLS_102));
__VLS_104.slots.default;
/** @type {[typeof UploadImage, ]} */ ;
// @ts-ignore
const __VLS_105 = __VLS_asFunctionalComponent(UploadImage, new UploadImage({
    modelValue: (__VLS_ctx.form.images),
    multiple: true,
    limit: (9),
}));
const __VLS_106 = __VLS_105({
    modelValue: (__VLS_ctx.form.images),
    multiple: true,
    limit: (9),
}, ...__VLS_functionalComponentArgsRest(__VLS_105));
var __VLS_104;
const __VLS_108 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_109 = __VLS_asFunctionalComponent(__VLS_108, new __VLS_108({
    label: "视频 URL",
}));
const __VLS_110 = __VLS_109({
    label: "视频 URL",
}, ...__VLS_functionalComponentArgsRest(__VLS_109));
__VLS_111.slots.default;
const __VLS_112 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_113 = __VLS_asFunctionalComponent(__VLS_112, new __VLS_112({
    modelValue: (__VLS_ctx.form.video_url),
    placeholder: "视频 URL，可选",
    ...{ style: {} },
}));
const __VLS_114 = __VLS_113({
    modelValue: (__VLS_ctx.form.video_url),
    placeholder: "视频 URL，可选",
    ...{ style: {} },
}, ...__VLS_functionalComponentArgsRest(__VLS_113));
var __VLS_111;
var __VLS_93;
const __VLS_116 = {}.ElCard;
/** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
// @ts-ignore
const __VLS_117 = __VLS_asFunctionalComponent(__VLS_116, new __VLS_116({
    ...{ style: {} },
}));
const __VLS_118 = __VLS_117({
    ...{ style: {} },
}, ...__VLS_functionalComponentArgsRest(__VLS_117));
__VLS_119.slots.default;
{
    const { header: __VLS_thisSlot } = __VLS_119.slots;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ style: {} },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
    const __VLS_120 = {}.ElTooltip;
    /** @type {[typeof __VLS_components.ElTooltip, typeof __VLS_components.elTooltip, typeof __VLS_components.ElTooltip, typeof __VLS_components.elTooltip, ]} */ ;
    // @ts-ignore
    const __VLS_121 = __VLS_asFunctionalComponent(__VLS_120, new __VLS_120({
        content: (__VLS_ctx.form.specs.length >= 3 ? '最多 3 层规格' : '添加规格'),
        placement: "top",
    }));
    const __VLS_122 = __VLS_121({
        content: (__VLS_ctx.form.specs.length >= 3 ? '最多 3 层规格' : '添加规格'),
        placement: "top",
    }, ...__VLS_functionalComponentArgsRest(__VLS_121));
    __VLS_123.slots.default;
    const __VLS_124 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_125 = __VLS_asFunctionalComponent(__VLS_124, new __VLS_124({
        ...{ 'onClick': {} },
        size: "small",
        disabled: (__VLS_ctx.form.specs.length >= 3),
    }));
    const __VLS_126 = __VLS_125({
        ...{ 'onClick': {} },
        size: "small",
        disabled: (__VLS_ctx.form.specs.length >= 3),
    }, ...__VLS_functionalComponentArgsRest(__VLS_125));
    let __VLS_128;
    let __VLS_129;
    let __VLS_130;
    const __VLS_131 = {
        onClick: (__VLS_ctx.addSpec)
    };
    __VLS_127.slots.default;
    var __VLS_127;
    var __VLS_123;
}
if (__VLS_ctx.form.specs.length === 0) {
    if (__VLS_ctx.singleSku) {
        const __VLS_132 = {}.ElTable;
        /** @type {[typeof __VLS_components.ElTable, typeof __VLS_components.elTable, typeof __VLS_components.ElTable, typeof __VLS_components.elTable, ]} */ ;
        // @ts-ignore
        const __VLS_133 = __VLS_asFunctionalComponent(__VLS_132, new __VLS_132({
            data: ([__VLS_ctx.singleSku]),
            border: true,
        }));
        const __VLS_134 = __VLS_133({
            data: ([__VLS_ctx.singleSku]),
            border: true,
        }, ...__VLS_functionalComponentArgsRest(__VLS_133));
        __VLS_135.slots.default;
        const __VLS_136 = {}.ElTableColumn;
        /** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
        // @ts-ignore
        const __VLS_137 = __VLS_asFunctionalComponent(__VLS_136, new __VLS_136({
            label: "售价(分)",
            width: "160",
        }));
        const __VLS_138 = __VLS_137({
            label: "售价(分)",
            width: "160",
        }, ...__VLS_functionalComponentArgsRest(__VLS_137));
        __VLS_139.slots.default;
        {
            const { default: __VLS_thisSlot } = __VLS_139.slots;
            const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
            /** @type {[typeof PriceInput, ]} */ ;
            // @ts-ignore
            const __VLS_140 = __VLS_asFunctionalComponent(PriceInput, new PriceInput({
                modelValue: (row.price_cents),
            }));
            const __VLS_141 = __VLS_140({
                modelValue: (row.price_cents),
            }, ...__VLS_functionalComponentArgsRest(__VLS_140));
        }
        var __VLS_139;
        const __VLS_143 = {}.ElTableColumn;
        /** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
        // @ts-ignore
        const __VLS_144 = __VLS_asFunctionalComponent(__VLS_143, new __VLS_143({
            label: "原价(分)",
            width: "160",
        }));
        const __VLS_145 = __VLS_144({
            label: "原价(分)",
            width: "160",
        }, ...__VLS_functionalComponentArgsRest(__VLS_144));
        __VLS_146.slots.default;
        {
            const { default: __VLS_thisSlot } = __VLS_146.slots;
            const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
            /** @type {[typeof PriceInput, ]} */ ;
            // @ts-ignore
            const __VLS_147 = __VLS_asFunctionalComponent(PriceInput, new PriceInput({
                modelValue: (row.original_price_cents),
            }));
            const __VLS_148 = __VLS_147({
                modelValue: (row.original_price_cents),
            }, ...__VLS_functionalComponentArgsRest(__VLS_147));
        }
        var __VLS_146;
        const __VLS_150 = {}.ElTableColumn;
        /** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
        // @ts-ignore
        const __VLS_151 = __VLS_asFunctionalComponent(__VLS_150, new __VLS_150({
            label: "库存",
            width: "120",
        }));
        const __VLS_152 = __VLS_151({
            label: "库存",
            width: "120",
        }, ...__VLS_functionalComponentArgsRest(__VLS_151));
        __VLS_153.slots.default;
        {
            const { default: __VLS_thisSlot } = __VLS_153.slots;
            const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
            const __VLS_154 = {}.ElInputNumber;
            /** @type {[typeof __VLS_components.ElInputNumber, typeof __VLS_components.elInputNumber, ]} */ ;
            // @ts-ignore
            const __VLS_155 = __VLS_asFunctionalComponent(__VLS_154, new __VLS_154({
                modelValue: (row.stock),
                min: (0),
                controlsPosition: "right",
                ...{ style: {} },
            }));
            const __VLS_156 = __VLS_155({
                modelValue: (row.stock),
                min: (0),
                controlsPosition: "right",
                ...{ style: {} },
            }, ...__VLS_functionalComponentArgsRest(__VLS_155));
        }
        var __VLS_153;
        const __VLS_158 = {}.ElTableColumn;
        /** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
        // @ts-ignore
        const __VLS_159 = __VLS_asFunctionalComponent(__VLS_158, new __VLS_158({
            label: "重量(g)",
            width: "120",
        }));
        const __VLS_160 = __VLS_159({
            label: "重量(g)",
            width: "120",
        }, ...__VLS_functionalComponentArgsRest(__VLS_159));
        __VLS_161.slots.default;
        {
            const { default: __VLS_thisSlot } = __VLS_161.slots;
            const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
            const __VLS_162 = {}.ElInputNumber;
            /** @type {[typeof __VLS_components.ElInputNumber, typeof __VLS_components.elInputNumber, ]} */ ;
            // @ts-ignore
            const __VLS_163 = __VLS_asFunctionalComponent(__VLS_162, new __VLS_162({
                modelValue: (row.weight_g),
                min: (0),
                controlsPosition: "right",
                ...{ style: {} },
            }));
            const __VLS_164 = __VLS_163({
                modelValue: (row.weight_g),
                min: (0),
                controlsPosition: "right",
                ...{ style: {} },
            }, ...__VLS_functionalComponentArgsRest(__VLS_163));
        }
        var __VLS_161;
        const __VLS_166 = {}.ElTableColumn;
        /** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
        // @ts-ignore
        const __VLS_167 = __VLS_asFunctionalComponent(__VLS_166, new __VLS_166({
            label: "SKU码",
            width: "160",
        }));
        const __VLS_168 = __VLS_167({
            label: "SKU码",
            width: "160",
        }, ...__VLS_functionalComponentArgsRest(__VLS_167));
        __VLS_169.slots.default;
        {
            const { default: __VLS_thisSlot } = __VLS_169.slots;
            const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
            const __VLS_170 = {}.ElInput;
            /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
            // @ts-ignore
            const __VLS_171 = __VLS_asFunctionalComponent(__VLS_170, new __VLS_170({
                modelValue: (row.sku_code),
                placeholder: "可选",
            }));
            const __VLS_172 = __VLS_171({
                modelValue: (row.sku_code),
                placeholder: "可选",
            }, ...__VLS_functionalComponentArgsRest(__VLS_171));
        }
        var __VLS_169;
        var __VLS_135;
    }
}
else {
    for (const [spec, si] of __VLS_getVForSourceType((__VLS_ctx.form.specs))) {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            key: (si),
            ...{ style: {} },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ style: {} },
        });
        const __VLS_174 = {}.ElInput;
        /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
        // @ts-ignore
        const __VLS_175 = __VLS_asFunctionalComponent(__VLS_174, new __VLS_174({
            modelValue: (spec.name),
            placeholder: "规格名（如：颜色）",
            ...{ style: {} },
        }));
        const __VLS_176 = __VLS_175({
            modelValue: (spec.name),
            placeholder: "规格名（如：颜色）",
            ...{ style: {} },
        }, ...__VLS_functionalComponentArgsRest(__VLS_175));
        const __VLS_178 = {}.ElButton;
        /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
        // @ts-ignore
        const __VLS_179 = __VLS_asFunctionalComponent(__VLS_178, new __VLS_178({
            ...{ 'onClick': {} },
            text: true,
            type: "danger",
            size: "small",
        }));
        const __VLS_180 = __VLS_179({
            ...{ 'onClick': {} },
            text: true,
            type: "danger",
            size: "small",
        }, ...__VLS_functionalComponentArgsRest(__VLS_179));
        let __VLS_182;
        let __VLS_183;
        let __VLS_184;
        const __VLS_185 = {
            onClick: (...[$event]) => {
                if (!!(__VLS_ctx.form.specs.length === 0))
                    return;
                __VLS_ctx.removeSpec(si);
            }
        };
        __VLS_181.slots.default;
        var __VLS_181;
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ style: {} },
        });
        for (const [val, vi] of __VLS_getVForSourceType((spec.values))) {
            const __VLS_186 = {}.ElTag;
            /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
            // @ts-ignore
            const __VLS_187 = __VLS_asFunctionalComponent(__VLS_186, new __VLS_186({
                ...{ 'onClose': {} },
                key: (vi),
                closable: true,
            }));
            const __VLS_188 = __VLS_187({
                ...{ 'onClose': {} },
                key: (vi),
                closable: true,
            }, ...__VLS_functionalComponentArgsRest(__VLS_187));
            let __VLS_190;
            let __VLS_191;
            let __VLS_192;
            const __VLS_193 = {
                onClose: (...[$event]) => {
                    if (!!(__VLS_ctx.form.specs.length === 0))
                        return;
                    __VLS_ctx.removeSpecValue(si, vi);
                }
            };
            __VLS_189.slots.default;
            (val);
            var __VLS_189;
        }
        const __VLS_194 = {}.ElInput;
        /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
        // @ts-ignore
        const __VLS_195 = __VLS_asFunctionalComponent(__VLS_194, new __VLS_194({
            ...{ 'onKeyup': {} },
            modelValue: (__VLS_ctx.specInputs[si]),
            placeholder: "输入值后回车",
            ...{ style: {} },
        }));
        const __VLS_196 = __VLS_195({
            ...{ 'onKeyup': {} },
            modelValue: (__VLS_ctx.specInputs[si]),
            placeholder: "输入值后回车",
            ...{ style: {} },
        }, ...__VLS_functionalComponentArgsRest(__VLS_195));
        let __VLS_198;
        let __VLS_199;
        let __VLS_200;
        const __VLS_201 = {
            onKeyup: (...[$event]) => {
                if (!!(__VLS_ctx.form.specs.length === 0))
                    return;
                __VLS_ctx.addSpecValue(si, __VLS_ctx.specInputs[si]);
            }
        };
        var __VLS_197;
    }
    if (__VLS_ctx.form.skus.length) {
        const __VLS_202 = {}.ElTable;
        /** @type {[typeof __VLS_components.ElTable, typeof __VLS_components.elTable, typeof __VLS_components.ElTable, typeof __VLS_components.elTable, ]} */ ;
        // @ts-ignore
        const __VLS_203 = __VLS_asFunctionalComponent(__VLS_202, new __VLS_202({
            data: (__VLS_ctx.form.skus),
            border: true,
            ...{ style: {} },
        }));
        const __VLS_204 = __VLS_203({
            data: (__VLS_ctx.form.skus),
            border: true,
            ...{ style: {} },
        }, ...__VLS_functionalComponentArgsRest(__VLS_203));
        __VLS_205.slots.default;
        const __VLS_206 = {}.ElTableColumn;
        /** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
        // @ts-ignore
        const __VLS_207 = __VLS_asFunctionalComponent(__VLS_206, new __VLS_206({
            label: "规格图",
            width: "100",
        }));
        const __VLS_208 = __VLS_207({
            label: "规格图",
            width: "100",
        }, ...__VLS_functionalComponentArgsRest(__VLS_207));
        __VLS_209.slots.default;
        {
            const { default: __VLS_thisSlot } = __VLS_209.slots;
            const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
            /** @type {[typeof UploadImage, ]} */ ;
            // @ts-ignore
            const __VLS_210 = __VLS_asFunctionalComponent(UploadImage, new UploadImage({
                modelValue: (row.image),
            }));
            const __VLS_211 = __VLS_210({
                modelValue: (row.image),
            }, ...__VLS_functionalComponentArgsRest(__VLS_210));
        }
        var __VLS_209;
        const __VLS_213 = {}.ElTableColumn;
        /** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
        // @ts-ignore
        const __VLS_214 = __VLS_asFunctionalComponent(__VLS_213, new __VLS_213({
            label: "规格组合",
            minWidth: "120",
        }));
        const __VLS_215 = __VLS_214({
            label: "规格组合",
            minWidth: "120",
        }, ...__VLS_functionalComponentArgsRest(__VLS_214));
        __VLS_216.slots.default;
        {
            const { default: __VLS_thisSlot } = __VLS_216.slots;
            const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
            for (const [v, i] of __VLS_getVForSourceType((row.spec_values))) {
                const __VLS_217 = {}.ElTag;
                /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
                // @ts-ignore
                const __VLS_218 = __VLS_asFunctionalComponent(__VLS_217, new __VLS_217({
                    key: (i),
                    size: "small",
                    ...{ style: {} },
                }));
                const __VLS_219 = __VLS_218({
                    key: (i),
                    size: "small",
                    ...{ style: {} },
                }, ...__VLS_functionalComponentArgsRest(__VLS_218));
                __VLS_220.slots.default;
                (v);
                var __VLS_220;
            }
        }
        var __VLS_216;
        const __VLS_221 = {}.ElTableColumn;
        /** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
        // @ts-ignore
        const __VLS_222 = __VLS_asFunctionalComponent(__VLS_221, new __VLS_221({
            label: "售价(分)",
            width: "150",
        }));
        const __VLS_223 = __VLS_222({
            label: "售价(分)",
            width: "150",
        }, ...__VLS_functionalComponentArgsRest(__VLS_222));
        __VLS_224.slots.default;
        {
            const { default: __VLS_thisSlot } = __VLS_224.slots;
            const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
            /** @type {[typeof PriceInput, ]} */ ;
            // @ts-ignore
            const __VLS_225 = __VLS_asFunctionalComponent(PriceInput, new PriceInput({
                modelValue: (row.price_cents),
            }));
            const __VLS_226 = __VLS_225({
                modelValue: (row.price_cents),
            }, ...__VLS_functionalComponentArgsRest(__VLS_225));
        }
        var __VLS_224;
        const __VLS_228 = {}.ElTableColumn;
        /** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
        // @ts-ignore
        const __VLS_229 = __VLS_asFunctionalComponent(__VLS_228, new __VLS_228({
            label: "原价(分)",
            width: "150",
        }));
        const __VLS_230 = __VLS_229({
            label: "原价(分)",
            width: "150",
        }, ...__VLS_functionalComponentArgsRest(__VLS_229));
        __VLS_231.slots.default;
        {
            const { default: __VLS_thisSlot } = __VLS_231.slots;
            const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
            /** @type {[typeof PriceInput, ]} */ ;
            // @ts-ignore
            const __VLS_232 = __VLS_asFunctionalComponent(PriceInput, new PriceInput({
                modelValue: (row.original_price_cents),
            }));
            const __VLS_233 = __VLS_232({
                modelValue: (row.original_price_cents),
            }, ...__VLS_functionalComponentArgsRest(__VLS_232));
        }
        var __VLS_231;
        const __VLS_235 = {}.ElTableColumn;
        /** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
        // @ts-ignore
        const __VLS_236 = __VLS_asFunctionalComponent(__VLS_235, new __VLS_235({
            label: "库存",
            width: "100",
        }));
        const __VLS_237 = __VLS_236({
            label: "库存",
            width: "100",
        }, ...__VLS_functionalComponentArgsRest(__VLS_236));
        __VLS_238.slots.default;
        {
            const { default: __VLS_thisSlot } = __VLS_238.slots;
            const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
            const __VLS_239 = {}.ElInputNumber;
            /** @type {[typeof __VLS_components.ElInputNumber, typeof __VLS_components.elInputNumber, ]} */ ;
            // @ts-ignore
            const __VLS_240 = __VLS_asFunctionalComponent(__VLS_239, new __VLS_239({
                modelValue: (row.stock),
                min: (0),
                controlsPosition: "right",
                ...{ style: {} },
            }));
            const __VLS_241 = __VLS_240({
                modelValue: (row.stock),
                min: (0),
                controlsPosition: "right",
                ...{ style: {} },
            }, ...__VLS_functionalComponentArgsRest(__VLS_240));
        }
        var __VLS_238;
        const __VLS_243 = {}.ElTableColumn;
        /** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
        // @ts-ignore
        const __VLS_244 = __VLS_asFunctionalComponent(__VLS_243, new __VLS_243({
            label: "重量(g)",
            width: "100",
        }));
        const __VLS_245 = __VLS_244({
            label: "重量(g)",
            width: "100",
        }, ...__VLS_functionalComponentArgsRest(__VLS_244));
        __VLS_246.slots.default;
        {
            const { default: __VLS_thisSlot } = __VLS_246.slots;
            const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
            const __VLS_247 = {}.ElInputNumber;
            /** @type {[typeof __VLS_components.ElInputNumber, typeof __VLS_components.elInputNumber, ]} */ ;
            // @ts-ignore
            const __VLS_248 = __VLS_asFunctionalComponent(__VLS_247, new __VLS_247({
                modelValue: (row.weight_g),
                min: (0),
                controlsPosition: "right",
                ...{ style: {} },
            }));
            const __VLS_249 = __VLS_248({
                modelValue: (row.weight_g),
                min: (0),
                controlsPosition: "right",
                ...{ style: {} },
            }, ...__VLS_functionalComponentArgsRest(__VLS_248));
        }
        var __VLS_246;
        const __VLS_251 = {}.ElTableColumn;
        /** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
        // @ts-ignore
        const __VLS_252 = __VLS_asFunctionalComponent(__VLS_251, new __VLS_251({
            label: "SKU码",
            width: "140",
        }));
        const __VLS_253 = __VLS_252({
            label: "SKU码",
            width: "140",
        }, ...__VLS_functionalComponentArgsRest(__VLS_252));
        __VLS_254.slots.default;
        {
            const { default: __VLS_thisSlot } = __VLS_254.slots;
            const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
            const __VLS_255 = {}.ElInput;
            /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
            // @ts-ignore
            const __VLS_256 = __VLS_asFunctionalComponent(__VLS_255, new __VLS_255({
                modelValue: (row.sku_code),
                placeholder: "可选",
            }));
            const __VLS_257 = __VLS_256({
                modelValue: (row.sku_code),
                placeholder: "可选",
            }, ...__VLS_functionalComponentArgsRest(__VLS_256));
        }
        var __VLS_254;
        const __VLS_259 = {}.ElTableColumn;
        /** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
        // @ts-ignore
        const __VLS_260 = __VLS_asFunctionalComponent(__VLS_259, new __VLS_259({
            label: "条形码",
            width: "140",
        }));
        const __VLS_261 = __VLS_260({
            label: "条形码",
            width: "140",
        }, ...__VLS_functionalComponentArgsRest(__VLS_260));
        __VLS_262.slots.default;
        {
            const { default: __VLS_thisSlot } = __VLS_262.slots;
            const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
            const __VLS_263 = {}.ElInput;
            /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
            // @ts-ignore
            const __VLS_264 = __VLS_asFunctionalComponent(__VLS_263, new __VLS_263({
                modelValue: (row.barcode),
                placeholder: "可选",
            }));
            const __VLS_265 = __VLS_264({
                modelValue: (row.barcode),
                placeholder: "可选",
            }, ...__VLS_functionalComponentArgsRest(__VLS_264));
        }
        var __VLS_262;
        const __VLS_267 = {}.ElTableColumn;
        /** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
        // @ts-ignore
        const __VLS_268 = __VLS_asFunctionalComponent(__VLS_267, new __VLS_267({
            label: "状态",
            width: "90",
            align: "center",
        }));
        const __VLS_269 = __VLS_268({
            label: "状态",
            width: "90",
            align: "center",
        }, ...__VLS_functionalComponentArgsRest(__VLS_268));
        __VLS_270.slots.default;
        {
            const { default: __VLS_thisSlot } = __VLS_270.slots;
            const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
            const __VLS_271 = {}.ElSwitch;
            /** @type {[typeof __VLS_components.ElSwitch, typeof __VLS_components.elSwitch, ]} */ ;
            // @ts-ignore
            const __VLS_272 = __VLS_asFunctionalComponent(__VLS_271, new __VLS_271({
                modelValue: (row.status),
                activeValue: "active",
                inactiveValue: "disabled",
            }));
            const __VLS_273 = __VLS_272({
                modelValue: (row.status),
                activeValue: "active",
                inactiveValue: "disabled",
            }, ...__VLS_functionalComponentArgsRest(__VLS_272));
        }
        var __VLS_270;
        var __VLS_205;
    }
    else {
        const __VLS_275 = {}.ElEmpty;
        /** @type {[typeof __VLS_components.ElEmpty, typeof __VLS_components.elEmpty, ]} */ ;
        // @ts-ignore
        const __VLS_276 = __VLS_asFunctionalComponent(__VLS_275, new __VLS_275({
            description: "请先添加规格值以生成 SKU",
            imageSize: (60),
            ...{ style: {} },
        }));
        const __VLS_277 = __VLS_276({
            description: "请先添加规格值以生成 SKU",
            imageSize: (60),
            ...{ style: {} },
        }, ...__VLS_functionalComponentArgsRest(__VLS_276));
    }
}
var __VLS_119;
const __VLS_279 = {}.ElCard;
/** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
// @ts-ignore
const __VLS_280 = __VLS_asFunctionalComponent(__VLS_279, new __VLS_279({
    ...{ style: {} },
}));
const __VLS_281 = __VLS_280({
    ...{ style: {} },
}, ...__VLS_functionalComponentArgsRest(__VLS_280));
__VLS_282.slots.default;
{
    const { header: __VLS_thisSlot } = __VLS_282.slots;
}
const __VLS_283 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_284 = __VLS_asFunctionalComponent(__VLS_283, new __VLS_283({
    label: "虚拟商品",
}));
const __VLS_285 = __VLS_284({
    label: "虚拟商品",
}, ...__VLS_functionalComponentArgsRest(__VLS_284));
__VLS_286.slots.default;
const __VLS_287 = {}.ElSwitch;
/** @type {[typeof __VLS_components.ElSwitch, typeof __VLS_components.elSwitch, ]} */ ;
// @ts-ignore
const __VLS_288 = __VLS_asFunctionalComponent(__VLS_287, new __VLS_287({
    modelValue: (__VLS_ctx.form.is_virtual),
    activeText: "虚拟商品（数字/卡券，无需发货）",
}));
const __VLS_289 = __VLS_288({
    modelValue: (__VLS_ctx.form.is_virtual),
    activeText: "虚拟商品（数字/卡券，无需发货）",
}, ...__VLS_functionalComponentArgsRest(__VLS_288));
var __VLS_286;
if (!__VLS_ctx.form.is_virtual) {
    const __VLS_291 = {}.ElFormItem;
    /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
    // @ts-ignore
    const __VLS_292 = __VLS_asFunctionalComponent(__VLS_291, new __VLS_291({
        label: "运费模板",
    }));
    const __VLS_293 = __VLS_292({
        label: "运费模板",
    }, ...__VLS_functionalComponentArgsRest(__VLS_292));
    __VLS_294.slots.default;
    const __VLS_295 = {}.ElSelect;
    /** @type {[typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, ]} */ ;
    // @ts-ignore
    const __VLS_296 = __VLS_asFunctionalComponent(__VLS_295, new __VLS_295({
        modelValue: (__VLS_ctx.form.freight_template_id),
        placeholder: "请选择运费模板",
        ...{ style: {} },
    }));
    const __VLS_297 = __VLS_296({
        modelValue: (__VLS_ctx.form.freight_template_id),
        placeholder: "请选择运费模板",
        ...{ style: {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_296));
    __VLS_298.slots.default;
    const __VLS_299 = {}.ElOption;
    /** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
    // @ts-ignore
    const __VLS_300 = __VLS_asFunctionalComponent(__VLS_299, new __VLS_299({
        label: "包邮（不设运费模板）",
        value: (null),
    }));
    const __VLS_301 = __VLS_300({
        label: "包邮（不设运费模板）",
        value: (null),
    }, ...__VLS_functionalComponentArgsRest(__VLS_300));
    for (const [tpl] of __VLS_getVForSourceType((__VLS_ctx.freightTemplates))) {
        const __VLS_303 = {}.ElOption;
        /** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
        // @ts-ignore
        const __VLS_304 = __VLS_asFunctionalComponent(__VLS_303, new __VLS_303({
            key: (tpl.id),
            label: (tpl.name),
            value: (tpl.id),
        }));
        const __VLS_305 = __VLS_304({
            key: (tpl.id),
            label: (tpl.name),
            value: (tpl.id),
        }, ...__VLS_functionalComponentArgsRest(__VLS_304));
    }
    var __VLS_298;
    var __VLS_294;
}
var __VLS_282;
const __VLS_307 = {}.ElCard;
/** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
// @ts-ignore
const __VLS_308 = __VLS_asFunctionalComponent(__VLS_307, new __VLS_307({
    ...{ style: {} },
}));
const __VLS_309 = __VLS_308({
    ...{ style: {} },
}, ...__VLS_functionalComponentArgsRest(__VLS_308));
__VLS_310.slots.default;
{
    const { header: __VLS_thisSlot } = __VLS_310.slots;
}
const __VLS_311 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_312 = __VLS_asFunctionalComponent(__VLS_311, new __VLS_311({
    labelWidth: "0",
}));
const __VLS_313 = __VLS_312({
    labelWidth: "0",
}, ...__VLS_functionalComponentArgsRest(__VLS_312));
__VLS_314.slots.default;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ style: {} },
});
const __VLS_315 = {}.Toolbar;
/** @type {[typeof __VLS_components.Toolbar, ]} */ ;
// @ts-ignore
const __VLS_316 = __VLS_asFunctionalComponent(__VLS_315, new __VLS_315({
    editor: (__VLS_ctx.editorRef),
    defaultConfig: ({}),
    mode: "default",
    ...{ style: {} },
}));
const __VLS_317 = __VLS_316({
    editor: (__VLS_ctx.editorRef),
    defaultConfig: ({}),
    mode: "default",
    ...{ style: {} },
}, ...__VLS_functionalComponentArgsRest(__VLS_316));
const __VLS_319 = {}.Editor;
/** @type {[typeof __VLS_components.Editor, ]} */ ;
// @ts-ignore
const __VLS_320 = __VLS_asFunctionalComponent(__VLS_319, new __VLS_319({
    ...{ 'onOnCreated': {} },
    modelValue: (__VLS_ctx.form.detail_html),
    defaultConfig: (__VLS_ctx.editorConfig),
    mode: "default",
    ...{ style: {} },
}));
const __VLS_321 = __VLS_320({
    ...{ 'onOnCreated': {} },
    modelValue: (__VLS_ctx.form.detail_html),
    defaultConfig: (__VLS_ctx.editorConfig),
    mode: "default",
    ...{ style: {} },
}, ...__VLS_functionalComponentArgsRest(__VLS_320));
let __VLS_323;
let __VLS_324;
let __VLS_325;
const __VLS_326 = {
    onOnCreated: ((e) => (__VLS_ctx.editorRef = e))
};
var __VLS_322;
var __VLS_314;
var __VLS_310;
const __VLS_327 = {}.ElCard;
/** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
// @ts-ignore
const __VLS_328 = __VLS_asFunctionalComponent(__VLS_327, new __VLS_327({
    ...{ style: {} },
}));
const __VLS_329 = __VLS_328({
    ...{ style: {} },
}, ...__VLS_functionalComponentArgsRest(__VLS_328));
__VLS_330.slots.default;
{
    const { header: __VLS_thisSlot } = __VLS_330.slots;
}
const __VLS_331 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_332 = __VLS_asFunctionalComponent(__VLS_331, new __VLS_331({
    label: "虚拟销量",
}));
const __VLS_333 = __VLS_332({
    label: "虚拟销量",
}, ...__VLS_functionalComponentArgsRest(__VLS_332));
__VLS_334.slots.default;
const __VLS_335 = {}.ElInputNumber;
/** @type {[typeof __VLS_components.ElInputNumber, typeof __VLS_components.elInputNumber, ]} */ ;
// @ts-ignore
const __VLS_336 = __VLS_asFunctionalComponent(__VLS_335, new __VLS_335({
    modelValue: (__VLS_ctx.form.virtual_sales),
    min: (0),
    controlsPosition: "right",
}));
const __VLS_337 = __VLS_336({
    modelValue: (__VLS_ctx.form.virtual_sales),
    min: (0),
    controlsPosition: "right",
}, ...__VLS_functionalComponentArgsRest(__VLS_336));
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
    ...{ style: {} },
});
var __VLS_334;
if (__VLS_ctx.form.status === 'draft') {
    const __VLS_339 = {}.ElFormItem;
    /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
    // @ts-ignore
    const __VLS_340 = __VLS_asFunctionalComponent(__VLS_339, new __VLS_339({
        label: "定时上架",
    }));
    const __VLS_341 = __VLS_340({
        label: "定时上架",
    }, ...__VLS_functionalComponentArgsRest(__VLS_340));
    __VLS_342.slots.default;
    const __VLS_343 = {}.ElDatePicker;
    /** @type {[typeof __VLS_components.ElDatePicker, typeof __VLS_components.elDatePicker, ]} */ ;
    // @ts-ignore
    const __VLS_344 = __VLS_asFunctionalComponent(__VLS_343, new __VLS_343({
        modelValue: (__VLS_ctx.form.on_sale_at),
        type: "datetime",
        placeholder: "选择定时上架时间（可选）",
        valueFormat: "YYYY-MM-DDTHH:mm:ssZ",
        ...{ style: {} },
    }));
    const __VLS_345 = __VLS_344({
        modelValue: (__VLS_ctx.form.on_sale_at),
        type: "datetime",
        placeholder: "选择定时上架时间（可选）",
        valueFormat: "YYYY-MM-DDTHH:mm:ssZ",
        ...{ style: {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_344));
    var __VLS_342;
}
const __VLS_347 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_348 = __VLS_asFunctionalComponent(__VLS_347, new __VLS_347({
    label: "状态",
}));
const __VLS_349 = __VLS_348({
    label: "状态",
}, ...__VLS_functionalComponentArgsRest(__VLS_348));
__VLS_350.slots.default;
const __VLS_351 = {}.ElRadioGroup;
/** @type {[typeof __VLS_components.ElRadioGroup, typeof __VLS_components.elRadioGroup, typeof __VLS_components.ElRadioGroup, typeof __VLS_components.elRadioGroup, ]} */ ;
// @ts-ignore
const __VLS_352 = __VLS_asFunctionalComponent(__VLS_351, new __VLS_351({
    modelValue: (__VLS_ctx.form.status),
}));
const __VLS_353 = __VLS_352({
    modelValue: (__VLS_ctx.form.status),
}, ...__VLS_functionalComponentArgsRest(__VLS_352));
__VLS_354.slots.default;
const __VLS_355 = {}.ElRadio;
/** @type {[typeof __VLS_components.ElRadio, typeof __VLS_components.elRadio, typeof __VLS_components.ElRadio, typeof __VLS_components.elRadio, ]} */ ;
// @ts-ignore
const __VLS_356 = __VLS_asFunctionalComponent(__VLS_355, new __VLS_355({
    value: "draft",
}));
const __VLS_357 = __VLS_356({
    value: "draft",
}, ...__VLS_functionalComponentArgsRest(__VLS_356));
__VLS_358.slots.default;
var __VLS_358;
const __VLS_359 = {}.ElRadio;
/** @type {[typeof __VLS_components.ElRadio, typeof __VLS_components.elRadio, typeof __VLS_components.ElRadio, typeof __VLS_components.elRadio, ]} */ ;
// @ts-ignore
const __VLS_360 = __VLS_asFunctionalComponent(__VLS_359, new __VLS_359({
    value: "onsale",
}));
const __VLS_361 = __VLS_360({
    value: "onsale",
}, ...__VLS_functionalComponentArgsRest(__VLS_360));
__VLS_362.slots.default;
var __VLS_362;
const __VLS_363 = {}.ElRadio;
/** @type {[typeof __VLS_components.ElRadio, typeof __VLS_components.elRadio, typeof __VLS_components.ElRadio, typeof __VLS_components.elRadio, ]} */ ;
// @ts-ignore
const __VLS_364 = __VLS_asFunctionalComponent(__VLS_363, new __VLS_363({
    value: "offsale",
}));
const __VLS_365 = __VLS_364({
    value: "offsale",
}, ...__VLS_functionalComponentArgsRest(__VLS_364));
__VLS_366.slots.default;
var __VLS_366;
var __VLS_354;
var __VLS_350;
var __VLS_330;
var __VLS_27;
/** @type {__VLS_StyleScopedClasses['page-card']} */ ;
// @ts-ignore
var __VLS_29 = __VLS_28;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            Editor: Editor,
            Toolbar: Toolbar,
            UploadImage: UploadImage,
            PriceInput: PriceInput,
            router: router,
            isEdit: isEdit,
            formRef: formRef,
            saving: saving,
            categories: categories,
            freightTemplates: freightTemplates,
            editorRef: editorRef,
            editorConfig: editorConfig,
            UNIT_OPTIONS: UNIT_OPTIONS,
            form: form,
            addSpec: addSpec,
            removeSpec: removeSpec,
            specInputs: specInputs,
            addSpecValue: addSpecValue,
            removeSpecValue: removeSpecValue,
            singleSku: singleSku,
            formRules: formRules,
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
