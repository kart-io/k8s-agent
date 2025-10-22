package server

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/kart-io/k8s-agent/common/config"
	"github.com/kart-io/logger/core"
)

// InterceptorConfig 拦截器配置
type InterceptorConfig struct {
	Logger        core.Logger
	LoggingConfig *config.LoggingOptions
}

// LoggingUnaryInterceptor 一元 RPC 日志拦截器
func LoggingUnaryInterceptor(cfg *InterceptorConfig) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()

		// 调用处理器
		resp, err := handler(ctx, req)

		// 记录日志
		duration := time.Since(start)
		code := codes.OK
		if err != nil {
			code = status.Code(err)
		}

		// 根据配置决定是否记录日志
		if shouldLog(cfg.LoggingConfig, code) {
			// 构建日志字段
			fields := []interface{}{
				"method", info.FullMethod,
				"duration", duration.String(),
				"code", code.String(),
			}

			// 添加请求ID（如果存在）
			if requestID := ctx.Value("request_id"); requestID != nil {
				fields = append(fields, "request_id", requestID.(string))
			}

			// 根据状态码选择日志级别
			if err != nil && code != codes.OK {
				cfg.Logger.Errorw("gRPC call failed", fields...)
			} else {
				cfg.Logger.Infow("gRPC call completed", fields...)
			}
		}

		return resp, err
	}
}

// LoggingStreamInterceptor 流式 RPC 日志拦截器
func LoggingStreamInterceptor(cfg *InterceptorConfig) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		start := time.Now()

		// 调用处理器
		err := handler(srv, ss)

		// 记录日志
		duration := time.Since(start)
		code := codes.OK
		if err != nil {
			code = status.Code(err)
		}

		// 根据配置决定是否记录日志
		if shouldLog(cfg.LoggingConfig, code) {
			fields := []interface{}{
				"method", info.FullMethod,
				"duration", duration.String(),
				"code", code.String(),
				"is_client_stream", info.IsClientStream,
				"is_server_stream", info.IsServerStream,
			}

			// 根据状态码选择日志级别
			if err != nil && code != codes.OK {
				cfg.Logger.Errorw("gRPC stream failed", fields...)
			} else {
				cfg.Logger.Infow("gRPC stream completed", fields...)
			}
		}

		return err
	}
}

// shouldLog 根据 LoggingOptions 判断是否应该记录日志
func shouldLog(cfg *config.LoggingOptions, code codes.Code) bool {
	if cfg == nil {
		return true // 默认记录所有日志
	}

	// 根据日志级别过滤
	level := cfg.Level
	switch level {
	case "debug", "DEBUG":
		return true // debug 级别记录所有
	case "info", "INFO":
		return true // info 级别也记录所有 gRPC 调用
	case "warn", "WARN":
		return code != codes.OK // warn 级别只记录错误
	case "error", "ERROR":
		return code != codes.OK && code != codes.Canceled // error 级别记录错误但忽略取消
	case "fatal", "FATAL":
		return code == codes.Internal || code == codes.DataLoss // fatal 级别只记录严重错误
	default:
		return true
	}
}

// RecoveryUnaryInterceptor 一元 RPC panic 恢复拦截器
func RecoveryUnaryInterceptor(logger core.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				logger.Errorw("panic recovered in gRPC handler",
					"method", info.FullMethod,
					"panic", r,
				)
				err = status.Errorf(codes.Internal, "internal server error")
			}
		}()

		return handler(ctx, req)
	}
}

// RecoveryStreamInterceptor 流式 RPC panic 恢复拦截器
func RecoveryStreamInterceptor(logger core.Logger) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		defer func() {
			if r := recover(); r != nil {
				logger.Errorw("panic recovered in gRPC stream handler",
					"method", info.FullMethod,
					"panic", r,
				)
				err = status.Errorf(codes.Internal, "internal server error")
			}
		}()

		return handler(srv, ss)
	}
}

// RequestIDUnaryInterceptor 一元 RPC 请求 ID 拦截器
func RequestIDUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// 从 metadata 中获取请求 ID
		requestID := extractRequestID(ctx)

		// 将请求 ID 添加到 context
		ctx = context.WithValue(ctx, "request_id", requestID)

		return handler(ctx, req)
	}
}

// RequestIDStreamInterceptor 流式 RPC 请求 ID 拦截器
func RequestIDStreamInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		// 从 metadata 中获取请求 ID
		requestID := extractRequestID(ss.Context())

		// 创建包装的 ServerStream 来传递 context
		wrapped := &wrappedStream{
			ServerStream: ss,
			ctx:          context.WithValue(ss.Context(), "request_id", requestID),
		}

		return handler(srv, wrapped)
	}
}

// wrappedStream 包装的 ServerStream，用于传递自定义 context
type wrappedStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedStream) Context() context.Context {
	return w.ctx
}

// extractRequestID 从 metadata 中提取请求 ID
func extractRequestID(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}

	requestIDs := md.Get("x-request-id")
	if len(requestIDs) == 0 {
		return ""
	}

	return requestIDs[0]
}
