# views/system

> 阶段 1（基础部分） + 阶段 5（FreightTemplate）。对应 arch 10。

## 文件
- `AdminList.vue`：员工账号 CRUD + 重置密码 + 停用
- `RoleList.vue`：角色 + 权限点勾选
- `AuditLog.vue`：操作日志查询（含 admin_username 快照）
- `SystemSetting.vue`：分组参数（基础/支付/物流/私域/消息）；敏感字段编辑需二次验证
- `FreightTemplate.vue`：运费模板（默认唯一）
