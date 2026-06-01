import request from '@/utils/request';
export const getAftersaleList = (params) => request.get('/admin/aftersales', { params });
export const getAftersaleDetail = (id) => request.get(`/admin/aftersales/${id}`);
export const agreeAftersale = (id, data = {}, idempotencyKey) => request.post(`/admin/aftersales/${id}/agree`, data, {
    headers: idempotencyKey ? { 'Idempotency-Key': idempotencyKey } : {},
});
export const rejectAftersale = (id, reason) => request.post(`/admin/aftersales/${id}/reject`, { reason });
export const confirmReceivedAftersale = (id, data = {}, idempotencyKey) => request.post(`/admin/aftersales/${id}/confirm-received`, data, {
    headers: idempotencyKey ? { 'Idempotency-Key': idempotencyKey } : {},
});
export const postAftersaleMessage = (id, data) => request.post(`/admin/aftersales/${id}/messages`, data);
export const closeAftersale = (id, reason) => request.post(`/admin/aftersales/${id}/close`, { reason });
