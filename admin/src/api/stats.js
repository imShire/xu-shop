import request from '@/utils/request';
export const getStatsOverview = (params) => request.get('/admin/stats/overview', { params });
export const getStatsTrend = (params) => request.get('/admin/stats/sales-trend', { params });
export const getCategoryStats = (params) => request.get('/admin/stats/category-pie', { params });
export const getProductRanking = (params) => request.get('/admin/stats/products', { params });
export const getUserStats = (params) => request.get('/admin/stats/users', { params });
export const getChannelStats = (params) => request.get('/admin/stats/channels', { params });
export const exportStats = (params) => request.get('/admin/stats/products/export', { params, responseType: 'blob' });
export const getWorkbenchStats = () => request.get('/admin/stats/workbench');
