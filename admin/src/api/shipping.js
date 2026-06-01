import request from '@/utils/request';
// 发件地址
export const getSenderAddresses = () => request.get('/admin/sender-addresses');
export const createSenderAddress = (data) => request.post('/admin/sender-addresses', data);
export const updateSenderAddress = (id, data) => request.put(`/admin/sender-addresses/${id}`, data);
export const deleteSenderAddress = (id) => request.delete(`/admin/sender-addresses/${id}`);
export const setDefaultSenderAddress = (id) => request.post(`/admin/sender-addresses/${id}/default`);
// 快递商
export const getCarriers = () => request.get('/admin/carriers');
export const updateCarrier = (code, data) => request.put(`/admin/carriers/${code}`, data);
// 发货
export const shipOrder = (orderId, data) => request.post(`/admin/orders/${orderId}/ship`, data);
export const batchShipOrders = (data) => request.post('/admin/orders/batch-ship', data);
export const getBatchShipStatus = (taskId) => request.get(`/admin/orders/batch-ship/${taskId}`);
export const getShipmentList = (params) => request.get('/admin/shipments', { params });
export const getShipmentDetail = (id) => request.get(`/admin/shipments/${id}`);
