import request from '@/utils/request';
export const getUserList = (params) => request.get('/admin/users', { params });
export const createUser = (data) => request.post('/admin/users', data);
export const disableUser = (id) => request.post(`/admin/users/${id}/disable`);
export const enableUser = (id) => request.post(`/admin/users/${id}/enable`);
export const getUserDetail = (id) => request.get(`/admin/users/${id}`);
