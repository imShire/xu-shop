export const AFTERSALE_TYPE_LABEL = {
    refund_only: '仅退款',
    refund_return: '退货退款',
    exchange: '换货',
};
export const AFTERSALE_STATUS_LABEL = {
    applying: { label: '待商家处理', type: 'warning' },
    seller_agreed: { label: '商家已同意', type: 'primary' },
    buyer_returned: { label: '买家已寄回', type: 'primary' },
    seller_received: { label: '商家已收货', type: 'primary' },
    completed: { label: '已完成', type: 'success' },
    seller_rejected: { label: '商家已拒绝', type: 'danger' },
    cancelled: { label: '已取消', type: 'info' },
    closed: { label: '已关闭', type: 'info' },
};
export const NEGOTIATION_ROLE_LABEL = {
    buyer: '买家',
    seller: '商家',
    system: '系统',
};
/** 终态：不允许再发起任何业务操作（仅协商留言/手动关闭除外，见 ACTION_MATRIX） */
export const TERMINAL_STATUSES = ['completed', 'seller_rejected', 'cancelled', 'closed'];
export function isTerminal(status) {
    return TERMINAL_STATUSES.includes(status);
}
