package utils

import "github.com/golang-jwt/jwt/v4"

// GenerateJwtToken 生成 Token 的辅助方法
// 实际的userId是雪花算法生成的int64，但是在go-zero框架下前端的token会被解析为json.Number类型，导致精度丢失，所以这里直接使用字符串类型的userId
func GenerateJwtToken(secretKey string, iat, seconds int64, userId string) (string, error) {
	claims := make(jwt.MapClaims)
	claims["exp"] = iat + seconds
	claims["iat"] = iat
	// 自定义字段：存入 UserId，后续在受保护接口中可直接获取
	claims["userId"] = userId

	token := jwt.New(jwt.SigningMethodHS256)
	token.Claims = claims

	return token.SignedString([]byte(secretKey))
}

// GenerateRefreshToken 生成 Refresh Token
// 生产环境中，Refresh Token 既可以是 JWT，也可以是随机字符串存 Redis
// 这里为了演示方便，使用 JWT，但 Payload 里标记它是 refresh 类型
func GenerateRefreshToken(secretKey string, iat, seconds, userId int64) (string, error) {
	claims := make(jwt.MapClaims)
	claims["exp"] = iat + seconds
	claims["iat"] = iat
	claims["userId"] = userId
	claims["type"] = "refresh" // 关键：标记类型，防止混用
	token := jwt.New(jwt.SigningMethodHS256)
	token.Claims = claims
	return token.SignedString([]byte(secretKey))
}
