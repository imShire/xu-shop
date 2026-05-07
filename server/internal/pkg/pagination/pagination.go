// Package pagination 提供通用分页请求/响应结构。
package pagination

// Req 分页查询参数。
type Req struct {
	Page     int `form:"page"      binding:"min=1"`
	PageSize int `form:"page_size" binding:"min=1,max=100"`
}

// Offset 计算 SQL OFFSET；Page/PageSize 无效时使用默认值。
func (r *Req) Offset() int {
	if r.Page < 1 {
		r.Page = 1
	}
	if r.PageSize < 1 {
		r.PageSize = 20
	}
	return (r.Page - 1) * r.PageSize
}

// Limit 返回 LIMIT 值，PageSize 无效时返回 20。
func (r *Req) Limit() int {
	if r.PageSize < 1 {
		return 20
	}
	return r.PageSize
}

// DefaultReq 返回默认分页参数（第 1 页，每页 20）。
func DefaultReq() Req {
	return Req{Page: 1, PageSize: 20}
}

// Resp 泛型分页响应。
type Resp[T any] struct {
	List     []T   `json:"list"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
	Total    int64 `json:"total"`
}
