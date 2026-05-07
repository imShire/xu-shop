import type { UserConfigExport } from "@tarojs/cli"

const isH5Build = process.env.TARO_ENV === 'h5'

export default {
  logger: {
    quiet: false,
    stats: true
  },
  // H5 dev 使用相对路径，走 devServer proxy；weapp 直连后端
  defineConstants: {
    'process.env.TARO_APP_API_BASE': JSON.stringify(
      isH5Build ? '/api/v1' : 'http://localhost:8080/api/v1'
    ),
  },
  mini: {},
  h5: {
    devServer: {
      host: '0.0.0.0',
      port: Number.parseInt(process.env.TARO_H5_PORT ?? '10086', 10),
      proxy: {
        '/api': {
          target: 'http://localhost:8080',
          changeOrigin: true,
        },
      },
    },
  }
} satisfies UserConfigExport<'webpack5'>
