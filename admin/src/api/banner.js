import request from '@/utils/request';
export const getBanners = () => request.get('/admin/banners');
export const createBanner = (data) => request.post('/admin/banners', data);
export const updateBanner = (id, data) => request.put(`/admin/banners/${id}`, data);
export const deleteBanner = (id) => request.delete(`/admin/banners/${id}`);
export const toggleBanner = (id) => request.patch(`/admin/banners/${id}/toggle`);
export const sortBanners = (items) => request.patch('/admin/banners/sort', { items });
