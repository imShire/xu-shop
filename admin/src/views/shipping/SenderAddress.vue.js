import { ref, onMounted } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import { getSenderAddresses, createSenderAddress, updateSenderAddress, deleteSenderAddress, setDefaultSenderAddress, } from '@/api/shipping';
const loading = ref(false);
const addresses = ref([]);
const dialogVisible = ref(false);
const saving = ref(false);
const formRef = ref();
const isEdit = ref(false);
const form = ref({
    id: '',
    name: '',
    phone: '',
    province: '',
    city: '',
    district: '',
    address: '',
});
async function loadData() {
    loading.value = true;
    try {
        addresses.value = await getSenderAddresses();
    }
    finally {
        loading.value = false;
    }
}
function openCreate() {
    form.value = { id: '', name: '', phone: '', province: '', city: '', district: '', address: '' };
    isEdit.value = false;
    dialogVisible.value = true;
}
function openEdit(row) {
    Object.assign(form.value, row);
    isEdit.value = true;
    dialogVisible.value = true;
}
async function handleSave() {
    await formRef.value?.validate();
    saving.value = true;
    try {
        if (isEdit.value) {
            await updateSenderAddress(form.value.id, form.value);
        }
        else {
            await createSenderAddress(form.value);
        }
        ElMessage.success('保存成功');
        dialogVisible.value = false;
        await loadData();
    }
    catch (e) {
        ElMessage.error(e?.message || '操作失败');
    }
    finally {
        saving.value = false;
    }
}
async function handleDelete(row) {
    try {
        await ElMessageBox.confirm('确认删除该发件地址？', '提示', { type: 'warning' });
        await deleteSenderAddress(row.id);
        ElMessage.success('删除成功');
        await loadData();
    }
    catch (e) {
        if (e !== 'cancel')
            ElMessage.error(e?.message || '操作失败');
    }
}
async function handleSetDefault(row) {
    try {
        await setDefaultSenderAddress(row.id);
        ElMessage.success('已设为默认');
        await loadData();
    }
    catch (e) {
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
    data: (__VLS_ctx.addresses),
    border: true,
}));
const __VLS_10 = __VLS_9({
    data: (__VLS_ctx.addresses),
    border: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_9));
