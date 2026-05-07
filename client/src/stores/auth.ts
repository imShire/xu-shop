import { create } from 'zustand'
import { clearAccessToken, setAccessToken } from '@/services/api'
import { getMe, logout } from '@/services/auth'
import type { User } from '@/types/biz'
import { persistShareUserId } from '@/utils/share-attr'

interface AuthState {
  user: User | null
  isLoggedIn: boolean
  isHydrating: boolean
  hydrate: (scene?: string) => Promise<void>
  setAuth: (token: string, user: User) => void
  updateUser: (partial: Partial<User>) => void
  logout: () => Promise<void>
}

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  isLoggedIn: false,
  isHydrating: false,
  async hydrate(scene) {
    if (scene) {
      persistShareUserId(scene)
    }

    set({ isHydrating: true })
    try {
      const user = await getMe()
      // getMe 未登录时返回 null（200 + data:null），已登录返回 User
      if (user) {
        set({ user, isLoggedIn: true })
      } else {
        clearAccessToken()
        set({ user: null, isLoggedIn: false })
      }
    } catch {
      clearAccessToken()
      set({ user: null, isLoggedIn: false })
    } finally {
      set({ isHydrating: false })
    }
  },
  setAuth(token: string, user: User) {
    setAccessToken(token)
    set({ user, isLoggedIn: true })
  },
  updateUser(partial: Partial<User>) {
    set((state) => ({ user: state.user ? { ...state.user, ...partial } : state.user }))
  },
  async logout() {
    await logout()
    clearAccessToken()
    set({ user: null, isLoggedIn: false })
  },
}))
