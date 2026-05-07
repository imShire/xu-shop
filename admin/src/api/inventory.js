import request from '@/utils/request';
export const getInventoryAlerts = (params) => request.get('/admin/inventory/alerts', { params });
export const markAlertRead = (id) => request.post(`/admin/inventory/alerts/${id}/read`);
export const markAllAlertsRead = () => request.post('/admin/inventory/alerts/read-all');
export const getInventoryLogs = (params) => request.get('/admin/inventory/logs', { params });
export const adjustInventory = (skuId, data) => request.post(`/admin/skus/${skuId}/adjust`, data);
