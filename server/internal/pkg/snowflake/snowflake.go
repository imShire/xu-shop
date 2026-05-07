// Package snowflake 封装 bwmarrin/snowflake，worker_id 从 INSTANCE_ID 环境变量读取。
package snowflake

import (
	"os"
	"strconv"
	"sync"

	"github.com/bwmarrin/snowflake"
)

var (
	node *snowflake.Node
	once sync.Once
)

// Init 初始化雪花节点，workerID 范围 0-1023。
func Init(workerID int64) {
	once.Do(func() {
		var err error
		node, err = snowflake.NewNode(workerID)
		if err != nil {
			panic(err)
		}
	})
}

// initDefault 从环境变量 INSTANCE_ID 读取 worker_id（默认 1）后初始化。
func initDefault() {
	id := int64(1)
	if v := os.Getenv("INSTANCE_ID"); v != "" {
		if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
			id = parsed
		}
	}
	Init(id)
}

// NextID 返回下一个雪花 ID。
func NextID() int64 {
	if node == nil {
		initDefault()
	}
	return node.Generate().Int64()
}
