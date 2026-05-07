# views/login

> 阶段 1。对应 arch 01。

## Login.vue
- 表单：username / password / captcha
- 拉 `/admin/auth/captcha` → 显示 base64 图
- 提交：POST `/admin/auth/login`，写 auth store，跳 `/workbench`
- "记住我"勾选 → remember=true
- 失败展示后端 message；锁定提示 15 分钟后再试
