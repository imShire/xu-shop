import Taro from '@tarojs/taro'
import { create } from 'zustand'
import { getCartCount } from '@/services/cart'
import { useAuthStore } from '@/stores/auth'

interface CartState {
  count: number
  refreshCount: () => Promise<void>
}

export const useCartStore = create<CartState>((set) => ({
  count: 0,
  async refreshCount() {
    if (!useAuthStore.getState().isLoggedIn) {
      set({ count: 0 })
      void Taro.removeTabBarBadge({ index: 2 }).catch(() => {})
      return
    }

    try {
      const result = await getCartCount()
      const count = result.count
      set({ count })
      if (count > 0) {
        void Taro.setTabBarBadge({ index: 2, text: String(count) }).catch(() => {})
      } else {
        void Taro.removeTabBarBadge({ index: 2 }).catch(() => {})
      }
    } catch {
      set({ count: 0 })
    }
  },
}))
