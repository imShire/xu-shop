// Package snowflake 封装 bwmarrin/snowflake，worker_id 从 INSTANCE_ID 环境变量读取。
package snowflake

import (
	"log"
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

// initDefault 从环境变量 INSTANCE_ID 读取 worker_id（默认 1，范围 0-1023）后初始化。
// 若值不合法，记 warn 日志并回退到 1。
// 注意：此函数已在 once.Do 内部调用，不能再调用 Init（会重入死锁），直接操作 node。
func initDefault() {
	id := int64(1)
	if v := os.Getenv("INSTANCE_ID"); v != "" {
		if parsed, err := strconv.ParseInt(v, 10, 64); err == nil && parsed >= 0 && parsed <= 1023 {
			id = parsed
		} else {
			log.Printf("[WARN] snowflake: invalid INSTANCE_ID %q (must be integer 0-1023), falling back to 1", v)
		}
	}
	var err error
	node, err = snowflake.NewNode(id)
	if err != nil {
		panic(err)
	}
}

// NextID 返回下一个雪花 ID。
func NextID() int64 {
	once.Do(initDefault)
	return node.Generate().Int64()
}
