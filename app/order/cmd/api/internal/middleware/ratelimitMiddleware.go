// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package middleware

import (
	"net/http"

	"github.com/zeromicro/go-zero/core/limit"
)

type RateLimitMiddleware struct {
	limiter *limit.TokenLimiter
}

func NewRateLimitMiddleware(limiter *limit.TokenLimiter) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		limiter: limiter,
	}
}

func (m *RateLimitMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		//if !m.limiter.Allow() {
		//	http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
		//}

		next(w, r)
	}
}
