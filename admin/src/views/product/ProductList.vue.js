import { ref, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { ElMessage, ElMessageBox } from 'element-plus';
import ProTable from '@/components/ProTable/index.vue';
import { useTable } from '@/composables/useTable';
import { getProductList, putOnSale, putOffSale, copyProduct, deleteProduct, batchOnSale, batchOffSale, } from '@/api/product';
import { getCategoryList } from '@/api/product';
import { formatAmount, formatTime, productStatusMap } from '@/utils/format';
const router = useRouter();
const searchForm = ref({
    title: '',
    category_id: undefined,
    status: '',
});
const categories = ref([]);
const selectedRows = ref([]);
const { list, total, page, pageSize, loading, fetch } = useTable((params) => getProductList({ ...params, ...searchForm.value }));
const columns = [
    { label: '商品', slot: 'product', minWidth: 200 },
    { label: '分类', prop: 'category_name', width: 100 },
    {
        label: '价格',
        slot: 'price',
        width: 140,
    },
    { label: '库存', prop: 'total_stock', width: 80 },
    { label: '销量', slot: 'sales', width: 80 },
    { label: '运费', slot: 'freight', width: 100 },
    {
        label: '状态',
        slot: 'status',
        width: 90,
        align: 'center',
    },
    {
        label: '创建时间',
        prop: 'created_at',
        width: 150,
        formatter: (row) => formatTime(row.created_at),
    },
];
async function handleOnSale(row) {
    await putOnSale(row.id);
    ElMessage.success('已上架');
    fetch(searchForm.value);
}
async function handleOffSale(row) {
    await putOffSale(row.id);
    ElMessage.success('已下架');
    fetch(searchForm.value);
}
async function handleCopy(row) {
    await copyProduct(row.id);
    ElMessage.success('复制成功');
    fetch(searchForm.value);
}
async function handleDelete(row) {
    await ElMessageBox.confirm(`确认删除「${row.title}」？`, '提示', { type: 'warning' });
    await deleteProduct(row.id);
    ElMessage.success('删除成功');
    fetch(searchForm.value);
}
async function handleBatchOnSale() {
    if (!selectedRows.value.length)
        return ElMessage.warning('请先选择商品');
    await batchOnSale(selectedRows.value.map((r) => r.id));
    ElMessage.success('批量上架成功');
    fetch(searchForm.value);
}
async function handleBatchOffSale() {
    if (!selectedRows.value.length)
        return ElMessage.warning('请先选择商品');
    await batchOffSale(selectedRows.value.map((r) => r.id));
    ElMessage.success('批量下架成功');
    fetch(searchForm.value);
}
function handleSearch() {
    page.value = 1;
    fetch(searchForm.value);
}
function handleReset() {
    searchForm.value = { title: '', category_id: undefined, status: '' };
    handleSearch();
}
onMounted(async () => {
    const cats = await getCategoryList().catch(() => []);
    categories.value = cats;
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
    ...{ 'onSelectionChange': {} },
    columns: (__VLS_ctx.columns),
    data: (__VLS_ctx.list),
    total: (__VLS_ctx.total),
    loading: (__VLS_ctx.loading),
    page: (__VLS_ctx.page),
    pageSize: (__VLS_ctx.pageSize),
    selection: true,
}));
const __VLS_1 = __VLS_0({
    ...{ 'onRefresh': {} },
    ...{ 'onSelectionChange': {} },
    columns: (__VLS_ctx.columns),
    data: (__VLS_ctx.list),
    total: (__VLS_ctx.total),
    loading: (__VLS_ctx.loading),
    page: (__VLS_ctx.page),
    pageSize: (__VLS_ctx.pageSize),
    selection: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_0));
