package main

import (
	"context"
	"github.com/bytedance/gopkg/cloud/metainfo"
	"github.com/cloudwego/kitex/pkg/endpoint"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/pkg/stats"
	"github.com/cloudwego/kitex/pkg/transmeta"
	"github.com/cloudwego/kitex/server"
	"github.com/cloudwego/kitex/server/genericserver"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"net"
	"time"
)

var (
	proxyAddr  net.Addr
	proxySvr   server.Server
	backendSvr server.Server
)

var sugar *zap.SugaredLogger

func init() {
	// 创建 zap logger（生产建议用 NewProduction）
	logger, err := zap.NewDevelopment(
		zap.AddStacktrace(zapcore.FatalLevel), // 只在 Fatal 时打印堆栈
	)
	if err != nil {
		panic(err)
	}

	sugar = logger.Sugar()

	// 👇 关键：将 klog 的输出重定向到 zap
	klog.SetOutput(&zapWriter{logger: sugar})
	klog.SetLevel(klog.LevelInfo) // 可选：设置日志级别
}

// zapWriter 实现 io.Writer，用于 klog.SetOutput
type zapWriter struct {
	logger *zap.SugaredLogger
}

func (z *zapWriter) Write(p []byte) (n int, err error) {
	// 去掉末尾的换行符（klog 会自动加 \n）
	msg := string(p)
	if len(msg) > 0 && msg[len(msg)-1] == '\n' {
		msg = msg[:len(msg)-1]
	}
	z.logger.Info(msg)
	return len(p), nil
}

// 提供全局 sugar 给业务代码使用
func GetLogger() *zap.SugaredLogger {
	return sugar
}

func main() {
	// 1. 创建 Listener（ln）
	ln, err := net.Listen("tcp", ":8080") // 监听 0.0.0.0:8888
	if err != nil {
		panic(err)
	}

	baseStats := server.WithStatsLevel(stats.LevelDetailed)

	// 2. 初始化 opts 切片
	var opts []server.Option

	// 3. 添加选项
	opts = append(opts,
		baseStats,
		server.WithMiddleware(AccessLogMiddleware(GetLogger())),
		server.WithListener(ln),
		server.WithMetaHandler(transmeta.ServerTTHeaderHandler),
		server.WithMetaHandler(transmeta.ServerHTTP2Handler),
	)

	// 4. 创建 Server
	svr := server.NewServer(opts...)

	addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:8888")
	if err != nil {
		panic(err)
	}

	// 5. 注册未知服务/方法的处理器（泛化调用关键）
	err = genericserver.RegisterUnknownServiceOrMethodHandler(svr, &genericserver.UnknownServiceOrMethodHandler{
		DefaultHandler:   defaultUnknownHandler(addr),
		StreamingHandler: streamingUnknownHandler(addr),
	})
	if err != nil {
		panic(err)
	}

	// 6. 启动服务
	if err := svr.Run(); err != nil {
		panic(err)
	}
}

func AccessLogMiddleware(logger *zap.SugaredLogger) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, req, resp interface{}) error {
			start := time.Now()

			// 执行实际调用
			err := next(ctx, req, resp)

			// 获取 RPC 信息
			ri := rpcinfo.GetRPCInfo(ctx)
			if ri == nil {
				logger.Warnw("missing rpcinfo in context")
				return err
			}

			invocation := ri.Invocation()
			//from := ri.From()
			//to := ri.To()
			uuid, _ := metainfo.GetValue(ctx, "uuid")

			// 输出结构化日志
			logger.Infow("rpc_access_log",
				"service", invocation.ServiceName(),
				"method", invocation.MethodName(),
				//"caller_addr", from.Address().String(),
				//"local_addr", to.Address().String(),
				"duration_ms", time.Since(start).Milliseconds(),
				"uuid", uuid,
				"success", err == nil,
				"error", func() string {
					if err != nil {
						return err.Error()
					}
					return ""
				}(),
			)

			return err
		}
	}
}
