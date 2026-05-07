import request from '@/utils/request';
export const getAftersaleList = (params) => request.get('/admin/aftersales', { params });
export const approveAftersale = (id) => request.post(`/admin/aftersales/${id}/approve`);
export const rejectAftersale = (id, reason) => request.post(`/admin/aftersales/${id}/reject`, { reason });
export const directRefund = (id, amount) => request.post(`/admin/aftersales/${id}/refund`, { amount });
