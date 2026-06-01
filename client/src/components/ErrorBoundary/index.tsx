import { Component, ErrorInfo, ReactNode } from 'react'
import { View, Text } from '@tarojs/components'
import Taro from '@tarojs/taro'
import { Button } from '@/ui/nutui'
import { report } from '@/utils/clog'

interface Props {
  children: ReactNode
}

interface State {
  hasError: boolean
}

export default class ErrorBoundary extends Component<Props, State> {
  state: State = { hasError: false }

  static getDerivedStateFromError(): State {
    return { hasError: true }
  }

  componentDidCatch(error: Error, info: ErrorInfo) {
    console.error('[ErrorBoundary]', error, info.componentStack)
    try {
      report('error', `react_render: ${error?.message ?? String(error)}`, {
        stack: [error?.stack, info?.componentStack].filter(Boolean).join('\n\n'),
      })
    } catch {
      // 静默
    }
  }

  handleRetry = () => {
    this.setState({ hasError: false })
  }

  handleBackHome = () => {
    this.setState({ hasError: false })
    void Taro.switchTab({ url: '/pages/home/index' }).catch(() => {
      void Taro.reLaunch({ url: '/pages/home/index' })
    })
  }

  render() {
    if (!this.state.hasError) {
      return this.props.children
    }

    return (
      <View
        style={{
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          justifyContent: 'center',
          minHeight: '600rpx',
          padding: '48px 24px',
          textAlign: 'center',
        }}
      >
        <Text style={{ fontSize: '32px', marginBottom: '16px' }}>😵</Text>
        <Text style={{ fontSize: '17px', color: '#333', marginBottom: '8px' }}>
          页面出了点小问题
        </Text>
        <Text style={{ fontSize: '13px', color: '#999', marginBottom: '32px' }}>
          请稍后再试，或返回首页继续浏览
        </Text>
        <Button type='primary' onClick={this.handleBackHome}>
          返回首页
        </Button>
        <View style={{ height: '12px' }} />
        <Button onClick={this.handleRetry}>重试</Button>
      </View>
    )
  }
}
