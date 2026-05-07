// Package main 是运维 CLI 入口（cobra）。
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/xushop/xu-shop/internal/config"
	"github.com/xushop/xu-shop/internal/modules/account"
	"github.com/xushop/xu-shop/internal/pkg/snowflake"
	pkgvalidator "github.com/xushop/xu-shop/internal/pkg/validator"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "xu-cli",
		Short: "xu-shop 运维命令行工具",
	}

	rootCmd.AddCommand(adminCmd())
	rootCmd.AddCommand(seedCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// ---- admin 命令组 ----

func adminCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "admin",
		Short: "管理员相关命令",
	}
	cmd.AddCommand(createSuperCmd())
	return cmd
}

func createSuperCmd() *cobra.Command {
	var username, password, realName string

	cmd := &cobra.Command{
		Use:   "create-super",
		Short: "创建超级管理员",
		RunE: func(_ *cobra.Command, _ []string) error {
			// 密码强度校验
			v := pkgvalidator.New()
			if err := v.Var(password, "strongpwd"); err != nil {
				return fmt.Errorf("密码强度不足：需 ≥12 字符，含字母+数字+特殊符号")
			}

			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("加载配置失败: %w", err)
			}

			db, err := gorm.Open(postgres.Open(cfg.DB.DSN), &gorm.Config{})
			if err != nil {
				return fmt.Errorf("连接数据库失败: %w", err)
			}

			snowflake.Init(cfg.App.InstanceID)

			hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
			if err != nil {
				return fmt.Errorf("bcrypt 失败: %w", err)
			}

			admin := &account.Admin{
				ID:           snowflake.NextID(),
				Username:     username,
				PasswordHash: string(hash),
				Status:       "active",
			}
			if realName != "" {
				admin.RealName = &realName
			}

			if err := db.Create(admin).Error; err != nil {
				return fmt.Errorf("创建管理员失败: %w", err)
			}
			fmt.Printf("超级管理员 %s 创建成功，ID=%d\n", username, admin.ID)
			return nil
		},
	}

	cmd.Flags().StringVar(&username, "username", "", "用户名（必填）")
	cmd.Flags().StringVar(&password, "password", "", "密码（必填，强密码）")
	cmd.Flags().StringVar(&realName, "real-name", "", "真实姓名")
	_ = cmd.MarkFlagRequired("username")
	_ = cmd.MarkFlagRequired("password")
	return cmd
}

// ---- seed 命令组 ----

func seedCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "seed",
		Short: "数据初始化命令",
	}
	cmd.AddCommand(seedPermissionsCmd())
	cmd.AddCommand(seedRegionsCmd())
	cmd.AddCommand(seedSuperAdminRoleCmd())
	cmd.AddCommand(seedRolesCmd())
	cmd.AddCommand(seedAllCmd())
	return cmd
}

