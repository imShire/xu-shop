import Taro from '@tarojs/taro'
import { request } from '@/services/api'
import { isWeapp, isH5 } from '@/utils/platform'

export interface PrepayParams {
  order_id: string
}

export interface WeappPrepay {
  timeStamp: string
  nonceStr: string
  package: string
  signType: 'MD5' | 'RSA'
  paySign: string
}

export interface H5PrepayResult {
  mweb_url?: string
  jsapi?: WeappPrepay
}

export function createPrepay(params: PrepayParams) {
  return request<WeappPrepay | H5PrepayResult>('/c/pay/wxpay/prepay', {
    method: 'POST',
    auth: true,
    data: params,
  })
}

export function pollOrderStatus(orderId: string) {
  return request<{ status: string }>(`/c/orders/${orderId}/pay-status`, { auth: true })
}

export async function invokePay(prepay: WeappPrepay): Promise<void> {
  if (isWeapp) {
    await Taro.requestPayment({
      timeStamp: prepay.timeStamp,
      nonceStr: prepay.nonceStr,
      package: prepay.package,
      signType: prepay.signType,
      paySign: prepay.paySign,
    })
    return
  }

  if (isH5) {
    const h5Prepay = prepay as unknown as H5PrepayResult
    if (h5Prepay.mweb_url) {
      window.location.href = h5Prepay.mweb_url
    } else if (h5Prepay.jsapi) {
      await invokeJsapi(h5Prepay.jsapi)
    }
  }
}

function invokeJsapi(params: WeappPrepay): Promise<void> {
  return new Promise((resolve, reject) => {
    if (!('WeixinJSBridge' in window)) {
      reject(new Error('请在微信浏览器内打开'))
      return
    }
    const bridge = (window as Record<string, unknown>)['WeixinJSBridge'] as {
      invoke: (api: string, params: unknown, cb: (res: { err_msg: string }) => void) => void
    }
    bridge.invoke(
      'getBrandWCPayRequest',
      {
        appId: '',
        timeStamp: params.timeStamp,
        nonceStr: params.nonceStr,
        package: params.package,
        signType: params.signType,
        paySign: params.paySign,
      },
      (res) => {
        if (res.err_msg === 'get_brand_wcpay_request:ok') {
          resolve()
        } else {
          reject(new Error(res.err_msg))
        }
      }
    )
  })
}
