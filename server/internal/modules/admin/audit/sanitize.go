// Package audit 内 sanitize 现已迁移到 internal/pkg/sensitive，
// 此处仅保留 re-export 以兼容旧调用方。
package audit

import "github.com/xushop/xu-shop/internal/pkg/sensitive"

// SensitiveKeys 为兼容保留；以 pkg/sensitive 为权威清单。
var SensitiveKeys = sensitive.SensitiveKeys

// Sanitize 转发到 pkg/sensitive.Sanitize。
func Sanitize(in []byte) []byte { return sensitive.Sanitize(in) }
