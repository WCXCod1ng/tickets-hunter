package delay_queue

import (
	"context"
	_ "embed"
	"errors"
	"strconv"
	"tickets-hunter/common/luaexec"
	"time"

	errors2 "github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

// 内嵌Lua脚本
//
//go:embed popDelayTaskScript.lua
var popDelayTaskScriptContent string

var popDelayTaskScript *luaexec.LuaScript

//go:embed popDelayTaskBatchScript.lua
var popDelayTaskBatchScriptContent string

var popDelayTaskBatchScript *luaexec.LuaScript

var ErrNoDelayTask = errors.New("no delay task")

func init() {
	popDelayTaskScript = luaexec.NewLuaScript(popDelayTaskScriptContent)
	popDelayTaskBatchScript = luaexec.NewLuaScript(popDelayTaskBatchScriptContent)
}

// ZSetDelayQueue 基于Redis Sorted Set实现的延迟队列
type ZSetDelayQueue struct {
	redisClient *redis.Redis
	queueKey    string
}

// NewZSetDelayQueue 创建一个新的ZSetDelayQueue实例
func NewZSetDelayQueue(redisClient *redis.Redis, queueKey string) *ZSetDelayQueue {
	return &ZSetDelayQueue{
		redisClient: redisClient,
		queueKey:    queueKey,
	}
}

// 添加一个任务到延迟队列，delay为延迟时间（单位：秒）
// score为当前时间戳加上延迟时间，表示任务的执行时间
func (q *ZSetDelayQueue) Add(ctx context.Context, taskID string, delay time.Duration) error {
	score := time.Now().Add(delay).Unix()
	_, err := q.redisClient.ZaddCtx(ctx, q.queueKey, score, taskID)
	// 返回值表示是否新增了成员（而不是score是否发生了变化），这里不关心这个结果，所以直接忽略
	if err != nil {
		return err
	}
	return nil
}

// 获取并移除到期的一个任务，返回任务ID
func (q *ZSetDelayQueue) GetDelayTask(ctx context.Context) (string, error) {
	now := time.Now().Unix()
	// 执行Lua脚本，原子地获取并移除到期的任务
	keys := []string{q.queueKey}
	args := []any{strconv.FormatInt(now, 10)}
	res, err := popDelayTaskScript.Exec(ctx, q.redisClient, keys, args...)
	if errors.Is(err, redis.Nil) {
		// 没有到期的任务，返回空列表
		return "", ErrNoDelayTask
	} else if err != nil {
		return "", errors2.WithStack(err)
	}
	// 返回结果
	taskID, ok := res.(string)
	if !ok {
		return "", errors2.WithStack(errors.New("unexpected script result type"))
	}
	return taskID, nil
}

// 获取并移除到期的多个任务，返回任务ID列表
// @param batchSize 一次获取的最大任务数量，实际返回的数量可能少于这个值，取决于当前到期的任务数量
func (q *ZSetDelayQueue) GetDelayTaskBatch(ctx context.Context, batchSize int) ([]string, error) {
	now := time.Now().Unix()
	// 执行Lua脚本，原子地获取并移除到期的任务
	keys := []string{q.queueKey}
	args := []any{strconv.FormatInt(now, 10), batchSize}
	res, err := popDelayTaskBatchScript.Exec(ctx, q.redisClient, keys, args...)
	if err != nil { // 不区分没有任务和其他错误，统一返回空列表
		return []string{}, errors2.WithStack(err)
	}
	// 返回结果
	taskIDs, ok := res.([]any)
	if !ok {
		return []string{}, errors2.WithStack(errors.New("unexpected script result type"))
	}
	result := make([]string, 0, len(taskIDs))
	for _, id := range taskIDs {
		strID, ok := id.(string)
		if !ok {
			return []string{}, errors2.WithStack(errors.New("unexpected script result type"))
		}
		result = append(result, strID)
	}
	return result, nil
}
