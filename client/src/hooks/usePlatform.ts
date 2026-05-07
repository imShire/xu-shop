import { isH5, isWeapp } from '@/utils/platform'

export function usePlatform() {
  return { isH5, isWeapp }
}
