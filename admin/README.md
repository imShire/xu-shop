# admin — Vue3 + Element Plus 后台

> 对应文档：`docs/arch/10-admin.md`

## 实施阶段
阶段 1 起骨架（登录 + 布局），阶段 2-5 各业务模块逐步填充。

## 目录结构

```
admin/src/
├── main.ts
├── App.vue
├── router/
│   ├── index.ts                # vue-router 实例
│   ├── routes.static.ts        # 公共路由（login / 404 / 403）
│   ├── routes.dynamic.ts       # 业务路由全集（含 meta.perms）
│   └── permission.ts           # 守卫：未登录跳 login + 按 perm 过滤
├── stores/
│   ├── auth.ts                 # token / user / roles / perms / csrf
│   ├── permission.ts           # 已注入的动态路由
│   ├── app.ts                  # 折叠 / 主题
│   └── dict.ts                 # 字典缓存
├── api/
│   ├── http.ts                 # axios 实例 + 拦截器（注入 Bearer + X-CSRF-Token + X-Request-Id；统一错误 toast；401 跳 login）
│   ├── account.ts              # /admin/auth/* /admin/me /admin/admins /admin/roles /admin/permissions
│   ├── product.ts category.ts inventory.ts cart.ts
│   ├── order.ts payment.ts shipping.ts aftersale.ts
│   ├── notification.ts privateDomain.ts stats.ts
│   ├── upload.ts               # STS 凭证 + confirm
│   └── system.ts               # audit / settings / freight templates
├── views/
│   ├── login/Login.vue
│   ├── workbench/Workbench.vue
│   ├── product/{ProductList,ProductEdit,CategoryTree}.vue
│   ├── inventory/{AlertList,LogList}.vue
│   ├── order/{OrderList,OrderDetail}.vue + components/{StatusTag,Timeline,Remarks}.vue
│   ├── aftersale/AftersaleList.vue
│   ├── shipping/{PendingList,ShippedList,SenderAddress,CarrierConfig,BatchShipDialog}.vue
│   ├── user/{UserList,UserDetail}.vue
│   ├── private-domain/{ChannelCodeList,TagList}.vue
│   ├── stats/{SalesOverview,ProductSales,ChannelStats,UserStats}.vue
│   ├── system/{AdminList,RoleList,AuditLog,SystemSetting,FreightTemplate}.vue
│   └── error/{NotFound,Forbidden}.vue
├── layouts/
│   ├── BasicLayout.vue
│   └── components/{Sidebar,Header,Breadcrumb,Tags}.vue
├── components/
│   ├── ProTable/               # el-table 二次封装：分页 / 刷新 / 列设置 / 导出
│   ├── ProForm/
│   ├── UploadImage/            # 走 STS 直传 OSS（v1.1 #22）
│   ├── PriceInput/             # 元 ↔ 分
│   └── PermButton/             # v-permission 简写
├── directives/
│   └── permission.ts           # v-permission="'order.ship'"
├── composables/
│   ├── useTable.ts useExport.ts useDict.ts usePermission.ts
├── utils/
│   ├── request.ts errorCode.ts dayjs.ts price.ts download.ts
├── types/
│   └── api.d.ts                # openapi-typescript 生成（CI 自动）
└── styles/
    ├── index.scss
    └── element-overrides.scss  # 主色覆盖（拒绝默认蓝）
```

## 主题
按 arch 10 + 80 约定：主色 `#0F4C36`，强调色 `#F08C44`。在 `element-overrides.scss` 通过 SCSS 变量覆盖 Element Plus。

## 模块顺序
1. 阶段 1：login + BasicLayout + auth store + permission 守卫 + http + system/AdminList + RoleList + AuditLog
2. 阶段 2：product / category / inventory + UploadImage 组件
3. 阶段 3：order list + detail + freight template
4. 阶段 4：payment / aftersale / shipping
5. 阶段 5：notification / private-domain / stats / workbench

## 启动
```
pnpm install
cp .env.example .env.local        # 配 VITE_API_BASE
pnpm dev
```
