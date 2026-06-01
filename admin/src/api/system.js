import request from '@/utils/request';
export const getAuditLogs = (params) => request.get('/admin/audit-logs', { params });
export const getUploadSettings = () => request.get('/admin/settings/upload');
export const updateUploadSettings = (data) => request.put('/admin/settings/upload', data);
export const testUploadSettings = (data) => request.post('/admin/settings/upload/test', data);
export const probeUploadSettings = (file) => {
    const form = new FormData();
    form.append('file', file);
    return request.post('/admin/settings/upload/probe', form, {
        headers: { 'Content-Type': 'multipart/form-data' },
    });
};
export const getSettings = (group) => request.get(`/admin/settings/${group}`);
export const updateSettings = (group, data) => request.put(`/admin/settings/${group}`, data);
