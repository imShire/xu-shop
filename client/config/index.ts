import path from 'node:path'
import { defineConfig, type UserConfigExport } from '@tarojs/cli'
import TsconfigPathsPlugin from 'tsconfig-paths-webpack-plugin'
import devConfig from './dev'
import prodConfig from './prod'

const srcPath = path.resolve(__dirname, '..', 'src')

export default defineConfig<'webpack5'>(async (merge) => {
  const baseConfig: UserConfigExport<'webpack5'> = {
    projectName: 'xu-shop-client',
    date: '2026-4-30',
    designWidth: 750,
    deviceRatio: {
      640: 2.34 / 2,
      750: 1,
      375: 2,
      828: 1.81 / 2
    },
    sourceRoot: 'src',
    outputRoot: 'dist',
    alias: {
      '@': srcPath,
    },
    plugins: ['@tarojs/plugin-html'],
    defineConstants: {
      'process.env.TARO_APP_API_BASE': JSON.stringify(process.env.TARO_APP_API_BASE ?? ''),
    },
    copy: {
      patterns: [],
      options: {},
    },
    framework: 'react',
    compiler: 'webpack5',
    cache: {
      enable: true,
    },
    mini: {
      miniCssExtractPluginOption: {
        ignoreOrder: true,
      },
      postcss: {
        pxtransform: {
          enable: true,
          config: {},
        },
        cssModules: {
          enable: false,
          config: {
            namingPattern: 'module',
            generateScopedName: '[name]__[local]___[hash:base64:5]',
          },
        },
      },
      webpackChain(chain) {
        chain.resolve.plugin('tsconfig-paths').use(TsconfigPathsPlugin)
      },
    },
    h5: {
      router: {
        mode: 'browser',
      },
      publicPath: '/',
      staticDirectory: 'static',
      output: {
        filename: 'js/[name].[hash:8].js',
        chunkFilename: 'js/[name].[chunkhash:8].js',
      },
      miniCssExtractPluginOption: {
        ignoreOrder: true,
        filename: 'css/[name].[hash].css',
        chunkFilename: 'css/[name].[chunkhash].css',
      },
      postcss: {
        autoprefixer: {
          enable: true,
          config: {},
        },
        cssModules: {
          enable: false,
          config: {
            namingPattern: 'module',
            generateScopedName: '[name]__[local]___[hash:base64:5]',
          },
        },
      },
      webpackChain(chain) {
        chain.resolve.plugin('tsconfig-paths').use(TsconfigPathsPlugin)
      },
    },
  }

  if (process.env.NODE_ENV === 'development') {
    return merge({}, baseConfig, devConfig)
  }

  return merge({}, baseConfig, prodConfig)
})
