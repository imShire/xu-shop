package marketing

import "github.com/xushop/xu-shop/internal/pkg/snowflake"

// snowID 雪花 ID。
func snowID() int64 { return snowflake.NextID() }
