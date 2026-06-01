import request from '@/utils/request';
export const getNotificationList = (params) => request.get('/admin/notifications', { params });
export const getTemplateList = () => request.get('/admin/notification-templates');
export const updateTemplate = (code, data) => request.put(`/admin/notification-templates/${code}`, data);
export const testSendNotification = (code, data) => request.post(`/admin/notification-templates/${code}/test`, data);
