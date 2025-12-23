package main

import (
	"context"
	"encoding/json"
	"github.com/bytedance/gopkg/cloud/metainfo"
	"github.com/cloudwego/kitex/pkg/endpoint"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/pkg/transmeta"
	"github.com/cloudwego/kitex/server"
	etcd "github.com/kitex-contrib/registry-etcd"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	service "kit/kitex_gen/kit/service/testservice"
	"log"
	"net"
	"sync"
	"time"
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
	wg := sync.WaitGroup{}
	wg.Add(3)
	go func() {
		defer func() {
			wg.Done()
		}()
		r, err := etcd.NewEtcdRegistry([]string{"127.0.0.1:2379"}) // r should not be reused.
		if err != nil {
			log.Fatal(err)
		}

		ipPort := "127.0.0.1:8889"

		ln, err := net.Listen("tcp", ipPort)

		svr := service.NewServer(&TestServiceImpl{Port: "8889"},
			server.WithMiddleware(AccessLogMiddleware(GetLogger())),
			server.WithMetaHandler(transmeta.MetainfoServerHandler),
			server.WithListener(ln),
			server.WithRegistry(r), // 注册到 etcd
			server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{
				ServiceName: "TestService", // 服务名必须与客户端一致
			}),
		)

		err = svr.Run()

		if err != nil {
			log.Println(err.Error())
		}
	}()

	go func() {
		defer func() {
			wg.Done()
		}()
		r, err := etcd.NewEtcdRegistry([]string{"127.0.0.1:2379"}) // r should not be reused.
		if err != nil {
			log.Fatal(err)
		}

		ipPort := "127.0.0.1:8888"

		ln, err := net.Listen("tcp", ipPort)

		svr := service.NewServer(&TestServiceImpl{Port: "8888"}, server.WithListener(ln),
			server.WithMetaHandler(transmeta.MetainfoServerHandler),
			server.WithMiddleware(AccessLogMiddleware(GetLogger())),
			server.WithRegistry(r), // 注册到 etcd
			server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{
				ServiceName: "TestService", // 服务名必须与客户端一致
			}))

		err = svr.Run()

		if err != nil {
			log.Println(err.Error())
		}
	}()

	go func() {
		defer func() {
			wg.Done()
		}()

		time.Sleep(time.Second * 20)
		r, err := etcd.NewEtcdRegistry([]string{"127.0.0.1:2379"}) // r should not be reused.
		if err != nil {
			log.Fatal(err)
		}

		ipPort := "127.0.0.1:8890"

		ln, err := net.Listen("tcp", ipPort)

		svr := service.NewServer(&TestServiceImpl{Port: "8890"},
			server.WithMiddleware(AccessLogMiddleware(GetLogger())),
			server.WithMetaHandler(transmeta.MetainfoServerHandler),
			server.WithListener(ln),
			server.WithRegistry(r), // 注册到 etcd
			server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{
				ServiceName: "TestService", // 服务名必须与客户端一致
			}),
		)

		err = svr.Run()

		if err != nil {
			log.Println(err.Error())
		}
	}()

	wg.Wait()
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

			logB, _ := json.Marshal(map[string]interface{}{
				"service": invocation.ServiceName(),
				"method":  invocation.MethodName(),
				//"caller_addr", from.Address().String(),
				//"local_addr", to.Address().String(),
				"duration_ms": time.Since(start).Milliseconds(),
				"uuid":        uuid,
				"success":     err == nil,
				"error": func() string {
					if err != nil {
						return err.Error()
					}
					return ""
				}(),
				"request":  req,  //
				"response": resp, //
			})
			// 输出结构化日志
			logger.Info(string(logB))

			return err
		}
	}
}
