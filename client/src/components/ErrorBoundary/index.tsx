import { Component, ErrorInfo, ReactNode } from 'react'
import { View, Text } from '@tarojs/components'
import Taro from '@tarojs/taro'
import { Button } from '@/ui/nutui'

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
    // Sentry hook 占位
    console.error('[ErrorBoundary]', error, info.componentStack)
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
      </View>
    )
  }
}
