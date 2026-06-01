import request from '@/utils/request';
export const tagCategoryLabels = {
    rfm: 'RFM',
    lifecycle: '生命周期',
    category_pref: '品类偏好',
    price_band: '价格带',
    source: '来源',
    business: '业务',
    member: '会员',
    system: '系统',
};
export const listUserTags = (params) => request.get('/admin/user-tags', { params });
export const createUserTag = (data) => request.post('/admin/user-tags', data);
export const updateUserTag = (code, data) => request.put(`/admin/user-tags/${code}`, data);
export const deleteUserTag = (code) => request.delete(`/admin/user-tags/${code}`);
export const getUserTags = (userId) => request.get(`/admin/users/${userId}/tags`);
export const grantUserTag = (userId, data) => request.post(`/admin/users/${userId}/tags`, data);
export const revokeUserTag = (userId, tagCode) => request.delete(`/admin/users/${userId}/tags/${tagCode}`);
export const previewAudience = (filter) => request.post('/admin/audience/preview', { filter });
// 用户搜索（用于"按 user_id/手机号"查标签）
export const searchUserBrief = (params) => request.get('/admin/users', { params });
