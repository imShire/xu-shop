import request from '@/utils/request';
export function getArticles(params) {
    return request.get('/admin/articles', { params });
}
export function getArticle(id) {
    return request.get(`/admin/articles/${id}`);
}
export function createArticle(data) {
    return request.post('/admin/articles', data);
}
export function updateArticle(id, data) {
    return request.put(`/admin/articles/${id}`, data);
}
export function deleteArticle(id) {
    return request.delete(`/admin/articles/${id}`);
}
