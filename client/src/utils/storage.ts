import Taro from '@tarojs/taro'

export const storage = {
  get<T>(key: string): T | null {
    try {
      return Taro.getStorageSync(key) as T
    } catch {
      return null
    }
  },
  set<T>(key: string, value: T): void {
    Taro.setStorageSync(key, value)
  },
  remove(key: string): void {
    Taro.removeStorageSync(key)
  },
}
