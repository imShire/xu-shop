import request from '@/utils/request';
export const getPaymentList = (params) => request.get('/admin/payments', { params });
export const getRefundList = (params) => request.get('/admin/refunds', { params });
export const createRefund = (orderId, data) => request.post(`/admin/orders/${orderId}/refund`, data);
export const getReconcileList = (params) => request.get('/admin/reconciliation', { params });
export const resolveReconcile = (id) => request.post(`/admin/reconciliation/${id}/resolve`);
