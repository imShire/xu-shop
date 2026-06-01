import request from '@/utils/request';
export function getPageVersions(pageKey = 'home') {
    return request.get('/admin/decorate/versions', {
        params: { page_key: pageKey },
    });
}
export function savePageConfig(data) {
    return request.post('/admin/decorate/save', data);
}
export function activatePageConfig(id, pageKey = 'home') {
    return request.post(`/admin/decorate/activate/${id}`, {}, {
        params: { page_key: pageKey },
    });
}
