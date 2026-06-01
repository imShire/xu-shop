import request from '@/utils/request';
// 渠道码
export const getChannelCodeList = (params) => request.get('/admin/channel-codes', { params });
export const createChannelCode = (data) => request.post('/admin/channel-codes', data);
export const updateChannelCode = (id, data) => request.put(`/admin/channel-codes/${id}`, data);
export const deleteChannelCode = (id) => request.delete(`/admin/channel-codes/${id}`);
// 客户标签
export const getTagList = (params) => request.get('/admin/tags', { params });
export const createTag = (data) => request.post('/admin/tags', data);
export const updateTag = (id, data) => request.put(`/admin/tags/${id}`, data);
export const deleteTag = (id) => request.delete(`/admin/tags/${id}`);
