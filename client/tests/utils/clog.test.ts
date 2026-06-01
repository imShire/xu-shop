import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

vi.mock('@tarojs/taro', () => ({
  default: {
    getCurrentInstance: () => ({ router: { path: '/pages/home/index' } }),
    getSystemInfoSync: () => ({ brand: 'devtools', model: 'sim', system: 'iOS 17', platform: 'ios', version: '8.0' }),
    request: vi.fn(),
  },
}))

import { ClogReporter } from '@/utils/clog'

describe('ClogReporter', () => {
  beforeEach(() => {
    vi.useFakeTimers()
  })
  afterEach(() => {
    vi.useRealTimers()
    vi.restoreAllMocks()
  })

  function makeFetch(status = 200) {
    return vi.fn(async () =>
      new Response('{}', { status, headers: { 'Content-Type': 'application/json' } }),
    )
  }

  it('入队后未达批量阈值，不立即发送；满 20 条触发 flush', async () => {
    const fetchImpl = makeFetch(200)
    const r = new ClogReporter({ platform: 'client_h5', fetchImpl, endpoint: '/api/v1/internal/clog/batch' })
    for (let i = 0; i < 19; i++) {
      r.report('info', `m${i}`)
    }
    expect(fetchImpl).not.toHaveBeenCalled()
    expect(r._peekQueue().length).toBe(19)
    r.report('info', 'last')
    // 触发 flush，是异步的
    await vi.runOnlyPendingTimersAsync()
    await Promise.resolve()
    expect(fetchImpl).toHaveBeenCalledTimes(1)
  })

  it('定时器 5s 后自动 flush', async () => {
    const fetchImpl = makeFetch(200)
    const r = new ClogReporter({ platform: 'client_h5', fetchImpl })
    r.report('warn', 'tick')
    expect(fetchImpl).not.toHaveBeenCalled()
    vi.advanceTimersByTime(5000)
    await Promise.resolve()
    await Promise.resolve()
    expect(fetchImpl).toHaveBeenCalledTimes(1)
  })

  it('message 超长被截断', () => {
    const r = new ClogReporter({ platform: 'client_h5', fetchImpl: makeFetch() })
    const big = 'a'.repeat(10000)
    r.report('error', big)
    const queued = r._peekQueue()
    expect(queued[0].message.length).toBeLessThanOrEqual(4 * 1024 + 32)
    expect(queued[0].message.endsWith('...[truncated]')).toBe(true)
  })

  it('失败重试一次后丢弃，不抛错', async () => {
    const fetchImpl = vi.fn(async () => new Response('err', { status: 500 }))
    const r = new ClogReporter({ platform: 'client_h5', fetchImpl })
    r.report('error', 'boom')
    await r.flush()
    // 一次原始 + 一次重试
    expect(fetchImpl).toHaveBeenCalledTimes(2)
    // 队列已清空
    expect(r._peekQueue().length).toBe(0)
  })

  it('extra 含敏感字段会脱敏', async () => {
    const fetchImpl = makeFetch()
    const r = new ClogReporter({ platform: 'client_h5', fetchImpl })
    r.report('warn', 'API_FAIL', { extra: { password: 'p4ssw0rd', token: 'abc.def.ghi', q: 'ok' } })
    const queued = r._peekQueue()
    const serialized = JSON.stringify(queued[0].extra)
    expect(serialized).not.toContain('p4ssw0rd')
    expect(serialized).not.toContain('abc.def.ghi')
    expect(serialized).toContain('ok')
  })

  it('flushSync 走 sendBeacon', () => {
    const beacon = vi.fn(() => true)
    const r = new ClogReporter({ platform: 'client_h5', fetchImpl: makeFetch(), beaconImpl: beacon })
    r.report('info', 'bye')
    r.flushSync()
    expect(beacon).toHaveBeenCalledTimes(1)
    expect(r._peekQueue().length).toBe(0)
  })

  it('report 本身不抛错（即便 fetch 抛）', async () => {
    const fetchImpl = vi.fn(async () => {
      throw new Error('network down')
    })
    const r = new ClogReporter({ platform: 'client_h5', fetchImpl })
    expect(() => r.report('error', 'x')).not.toThrow()
    await r.flush()
    expect(fetchImpl).toHaveBeenCalledTimes(2)
  })

  it('user_id 通过 getUserId 注入', () => {
    const r = new ClogReporter({
      platform: 'client_h5',
      fetchImpl: makeFetch(),
      getUserId: () => '12345',
    })
    r.report('info', 'x')
    expect(r._peekQueue()[0].user_id).toBe('12345')
  })

  it('weapp 平台用 requestImpl', async () => {
    const requestImpl = vi.fn(async (_opts: { url: string; method: 'POST'; data: string; header: Record<string, string> }) => ({ statusCode: 200 }))
    const r = new ClogReporter({ platform: 'client_weapp', requestImpl })
    r.report('info', 'm')
    await r.flush()
    expect(requestImpl).toHaveBeenCalledTimes(1)
    const firstCall = requestImpl.mock.calls[0]
    expect(firstCall?.[0]?.method).toBe('POST')
  })
})
