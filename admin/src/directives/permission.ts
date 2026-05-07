import type { App, DirectiveBinding } from 'vue'
import { useAuthStore } from '@/stores/auth'

export function setupPermissionDirective(app: App) {
  app.directive('permission', {
    mounted(el: HTMLElement, binding: DirectiveBinding) {
      const auth = useAuthStore()
      const perm = binding.value as string
      if (!auth.isSuperAdmin && !auth.perms.includes(perm)) {
        el.parentNode?.removeChild(el)
      }
    },
  })
}