let __VLS_3;
let __VLS_4;
let __VLS_5;
const __VLS_6 = {
    onRefresh: (...[$event]) => {
        __VLS_ctx.fetch(__VLS_ctx.searchForm);
    }
};
const __VLS_7 = {
    onSelectionChange: ((rows) => (__VLS_ctx.selectedRows = rows))
};
__VLS_2.slots.default;
{
    const { search: __VLS_thisSlot } = __VLS_2.slots;
    const __VLS_8 = {}.ElInput;
    /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
    // @ts-ignore
    const __VLS_9 = __VLS_asFunctionalComponent(__VLS_8, new __VLS_8({
        ...{ 'onKeyup': {} },
        modelValue: (__VLS_ctx.searchForm.title),
        placeholder: "商品名称",
        clearable: true,
        ...{ style: {} },
    }));
    const __VLS_10 = __VLS_9({
        ...{ 'onKeyup': {} },
        modelValue: (__VLS_ctx.searchForm.title),
        placeholder: "商品名称",
        clearable: true,
        ...{ style: {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_9));
    let __VLS_12;
    let __VLS_13;
    let __VLS_14;
    const __VLS_15 = {
        onKeyup: (__VLS_ctx.handleSearch)
    };
    var __VLS_11;
    const __VLS_16 = {}.ElSelect;
    /** @type {[typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, ]} */ ;
    // @ts-ignore
    const __VLS_17 = __VLS_asFunctionalComponent(__VLS_16, new __VLS_16({
        modelValue: (__VLS_ctx.searchForm.category_id),
        placeholder: "商品分类",
        clearable: true,
        ...{ style: {} },
    }));
    const __VLS_18 = __VLS_17({
        modelValue: (__VLS_ctx.searchForm.category_id),
        placeholder: "商品分类",
        clearable: true,
        ...{ style: {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_17));
    __VLS_19.slots.default;
    for (const [cat] of __VLS_getVForSourceType((__VLS_ctx.categories))) {
        const __VLS_20 = {}.ElOption;
        /** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
        // @ts-ignore
        const __VLS_21 = __VLS_asFunctionalComponent(__VLS_20, new __VLS_20({
            key: (cat.id),
            label: (cat.name),
            value: (cat.id),
        }));
        const __VLS_22 = __VLS_21({
            key: (cat.id),
            label: (cat.name),
            value: (cat.id),
        }, ...__VLS_functionalComponentArgsRest(__VLS_21));
    }
    var __VLS_19;
    const __VLS_24 = {}.ElSelect;
    /** @type {[typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, ]} */ ;
    // @ts-ignore
    const __VLS_25 = __VLS_asFunctionalComponent(__VLS_24, new __VLS_24({
        modelValue: (__VLS_ctx.searchForm.status),
        placeholder: "商品状态",
        clearable: true,
        ...{ style: {} },
    }));
    const __VLS_26 = __VLS_25({
        modelValue: (__VLS_ctx.searchForm.status),
        placeholder: "商品状态",
        clearable: true,
        ...{ style: {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_25));
    __VLS_27.slots.default;
    const __VLS_28 = {}.ElOption;
    /** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
    // @ts-ignore
    const __VLS_29 = __VLS_asFunctionalComponent(__VLS_28, new __VLS_28({
        label: "草稿",
        value: "draft",
    }));
    const __VLS_30 = __VLS_29({
        label: "草稿",
        value: "draft",
    }, ...__VLS_functionalComponentArgsRest(__VLS_29));
    const __VLS_32 = {}.ElOption;
    /** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
    // @ts-ignore
    const __VLS_33 = __VLS_asFunctionalComponent(__VLS_32, new __VLS_32({
        label: "在售",
        value: "onsale",
    }));
    const __VLS_34 = __VLS_33({
        label: "在售",
        value: "onsale",
    }, ...__VLS_functionalComponentArgsRest(__VLS_33));
    const __VLS_36 = {}.ElOption;
    /** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
    // @ts-ignore
    const __VLS_37 = __VLS_asFunctionalComponent(__VLS_36, new __VLS_36({
        label: "下架",
        value: "offsale",
    }));
    const __VLS_38 = __VLS_37({
        label: "下架",
        value: "offsale",
    }, ...__VLS_functionalComponentArgsRest(__VLS_37));
    var __VLS_27;
    const __VLS_40 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_41 = __VLS_asFunctionalComponent(__VLS_40, new __VLS_40({
        ...{ 'onClick': {} },
        type: "primary",
    }));
    const __VLS_42 = __VLS_41({
        ...{ 'onClick': {} },
        type: "primary",
    }, ...__VLS_functionalComponentArgsRest(__VLS_41));
    let __VLS_44;
    let __VLS_45;
    let __VLS_46;
    const __VLS_47 = {
        onClick: (__VLS_ctx.handleSearch)
    };
    __VLS_43.slots.default;
    var __VLS_43;
    const __VLS_48 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_49 = __VLS_asFunctionalComponent(__VLS_48, new __VLS_48({
        ...{ 'onClick': {} },
    }));
    const __VLS_50 = __VLS_49({
        ...{ 'onClick': {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_49));
    let __VLS_52;
    let __VLS_53;
    let __VLS_54;
    const __VLS_55 = {
        onClick: (__VLS_ctx.handleReset)
    };
    __VLS_51.slots.default;
    var __VLS_51;
}
{
    const { toolbar: __VLS_thisSlot } = __VLS_2.slots;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ style: {} },
    });
    const __VLS_56 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_57 = __VLS_asFunctionalComponent(__VLS_56, new __VLS_56({
        ...{ 'onClick': {} },
        type: "primary",
    }));
    const __VLS_58 = __VLS_57({
        ...{ 'onClick': {} },
        type: "primary",
    }, ...__VLS_functionalComponentArgsRest(__VLS_57));
    let __VLS_60;
    let __VLS_61;
    let __VLS_62;
    const __VLS_63 = {
        onClick: (...[$event]) => {
            __VLS_ctx.router.push('/product/edit');
        }
    };
    __VLS_59.slots.default;
    var __VLS_59;
    const __VLS_64 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_65 = __VLS_asFunctionalComponent(__VLS_64, new __VLS_64({
        ...{ 'onClick': {} },
    }));
    const __VLS_66 = __VLS_65({
        ...{ 'onClick': {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_65));
    let __VLS_68;
    let __VLS_69;
    let __VLS_70;
    const __VLS_71 = {
        onClick: (__VLS_ctx.handleBatchOnSale)
    };
    __VLS_67.slots.default;
    var __VLS_67;
    const __VLS_72 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_73 = __VLS_asFunctionalComponent(__VLS_72, new __VLS_72({
        ...{ 'onClick': {} },
    }));
    const __VLS_74 = __VLS_73({
        ...{ 'onClick': {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_73));
    let __VLS_76;
    let __VLS_77;
    let __VLS_78;
    const __VLS_79 = {
        onClick: (__VLS_ctx.handleBatchOffSale)
    };
    __VLS_75.slots.default;
    var __VLS_75;
}
{
    const { product: __VLS_thisSlot } = __VLS_2.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ style: {} },
    });
    if (row.main_image) {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.img)({
            src: (row.main_image),
            ...{ style: {} },
        });
    }
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ style: {} },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "text-truncate" },
        ...{ style: {} },
    });
    (row.title);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "text-truncate" },
        ...{ style: {} },
    });
    (row.subtitle);
}
{
    const { price: __VLS_thisSlot } = __VLS_2.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({});
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ style: {} },
    });
    (__VLS_ctx.formatAmount(row.price_min_cents / 100));
    if (row.price_max_cents !== row.price_min_cents) {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
        (__VLS_ctx.formatAmount(row.price_max_cents / 100));
    }
}
{
    const { sales: __VLS_thisSlot } = __VLS_2.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    ((row.virtual_sales || 0) + (row.sales || 0));
}
{
    const { freight: __VLS_thisSlot } = __VLS_2.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    if (!row.freight_template_name) {
        const __VLS_80 = {}.ElTag;
        /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
        // @ts-ignore
        const __VLS_81 = __VLS_asFunctionalComponent(__VLS_80, new __VLS_80({
            type: "success",
            size: "small",
        }));
        const __VLS_82 = __VLS_81({
            type: "success",
            size: "small",
        }, ...__VLS_functionalComponentArgsRest(__VLS_81));
        __VLS_83.slots.default;
        var __VLS_83;
    }
    else {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
        (row.freight_template_name);
    }
}
{
    const { status: __VLS_thisSlot } = __VLS_2.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_84 = {}.ElTag;
    /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
    // @ts-ignore
    const __VLS_85 = __VLS_asFunctionalComponent(__VLS_84, new __VLS_84({
        type: (__VLS_ctx.productStatusMap[row.status]?.type || ''),
        size: "small",
    }));
    const __VLS_86 = __VLS_85({
        type: (__VLS_ctx.productStatusMap[row.status]?.type || ''),
        size: "small",
    }, ...__VLS_functionalComponentArgsRest(__VLS_85));
    __VLS_87.slots.default;
    (__VLS_ctx.productStatusMap[row.status]?.label || row.status);
    var __VLS_87;
}
{
    const { actions: __VLS_thisSlot } = __VLS_2.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_88 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_89 = __VLS_asFunctionalComponent(__VLS_88, new __VLS_88({
        ...{ 'onClick': {} },
        text: true,
        type: "primary",
        size: "small",
    }));
    const __VLS_90 = __VLS_89({
        ...{ 'onClick': {} },
        text: true,
        type: "primary",
        size: "small",
    }, ...__VLS_functionalComponentArgsRest(__VLS_89));
    let __VLS_92;
    let __VLS_93;
    let __VLS_94;
    const __VLS_95 = {
        onClick: (...[$event]) => {
            __VLS_ctx.router.push(`/product/edit/${row.id}`);
        }
    };
    __VLS_91.slots.default;
    var __VLS_91;
    if (row.status !== 'onsale') {
        const __VLS_96 = {}.ElButton;
        /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
        // @ts-ignore
        const __VLS_97 = __VLS_asFunctionalComponent(__VLS_96, new __VLS_96({
            ...{ 'onClick': {} },
            text: true,
            type: "success",
            size: "small",
        }));
        const __VLS_98 = __VLS_97({
            ...{ 'onClick': {} },
            text: true,
            type: "success",
            size: "small",
        }, ...__VLS_functionalComponentArgsRest(__VLS_97));
        let __VLS_100;
        let __VLS_101;
        let __VLS_102;
        const __VLS_103 = {
            onClick: (...[$event]) => {
                if (!(row.status !== 'onsale'))
                    return;
                __VLS_ctx.handleOnSale(row);
            }
        };
        __VLS_99.slots.default;
        var __VLS_99;
    }
    else {
        const __VLS_104 = {}.ElButton;
        /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
        // @ts-ignore
        const __VLS_105 = __VLS_asFunctionalComponent(__VLS_104, new __VLS_104({
            ...{ 'onClick': {} },
            text: true,
            type: "warning",
            size: "small",
        }));
        const __VLS_106 = __VLS_105({
            ...{ 'onClick': {} },
            text: true,
            type: "warning",
            size: "small",
        }, ...__VLS_functionalComponentArgsRest(__VLS_105));
        let __VLS_108;
        let __VLS_109;
        let __VLS_110;
        const __VLS_111 = {
            onClick: (...[$event]) => {
                if (!!(row.status !== 'onsale'))
                    return;
                __VLS_ctx.handleOffSale(row);
            }
        };
        __VLS_107.slots.default;
        var __VLS_107;
    }
    const __VLS_112 = {}.ElDropdown;
    /** @type {[typeof __VLS_components.ElDropdown, typeof __VLS_components.elDropdown, typeof __VLS_components.ElDropdown, typeof __VLS_components.elDropdown, ]} */ ;
    // @ts-ignore
    const __VLS_113 = __VLS_asFunctionalComponent(__VLS_112, new __VLS_112({}));
    const __VLS_114 = __VLS_113({}, ...__VLS_functionalComponentArgsRest(__VLS_113));
    __VLS_115.slots.default;
    const __VLS_116 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_117 = __VLS_asFunctionalComponent(__VLS_116, new __VLS_116({
        text: true,
        size: "small",
    }));
    const __VLS_118 = __VLS_117({
        text: true,
        size: "small",
    }, ...__VLS_functionalComponentArgsRest(__VLS_117));
    __VLS_119.slots.default;
    var __VLS_119;
    {
        const { dropdown: __VLS_thisSlot } = __VLS_115.slots;
        const __VLS_120 = {}.ElDropdownMenu;
        /** @type {[typeof __VLS_components.ElDropdownMenu, typeof __VLS_components.elDropdownMenu, typeof __VLS_components.ElDropdownMenu, typeof __VLS_components.elDropdownMenu, ]} */ ;
        // @ts-ignore
        const __VLS_121 = __VLS_asFunctionalComponent(__VLS_120, new __VLS_120({}));
        const __VLS_122 = __VLS_121({}, ...__VLS_functionalComponentArgsRest(__VLS_121));
        __VLS_123.slots.default;
        const __VLS_124 = {}.ElDropdownItem;
        /** @type {[typeof __VLS_components.ElDropdownItem, typeof __VLS_components.elDropdownItem, typeof __VLS_components.ElDropdownItem, typeof __VLS_components.elDropdownItem, ]} */ ;
        // @ts-ignore
        const __VLS_125 = __VLS_asFunctionalComponent(__VLS_124, new __VLS_124({
            ...{ 'onClick': {} },
        }));
        const __VLS_126 = __VLS_125({
            ...{ 'onClick': {} },
        }, ...__VLS_functionalComponentArgsRest(__VLS_125));
        let __VLS_128;
        let __VLS_129;
        let __VLS_130;
        const __VLS_131 = {
            onClick: (...[$event]) => {
                __VLS_ctx.handleCopy(row);
            }
        };
        __VLS_127.slots.default;
        var __VLS_127;
        const __VLS_132 = {}.ElDropdownItem;
        /** @type {[typeof __VLS_components.ElDropdownItem, typeof __VLS_components.elDropdownItem, typeof __VLS_components.ElDropdownItem, typeof __VLS_components.elDropdownItem, ]} */ ;
        // @ts-ignore
        const __VLS_133 = __VLS_asFunctionalComponent(__VLS_132, new __VLS_132({
            ...{ 'onClick': {} },
            ...{ style: {} },
        }));
        const __VLS_134 = __VLS_133({
            ...{ 'onClick': {} },
            ...{ style: {} },
        }, ...__VLS_functionalComponentArgsRest(__VLS_133));
        let __VLS_136;
        let __VLS_137;
        let __VLS_138;
        const __VLS_139 = {
            onClick: (...[$event]) => {
                __VLS_ctx.handleDelete(row);
            }
        };
        __VLS_135.slots.default;
        var __VLS_135;
        var __VLS_123;
    }
    var __VLS_115;
}
var __VLS_2;
/** @type {__VLS_StyleScopedClasses['page-card']} */ ;
/** @type {__VLS_StyleScopedClasses['text-truncate']} */ ;
/** @type {__VLS_StyleScopedClasses['text-truncate']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            ProTable: ProTable,
            formatAmount: formatAmount,
            productStatusMap: productStatusMap,
            router: router,
            searchForm: searchForm,
            categories: categories,
            selectedRows: selectedRows,
            list: list,
            total: total,
            page: page,
            pageSize: pageSize,
            loading: loading,
            fetch: fetch,
            columns: columns,
            handleOnSale: handleOnSale,
            handleOffSale: handleOffSale,
            handleCopy: handleCopy,
            handleDelete: handleDelete,
            handleBatchOnSale: handleBatchOnSale,
            handleBatchOffSale: handleBatchOffSale,
            handleSearch: handleSearch,
            handleReset: handleReset,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
