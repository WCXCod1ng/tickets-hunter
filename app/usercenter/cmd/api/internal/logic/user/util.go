package user

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
)

func GetUserIdFromToken(ctx context.Context) (int64, error) {
	// 从上下文中获取 userId，前提是中间件已经将其存入
	userIdValue := ctx.Value("userId")
	if userIdValue == nil {
		return 0, errors.New("userId not found in context")
	}

	userId, ok := userIdValue.(string)
	if !ok {
		return 0, errors.New("userId in context is not a string")
	}

	// 这里我们之前在生成 Token 时将 userId 作为字符串存入，所以直接转换为 int64
	var userIdInt int64
	if _, err := fmt.Sscanf(userId, "%d", &userIdInt); err != nil {
		return 0, errors.Errorf("userId in context is not a number: %v", err)
	}

	return userIdInt, nil
}
