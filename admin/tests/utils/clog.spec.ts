import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { ClogReporter, setReporter } from '@/utils/clog'

function okFetch() {
  return vi.fn(async () => new Response(null, { status: 200 }))
}

beforeEach(() => {
  setReporter(null)
  vi.useRealTimers()
})

afterEach(() => {
  vi.useRealTimers()
})

describe('ClogReporter', () => {
  it('queues entries until threshold', () => {
    const fetchImpl = okFetch()
    const r = new ClogReporter({ fetchImpl })
    r.report('info', 'hello')
    expect(r._peekQueue().length).toBe(1)
    expect(fetchImpl).not.toHaveBeenCalled()
  })

  it('flushes when batch size reached (20)', async () => {
    const fetchImpl = okFetch()
    const r = new ClogReporter({ fetchImpl })
    for (let i = 0; i < 20; i++) r.report('info', `m${i}`)
    // 微任务排空
    await Promise.resolve()
    await Promise.resolve()
    expect(fetchImpl).toHaveBeenCalledTimes(1)
    const call = fetchImpl.mock.calls[0]
    const body = JSON.parse(call[1]!.body as string)
    expect(body.logs.length).toBe(20)
    expect(body.logs[0].source).toBe('admin')
    expect(body.logs[0].release).toBe('test-0')
  })

  it('flushes after 5s interval', async () => {
    vi.useFakeTimers()
    const fetchImpl = okFetch()
    const r = new ClogReporter({ fetchImpl })
    r.report('warn', 'tick')
    expect(fetchImpl).not.toHaveBeenCalled()
    await vi.advanceTimersByTimeAsync(5000)
    expect(fetchImpl).toHaveBeenCalledTimes(1)
  })

  it('truncates message/stack/extra', () => {
    const fetchImpl = okFetch()
    const r = new ClogReporter({ fetchImpl })
    const longMsg = 'x'.repeat(10000)
    const longStack = 'y'.repeat(20000)
    r.report('error', longMsg, { stack: longStack, extra: { huge: 'z'.repeat(20000) } })
    const e = r._peekQueue()[0]
    expect(e.message.length).toBeLessThanOrEqual(4 * 1024 + 32)
    expect(e.stack!.length).toBeLessThanOrEqual(16 * 1024 + 32)
    expect(JSON.stringify(e.extra).length).toBeLessThanOrEqual(8 * 1024 + 64)
  })

  it('retries once on failure then drops', async () => {
    let calls = 0
    const fetchImpl = vi.fn(async () => {
      calls++
      return new Response(null, { status: 500 })
    })
    const r = new ClogReporter({ fetchImpl })
    r.report('warn', 'x')
    await r.flush()
    expect(calls).toBe(2)
    expect(r._peekQueue().length).toBe(0)
  })

  it('sanitizes sensitive fields', () => {
    const fetchImpl = okFetch()
    const r = new ClogReporter({ fetchImpl })
    r.report('warn', 'password=secret123 token=abcdef', {
      extra: { authorization: 'Bearer xxx', other: 'safe' },
    })
    const e = r._peekQueue()[0]
    expect(e.message).not.toContain('secret123')
    expect(e.message).not.toContain('abcdef')
    expect(JSON.stringify(e.extra)).not.toContain('Bearer xxx')
  })

  it('flushSync uses sendBeacon', () => {
    const beaconImpl = vi.fn(() => true)
    const r = new ClogReporter({ fetchImpl: okFetch(), beaconImpl })
    r.report('error', 'boom')
    r.flushSync()
    expect(beaconImpl).toHaveBeenCalledTimes(1)
    expect(beaconImpl.mock.calls[0][0]).toBe('/api/v1/internal/clog/batch')
  })

  it('flushSync falls back to fetch when beacon fails', () => {
    const beaconImpl = vi.fn(() => false)
    const fetchImpl = okFetch()
    const r = new ClogReporter({ fetchImpl, beaconImpl })
    r.report('error', 'boom')
    r.flushSync()
    expect(beaconImpl).toHaveBeenCalledTimes(1)
    expect(fetchImpl).toHaveBeenCalledTimes(1)
  })

  it('uses admin_id getter when provided', () => {
    const fetchImpl = okFetch()
    const r = new ClogReporter({ fetchImpl, getAdminId: () => '12345' })
    r.report('info', 'x')
    expect(r._peekQueue()[0].admin_id).toBe('12345')
  })

  it('never throws even with bad input', () => {
    const r = new ClogReporter({ fetchImpl: okFetch() })
    expect(() => r.report('info', null as unknown as string)).not.toThrow()
  })
})
