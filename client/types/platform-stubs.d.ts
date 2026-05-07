declare module 'react-native' {
  const ReactNative: Record<string, unknown>
  export = ReactNative
}

declare module 'vue' {
  export interface App<Element = unknown> {
    mount: (rootContainer: Element | string) => void
  }
}

declare module '@tarojs/taro-rn/types/overlay' {}
