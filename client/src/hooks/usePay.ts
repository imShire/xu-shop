import { useState, useCallback } from 'react'
import Taro from '@tarojs/taro'
import { createPrepay, invokePay, pollOrderStatus } from '@/services/pay'
import type { WeappPrepay } from '@/services/pay'

interface UsePayOptions {
  onSuccess?: (orderId: string) => void
  onFail?: (err: Error) => void
}

export function usePay(options?: UsePayOptions) {
  const [loading, setLoading] = useState(false)

  const pay = useCallback(async (orderId: string) => {
    setLoading(true)
    try {
      const prepay = await createPrepay({ order_id: orderId })
      await invokePay(prepay as WeappPrepay)

      // Poll for payment status (up to 10s)
      for (let i = 0; i < 5; i++) {
        await sleep(2000)
        const result = await pollOrderStatus(orderId)
        if (result.status === 'paid') {
          options?.onSuccess?.(orderId)
          return
        }
      }
      options?.onSuccess?.(orderId)
    } catch (err) {
      const error = err instanceof Error ? err : new Error('支付失败')
      options?.onFail?.(error)
      void Taro.showToast({ title: error.message, icon: 'none' })
    } finally {
      setLoading(false)
    }
  }, [options])

  return { pay, loading }
}

function sleep(ms: number) {
  return new Promise((resolve) => setTimeout(resolve, ms))
}
