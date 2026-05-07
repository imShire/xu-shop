// Package types 提供通用自定义类型。
package types

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// Int64Str 序列化为 JSON 字符串，避免 JS 精度丢失（JS Number.MAX_SAFE_INTEGER = 2^53-1）。
// 反序列化时同时支持 JSON 字符串和 JSON 数字。
type Int64Str int64

func (v Int64Str) MarshalJSON() ([]byte, error) {
	return json.Marshal(strconv.FormatInt(int64(v), 10))
}

func (v *Int64Str) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err == nil {
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return err
		}
		*v = Int64Str(n)
		return nil
	}
	var n int64
	if err := json.Unmarshal(b, &n); err != nil {
		return err
	}
	*v = Int64Str(n)
	return nil
}

// Int64 返回底层 int64 值。
func (v Int64Str) Int64() int64 { return int64(v) }

// String 返回十进制字符串。
func (v Int64Str) String() string { return strconv.FormatInt(int64(v), 10) }

func (v Int64Str) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *Int64Str) UnmarshalText(b []byte) error {
	n, err := strconv.ParseInt(string(b), 10, 64)
	if err != nil {
		return fmt.Errorf("Int64Str: invalid value %q", b)
	}
	*v = Int64Str(n)
	return nil
}
