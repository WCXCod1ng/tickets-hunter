package interceptor

import (
	"context"

	"tickets-hunter/common/xerr" // 替换为你实际的项目路径

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ServerErrorInterceptor RPC服务端错误处理拦截器
// 功能：
// 1. 统一处理错误日志（区分业务错误和系统错误）
// 2. 隐藏系统错误的细节（防止敏感信息泄露到API层）
// 3. 提取 TraceID 关联日志
func ServerErrorInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	// --- 【阶段 1：前置处理】 ---
	// (此处可以添加鉴权、参数打印等逻辑，但在本例中略过)

	// --- 【阶段 2：执行核心业务逻辑】 ---
	// 调用 handler 进入 Logic 层
	resp, err = handler(ctx, req)

	// --- 【阶段 3：后置处理 (错误分类与清洗)】 ---
	if err != nil {
		// 获取错误的根本原因
		cause := errors.Cause(err)

		s, _ := status.FromError(cause)
		// 获取根因错误码
		errCode := s.Code()

		// 不应当拦截codes.Aborted错误，因为它是DTM分布式事务框架用来触发补偿逻辑的特定错误码，拦截后会导致DTM无法正确识别业务错误，无法执行补偿逻辑
		if errCode == codes.Aborted {
			return resp, err
		}

		// 利用根因错误码判断错误类型
		if xerr.IsServerError(uint32(errCode)) {
			// === 场景 A：系统错误 (如：DB连接失败、空指针、Redis超时) ===
			// 策略：
			// 1. 记录 Error 日志，并【打印堆栈】(核心！)
			//    %+v 会利用 pkg/errors 打印出 Logic 层代码崩溃的具体行号。注意这里的 err 是经过 pkg/errors 包装过的，才会有堆栈信息
			logx.WithContext(ctx).Errorf("[RPC-System-Error] Method:%s Error:%+v", info.FullMethod, err)

			// 2. 【篡改】返回给 API 层的错误
			//    为了安全，不能把 "DB connection failed ip:10.0.1.5" 这种信息传出去
			//    统一转换为 "服务器繁忙" 或者对应的通用系统错误码
			//    注意：这里返回的新 error
			return nil, status.Error(codes.Internal, "RPC服务内部错误")
		} else {
			// === 场景 B：业务错误 (如：账号密码错误、库存不足) ===
			// 策略：
			// 1. 记录 Info 日志 (通常不需要堆栈信息，只需要知道发生了什么)
			// 	【保留】原始错误返回给 API 层，因为 API 层需要根据这个 code 做判断。注意这里的s.Message()才是业务错误的具体信息，比如 "账号密码错误"、"库存不足"，这些信息是可以直接返回给前端的，不存在安全问题，不包括堆栈等实现细节
			logx.WithContext(ctx).Infof("[RPC-Business-Log] Code:%d Msg:%s", errCode, s.Message())

			// 原样返回，不做修改
			return resp, err
		}
	}

	// 当然，如果err为nil，那么说明没有错误，原样返回
	return resp, nil
}
