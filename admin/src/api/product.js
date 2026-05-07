import request from '@/utils/request';
// 分类
export const getCategoryList = () => request.get('/admin/categories');
export const createCategory = (data) => request.post('/admin/categories', data);
export const updateCategory = (id, data) => request.put(`/admin/categories/${id}`, data);
export const deleteCategory = (id) => request.delete(`/admin/categories/${id}`);
// 商品
export const getProductList = (params) => request.get('/admin/products', { params });
export const getProductDetail = (id) => request.get(`/admin/products/${id}`);
export const getProduct = (id) => request.get(`/admin/products/${id}`);
export const createProduct = (data) => request.post('/admin/products', data);
export const updateProduct = (id, data) => request.put(`/admin/products/${id}`, data);
export const deleteProduct = (id) => request.delete(`/admin/products/${id}`);
export const putOnSale = (id) => request.post(`/admin/products/${id}/onsale`);
export const putOffSale = (id) => request.post(`/admin/products/${id}/offsale`);
export const copyProduct = (id) => request.post(`/admin/products/${id}/copy`);
export const batchOnSale = (ids) => request.post('/admin/products/batch-status', { ids, status: 'onsale' });
export const batchOffSale = (ids) => request.post('/admin/products/batch-status', { ids, status: 'offsale' });
// SKU
export const batchUpdateSkuPrice = (productId, data) => request.put(`/admin/products/${productId}/skus/batch-price`, data);
// 运费模板（供商品编辑页选择）
export const getFreightTemplates = () => request.get('/admin/freight-templates');
// 图片上传
export const uploadImage = (file) => {
    const form = new FormData();
    form.append('file', file);
    return request.post('/admin/upload/image', form, {
        headers: { 'Content-Type': 'multipart/form-data' },
    });
};
