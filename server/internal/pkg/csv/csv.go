// Package csv 提供防公式注入的 CSV 写入器。
package csv

import (
	"encoding/csv"
	"io"
	"strings"
)

// 可能触发电子表格公式的前缀字符。
const dangerousChars = "=+-@\t\r"

// SafeWriter 防公式注入的 CSV 写入器。
type SafeWriter struct {
	w *csv.Writer
}

// NewSafeWriter 创建 SafeWriter。
func NewSafeWriter(w io.Writer) *SafeWriter {
	return &SafeWriter{w: csv.NewWriter(w)}
}

// Write 写入一行，对危险前缀自动添加 ' 前缀。
func (s *SafeWriter) Write(record []string) error {
	sanitized := make([]string, len(record))
	for i, cell := range record {
		if len(cell) > 0 && strings.ContainsRune(dangerousChars, rune(cell[0])) {
			sanitized[i] = "'" + cell
		} else {
			sanitized[i] = cell
		}
	}
	return s.w.Write(sanitized)
}

// Flush 刷新缓冲区。
func (s *SafeWriter) Flush() {
	s.w.Flush()
}

// Error 返回写入过程中的错误。
func (s *SafeWriter) Error() error {
	return s.w.Error()
}
