import request from '@/utils/request';
export function getUserBalanceLogs(userId, params) {
    return request.get(`/admin/users/${userId}/balance-logs`, { params });
}
export function rechargeBalance(userId, data) {
    return request.post(`/admin/users/${userId}/recharge`, data);
}