// allRoles 返回 4 种内置角色定义及其权限 code 列表。
func allRoles() []struct {
	Code  string
	Name  string
	Perms []string
} {
	all := func(perms ...string) []string { return perms }
	return []struct {
		Code  string
		Name  string
		Perms []string
	}{
		{
			Code: "super_admin",
			Name: "超级管理员",
			Perms: all(
				"user.view", "user.disable", "user.create", "user.recharge",
				"product.view", "product.create", "product.edit", "product.delete",
				"category.view", "category.create", "category.edit", "category.delete",
				"inventory.view", "inventory.adjust",
				"order.view", "order.export", "order.remark", "order.cancel",
				"payment.view", "refund.create", "reconcile.view",
				"shipment.view", "shipment.ship", "shipment.batch_ship", "shipment.update",
				"aftersale.view", "aftersale.process",
				"notif.view", "notif.config",
				"channel.view", "channel.create",
				"tag.view", "tag.create", "tag.edit", "tag.delete",
				"banner.view", "banner.edit",
				"nav_icon.view", "nav_icon.edit",
				"decorate.view", "decorate.edit",
				"cms.article.view", "cms.article.edit",
				"stats.view", "stats.export",
				"system.admin.view", "system.admin.create", "system.admin.edit",
				"system.admin.disable", "system.admin.reset_pwd",
				"system.role.view", "system.role.create", "system.role.edit", "system.role.delete",
				"system.audit.view",
				"system.setting.view", "system.setting.edit",
				"system.upload.view", "system.upload.edit",
			),
		},
		{
			Code: "operator",
			Name: "运营",
			Perms: all(
				"user.view", "user.disable", "user.recharge",
				"product.view", "product.create", "product.edit", "product.delete",
				"category.view", "category.create", "category.edit", "category.delete",
				"inventory.view", "inventory.adjust",
				"order.view", "order.export", "order.remark", "order.cancel",
				"payment.view",
				"shipment.view", "shipment.ship", "shipment.batch_ship", "shipment.update",
				"aftersale.view", "aftersale.process",
				"notif.view", "notif.config",
				"channel.view", "channel.create",
				"tag.view", "tag.create", "tag.edit", "tag.delete",
				"banner.view", "banner.edit",
				"nav_icon.view", "nav_icon.edit",
				"decorate.view", "decorate.edit",
				"cms.article.view", "cms.article.edit",
				"stats.view", "stats.export",
			),
		},
		{
			Code: "cs",
			Name: "客服",
			Perms: all(
				"order.view", "order.remark",
				"aftersale.view", "aftersale.process",
				"refund.create",
				"user.view",
			),
		},
		{
			Code: "fulfillment",
			Name: "仓库/打单",
			Perms: all(
				"order.view",
				"shipment.view", "shipment.ship", "shipment.batch_ship", "shipment.update",
			),
		},
	}
}

func seedRolesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "roles",
		Short: "创建 4 种内置角色并绑定权限",
		RunE: func(_ *cobra.Command, _ []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("加载配置失败: %w", err)
			}
			db, err := gorm.Open(postgres.Open(cfg.DB.DSN), &gorm.Config{})
			if err != nil {
				return fmt.Errorf("连接数据库失败: %w", err)
			}
			snowflake.Init(cfg.App.InstanceID)
			return doSeedRoles(db)
		},
	}
}

func doSeedRoles(db *gorm.DB) error {
	for _, rd := range allRoles() {
		// upsert role
		role := &account.Role{}
		db.Where("code = ?", rd.Code).First(role)
		if role.ID == 0 {
			role.ID = snowflake.NextID()
			role.Code = rd.Code
			role.Name = rd.Name
			role.IsSystem = true
			if err := db.Create(role).Error; err != nil {
				return fmt.Errorf("创建角色 %s 失败: %w", rd.Code, err)
			}
			fmt.Printf("  角色 %-20s 创建成功 ID=%d\n", rd.Code, role.ID)
		} else {
			fmt.Printf("  角色 %-20s 已存在   ID=%d\n", rd.Code, role.ID)
		}

		// upsert role_permission bindings
		for _, permCode := range rd.Perms {
			rp := account.RolePermission{RoleID: role.ID, PermissionCode: permCode}
			if err := db.FirstOrCreate(&rp, rp).Error; err != nil {
				return fmt.Errorf("绑定权限 %s -> %s 失败: %w", rd.Code, permCode, err)
			}
		}
		fmt.Printf("    └─ 绑定 %d 个权限点\n", len(rd.Perms))
	}
	return nil
}

func seedAllCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "all",
		Short: "一键初始化：权限点 + 4种内置角色 + 内置行政区划",
		RunE: func(_ *cobra.Command, _ []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("加载配置失败: %w", err)
			}
			db, err := gorm.Open(postgres.Open(cfg.DB.DSN), &gorm.Config{})
			if err != nil {
				return fmt.Errorf("连接数据库失败: %w", err)
			}
			snowflake.Init(cfg.App.InstanceID)

			fmt.Println("==> [1/3] 写入权限点...")
			perms := allPermissions()
			for _, p := range perms {
				if err := db.Save(&p).Error; err != nil {
					return fmt.Errorf("插入权限 %s 失败: %w", p.Code, err)
				}
			}
			fmt.Printf("    写入 %d 个权限点\n", len(perms))

			fmt.Println("==> [2/3] 创建内置角色...")
			if err := doSeedRoles(db); err != nil {
				return err
			}

			fmt.Println("==> [3/3] 导入内置行政区划...")
			regionRows, _, err := loadSeedRegions(nil)
			if err != nil {
				return err
			}
			if err := seedRegionsIntoDB(db, regionRows); err != nil {
				return err
			}
			fmt.Printf("    写入 %d 条行政区划\n", len(regionRows))

			fmt.Println("\n初始化完成。")
			return nil
		},
	}
}

