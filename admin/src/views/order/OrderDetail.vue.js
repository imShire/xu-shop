import { ref, onMounted } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { ElMessage, ElMessageBox } from 'element-plus';
import { getOrderDetail, addOrderRemark } from '@/api/order';
import { createRefund } from '@/api/payment';
import { shipOrder } from '@/api/shipping';
import { formatAmount, formatTime, orderStatusMap } from '@/utils/format';
import PriceInput from '@/components/PriceInput/index.vue';
const route = useRoute();
const router = useRouter();
const orderId = String(route.params.id);
const loading = ref(false);
const order = ref(null);
const remarkDialogVisible = ref(false);
const remarkContent = ref('');
const remarkLoading = ref(false);
const refundDialogVisible = ref(false);
const refundAmount = ref(0);
const refundReason = ref('');
const refundLoading = ref(false);
const shipDialogVisible = ref(false);
const shipForm = ref({ carrier_code: '', tracking_no: '' });
const shipLoading = ref(false);
async function loadOrder() {
    loading.value = true;
    try {
        order.value = await getOrderDetail(orderId);
    }
    finally {
        loading.value = false;
    }
}
async function handleAddRemark() {
    if (!remarkContent.value.trim())
        return ElMessage.warning('请输入备注内容');
    remarkLoading.value = true;
    try {
        await addOrderRemark(orderId, remarkContent.value);
        ElMessage.success('备注添加成功');
        remarkDialogVisible.value = false;
        await loadOrder();
    }
    catch (e) {
        ElMessage.error(e?.message || '操作失败');
    }
    finally {
        remarkLoading.value = false;
    }
}
async function handleRefund() {
    refundLoading.value = true;
    try {
        await ElMessageBox.confirm(`确认退款 ${formatAmount(refundAmount.value)} 元？此操作不可撤销`, '退款确认', { type: 'warning' });
        await createRefund(orderId, { amount_cents: refundAmount.value, reason: refundReason.value });
        ElMessage.success('退款申请已提交');
        refundDialogVisible.value = false;
        await loadOrder();
    }
    catch (e) {
        if (e !== 'cancel')
            ElMessage.error(e?.message || '操作失败');
    }
    finally {
        refundLoading.value = false;
    }
}
async function handleShip() {
    if (!shipForm.value.carrier_code || !shipForm.value.tracking_no) {
        return ElMessage.warning('请填写快递信息');
    }
    shipLoading.value = true;
    try {
        await ElMessageBox.confirm('确认发货？发货后无法撤回', '发货确认', { type: 'warning' });
        await shipOrder(orderId, shipForm.value);
        ElMessage.success('发货成功');
        shipDialogVisible.value = false;
        await loadOrder();
    }
    catch (e) {
        if (e !== 'cancel')
            ElMessage.error(e?.message || '操作失败');
    }
    finally {
        shipLoading.value = false;
    }
}
const logSteps = [
    { status: 'pending', label: '待付款' },
    { status: 'paid', label: '已付款' },
    { status: 'shipped', label: '已发货' },
    { status: 'completed', label: '已完成' },
];
function getActiveStep(status) {
    const idx = logSteps.findIndex((s) => s.status === status);
    return idx >= 0 ? idx : 0;
}
onMounted(loadOrder);
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({});
__VLS_asFunctionalDirective(__VLS_directives.vLoading)(null, { ...__VLS_directiveBindingRestFields, value: (__VLS_ctx.loading) }, null, null);
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ style: {} },
});
const __VLS_0 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_1 = __VLS_asFunctionalComponent(__VLS_0, new __VLS_0({
    ...{ 'onClick': {} },
    text: true,
    icon: ('ArrowLeft'),
}));
const __VLS_2 = __VLS_1({
    ...{ 'onClick': {} },
    text: true,
    icon: ('ArrowLeft'),
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
if (__VLS_ctx.order) {
    const __VLS_8 = {}.ElCard;
    /** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
    // @ts-ignore
    const __VLS_9 = __VLS_asFunctionalComponent(__VLS_8, new __VLS_8({
        ...{ style: {} },
    }));
    const __VLS_10 = __VLS_9({
        ...{ style: {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_9));
    __VLS_11.slots.default;
    const __VLS_12 = {}.ElSteps;
    /** @type {[typeof __VLS_components.ElSteps, typeof __VLS_components.elSteps, typeof __VLS_components.ElSteps, typeof __VLS_components.elSteps, ]} */ ;
    // @ts-ignore
    const __VLS_13 = __VLS_asFunctionalComponent(__VLS_12, new __VLS_12({
        active: (__VLS_ctx.getActiveStep(__VLS_ctx.order.status)),
        alignCenter: true,
    }));
    const __VLS_14 = __VLS_13({
        active: (__VLS_ctx.getActiveStep(__VLS_ctx.order.status)),
        alignCenter: true,
    }, ...__VLS_functionalComponentArgsRest(__VLS_13));
    __VLS_15.slots.default;
    for (const [step] of __VLS_getVForSourceType((__VLS_ctx.logSteps))) {
        const __VLS_16 = {}.ElStep;
        /** @type {[typeof __VLS_components.ElStep, typeof __VLS_components.elStep, ]} */ ;
        // @ts-ignore
        const __VLS_17 = __VLS_asFunctionalComponent(__VLS_16, new __VLS_16({
            key: (step.status),
            title: (step.label),
        }));
        const __VLS_18 = __VLS_17({
            key: (step.status),
            title: (step.label),
        }, ...__VLS_functionalComponentArgsRest(__VLS_17));
    }
    var __VLS_15;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ style: {} },
    });
    const __VLS_20 = {}.ElTag;
    /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
    // @ts-ignore
    const __VLS_21 = __VLS_asFunctionalComponent(__VLS_20, new __VLS_20({
        type: (__VLS_ctx.orderStatusMap[__VLS_ctx.order.status]?.type || ''),
        size: "large",
    }));
    const __VLS_22 = __VLS_21({
        type: (__VLS_ctx.orderStatusMap[__VLS_ctx.order.status]?.type || ''),
        size: "large",
    }, ...__VLS_functionalComponentArgsRest(__VLS_21));
    __VLS_23.slots.default;
    (__VLS_ctx.orderStatusMap[__VLS_ctx.order.status]?.label || __VLS_ctx.order.status);
    var __VLS_23;
    if (__VLS_ctx.order.status === 'paid') {
        const __VLS_24 = {}.ElButton;
        /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
        // @ts-ignore
        const __VLS_25 = __VLS_asFunctionalComponent(__VLS_24, new __VLS_24({
            ...{ 'onClick': {} },
            type: "primary",
            size: "small",
        }));
        const __VLS_26 = __VLS_25({
            ...{ 'onClick': {} },
            type: "primary",
            size: "small",
        }, ...__VLS_functionalComponentArgsRest(__VLS_25));
        let __VLS_28;
        let __VLS_29;
        let __VLS_30;
        const __VLS_31 = {
            onClick: (...[$event]) => {
                if (!(__VLS_ctx.order))
                    return;
                if (!(__VLS_ctx.order.status === 'paid'))
                    return;
                __VLS_ctx.shipDialogVisible = true;
            }
        };
        __VLS_27.slots.default;
        var __VLS_27;
    }
    if (['paid', 'shipped'].includes(__VLS_ctx.order.status)) {
        const __VLS_32 = {}.ElButton;
        /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
        // @ts-ignore
        const __VLS_33 = __VLS_asFunctionalComponent(__VLS_32, new __VLS_32({
            ...{ 'onClick': {} },
            type: "warning",
            size: "small",
        }));
        const __VLS_34 = __VLS_33({
            ...{ 'onClick': {} },
            type: "warning",
            size: "small",
        }, ...__VLS_functionalComponentArgsRest(__VLS_33));
        let __VLS_36;
        let __VLS_37;
        let __VLS_38;
        const __VLS_39 = {
            onClick: (...[$event]) => {
                if (!(__VLS_ctx.order))
                    return;
                if (!(['paid', 'shipped'].includes(__VLS_ctx.order.status)))
                    return;
                __VLS_ctx.refundDialogVisible = true;
                __VLS_ctx.refundAmount = __VLS_ctx.order.pay_amount;
            }
        };
        __VLS_35.slots.default;
        var __VLS_35;
    }
    const __VLS_40 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_41 = __VLS_asFunctionalComponent(__VLS_40, new __VLS_40({
        ...{ 'onClick': {} },
        size: "small",
    }));
    const __VLS_42 = __VLS_41({
        ...{ 'onClick': {} },
        size: "small",
    }, ...__VLS_functionalComponentArgsRest(__VLS_41));
    let __VLS_44;
    let __VLS_45;
    let __VLS_46;
    const __VLS_47 = {
        onClick: (...[$event]) => {
            if (!(__VLS_ctx.order))
                return;
            __VLS_ctx.remarkDialogVisible = true;
        }
    };
    __VLS_43.slots.default;
    var __VLS_43;
    var __VLS_11;
    const __VLS_48 = {}.ElRow;
    /** @type {[typeof __VLS_components.ElRow, typeof __VLS_components.elRow, typeof __VLS_components.ElRow, typeof __VLS_components.elRow, ]} */ ;
    // @ts-ignore
    const __VLS_49 = __VLS_asFunctionalComponent(__VLS_48, new __VLS_48({
        gutter: (16),
    }));
    const __VLS_50 = __VLS_49({
        gutter: (16),
    }, ...__VLS_functionalComponentArgsRest(__VLS_49));
    __VLS_51.slots.default;
    const __VLS_52 = {}.ElCol;
    /** @type {[typeof __VLS_components.ElCol, typeof __VLS_components.elCol, typeof __VLS_components.ElCol, typeof __VLS_components.elCol, ]} */ ;
    // @ts-ignore
    const __VLS_53 = __VLS_asFunctionalComponent(__VLS_52, new __VLS_52({
        span: (16),
    }));
    const __VLS_54 = __VLS_53({
        span: (16),
    }, ...__VLS_functionalComponentArgsRest(__VLS_53));
    __VLS_55.slots.default;
    const __VLS_56 = {}.ElCard;
    /** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
    // @ts-ignore
    const __VLS_57 = __VLS_asFunctionalComponent(__VLS_56, new __VLS_56({
        ...{ style: {} },
    }));
    const __VLS_58 = __VLS_57({
        ...{ style: {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_57));
    __VLS_59.slots.default;
    {
        const { header: __VLS_thisSlot } = __VLS_59.slots;
    }
    for (const [item] of __VLS_getVForSourceType((__VLS_ctx.order.items))) {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            key: (item.id),
            ...{ style: {} },
        });
        if (item.cover) {
            __VLS_asFunctionalElement(__VLS_intrinsicElements.img)({
                src: (item.cover),
                ...{ style: {} },
            });
        }
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ style: {} },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ style: {} },
        });
        (item.product_title);
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ style: {} },
        });
        (item.spec_values?.join(' / '));
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ style: {} },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
            ...{ style: {} },
        });
        (__VLS_ctx.formatAmount(item.price));
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
        (item.quantity);
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
            ...{ style: {} },
        });
        (__VLS_ctx.formatAmount(item.subtotal));
    }
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ style: {} },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ style: {} },
    });
    (__VLS_ctx.formatAmount(__VLS_ctx.order.shipping_fee));
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ style: {} },
    });
    (__VLS_ctx.formatAmount(__VLS_ctx.order.discount_amount));
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ style: {} },
    });
    (__VLS_ctx.formatAmount(__VLS_ctx.order.pay_amount));
    var __VLS_59;
    if (__VLS_ctx.order.logs?.length) {
        const __VLS_60 = {}.ElCard;
        /** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
        // @ts-ignore
        const __VLS_61 = __VLS_asFunctionalComponent(__VLS_60, new __VLS_60({}));
        const __VLS_62 = __VLS_61({}, ...__VLS_functionalComponentArgsRest(__VLS_61));
        __VLS_63.slots.default;
        {
            const { header: __VLS_thisSlot } = __VLS_63.slots;
        }
        const __VLS_64 = {}.ElTimeline;
        /** @type {[typeof __VLS_components.ElTimeline, typeof __VLS_components.elTimeline, typeof __VLS_components.ElTimeline, typeof __VLS_components.elTimeline, ]} */ ;
        // @ts-ignore
        const __VLS_65 = __VLS_asFunctionalComponent(__VLS_64, new __VLS_64({}));
        const __VLS_66 = __VLS_65({}, ...__VLS_functionalComponentArgsRest(__VLS_65));
        __VLS_67.slots.default;
        for (const [log] of __VLS_getVForSourceType((__VLS_ctx.order.logs))) {
            const __VLS_68 = {}.ElTimelineItem;
            /** @type {[typeof __VLS_components.ElTimelineItem, typeof __VLS_components.elTimelineItem, typeof __VLS_components.ElTimelineItem, typeof __VLS_components.elTimelineItem, ]} */ ;
            // @ts-ignore
            const __VLS_69 = __VLS_asFunctionalComponent(__VLS_68, new __VLS_68({
                key: (log.id),
                timestamp: (__VLS_ctx.formatTime(log.created_at)),
                placement: "top",
            }));
            const __VLS_70 = __VLS_69({
                key: (log.id),
                timestamp: (__VLS_ctx.formatTime(log.created_at)),
                placement: "top",
            }, ...__VLS_functionalComponentArgsRest(__VLS_69));
            __VLS_71.slots.default;
            __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({});
            (log.content);
            __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
                ...{ style: {} },
            });
            (log.operator);
            var __VLS_71;
        }
        var __VLS_67;
        var __VLS_63;
    }
    if (__VLS_ctx.order.remarks?.length) {
        const __VLS_72 = {}.ElCard;
        /** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
        // @ts-ignore
        const __VLS_73 = __VLS_asFunctionalComponent(__VLS_72, new __VLS_72({
            ...{ style: {} },
        }));
        const __VLS_74 = __VLS_73({
            ...{ style: {} },
        }, ...__VLS_functionalComponentArgsRest(__VLS_73));
        __VLS_75.slots.default;
        {
            const { header: __VLS_thisSlot } = __VLS_75.slots;
        }
        for (const [r] of __VLS_getVForSourceType((__VLS_ctx.order.remarks))) {
            __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
                key: (r.id),
                ...{ style: {} },
            });
            __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({});
            (r.content);
            __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
                ...{ style: {} },
            });
            (__VLS_ctx.formatTime(r.created_at));
        }
        var __VLS_75;
    }
    var __VLS_55;
    const __VLS_76 = {}.ElCol;
    /** @type {[typeof __VLS_components.ElCol, typeof __VLS_components.elCol, typeof __VLS_components.ElCol, typeof __VLS_components.elCol, ]} */ ;
    // @ts-ignore
    const __VLS_77 = __VLS_asFunctionalComponent(__VLS_76, new __VLS_76({
        span: (8),
    }));
    const __VLS_78 = __VLS_77({
        span: (8),
    }, ...__VLS_functionalComponentArgsRest(__VLS_77));
    __VLS_79.slots.default;
    const __VLS_80 = {}.ElCard;
    /** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
    // @ts-ignore
    const __VLS_81 = __VLS_asFunctionalComponent(__VLS_80, new __VLS_80({
        ...{ style: {} },
    }));
    const __VLS_82 = __VLS_81({
        ...{ style: {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_81));
    __VLS_83.slots.default;
    {
        const { header: __VLS_thisSlot } = __VLS_83.slots;
    }
    const __VLS_84 = {}.ElDescriptions;
    /** @type {[typeof __VLS_components.ElDescriptions, typeof __VLS_components.elDescriptions, typeof __VLS_components.ElDescriptions, typeof __VLS_components.elDescriptions, ]} */ ;
    // @ts-ignore
    const __VLS_85 = __VLS_asFunctionalComponent(__VLS_84, new __VLS_84({
        column: (1),
        size: "small",
    }));
    const __VLS_86 = __VLS_85({
        column: (1),
        size: "small",
    }, ...__VLS_functionalComponentArgsRest(__VLS_85));
    __VLS_87.slots.default;
    const __VLS_88 = {}.ElDescriptionsItem;
    /** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
    // @ts-ignore
    const __VLS_89 = __VLS_asFunctionalComponent(__VLS_88, new __VLS_88({
        label: "收货人",
    }));
    const __VLS_90 = __VLS_89({
        label: "收货人",
    }, ...__VLS_functionalComponentArgsRest(__VLS_89));
    __VLS_91.slots.default;
    (__VLS_ctx.order.address?.name);
    var __VLS_91;
    const __VLS_92 = {}.ElDescriptionsItem;
    /** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
    // @ts-ignore
    const __VLS_93 = __VLS_asFunctionalComponent(__VLS_92, new __VLS_92({
        label: "手机号",
    }));
    const __VLS_94 = __VLS_93({
        label: "手机号",
    }, ...__VLS_functionalComponentArgsRest(__VLS_93));
    __VLS_95.slots.default;
    (__VLS_ctx.order.address?.phone);
    var __VLS_95;
    const __VLS_96 = {}.ElDescriptionsItem;
    /** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
    // @ts-ignore
    const __VLS_97 = __VLS_asFunctionalComponent(__VLS_96, new __VLS_96({
        label: "地址",
    }));
    const __VLS_98 = __VLS_97({
        label: "地址",
    }, ...__VLS_functionalComponentArgsRest(__VLS_97));
    __VLS_99.slots.default;
    (__VLS_ctx.order.address?.province);
    (__VLS_ctx.order.address?.city);
    (__VLS_ctx.order.address?.district);
    (__VLS_ctx.order.address?.address);
    var __VLS_99;
    var __VLS_87;
    var __VLS_83;
    const __VLS_100 = {}.ElCard;
    /** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
    // @ts-ignore
    const __VLS_101 = __VLS_asFunctionalComponent(__VLS_100, new __VLS_100({
        ...{ style: {} },
    }));
    const __VLS_102 = __VLS_101({
        ...{ style: {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_101));
    __VLS_103.slots.default;
    {
        const { header: __VLS_thisSlot } = __VLS_103.slots;
    }
    const __VLS_104 = {}.ElDescriptions;
    /** @type {[typeof __VLS_components.ElDescriptions, typeof __VLS_components.elDescriptions, typeof __VLS_components.ElDescriptions, typeof __VLS_components.elDescriptions, ]} */ ;
    // @ts-ignore
    const __VLS_105 = __VLS_asFunctionalComponent(__VLS_104, new __VLS_104({
        column: (1),
        size: "small",
    }));
    const __VLS_106 = __VLS_105({
        column: (1),
        size: "small",
    }, ...__VLS_functionalComponentArgsRest(__VLS_105));
    __VLS_107.slots.default;
    const __VLS_108 = {}.ElDescriptionsItem;
    /** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
    // @ts-ignore
    const __VLS_109 = __VLS_asFunctionalComponent(__VLS_108, new __VLS_108({
        label: "下单时间",
    }));
    const __VLS_110 = __VLS_109({
        label: "下单时间",
    }, ...__VLS_functionalComponentArgsRest(__VLS_109));
    __VLS_111.slots.default;
    (__VLS_ctx.formatTime(__VLS_ctx.order.created_at));
    var __VLS_111;
    const __VLS_112 = {}.ElDescriptionsItem;
    /** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
    // @ts-ignore
    const __VLS_113 = __VLS_asFunctionalComponent(__VLS_112, new __VLS_112({
        label: "付款时间",
    }));
    const __VLS_114 = __VLS_113({
        label: "付款时间",
    }, ...__VLS_functionalComponentArgsRest(__VLS_113));
    __VLS_115.slots.default;
    (__VLS_ctx.formatTime(__VLS_ctx.order.paid_at));
    var __VLS_115;
    const __VLS_116 = {}.ElDescriptionsItem;
    /** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
    // @ts-ignore
    const __VLS_117 = __VLS_asFunctionalComponent(__VLS_116, new __VLS_116({
        label: "订单号",
    }));
    const __VLS_118 = __VLS_117({
        label: "订单号",
    }, ...__VLS_functionalComponentArgsRest(__VLS_117));
    __VLS_119.slots.default;
    (__VLS_ctx.order.order_no);
    var __VLS_119;
    var __VLS_107;
    var __VLS_103;
    if (__VLS_ctx.order.shipment) {
        const __VLS_120 = {}.ElCard;
        /** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
        // @ts-ignore
        const __VLS_121 = __VLS_asFunctionalComponent(__VLS_120, new __VLS_120({}));
        const __VLS_122 = __VLS_121({}, ...__VLS_functionalComponentArgsRest(__VLS_121));
        __VLS_123.slots.default;
        {
            const { header: __VLS_thisSlot } = __VLS_123.slots;
        }
        const __VLS_124 = {}.ElDescriptions;
        /** @type {[typeof __VLS_components.ElDescriptions, typeof __VLS_components.elDescriptions, typeof __VLS_components.ElDescriptions, typeof __VLS_components.elDescriptions, ]} */ ;
        // @ts-ignore
        const __VLS_125 = __VLS_asFunctionalComponent(__VLS_124, new __VLS_124({
            column: (1),
            size: "small",
        }));
        const __VLS_126 = __VLS_125({
            column: (1),
            size: "small",
        }, ...__VLS_functionalComponentArgsRest(__VLS_125));
        __VLS_127.slots.default;
        const __VLS_128 = {}.ElDescriptionsItem;
        /** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
        // @ts-ignore
        const __VLS_129 = __VLS_asFunctionalComponent(__VLS_128, new __VLS_128({
            label: "快递公司",
        }));
        const __VLS_130 = __VLS_129({
            label: "快递公司",
        }, ...__VLS_functionalComponentArgsRest(__VLS_129));
        __VLS_131.slots.default;
        (__VLS_ctx.order.shipment.carrier);
        var __VLS_131;
        const __VLS_132 = {}.ElDescriptionsItem;
        /** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
        // @ts-ignore
        const __VLS_133 = __VLS_asFunctionalComponent(__VLS_132, new __VLS_132({
            label: "运单号",
        }));
        const __VLS_134 = __VLS_133({
            label: "运单号",
        }, ...__VLS_functionalComponentArgsRest(__VLS_133));
        __VLS_135.slots.default;
        (__VLS_ctx.order.shipment.tracking_no);
        var __VLS_135;
        const __VLS_136 = {}.ElDescriptionsItem;
        /** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
        // @ts-ignore
        const __VLS_137 = __VLS_asFunctionalComponent(__VLS_136, new __VLS_136({
            label: "发货时间",
        }));
        const __VLS_138 = __VLS_137({
            label: "发货时间",
        }, ...__VLS_functionalComponentArgsRest(__VLS_137));
        __VLS_139.slots.default;
        (__VLS_ctx.formatTime(__VLS_ctx.order.shipment.shipped_at));
        var __VLS_139;
        var __VLS_127;
        if (__VLS_ctx.order.shipment.tracks?.length) {
            __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
                ...{ style: {} },
            });
            const __VLS_140 = {}.ElTimeline;
            /** @type {[typeof __VLS_components.ElTimeline, typeof __VLS_components.elTimeline, typeof __VLS_components.ElTimeline, typeof __VLS_components.elTimeline, ]} */ ;
            // @ts-ignore
            const __VLS_141 = __VLS_asFunctionalComponent(__VLS_140, new __VLS_140({}));
            const __VLS_142 = __VLS_141({}, ...__VLS_functionalComponentArgsRest(__VLS_141));
            __VLS_143.slots.default;
            for (const [track] of __VLS_getVForSourceType((__VLS_ctx.order.shipment.tracks))) {
                const __VLS_144 = {}.ElTimelineItem;
                /** @type {[typeof __VLS_components.ElTimelineItem, typeof __VLS_components.elTimelineItem, typeof __VLS_components.ElTimelineItem, typeof __VLS_components.elTimelineItem, ]} */ ;
                // @ts-ignore
                const __VLS_145 = __VLS_asFunctionalComponent(__VLS_144, new __VLS_144({
                    key: (track.time),
                    timestamp: (track.time),
                    size: "small",
                }));
                const __VLS_146 = __VLS_145({
                    key: (track.time),
                    timestamp: (track.time),
                    size: "small",
                }, ...__VLS_functionalComponentArgsRest(__VLS_145));
                __VLS_147.slots.default;
                (track.content);
                var __VLS_147;
            }
            var __VLS_143;
        }
        var __VLS_123;
    }
    var __VLS_79;
    var __VLS_51;
}
const __VLS_148 = {}.ElDialog;
/** @type {[typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, ]} */ ;
// @ts-ignore
const __VLS_149 = __VLS_asFunctionalComponent(__VLS_148, new __VLS_148({
    modelValue: (__VLS_ctx.remarkDialogVisible),
    title: "添加备注",
    width: "400px",
}));
const __VLS_150 = __VLS_149({
    modelValue: (__VLS_ctx.remarkDialogVisible),
    title: "添加备注",
    width: "400px",
}, ...__VLS_functionalComponentArgsRest(__VLS_149));
__VLS_151.slots.default;
const __VLS_152 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_153 = __VLS_asFunctionalComponent(__VLS_152, new __VLS_152({
    modelValue: (__VLS_ctx.remarkContent),
    type: "textarea",
    rows: (4),
    placeholder: "请输入备注内容",
}));
const __VLS_154 = __VLS_153({
    modelValue: (__VLS_ctx.remarkContent),
    type: "textarea",
    rows: (4),
    placeholder: "请输入备注内容",
}, ...__VLS_functionalComponentArgsRest(__VLS_153));
{
    const { footer: __VLS_thisSlot } = __VLS_151.slots;
    const __VLS_156 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_157 = __VLS_asFunctionalComponent(__VLS_156, new __VLS_156({
        ...{ 'onClick': {} },
    }));
    const __VLS_158 = __VLS_157({
        ...{ 'onClick': {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_157));
    let __VLS_160;
    let __VLS_161;
    let __VLS_162;
    const __VLS_163 = {
        onClick: (...[$event]) => {
            __VLS_ctx.remarkDialogVisible = false;
        }
    };
    __VLS_159.slots.default;
    var __VLS_159;
    const __VLS_164 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_165 = __VLS_asFunctionalComponent(__VLS_164, new __VLS_164({
        ...{ 'onClick': {} },
        type: "primary",
        loading: (__VLS_ctx.remarkLoading),
    }));
    const __VLS_166 = __VLS_165({
        ...{ 'onClick': {} },
        type: "primary",
        loading: (__VLS_ctx.remarkLoading),
    }, ...__VLS_functionalComponentArgsRest(__VLS_165));
    let __VLS_168;
    let __VLS_169;
    let __VLS_170;
    const __VLS_171 = {
        onClick: (__VLS_ctx.handleAddRemark)
    };
    __VLS_167.slots.default;
    var __VLS_167;
}
var __VLS_151;
const __VLS_172 = {}.ElDialog;
/** @type {[typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, ]} */ ;
// @ts-ignore
const __VLS_173 = __VLS_asFunctionalComponent(__VLS_172, new __VLS_172({
    modelValue: (__VLS_ctx.refundDialogVisible),
    title: "申请退款",
    width: "400px",
}));
const __VLS_174 = __VLS_173({
    modelValue: (__VLS_ctx.refundDialogVisible),
    title: "申请退款",
    width: "400px",
}, ...__VLS_functionalComponentArgsRest(__VLS_173));
__VLS_175.slots.default;
const __VLS_176 = {}.ElForm;
/** @type {[typeof __VLS_components.ElForm, typeof __VLS_components.elForm, typeof __VLS_components.ElForm, typeof __VLS_components.elForm, ]} */ ;
// @ts-ignore
const __VLS_177 = __VLS_asFunctionalComponent(__VLS_176, new __VLS_176({
    labelWidth: "80px",
}));
const __VLS_178 = __VLS_177({
    labelWidth: "80px",
}, ...__VLS_functionalComponentArgsRest(__VLS_177));
__VLS_179.slots.default;
const __VLS_180 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_181 = __VLS_asFunctionalComponent(__VLS_180, new __VLS_180({
    label: "退款金额",
}));
const __VLS_182 = __VLS_181({
    label: "退款金额",
}, ...__VLS_functionalComponentArgsRest(__VLS_181));
__VLS_183.slots.default;
/** @type {[typeof PriceInput, ]} */ ;
// @ts-ignore
const __VLS_184 = __VLS_asFunctionalComponent(PriceInput, new PriceInput({
    modelValue: (__VLS_ctx.refundAmount),
}));
const __VLS_185 = __VLS_184({
    modelValue: (__VLS_ctx.refundAmount),
}, ...__VLS_functionalComponentArgsRest(__VLS_184));
var __VLS_183;
const __VLS_187 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_188 = __VLS_asFunctionalComponent(__VLS_187, new __VLS_187({
    label: "退款原因",
}));
const __VLS_189 = __VLS_188({
    label: "退款原因",
}, ...__VLS_functionalComponentArgsRest(__VLS_188));
__VLS_190.slots.default;
const __VLS_191 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_192 = __VLS_asFunctionalComponent(__VLS_191, new __VLS_191({
    modelValue: (__VLS_ctx.refundReason),
    type: "textarea",
    rows: (3),
}));
const __VLS_193 = __VLS_192({
    modelValue: (__VLS_ctx.refundReason),
    type: "textarea",
    rows: (3),
}, ...__VLS_functionalComponentArgsRest(__VLS_192));
var __VLS_190;
var __VLS_179;
{
    const { footer: __VLS_thisSlot } = __VLS_175.slots;
    const __VLS_195 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_196 = __VLS_asFunctionalComponent(__VLS_195, new __VLS_195({
        ...{ 'onClick': {} },
    }));
    const __VLS_197 = __VLS_196({
        ...{ 'onClick': {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_196));
    let __VLS_199;
    let __VLS_200;
    let __VLS_201;
    const __VLS_202 = {
        onClick: (...[$event]) => {
            __VLS_ctx.refundDialogVisible = false;
        }
    };
    __VLS_198.slots.default;
    var __VLS_198;
    const __VLS_203 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_204 = __VLS_asFunctionalComponent(__VLS_203, new __VLS_203({
        ...{ 'onClick': {} },
        type: "warning",
        loading: (__VLS_ctx.refundLoading),
    }));
    const __VLS_205 = __VLS_204({
        ...{ 'onClick': {} },
        type: "warning",
        loading: (__VLS_ctx.refundLoading),
    }, ...__VLS_functionalComponentArgsRest(__VLS_204));
    let __VLS_207;
    let __VLS_208;
    let __VLS_209;
    const __VLS_210 = {
        onClick: (__VLS_ctx.handleRefund)
    };
    __VLS_206.slots.default;
    var __VLS_206;
}
var __VLS_175;
const __VLS_211 = {}.ElDialog;
/** @type {[typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, ]} */ ;
// @ts-ignore
const __VLS_212 = __VLS_asFunctionalComponent(__VLS_211, new __VLS_211({
    modelValue: (__VLS_ctx.shipDialogVisible),
    title: "手动发货",
    width: "400px",
}));
const __VLS_213 = __VLS_212({
    modelValue: (__VLS_ctx.shipDialogVisible),
    title: "手动发货",
    width: "400px",
}, ...__VLS_functionalComponentArgsRest(__VLS_212));
__VLS_214.slots.default;
const __VLS_215 = {}.ElForm;
/** @type {[typeof __VLS_components.ElForm, typeof __VLS_components.elForm, typeof __VLS_components.ElForm, typeof __VLS_components.elForm, ]} */ ;
// @ts-ignore
const __VLS_216 = __VLS_asFunctionalComponent(__VLS_215, new __VLS_215({
    model: (__VLS_ctx.shipForm),
    labelWidth: "80px",
}));
const __VLS_217 = __VLS_216({
    model: (__VLS_ctx.shipForm),
    labelWidth: "80px",
}, ...__VLS_functionalComponentArgsRest(__VLS_216));
__VLS_218.slots.default;
const __VLS_219 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_220 = __VLS_asFunctionalComponent(__VLS_219, new __VLS_219({
    label: "快递公司",
}));
const __VLS_221 = __VLS_220({
    label: "快递公司",
}, ...__VLS_functionalComponentArgsRest(__VLS_220));
__VLS_222.slots.default;
const __VLS_223 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_224 = __VLS_asFunctionalComponent(__VLS_223, new __VLS_223({
    modelValue: (__VLS_ctx.shipForm.carrier_code),
    placeholder: "如：顺丰",
}));
const __VLS_225 = __VLS_224({
    modelValue: (__VLS_ctx.shipForm.carrier_code),
    placeholder: "如：顺丰",
}, ...__VLS_functionalComponentArgsRest(__VLS_224));
var __VLS_222;
const __VLS_227 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_228 = __VLS_asFunctionalComponent(__VLS_227, new __VLS_227({
    label: "运单号",
}));
const __VLS_229 = __VLS_228({
    label: "运单号",
}, ...__VLS_functionalComponentArgsRest(__VLS_228));
__VLS_230.slots.default;
const __VLS_231 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_232 = __VLS_asFunctionalComponent(__VLS_231, new __VLS_231({
    modelValue: (__VLS_ctx.shipForm.tracking_no),
}));
const __VLS_233 = __VLS_232({
    modelValue: (__VLS_ctx.shipForm.tracking_no),
}, ...__VLS_functionalComponentArgsRest(__VLS_232));
var __VLS_230;
var __VLS_218;
{
    const { footer: __VLS_thisSlot } = __VLS_214.slots;
    const __VLS_235 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_236 = __VLS_asFunctionalComponent(__VLS_235, new __VLS_235({
        ...{ 'onClick': {} },
    }));
    const __VLS_237 = __VLS_236({
        ...{ 'onClick': {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_236));
    let __VLS_239;
    let __VLS_240;
    let __VLS_241;
    const __VLS_242 = {
        onClick: (...[$event]) => {
            __VLS_ctx.shipDialogVisible = false;
        }
    };
    __VLS_238.slots.default;
    var __VLS_238;
    const __VLS_243 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_244 = __VLS_asFunctionalComponent(__VLS_243, new __VLS_243({
        ...{ 'onClick': {} },
        type: "primary",
        loading: (__VLS_ctx.shipLoading),
    }));
    const __VLS_245 = __VLS_244({
        ...{ 'onClick': {} },
        type: "primary",
        loading: (__VLS_ctx.shipLoading),
    }, ...__VLS_functionalComponentArgsRest(__VLS_244));
    let __VLS_247;
    let __VLS_248;
    let __VLS_249;
    const __VLS_250 = {
        onClick: (__VLS_ctx.handleShip)
    };
    __VLS_246.slots.default;
    var __VLS_246;
}
var __VLS_214;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            formatAmount: formatAmount,
            formatTime: formatTime,
            orderStatusMap: orderStatusMap,
            PriceInput: PriceInput,
            router: router,
            loading: loading,
            order: order,
            remarkDialogVisible: remarkDialogVisible,
            remarkContent: remarkContent,
            remarkLoading: remarkLoading,
            refundDialogVisible: refundDialogVisible,
            refundAmount: refundAmount,
            refundReason: refundReason,
            refundLoading: refundLoading,
            shipDialogVisible: shipDialogVisible,
            shipForm: shipForm,
            shipLoading: shipLoading,
            handleAddRemark: handleAddRemark,
            handleRefund: handleRefund,
            handleShip: handleShip,
            logSteps: logSteps,
            getActiveStep: getActiveStep,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
