# views/product

> 阶段 2。对应 arch 03。

## 文件
- `ProductList.vue`：ProTable，按状态/分类/关键字筛选；行内"上下架/复制/编辑/删除"
- `ProductEdit.vue`：创建 + 编辑共用；分基础信息/规格 SKU/详情富文本/标签 4 个 Tab；wangEditor 富文本；UploadImage 走 STS 直传
- `CategoryTree.vue`：左右两栏，左 el-tree 展示分类树（两级），右编辑区
- `components/SkuMatrix.vue`：规格维度配置 + 自动笛卡尔积 + 批量改价 / 库存
- `components/BatchPriceDialog.vue`、`components/BatchStockDialog.vue`
