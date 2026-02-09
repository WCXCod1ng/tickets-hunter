package result

import (
	"errors"
	"log"
	"net/http"
	"tickets-hunter/common/xerr"

	errors2 "github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RegisterErrorHandler 在 main.go 中调用，注册全局错误处理
func RegisterErrorHandler() {
	httpx.SetErrorHandler(func(err error) (int, interface{}) {
		if err == nil {
			log.Panicf("execute GlobalErrorHandler while err is nil")
		}
		var code uint32 = uint32(codes.Internal)
		var msg string = "服务器端未知错误" // 默认消息体

		// 判断错误类型
		var validationError *xerr.ValidationError
		switch {
		case errors.As(err, &validationError):
			// 说明是我们自定义的校验错误，其Error()方法已经包含了具体的错误信息，可以直接返回
			code = validationError.Code()
			msg = validationError.Error()
			logx.Errorf("[Validation error]: %v", err)
		default:
			// 其余情况，尝试解析为 gRPC 的 status 错误
			// 首先剥离外壳，获取实际的错误原因
			cause := errors2.Cause(err)
			if s, ok := status.FromError(cause); ok {
				code = uint32(s.Code())
				// 特判是否是服务器内部错误，如果是则将消息输出到日志中，返回给前端的是默认的消息体
				if xerr.IsServerError(code) {
					logx.Errorf("[Server error]: %+v", err) // %+v会记录堆栈信息
				} else {
					// 否则将根因错误的消息返回到前端
					msg = s.Message()
					logx.Errorf("[Business error]: %v", err)
				}
			} else {
				// 这里可以记录一下未知系统错误的日志
				logx.Errorf("GlobalErrorHandler unknown error: %+v", err)
			}
		}
		// 返回给前端的 JSON 结构，我们约定除非是网络等错误，正常请求到状态码都为200，只根据Code字段来区分具体是否成功
		return http.StatusOK, map[string]interface{}{
			"code": code,
			"msg":  msg,
			"data": nil, // 出错时 data 为空
		}
	})
}

// Success 封装成功响应
func Success(data interface{}) map[string]interface{} {
	return map[string]interface{}{
		"code": codes.OK,
		"msg":  "success",
		"data": data,
	}
}

// HttpResult 【核心方法】统一处理 HTTP 响应
// 替代 goctl 生成代码中的 if err != nil { ... } else { ... }
func HttpResult(r *http.Request, w http.ResponseWriter, resp interface{}, err error) {
	if err == nil {
		// --- 1. 成功逻辑 ---
		// 这里手动调用 Success 包装
		data := Success(resp)
		httpx.OkJsonCtx(r.Context(), w, data)
	} else {
		// --- 2. 失败逻辑 ---
		// 这里调用 httpx.Error，它会自动触发你之前注册的 SetErrorHandler
		// 从而保持错误处理逻辑的一致性
		httpx.ErrorCtx(r.Context(), w, err)
	}
}