func seedSuperAdminRoleCmd() *cobra.Command {
	var username string
	cmd := &cobra.Command{
		Use:   "super-admin-role",
		Short: "创建 super_admin 角色并绑定到指定管理员",
		RunE: func(_ *cobra.Command, _ []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("加载配置失败: %w", err)
			}
			db, err := gorm.Open(postgres.Open(cfg.DB.DSN), &gorm.Config{})
			if err != nil {
				return fmt.Errorf("连接数据库失败: %w", err)
			}
			snowflake.Init(cfg.App.InstanceID)

			// upsert super_admin role
			role := &account.Role{}
			db.Where("code = ?", "super_admin").First(role)
			if role.ID == 0 {
				role.ID = snowflake.NextID()
				role.Code = "super_admin"
				role.Name = "超级管理员"
				role.IsSystem = true
				if err := db.Create(role).Error; err != nil {
					return fmt.Errorf("创建角色失败: %w", err)
				}
				fmt.Printf("角色 super_admin 创建成功，ID=%d\n", role.ID)
			} else {
				fmt.Printf("角色 super_admin 已存在，ID=%d\n", role.ID)
			}

			// find admin
			var admin account.Admin
			if err := db.Where("username = ? AND deleted_at IS NULL", username).First(&admin).Error; err != nil {
				return fmt.Errorf("管理员 %s 不存在: %w", username, err)
			}

			// upsert admin_role binding
			binding := account.AdminRole{AdminID: admin.ID, RoleID: role.ID}
			res := db.FirstOrCreate(&binding, account.AdminRole{AdminID: admin.ID, RoleID: role.ID})
			if res.Error != nil {
				return fmt.Errorf("绑定角色失败: %w", res.Error)
			}
			fmt.Printf("管理员 %s 已绑定 super_admin 角色\n", username)
			return nil
		},
	}
	cmd.Flags().StringVar(&username, "username", "", "管理员用户名（必填）")
	_ = cmd.MarkFlagRequired("username")
	return cmd
}

func seedPermissionsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "permissions",
		Short: "插入所有权限点",
		RunE: func(_ *cobra.Command, _ []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("加载配置失败: %w", err)
			}
			db, err := gorm.Open(postgres.Open(cfg.DB.DSN), &gorm.Config{})
			if err != nil {
				return fmt.Errorf("连接数据库失败: %w", err)
			}

			perms := allPermissions()
			for _, p := range perms {
				if err := db.Save(&p).Error; err != nil {
					return fmt.Errorf("插入权限 %s 失败: %w", p.Code, err)
				}
			}
			fmt.Printf("成功写入 %d 个权限点\n", len(perms))
			return nil
		},
	}
}

func seedRegionsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "regions [json-path|dir]",
		Short: "导入行政区划数据；默认使用内置预制数据",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("加载配置失败: %w", err)
			}
			db, err := gorm.Open(postgres.Open(cfg.DB.DSN), &gorm.Config{})
			if err != nil {
				return fmt.Errorf("连接数据库失败: %w", err)
			}

			rows, source, err := loadSeedRegions(args)
			if err != nil {
				return err
			}
			if err := seedRegionsIntoDB(db, rows); err != nil {
				return err
			}
			fmt.Printf("成功导入 %d 条行政区划（来源：%s）\n", len(rows), source)
			return nil
		},
	}
}

