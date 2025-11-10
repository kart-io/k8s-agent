package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Queue Redis 队列实现
// 使用 Redis List 数据结构实现 FIFO 队列
type Queue struct {
	client *Client
	name   string
}

// NewQueue 创建队列
//
// Parameters:
//   - client: Redis 客户端实例
//   - name: 队列名称
//
// Example usage:
//
//	queue := redis.NewQueue(client, "tasks")
//	err := queue.Push(ctx, myTask)
func NewQueue(client *Client, name string) *Queue {
	return &Queue{
		client: client,
		name:   fmt.Sprintf("queue:%s", name),
	}
}

// Push 推送消息到队列（从右侧）
// 消息会被序列化为 JSON 格式存储
//
// Parameters:
//   - ctx: Context for timeout control
//   - message: 要推送的消息（任意类型，会被 JSON 序列化）
//
// Returns:
//   - error: 序列化失败或 Redis 操作失败时返回错误
func (q *Queue) Push(ctx context.Context, message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	if err := q.client.Client().RPush(ctx, q.name, data).Err(); err != nil {
		return fmt.Errorf("failed to push to queue: %w", err)
	}

	q.client.logger.Debugw("Message pushed to queue",
		"queue", q.name,
		"size", len(data),
	)

	return nil
}

// Pop 从队列弹出消息（从左侧，阻塞）
// 如果队列为空，会阻塞等待直到有消息或超时
//
// Parameters:
//   - ctx: Context for timeout control
//   - timeout: 阻塞等待超时时间，0 表示无限等待
//
// Returns:
//   - []byte: 消息内容（JSON 格式）
//   - error: Redis 操作失败或超时时返回错误
func (q *Queue) Pop(ctx context.Context, timeout time.Duration) ([]byte, error) {
	result, err := q.client.Client().BLPop(ctx, timeout, q.name).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("queue empty or timeout: %s", q.name)
		}
		return nil, fmt.Errorf("failed to pop from queue: %w", err)
	}

	if len(result) < 2 {
		return nil, fmt.Errorf("invalid result from BLPop")
	}

	q.client.logger.Debugw("Message popped from queue",
		"queue", q.name,
		"size", len(result[1]),
	)

	return []byte(result[1]), nil
}

// PopNonBlocking 从队列弹出消息（非阻塞）
// 如果队列为空，立即返回 redis.Nil 错误
//
// Parameters:
//   - ctx: Context for timeout control
//
// Returns:
//   - []byte: 消息内容（JSON 格式）
//   - error: 队列为空时返回 redis.Nil，其他错误返回操作失败错误
func (q *Queue) PopNonBlocking(ctx context.Context) ([]byte, error) {
	result, err := q.client.Client().LPop(ctx, q.name).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, err // 队列为空，返回 redis.Nil 让调用者判断
		}
		return nil, fmt.Errorf("failed to pop from queue: %w", err)
	}

	q.client.logger.Debugw("Message popped from queue (non-blocking)",
		"queue", q.name,
		"size", len(result),
	)

	return []byte(result), nil
}

// Length 获取队列长度
//
// Returns:
//   - int64: 队列中消息数量
//   - error: Redis 操作失败时返回错误
func (q *Queue) Length(ctx context.Context) (int64, error) {
	length, err := q.client.Client().LLen(ctx, q.name).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get queue length: %w", err)
	}
	return length, nil
}

// Clear 清空队列
// 删除队列中的所有消息
//
// Returns:
//   - error: Redis 操作失败时返回错误
func (q *Queue) Clear(ctx context.Context) error {
	if err := q.client.Client().Del(ctx, q.name).Err(); err != nil {
		return fmt.Errorf("failed to clear queue: %w", err)
	}

	q.client.logger.Debugw("Queue cleared", "queue", q.name)
	return nil
}

// Peek 查看队列头部元素（不弹出）
// 返回队列前 N 个元素，但不从队列中移除它们
//
// Parameters:
//   - ctx: Context for timeout control
//   - count: 要查看的元素数量
//
// Returns:
//   - [][]byte: 消息列表（JSON 格式）
//   - error: Redis 操作失败时返回错误
func (q *Queue) Peek(ctx context.Context, count int64) ([][]byte, error) {
	results, err := q.client.Client().LRange(ctx, q.name, 0, count-1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to peek queue: %w", err)
	}

	data := make([][]byte, len(results))
	for i, result := range results {
		data[i] = []byte(result)
	}

	return data, nil
}

// PushBatch 批量推送消息到队列
// 使用 Redis Pipeline 提高性能
//
// Parameters:
//   - ctx: Context for timeout control
//   - messages: 要推送的消息列表
//
// Returns:
//   - error: 序列化失败或 Redis 操作失败时返回错误
func (q *Queue) PushBatch(ctx context.Context, messages []interface{}) error {
	if len(messages) == 0 {
		return nil
	}

	pipe := q.client.Client().Pipeline()

	for _, msg := range messages {
		data, err := json.Marshal(msg)
		if err != nil {
			return fmt.Errorf("failed to marshal message: %w", err)
		}
		pipe.RPush(ctx, q.name, data)
	}

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("failed to push batch to queue: %w", err)
	}

	q.client.logger.Debugw("Batch pushed to queue",
		"queue", q.name,
		"count", len(messages),
	)

	return nil
}
