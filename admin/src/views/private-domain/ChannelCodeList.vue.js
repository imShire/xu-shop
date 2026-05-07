import { ref, onMounted } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import ProTable from '@/components/ProTable/index.vue';
import { useTable } from '@/composables/useTable';
import { getChannelCodeList, createChannelCode, updateChannelCode, deleteChannelCode } from '@/api/private-domain';
import { formatTime } from '@/utils/format';
const searchForm = ref({ name: '' });
const { list, total, page, pageSize, loading, fetch } = useTable((params) => getChannelCodeList({ ...params, ...searchForm.value }));
const columns = [
    { label: '渠道名', prop: 'name', width: 150 },
    { label: '二维码', slot: 'qrcode', width: 80, align: 'center' },
    { label: '扫码数', prop: 'scan_count', width: 90, align: 'right' },
    { label: '加粉数', prop: 'follow_count', width: 90, align: 'right' },
    { label: '下单数', prop: 'order_count', width: 90, align: 'right' },
    { label: '创建时间', prop: 'created_at', width: 150, formatter: (r) => formatTime(r.created_at) },
];
const createDialogVisible = ref(false);
const saving = ref(false);
const formRef = ref();
const form = ref({ name: '' });
const editDialogVisible = ref(false);
const editSaving = ref(false);
const editFormRef = ref();
const editForm = ref({ id: '', name: '', remark: '' });
function openCreate() {
    form.value = { name: '' };
    createDialogVisible.value = true;
}
async function handleSave() {
    await formRef.value?.validate();
    saving.value = true;
    try {
        await createChannelCode(form.value);
        ElMessage.success('创建成功');
        createDialogVisible.value = false;
        fetch(searchForm.value);
    }
    catch (e) {
        ElMessage.error(e?.message || '操作失败');
    }
    finally {
        saving.value = false;
    }
}
function openEdit(row) {
    editForm.value = { id: row.id, name: row.name, remark: row.remark || '' };
    editDialogVisible.value = true;
}
async function handleEditSave() {
    await editFormRef.value?.validate();
    editSaving.value = true;
    try {
        await updateChannelCode(editForm.value.id, { name: editForm.value.name, remark: editForm.value.remark });
        ElMessage.success('更新成功');
        editDialogVisible.value = false;
        fetch(searchForm.value);
    }
    catch (e) {
        ElMessage.error(e?.message || '操作失败');
    }
    finally {
        editSaving.value = false;
    }
}
async function handleDelete(row) {
    try {
        await ElMessageBox.confirm(`确认删除渠道码「${row.name}」？`, '提示', { type: 'warning' });
        await deleteChannelCode(row.id);
        ElMessage.success('删除成功');
        fetch(searchForm.value);
    }
    catch (e) {
        if (e !== 'cancel')
            ElMessage.error(e?.message || '操作失败');
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
    const { search: __VLS_thisSlot } = __VLS_2.slots;
    const __VLS_7 = {}.ElInput;
    /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
    // @ts-ignore
    const __VLS_8 = __VLS_asFunctionalComponent(__VLS_7, new __VLS_7({
        modelValue: (__VLS_ctx.searchForm.name),
        placeholder: "渠道名",
        clearable: true,
        ...{ style: {} },
    }));
    const __VLS_9 = __VLS_8({
        modelValue: (__VLS_ctx.searchForm.name),
        placeholder: "渠道名",
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
    __VLS_22.slots.default;
    var __VLS_22;
}
{
    const { qrcode: __VLS_thisSlot } = __VLS_2.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    if (row.qrcode_url) {
        const __VLS_27 = {}.ElImage;
        /** @type {[typeof __VLS_components.ElImage, typeof __VLS_components.elImage, ]} */ ;
        // @ts-ignore
        const __VLS_28 = __VLS_asFunctionalComponent(__VLS_27, new __VLS_27({
            src: (row.qrcode_url),
            ...{ style: {} },
            previewSrcList: ([row.qrcode_url]),
            fit: "cover",
        }));
        const __VLS_29 = __VLS_28({
            src: (row.qrcode_url),
            ...{ style: {} },
            previewSrcList: ([row.qrcode_url]),
            fit: "cover",
        }, ...__VLS_functionalComponentArgsRest(__VLS_28));
    }
}
{
    const { actions: __VLS_thisSlot } = __VLS_2.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_31 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_32 = __VLS_asFunctionalComponent(__VLS_31, new __VLS_31({
        ...{ 'onClick': {} },
        text: true,
        type: "primary",
        size: "small",
    }));
    const __VLS_33 = __VLS_32({
        ...{ 'onClick': {} },
        text: true,
        type: "primary",
        size: "small",
    }, ...__VLS_functionalComponentArgsRest(__VLS_32));
    let __VLS_35;
    let __VLS_36;
    let __VLS_37;
    const __VLS_38 = {
        onClick: (...[$event]) => {
            __VLS_ctx.openEdit(row);
        }
    };
    __VLS_34.slots.default;
    var __VLS_34;
    const __VLS_39 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_40 = __VLS_asFunctionalComponent(__VLS_39, new __VLS_39({
        ...{ 'onClick': {} },
        text: true,
        type: "danger",
        size: "small",
    }));
    const __VLS_41 = __VLS_40({
        ...{ 'onClick': {} },
        text: true,
        type: "danger",
        size: "small",
    }, ...__VLS_functionalComponentArgsRest(__VLS_40));
    let __VLS_43;
    let __VLS_44;
    let __VLS_45;
    const __VLS_46 = {
        onClick: (...[$event]) => {
            __VLS_ctx.handleDelete(row);
        }
    };
    __VLS_42.slots.default;
    var __VLS_42;
}
var __VLS_2;
const __VLS_47 = {}.ElDialog;
/** @type {[typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, ]} */ ;
// @ts-ignore
const __VLS_48 = __VLS_asFunctionalComponent(__VLS_47, new __VLS_47({
    modelValue: (__VLS_ctx.createDialogVisible),
    title: "生成渠道码",
    width: "400px",
}));
const __VLS_49 = __VLS_48({
    modelValue: (__VLS_ctx.createDialogVisible),
    title: "生成渠道码",
    width: "400px",
}, ...__VLS_functionalComponentArgsRest(__VLS_48));
__VLS_50.slots.default;
const __VLS_51 = {}.ElForm;
/** @type {[typeof __VLS_components.ElForm, typeof __VLS_components.elForm, typeof __VLS_components.ElForm, typeof __VLS_components.elForm, ]} */ ;
// @ts-ignore
const __VLS_52 = __VLS_asFunctionalComponent(__VLS_51, new __VLS_51({
    ref: "formRef",
    model: (__VLS_ctx.form),
    labelWidth: "80px",
}));
const __VLS_53 = __VLS_52({
    ref: "formRef",
    model: (__VLS_ctx.form),
    labelWidth: "80px",
}, ...__VLS_functionalComponentArgsRest(__VLS_52));
/** @type {typeof __VLS_ctx.formRef} */ ;
var __VLS_55 = {};
__VLS_54.slots.default;
const __VLS_57 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_58 = __VLS_asFunctionalComponent(__VLS_57, new __VLS_57({
    label: "渠道名",
    prop: "name",
    rules: ([{ required: true }]),
}));
const __VLS_59 = __VLS_58({
    label: "渠道名",
    prop: "name",
    rules: ([{ required: true }]),
}, ...__VLS_functionalComponentArgsRest(__VLS_58));
__VLS_60.slots.default;
const __VLS_61 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_62 = __VLS_asFunctionalComponent(__VLS_61, new __VLS_61({
    modelValue: (__VLS_ctx.form.name),
    placeholder: "如：抖音广告",
}));
const __VLS_63 = __VLS_62({
    modelValue: (__VLS_ctx.form.name),
    placeholder: "如：抖音广告",
}, ...__VLS_functionalComponentArgsRest(__VLS_62));
var __VLS_60;
var __VLS_54;
{
    const { footer: __VLS_thisSlot } = __VLS_50.slots;
    const __VLS_65 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_66 = __VLS_asFunctionalComponent(__VLS_65, new __VLS_65({
        ...{ 'onClick': {} },
    }));
    const __VLS_67 = __VLS_66({
        ...{ 'onClick': {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_66));
    let __VLS_69;
    let __VLS_70;
    let __VLS_71;
    const __VLS_72 = {
        onClick: (...[$event]) => {
            __VLS_ctx.createDialogVisible = false;
        }
    };
    __VLS_68.slots.default;
    var __VLS_68;
    const __VLS_73 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_74 = __VLS_asFunctionalComponent(__VLS_73, new __VLS_73({
        ...{ 'onClick': {} },
        type: "primary",
        loading: (__VLS_ctx.saving),
    }));
    const __VLS_75 = __VLS_74({
        ...{ 'onClick': {} },
        type: "primary",
        loading: (__VLS_ctx.saving),
    }, ...__VLS_functionalComponentArgsRest(__VLS_74));
    let __VLS_77;
    let __VLS_78;
    let __VLS_79;
    const __VLS_80 = {
        onClick: (__VLS_ctx.handleSave)
    };
    __VLS_76.slots.default;
    var __VLS_76;
}
var __VLS_50;
const __VLS_81 = {}.ElDialog;
/** @type {[typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, ]} */ ;
// @ts-ignore
const __VLS_82 = __VLS_asFunctionalComponent(__VLS_81, new __VLS_81({
    modelValue: (__VLS_ctx.editDialogVisible),
    title: "编辑渠道码",
    width: "400px",
}));
const __VLS_83 = __VLS_82({
    modelValue: (__VLS_ctx.editDialogVisible),
    title: "编辑渠道码",
    width: "400px",
}, ...__VLS_functionalComponentArgsRest(__VLS_82));
__VLS_84.slots.default;
const __VLS_85 = {}.ElForm;
/** @type {[typeof __VLS_components.ElForm, typeof __VLS_components.elForm, typeof __VLS_components.ElForm, typeof __VLS_components.elForm, ]} */ ;
// @ts-ignore
const __VLS_86 = __VLS_asFunctionalComponent(__VLS_85, new __VLS_85({
    ref: "editFormRef",
    model: (__VLS_ctx.editForm),
    labelWidth: "80px",
}));
const __VLS_87 = __VLS_86({
    ref: "editFormRef",
    model: (__VLS_ctx.editForm),
    labelWidth: "80px",
}, ...__VLS_functionalComponentArgsRest(__VLS_86));
/** @type {typeof __VLS_ctx.editFormRef} */ ;
var __VLS_89 = {};
__VLS_88.slots.default;
const __VLS_91 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_92 = __VLS_asFunctionalComponent(__VLS_91, new __VLS_91({
    label: "渠道名",
    prop: "name",
    rules: ([{ required: true }]),
}));
const __VLS_93 = __VLS_92({
    label: "渠道名",
    prop: "name",
    rules: ([{ required: true }]),
}, ...__VLS_functionalComponentArgsRest(__VLS_92));
__VLS_94.slots.default;
const __VLS_95 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_96 = __VLS_asFunctionalComponent(__VLS_95, new __VLS_95({
    modelValue: (__VLS_ctx.editForm.name),
}));
const __VLS_97 = __VLS_96({
    modelValue: (__VLS_ctx.editForm.name),
}, ...__VLS_functionalComponentArgsRest(__VLS_96));
var __VLS_94;
const __VLS_99 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_100 = __VLS_asFunctionalComponent(__VLS_99, new __VLS_99({
    label: "备注",
}));
const __VLS_101 = __VLS_100({
    label: "备注",
}, ...__VLS_functionalComponentArgsRest(__VLS_100));
__VLS_102.slots.default;
const __VLS_103 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_104 = __VLS_asFunctionalComponent(__VLS_103, new __VLS_103({
    modelValue: (__VLS_ctx.editForm.remark),
    type: "textarea",
    rows: (2),
}));
const __VLS_105 = __VLS_104({
    modelValue: (__VLS_ctx.editForm.remark),
    type: "textarea",
    rows: (2),
}, ...__VLS_functionalComponentArgsRest(__VLS_104));
var __VLS_102;
var __VLS_88;
{
    const { footer: __VLS_thisSlot } = __VLS_84.slots;
    const __VLS_107 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_108 = __VLS_asFunctionalComponent(__VLS_107, new __VLS_107({
        ...{ 'onClick': {} },
    }));
    const __VLS_109 = __VLS_108({
        ...{ 'onClick': {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_108));
    let __VLS_111;
    let __VLS_112;
    let __VLS_113;
    const __VLS_114 = {
        onClick: (...[$event]) => {
            __VLS_ctx.editDialogVisible = false;
        }
    };
    __VLS_110.slots.default;
    var __VLS_110;
    const __VLS_115 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_116 = __VLS_asFunctionalComponent(__VLS_115, new __VLS_115({
        ...{ 'onClick': {} },
        type: "primary",
        loading: (__VLS_ctx.editSaving),
    }));
    const __VLS_117 = __VLS_116({
        ...{ 'onClick': {} },
        type: "primary",
        loading: (__VLS_ctx.editSaving),
    }, ...__VLS_functionalComponentArgsRest(__VLS_116));
    let __VLS_119;
    let __VLS_120;
    let __VLS_121;
    const __VLS_122 = {
        onClick: (__VLS_ctx.handleEditSave)
    };
    __VLS_118.slots.default;
    var __VLS_118;
}
var __VLS_84;
/** @type {__VLS_StyleScopedClasses['page-card']} */ ;
// @ts-ignore
var __VLS_56 = __VLS_55, __VLS_90 = __VLS_89;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            ProTable: ProTable,
            searchForm: searchForm,
            list: list,
            total: total,
            page: page,
            pageSize: pageSize,
            loading: loading,
            fetch: fetch,
            columns: columns,
            createDialogVisible: createDialogVisible,
            saving: saving,
            formRef: formRef,
            form: form,
            editDialogVisible: editDialogVisible,
            editSaving: editSaving,
            editFormRef: editFormRef,
            editForm: editForm,
            openCreate: openCreate,
            handleSave: handleSave,
            openEdit: openEdit,
            handleEditSave: handleEditSave,
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
