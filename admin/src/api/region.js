import request from '@/utils/request';
const cache = new Map();
export function getRegions(parentCode = '') {
    if (!cache.has(parentCode)) {
        cache.set(parentCode, request.get('/open/regions', {
            params: parentCode ? { parent_code: parentCode } : undefined,
        }));
    }
    return cache.get(parentCode);
}
