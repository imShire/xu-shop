import React from 'react'
import { describe, expect, it, vi } from 'vitest'
import { render, screen } from '@testing-library/react'

// 重型 Taro / NutUI 依赖直接桩掉
vi.mock('@tarojs/components', () => ({
  View: ({ children, ...rest }: { children?: React.ReactNode } & React.HTMLAttributes<HTMLDivElement>) =>
    React.createElement('div', rest, children),
  Text: ({ children, ...rest }: { children?: React.ReactNode } & React.HTMLAttributes<HTMLSpanElement>) =>
    React.createElement('span', rest, children),
}))

vi.mock('@tarojs/taro', () => ({
  default: {
    switchTab: vi.fn(async () => undefined),
    reLaunch: vi.fn(async () => undefined),
  },
}))

vi.mock('@/ui/nutui', () => ({
  Button: ({ children, onClick, ...rest }: { children?: React.ReactNode; onClick?: () => void }) =>
    React.createElement('button', { onClick, ...rest }, children),
}))

const reportSpy = vi.fn()
vi.mock('@/utils/clog', () => ({
  report: (...args: unknown[]) => reportSpy(...args),
}))

// 必须在 mock 之后再 import
import ErrorBoundary from '@/components/ErrorBoundary'

function Bomb(): JSX.Element {
  throw new Error('boom-inside')
}

describe('ErrorBoundary', () => {
  it('子组件抛错时渲染 fallback 并调用 report', () => {
    // 抑制 React 内部 console.error 噪音
    const origErr = console.error
    console.error = vi.fn()
    try {
      render(
        React.createElement(ErrorBoundary as unknown as React.ComponentType<{ children: React.ReactNode }>,
          { children: React.createElement(Bomb) },
        ),
      )
      expect(screen.getByText('页面出了点小问题')).toBeInTheDocument()
      expect(screen.getByText('返回首页')).toBeInTheDocument()
      expect(reportSpy).toHaveBeenCalledTimes(1)
      const [level, msg] = reportSpy.mock.calls[0]
      expect(level).toBe('error')
      expect(String(msg)).toContain('boom-inside')
    } finally {
      console.error = origErr
    }
  })

  it('未抛错时正常渲染子组件', () => {
    render(
      React.createElement(ErrorBoundary as unknown as React.ComponentType<{ children: React.ReactNode }>,
        { children: React.createElement('div', null, 'ok-child') },
      ),
    )
    expect(screen.getByText('ok-child')).toBeInTheDocument()
  })
})
