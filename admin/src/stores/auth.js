import { defineStore } from 'pinia';
export const useAuthStore = defineStore('auth', {
    state: () => ({
        token: localStorage.getItem('admin_token') || '',
        csrfToken: localStorage.getItem('admin_csrf') || '',
        user: null,
        roles: [],
        perms: [],
        ready: false,
    }),
    getters: {
        isLoggedIn: (state) => !!state.token,
        isSuperAdmin: (state) => state.roles.includes('super_admin'),
    },
    actions: {
        async init() {
            if (this.token) {
                try {
                    const { getAdminMe } = await import('@/api/account');
                    const user = await getAdminMe();
                    this.user = user;
                    this.roles = user.roles || [];
                    this.perms = user.perms || [];
                }
                catch {
                    this.logout();
                }
            }
            this.ready = true;
        },
        async login(form) {
            const { adminLogin } = await import('@/api/account');
            const res = await adminLogin(form);
            this.token = res.access_token;
            this.csrfToken = res.csrf_token || '';
            localStorage.setItem('admin_token', this.token);
            localStorage.setItem('admin_csrf', this.csrfToken);
            await this.init();
        },
        logout() {
            this.token = '';
            this.csrfToken = '';
            this.user = null;
            this.roles = [];
            this.perms = [];
            this.ready = false;
            localStorage.removeItem('admin_token');
            localStorage.removeItem('admin_csrf');
        },
    },
});
