// Package stock 封装 Redis Lua 原子库存操作。
package stock

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// LockItem 单个 SKU 锁定条目。
type LockItem struct {
	SKUID int64 `json:"sku_id"`
	Qty   int   `json:"qty"`
}

// Client Redis 库存操作封装。
type Client struct{ rdb *redis.Client }

// New 构造 Client。
func New(rdb *redis.Client) *Client { return &Client{rdb: rdb} }

func skuStockKey(skuID int64) string { return fmt.Sprintf("sku:stock:%d", skuID) }

// Load 按需加载（SETNX），availableStock = stock - lockedStock。
func (c *Client) Load(ctx context.Context, skuID int64, availableStock int) error {
	return c.rdb.SetNX(ctx, skuStockKey(skuID), availableStock, 0).Err()
}

// lockScript 锁定多 SKU（下单时）。
// 检查所有 SKU 余量 >= qty，全部足够才扣减。
// 返回 "" 表示成功；"insufficient:<index>:<remain>" 表示不足。
var lockScript = redis.NewScript(`
local n = #KEYS
for i=1,n do
  local stock = tonumber(redis.call('GET', KEYS[i]) or '0')
  local qty   = tonumber(ARGV[i])
  if stock < qty then
    return 'insufficient:' .. (i-1) .. ':' .. stock
  end
end
for i=1,n do
  redis.call('DECRBY', KEYS[i], ARGV[i])
end
local detail = ARGV[n+1]
local ttl    = tonumber(ARGV[n+2])
redis.call('SET', 'inv:lock:' .. ARGV[n+3], detail, 'EX', ttl)
return 'ok'
`)

// Lock 原子锁定多 SKU，成功返回 "ok"，库存不足返回 "insufficient:<idx>:<remain>"。
func (c *Client) Lock(ctx context.Context, orderNo string, items []LockItem) (string, error) {
	keys := make([]string, len(items))
	args := make([]any, len(items)+3)
	for i, it := range items {
		keys[i] = skuStockKey(it.SKUID)
		args[i] = it.Qty
	}
	detail, _ := json.Marshal(items)
	args[len(items)] = string(detail)
	args[len(items)+1] = 86400 // 1 day TTL
	args[len(items)+2] = orderNo
	result, err := lockScript.Run(ctx, c.rdb, keys, args...).Text()
	return result, err
}

// releaseScript 释放锁定（关单/取消），幂等。
var releaseScript = redis.NewScript(`
local detail = redis.call('GET', 'inv:lock:' .. KEYS[1])
if not detail then return 'noop' end
local items = cjson.decode(detail)
for _, it in ipairs(items) do
  redis.call('INCRBY', 'sku:stock:' .. it['sku_id'], it['qty'])
end
redis.call('DEL', 'inv:lock:' .. KEYS[1])
return 'ok'
`)

// Release 幂等释放订单锁定库存。
func (c *Client) Release(ctx context.Context, orderNo string) (string, error) {
	return releaseScript.Run(ctx, c.rdb, []string{orderNo}).Text()
}

// deductScript 支付成功后删锁（DB 异步扣减）。
var deductScript = redis.NewScript(`
local exists = redis.call('DEL', 'inv:lock:' .. KEYS[1])
if exists == 0 then return 'noop' end
return 'ok'
`)

// Deduct 支付成功，删除锁定记录（DB 扣减由 worker 处理）。
func (c *Client) Deduct(ctx context.Context, orderNo string) (string, error) {
	return deductScript.Run(ctx, c.rdb, []string{orderNo}).Text()
}

// AdjustIn 手动增库存（Redis INCRBY）。
func (c *Client) AdjustIn(ctx context.Context, skuID int64, change int) error {
	return c.rdb.IncrBy(ctx, skuStockKey(skuID), int64(change)).Err()
}

// adjustOutScript 手动减库存（Lua 保证不低于 0）。
var adjustOutScript = redis.NewScript(`
local stock = tonumber(redis.call('GET', KEYS[1]) or '0')
local change = tonumber(ARGV[1])
if stock < change then return 'insufficient' end
redis.call('DECRBY', KEYS[1], change)
return 'ok'
`)

// AdjustOut 手动减库存，返回 "ok" 或 "insufficient"。
func (c *Client) AdjustOut(ctx context.Context, skuID int64, change int) (string, error) {
	return adjustOutScript.Run(ctx, c.rdb, []string{skuStockKey(skuID)}, change).Text()
}

// setScript 覆盖设置可售库存。
var setScript = redis.NewScript(`
redis.call('SET', KEYS[1], ARGV[1])
return 'ok'
`)

// Set 覆盖设置可售库存（= stock - locked_stock）。
func (c *Client) Set(ctx context.Context, skuID int64, available int) error {
	return setScript.Run(ctx, c.rdb, []string{skuStockKey(skuID)}, available).Err()
}

// Get 获取 Redis 可售库存（key 不存在返回 0）。
func (c *Client) Get(ctx context.Context, skuID int64) (int, error) {
	v, err := c.rdb.Get(ctx, skuStockKey(skuID)).Int()
	if err == redis.Nil {
		return 0, nil
	}
	return v, err
}

// GetWithExists 获取 Redis 可售库存，并返回 key 是否存在。
func (c *Client) GetWithExists(ctx context.Context, skuID int64) (int, bool, error) {
	v, err := c.rdb.Get(ctx, skuStockKey(skuID)).Int()
	if err == redis.Nil {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	return v, true, nil
}
