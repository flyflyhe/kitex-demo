package server

import (
	"github.com/cloudwego/kitex/pkg/transmeta"
	"github.com/cloudwego/kitex/server"
	"github.com/cloudwego/kitex/server/genericserver"
	"net"
)

func main() {
	ln, err := net.Listen("tcp", ":9999") // 监听 0.0.0.0:9999
	if err != nil {
		panic(err)
	}

	// 2. 初始化 opts 切片
	var opts []server.Option

	opts = append(opts, server.WithListener(ln),
		server.WithMetaHandler(transmeta.ServerTTHeaderHandler),
		server.WithMetaHandler(transmeta.ServerHTTP2Handler))

	svr := server.NewServer(opts...)
	err := genericserver.RegisterUnknownServiceOrMethodHandler(svr, &genericserver.UnknownServiceOrMethodHandler{
		DefaultHandler:   defaultUnknownHandler,
		StreamingHandler: streamingUnknownHandler,
	})
}
