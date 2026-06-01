package reconciliation

import "time"

// bizLoc 业务自然日时区。
//
// 对账作业以 Asia/Shanghai 自然日为切分边界（私域电商面向国内商户）。
// 若运行环境缺少 zoneinfo（极少数 minimal 镜像），则 fallback 到 UTC+8 固定偏移。
var bizLoc = func() *time.Location {
	if loc, err := time.LoadLocation("Asia/Shanghai"); err == nil {
		return loc
	}
	return time.FixedZone("CST", 8*3600)
}()

// BizLocation 暴露给同包调用。
func BizLocation() *time.Location { return bizLoc }
