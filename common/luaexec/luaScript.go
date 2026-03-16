package luaexec

import (
	"context"
	"strings"
	"sync"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

type LuaScript struct {
	script string
	sha    string
	once   sync.Once
}

func NewLuaScript(script string) *LuaScript {
	return &LuaScript{
		script: script,
	}
}

// 获取Lua脚本的SHA1值，第一次调用时计算并缓存结果
func (l *LuaScript) loadSHA(ctx context.Context, r *redis.Redis) error {
	var err error
	// 使用sync.Once确保并发调用情况下也只计算一次SHA1值并缓存结果
	l.once.Do(func() {
		if l.sha != "" {
			return
		}
		l.sha, err = r.ScriptLoadCtx(ctx, l.script)
	})
	return err
}

// 执行Lua脚本
func (l *LuaScript) Exec(
	ctx context.Context,
	r *redis.Redis,
	keys []string,
	args ...any,
) (any, error) {
	// 0. 确保 SHA 已加载（只会一次）
	if err := l.loadSHA(ctx, r); err != nil {
		return nil, err
	}

	//1. 优先 EvalSha
	res, err := r.EvalShaCtx(ctx, l.sha, keys, args...)
	if err == nil {
		return res, nil
	}

	// 2. 如果是 NOSCRIPT，fallback 到 Eval
	if isNoScriptErr(err) {
		res, err = r.EvalCtx(ctx, l.script, keys, args...)
		if err != nil {
			return nil, err
		}

		// 3. 重新加载 SHA（可选，但推荐）
		l.once = sync.Once{} // 重置 sync.Once 以允许重新计算 SHA
		l.sha = ""           // 清空之前的 SHA，确保下一次调用时重新计算
		_ = l.loadSHA(ctx, r)
		return res, nil
	}

	return nil, err
}

func isNoScriptErr(err error) bool {
	if err == nil {
		return false
	}
	// 注意：只判断 NOSCRIPT
	return strings.HasPrefix(err.Error(), "NOSCRIPT")
}
