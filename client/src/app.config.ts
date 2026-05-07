export default defineAppConfig({
  pages: [
    'pages/home/index',
    'pages/category/index',
    'pages/cart/index',
    'pages/user/index/index',
  ],
  subPackages: [
    {
      root: 'pages/product',
      pages: ['detail/index', 'list/index'],
    },
    {
      root: 'pages/order',
      pages: ['confirm/index', 'list/index', 'detail/index'],
    },
    {
      root: 'pages/address',
      pages: ['list/index', 'edit/index'],
    },
    {
      root: 'pages/user',
      pages: ['favorite/index', 'history/index', 'settings/index', 'balance/index'],
    },
    {
      root: 'pages/auth',
      pages: [
        'h5-callback/index',
        'login/index',
        'register/index',
        'reset-password/index',
      ],
    },
    {
      root: 'pages/share',
      pages: ['poster/index'],
    },
    {
      root: 'pages/webview',
      pages: ['index'],
    },
  ],
  tabBar: {
    color: '#999999',
    selectedColor: '#e93323',
    backgroundColor: '#ffffff',
    borderStyle: 'black',
    list: [
      {
        pagePath: 'pages/home/index',
        text: '首页',
        iconPath: 'assets/tabbar/home.png',
        selectedIconPath: 'assets/tabbar/home-active.png',
      },
      {
        pagePath: 'pages/category/index',
        text: '分类',
        iconPath: 'assets/tabbar/class.png',
        selectedIconPath: 'assets/tabbar/class-active.png',
      },
      {
        pagePath: 'pages/cart/index',
        text: '购物车',
        iconPath: 'assets/tabbar/cart.png',
        selectedIconPath: 'assets/tabbar/cart-active.png',
      },
      {
        pagePath: 'pages/user/index/index',
        text: '我的',
        iconPath: 'assets/tabbar/user.png',
        selectedIconPath: 'assets/tabbar/user-active.png',
      },
    ],
  },
  window: {
    backgroundTextStyle: 'dark',
    navigationBarBackgroundColor: '#e93323',
    navigationBarTitleText: '徐记小铺',
    navigationBarTextStyle: 'white',
    backgroundColor: '#f5f5f5',
  },
  requiredPrivateInfos: ['chooseAddress'],
})