// allPermissions 返回所有预定义权限点。
func allPermissions() []account.Permission {
	defs := []struct {
		Code   string
		Module string
		Action string
		Name   string
	}{
		{"user.view", "user", "view", "查看用户"},
		{"user.disable", "user", "disable", "禁用用户"},
		{"product.view", "product", "view", "查看商品"},
		{"product.create", "product", "create", "创建商品"},
		{"product.edit", "product", "edit", "编辑商品"},
		{"product.delete", "product", "delete", "删除商品"},
		{"category.view", "category", "view", "查看分类"},
		{"category.create", "category", "create", "创建分类"},
		{"category.edit", "category", "edit", "编辑分类"},
		{"category.delete", "category", "delete", "删除分类"},
		{"inventory.view", "inventory", "view", "查看库存"},
		{"inventory.adjust", "inventory", "adjust", "调整库存"},
		{"order.view", "order", "view", "查看订单"},
		{"order.export", "order", "export", "导出订单"},
		{"order.remark", "order", "remark", "备注订单"},
		{"order.cancel", "order", "cancel", "取消订单"},
		{"payment.view", "payment", "view", "查看支付"},
		{"refund.create", "refund", "create", "发起退款"},
		{"reconcile.view", "reconcile", "view", "查看对账"},
		{"shipment.view", "shipment", "view", "查看物流"},
		{"shipment.ship", "shipment", "ship", "发货"},
		{"shipment.batch_ship", "shipment", "batch_ship", "批量发货"},
		{"shipment.update", "shipment", "update", "更新物流"},
		{"aftersale.view", "aftersale", "view", "查看售后"},
		{"aftersale.process", "aftersale", "process", "处理售后"},
		{"notif.view", "notif", "view", "查看通知"},
		{"notif.config", "notif", "config", "配置通知"},
		{"channel.view", "channel", "view", "查看渠道"},
		{"channel.create", "channel", "create", "创建渠道"},
		{"tag.view", "tag", "view", "查看标签"},
		{"tag.create", "tag", "create", "创建标签"},
		{"tag.edit", "tag", "edit", "编辑标签"},
		{"tag.delete", "tag", "delete", "删除标签"},
		{"banner.view", "banner", "view", "查看Banner"},
		{"banner.edit", "banner", "edit", "编辑Banner"},
		{"nav_icon.view", "nav_icon", "view", "查看金刚区"},
		{"nav_icon.edit", "nav_icon", "edit", "编辑金刚区"},
		{"decorate.view", "decorate", "view", "查看首页装修"},
		{"decorate.edit", "decorate", "edit", "编辑首页装修"},
		{"cms.article.view", "cms", "article.view", "查看文章"},
		{"cms.article.edit", "cms", "article.edit", "编辑文章"},
		{"user.create", "user", "create", "创建用户"},
		{"user.recharge", "user", "recharge", "用户充值"},
		{"stats.view", "stats", "view", "查看统计"},
		{"stats.export", "stats", "export", "导出统计"},
		{"system.admin.view", "system", "admin.view", "查看管理员"},
		{"system.admin.create", "system", "admin.create", "创建管理员"},
		{"system.admin.edit", "system", "admin.edit", "编辑管理员"},
		{"system.admin.disable", "system", "admin.disable", "禁用管理员"},
		{"system.admin.reset_pwd", "system", "admin.reset_pwd", "重置管理员密码"},
		{"system.role.view", "system", "role.view", "查看角色"},
		{"system.role.create", "system", "role.create", "创建角色"},
		{"system.role.edit", "system", "role.edit", "编辑角色"},
		{"system.role.delete", "system", "role.delete", "删除角色"},
		{"system.audit.view", "system", "audit.view", "查看审计日志"},
		{"system.setting.view", "system", "setting.view", "查看系统设置"},
		{"system.setting.edit", "system", "setting.edit", "编辑系统设置"},
		{"system.upload.view", "system", "upload.view", "查看上传设置"},
		{"system.upload.edit", "system", "upload.edit", "编辑上传设置"},
	}

	perms := make([]account.Permission, len(defs))
	for i, d := range defs {
		parts := strings.SplitN(d.Code, ".", 2)
		module := parts[0]
		action := d.Code
		if len(parts) == 2 {
			action = parts[1]
		}
		perms[i] = account.Permission{
			Code:   d.Code,
			Module: module,
			Action: action,
			Name:   d.Name,
		}
	}
	return perms
}
