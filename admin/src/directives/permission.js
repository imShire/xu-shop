import { useAuthStore } from '@/stores/auth';
export function setupPermissionDirective(app) {
    app.directive('permission', {
        mounted(el, binding) {
            const auth = useAuthStore();
            const perm = binding.value;
            if (!auth.isSuperAdmin && !auth.perms.includes(perm)) {
                el.parentNode?.removeChild(el);
            }
        },
    });
}
