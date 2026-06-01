import request from '@/utils/request';
export const listRecallCampaigns = (params) => request.get('/admin/recall-campaigns', { params });
export const getRecallCampaign = (id) => request.get(`/admin/recall-campaigns/${id}`);
export const createRecallCampaign = (data) => request.post('/admin/recall-campaigns', data);
export const updateRecallCampaign = (id, data) => request.put(`/admin/recall-campaigns/${id}`, data);
export const onlineRecallCampaign = (id) => request.post(`/admin/recall-campaigns/${id}/online`);
export const pauseRecallCampaign = (id) => request.post(`/admin/recall-campaigns/${id}/pause`);
export const closeRecallCampaign = (id) => request.post(`/admin/recall-campaigns/${id}/close`);
export const getRecallFunnel = (id, params) => request.get(`/admin/recall-campaigns/${id}/funnel`, { params });
export const listRecallLogs = (params) => request.get('/admin/recall-logs', { params });
// 测试触达 — 复用召回日志 / 召回活动接口；如果有专属接口可以再加
export const testSendRecall = (id, user_id) => request.post(`/admin/recall-campaigns/${id}/test-send`, { user_id });