__VLS_asFunctionalDirective(__VLS_directives.vLoading)(null, { ...__VLS_directiveBindingRestFields, value: (__VLS_ctx.loading) }, null, null);
__VLS_11.slots.default;
const __VLS_12 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_13 = __VLS_asFunctionalComponent(__VLS_12, new __VLS_12({
    prop: "name",
    label: "联系人",
    width: "100",
}));
const __VLS_14 = __VLS_13({
    prop: "name",
    label: "联系人",
    width: "100",
}, ...__VLS_functionalComponentArgsRest(__VLS_13));
const __VLS_16 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_17 = __VLS_asFunctionalComponent(__VLS_16, new __VLS_16({
    prop: "phone",
    label: "手机号",
    width: "130",
}));
const __VLS_18 = __VLS_17({
    prop: "phone",
    label: "手机号",
    width: "130",
}, ...__VLS_functionalComponentArgsRest(__VLS_17));
const __VLS_20 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_21 = __VLS_asFunctionalComponent(__VLS_20, new __VLS_20({
    label: "地址",
    minWidth: "200",
}));
const __VLS_22 = __VLS_21({
    label: "地址",
    minWidth: "200",
}, ...__VLS_functionalComponentArgsRest(__VLS_21));
__VLS_23.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_23.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    (row.province);
    (row.city);
    (row.district);
    (row.address);
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
    width: "180",
    fixed: "right",
}));
const __VLS_34 = __VLS_33({
    label: "操作",
    width: "180",
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
    if (!row.is_default) {
        const __VLS_44 = {}.ElButton;
        /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
        // @ts-ignore
        const __VLS_45 = __VLS_asFunctionalComponent(__VLS_44, new __VLS_44({
            ...{ 'onClick': {} },
            text: true,
            size: "small",
        }));
        const __VLS_46 = __VLS_45({
            ...{ 'onClick': {} },
            text: true,
            size: "small",
        }, ...__VLS_functionalComponentArgsRest(__VLS_45));
        let __VLS_48;
        let __VLS_49;
        let __VLS_50;
        const __VLS_51 = {
            onClick: (...[$event]) => {
                if (!(!row.is_default))
                    return;
                __VLS_ctx.handleSetDefault(row);
            }
        };
        __VLS_47.slots.default;
        var __VLS_47;
    }
    const __VLS_52 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_53 = __VLS_asFunctionalComponent(__VLS_52, new __VLS_52({
        ...{ 'onClick': {} },
        text: true,
        type: "danger",
        size: "small",
    }));
    const __VLS_54 = __VLS_53({
        ...{ 'onClick': {} },
        text: true,
        type: "danger",
        size: "small",
    }, ...__VLS_functionalComponentArgsRest(__VLS_53));
    let __VLS_56;
    let __VLS_57;
    let __VLS_58;
    const __VLS_59 = {
        onClick: (...[$event]) => {
            __VLS_ctx.handleDelete(row);
        }
    };
    __VLS_55.slots.default;
    var __VLS_55;
}
var __VLS_35;
var __VLS_11;
const __VLS_60 = {}.ElDialog;
/** @type {[typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, ]} */ ;
// @ts-ignore
const __VLS_61 = __VLS_asFunctionalComponent(__VLS_60, new __VLS_60({
    modelValue: (__VLS_ctx.dialogVisible),
    title: (__VLS_ctx.isEdit ? '编辑地址' : '新增地址'),
    width: "480px",
}));
const __VLS_62 = __VLS_61({
    modelValue: (__VLS_ctx.dialogVisible),
    title: (__VLS_ctx.isEdit ? '编辑地址' : '新增地址'),
    width: "480px",
}, ...__VLS_functionalComponentArgsRest(__VLS_61));
__VLS_63.slots.default;
const __VLS_64 = {}.ElForm;
/** @type {[typeof __VLS_components.ElForm, typeof __VLS_components.elForm, typeof __VLS_components.ElForm, typeof __VLS_components.elForm, ]} */ ;
// @ts-ignore
const __VLS_65 = __VLS_asFunctionalComponent(__VLS_64, new __VLS_64({
    ref: "formRef",
    model: (__VLS_ctx.form),
    labelWidth: "80px",
}));
const __VLS_66 = __VLS_65({
    ref: "formRef",
    model: (__VLS_ctx.form),
    labelWidth: "80px",
}, ...__VLS_functionalComponentArgsRest(__VLS_65));
/** @type {typeof __VLS_ctx.formRef} */ ;
var __VLS_68 = {};
__VLS_67.slots.default;
const __VLS_70 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_71 = __VLS_asFunctionalComponent(__VLS_70, new __VLS_70({
    label: "联系人",
    prop: "name",
    rules: ([{ required: true }]),
}));
const __VLS_72 = __VLS_71({
    label: "联系人",
    prop: "name",
    rules: ([{ required: true }]),
}, ...__VLS_functionalComponentArgsRest(__VLS_71));
__VLS_73.slots.default;
const __VLS_74 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_75 = __VLS_asFunctionalComponent(__VLS_74, new __VLS_74({
    modelValue: (__VLS_ctx.form.name),
}));
const __VLS_76 = __VLS_75({
    modelValue: (__VLS_ctx.form.name),
}, ...__VLS_functionalComponentArgsRest(__VLS_75));
var __VLS_73;
const __VLS_78 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_79 = __VLS_asFunctionalComponent(__VLS_78, new __VLS_78({
    label: "手机号",
    prop: "phone",
    rules: ([{ required: true }]),
}));
const __VLS_80 = __VLS_79({
    label: "手机号",
    prop: "phone",
    rules: ([{ required: true }]),
}, ...__VLS_functionalComponentArgsRest(__VLS_79));
__VLS_81.slots.default;
const __VLS_82 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_83 = __VLS_asFunctionalComponent(__VLS_82, new __VLS_82({
    modelValue: (__VLS_ctx.form.phone),
}));
const __VLS_84 = __VLS_83({
    modelValue: (__VLS_ctx.form.phone),
}, ...__VLS_functionalComponentArgsRest(__VLS_83));
var __VLS_81;
const __VLS_86 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_87 = __VLS_asFunctionalComponent(__VLS_86, new __VLS_86({
    label: "省份",
    prop: "province",
}));
const __VLS_88 = __VLS_87({
    label: "省份",
    prop: "province",
}, ...__VLS_functionalComponentArgsRest(__VLS_87));
__VLS_89.slots.default;
const __VLS_90 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_91 = __VLS_asFunctionalComponent(__VLS_90, new __VLS_90({
    modelValue: (__VLS_ctx.form.province),
}));
const __VLS_92 = __VLS_91({
    modelValue: (__VLS_ctx.form.province),
}, ...__VLS_functionalComponentArgsRest(__VLS_91));
var __VLS_89;
const __VLS_94 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_95 = __VLS_asFunctionalComponent(__VLS_94, new __VLS_94({
    label: "城市",
    prop: "city",
}));
const __VLS_96 = __VLS_95({
    label: "城市",
    prop: "city",
}, ...__VLS_functionalComponentArgsRest(__VLS_95));
__VLS_97.slots.default;
const __VLS_98 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_99 = __VLS_asFunctionalComponent(__VLS_98, new __VLS_98({
    modelValue: (__VLS_ctx.form.city),
}));
const __VLS_100 = __VLS_99({
    modelValue: (__VLS_ctx.form.city),
}, ...__VLS_functionalComponentArgsRest(__VLS_99));
var __VLS_97;
const __VLS_102 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_103 = __VLS_asFunctionalComponent(__VLS_102, new __VLS_102({
    label: "区/县",
    prop: "district",
}));
const __VLS_104 = __VLS_103({
    label: "区/县",
    prop: "district",
}, ...__VLS_functionalComponentArgsRest(__VLS_103));
__VLS_105.slots.default;
const __VLS_106 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_107 = __VLS_asFunctionalComponent(__VLS_106, new __VLS_106({
    modelValue: (__VLS_ctx.form.district),
}));
const __VLS_108 = __VLS_107({
    modelValue: (__VLS_ctx.form.district),
}, ...__VLS_functionalComponentArgsRest(__VLS_107));
var __VLS_105;
const __VLS_110 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_111 = __VLS_asFunctionalComponent(__VLS_110, new __VLS_110({
    label: "详细地址",
    prop: "address",
    rules: ([{ required: true }]),
}));
const __VLS_112 = __VLS_111({
    label: "详细地址",
    prop: "address",
    rules: ([{ required: true }]),
}, ...__VLS_functionalComponentArgsRest(__VLS_111));
__VLS_113.slots.default;
const __VLS_114 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_115 = __VLS_asFunctionalComponent(__VLS_114, new __VLS_114({
    modelValue: (__VLS_ctx.form.address),
    type: "textarea",
}));
const __VLS_116 = __VLS_115({
    modelValue: (__VLS_ctx.form.address),
    type: "textarea",
}, ...__VLS_functionalComponentArgsRest(__VLS_115));
var __VLS_113;
var __VLS_67;
{
    const { footer: __VLS_thisSlot } = __VLS_63.slots;
    const __VLS_118 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_119 = __VLS_asFunctionalComponent(__VLS_118, new __VLS_118({
        ...{ 'onClick': {} },
    }));
    const __VLS_120 = __VLS_119({
        ...{ 'onClick': {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_119));
    let __VLS_122;
    let __VLS_123;
    let __VLS_124;
    const __VLS_125 = {
        onClick: (...[$event]) => {
            __VLS_ctx.dialogVisible = false;
        }
    };
    __VLS_121.slots.default;
    var __VLS_121;
    const __VLS_126 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_127 = __VLS_asFunctionalComponent(__VLS_126, new __VLS_126({
        ...{ 'onClick': {} },
        type: "primary",
        loading: (__VLS_ctx.saving),
    }));
    const __VLS_128 = __VLS_127({
        ...{ 'onClick': {} },
        type: "primary",
        loading: (__VLS_ctx.saving),
    }, ...__VLS_functionalComponentArgsRest(__VLS_127));
    let __VLS_130;
    let __VLS_131;
    let __VLS_132;
    const __VLS_133 = {
        onClick: (__VLS_ctx.handleSave)
    };
    __VLS_129.slots.default;
    var __VLS_129;
}
var __VLS_63;
/** @type {__VLS_StyleScopedClasses['page-card']} */ ;
// @ts-ignore
var __VLS_69 = __VLS_68;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            loading: loading,
            addresses: addresses,
            dialogVisible: dialogVisible,
            saving: saving,
            formRef: formRef,
            isEdit: isEdit,
            form: form,
            openCreate: openCreate,
            openEdit: openEdit,
            handleSave: handleSave,
            handleDelete: handleDelete,
            handleSetDefault: handleSetDefault,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
