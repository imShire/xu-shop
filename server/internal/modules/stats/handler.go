package stats

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/xushop/xu-shop/internal/pkg/errs"
	srv "github.com/xushop/xu-shop/internal/server"
)

// Handler 数据看板 HTTP 处理器。
type Handler struct{ svc *Service }

// NewHandler 构造 Handler。
func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// parseDateRange 从 query 解析 from/to 日期（默认最近 30 天）。
func parseDateRange(c *gin.Context) (from, to time.Time) {
	layout := "2006-01-02"
	to = time.Now()
	from = to.AddDate(0, 0, -30)

	if v := c.Query("from"); v != "" {
		if t, err := time.Parse(layout, v); err == nil {
			from = t
		}
	}
	if v := c.Query("to"); v != "" {
		if t, err := time.Parse(layout, v); err == nil {
			to = t
		}
	}
	return
}

// AdminOverview 总览数据。
func (h *Handler) AdminOverview(c *gin.Context) {
	from, to := parseDateRange(c)
	resp, err := h.svc.GetOverview(c.Request.Context(), from, to)
	if err != nil {
		srv.Fail(c, errs.ErrInternal)
		return
	}
	srv.OK(c, resp)
}

// AdminSalesTrend 销售趋势折线图。
func (h *Handler) AdminSalesTrend(c *gin.Context) {
	from, to := parseDateRange(c)
	points, err := h.svc.GetSalesTrend(c.Request.Context(), from, to)
	if err != nil {
		srv.Fail(c, errs.ErrInternal)
		return
	}
	srv.OK(c, points)
}

// AdminCategoryPie 商品分类占比。
func (h *Handler) AdminCategoryPie(c *gin.Context) {
	from, to := parseDateRange(c)
	data, err := h.svc.GetCategoryPie(c.Request.Context(), from, to)
	if err != nil {
		srv.Fail(c, errs.ErrInternal)
		return
	}
	srv.OK(c, data)
}

// AdminTopProducts 商品销售排行。
func (h *Handler) AdminTopProducts(c *gin.Context) {
	from, to := parseDateRange(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	list, total, err := h.svc.GetTopProducts(c.Request.Context(), from, to, page, size)
	if err != nil {
		srv.Fail(c, errs.ErrInternal)
		return
	}
	srv.OK(c, gin.H{"list": list, "total": total})
}

// AdminUserStats 用户统计。
func (h *Handler) AdminUserStats(c *gin.Context) {
	from, to := parseDateRange(c)
	resp, err := h.svc.GetUserStats(c.Request.Context(), from, to)
	if err != nil {
		srv.Fail(c, errs.ErrInternal)
		return
	}
	srv.OK(c, resp)
}

// AdminChannelStats 渠道统计。
func (h *Handler) AdminChannelStats(c *gin.Context) {
	from, to := parseDateRange(c)
	data, err := h.svc.GetChannelStats(c.Request.Context(), from, to)
	if err != nil {
		srv.Fail(c, errs.ErrInternal)
		return
	}
	srv.OK(c, data)
}

// AdminExportProducts 导出商品销售 CSV。
func (h *Handler) AdminExportProducts(c *gin.Context) {
	from, to := parseDateRange(c)
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=products.csv")
	c.Header("Transfer-Encoding", "chunked")

	if err := h.svc.ExportProducts(c.Request.Context(), c.Writer, from, to); err != nil {
		c.Status(http.StatusInternalServerError)
	}
}

// AdminWorkbench 工作台实时概览。
func (h *Handler) AdminWorkbench(c *gin.Context) {
	resp, err := h.svc.GetWorkbench(c.Request.Context())
	if err != nil {
		srv.Fail(c, errs.ErrInternal)
		return
	}
	srv.OK(c, resp)
}
