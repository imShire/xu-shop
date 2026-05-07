import request from '@/utils/request';
export const getOrderList = (params) => request.get('/admin/orders', { params });
export const getOrderDetail = (id) => request.get(`/admin/orders/${id}`);
export const cancelOrder = (id, reason) => request.post(`/admin/orders/${id}/cancel`, { reason });
export const addOrderRemark = (id, content) => request.post(`/admin/orders/${id}/remarks`, { content });
export const exportOrders = (params) => request.get('/admin/orders/export', { params, responseType: 'blob' });
// 运费模板
export const getFreightTemplates = () => request.get('/admin/freight-templates');
export const createFreightTemplate = (data) => request.post('/admin/freight-templates', data);
export const updateFreightTemplate = (id, data) => request.put(`/admin/freight-templates/${id}`, data);
export const deleteFreightTemplate = (id) => request.delete(`/admin/freight-templates/${id}`);
