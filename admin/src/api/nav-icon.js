import request from '@/utils/request';
export const getNavIcons = () => request.get('/admin/nav-icons');
export const createNavIcon = (data) => request.post('/admin/nav-icons', data);
export const updateNavIcon = (id, data) => request.put(`/admin/nav-icons/${id}`, data);
export const deleteNavIcon = (id) => request.delete(`/admin/nav-icons/${id}`);
export const toggleNavIcon = (id) => request.patch(`/admin/nav-icons/${id}/toggle`);
export const sortNavIcons = (items) => request.patch('/admin/nav-icons/sort', { items });
