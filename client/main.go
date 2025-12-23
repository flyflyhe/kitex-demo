package main

import (
	"context"
	"fmt"
	"github.com/bytedance/gopkg/cloud/metainfo"
	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/pkg/transmeta"
	"github.com/cloudwego/kitex/transport"
	"io"
	"kit/kitex_gen/kit/common"
	"kit/kitex_gen/kit/service"
	"kit/kitex_gen/kit/service/testservice"
)

func main() {
	cli, err := testservice.NewClient("TestService",
		client.WithHostPorts("127.0.0.1:8080"),
		client.WithClientBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: "TestService"}),
		client.WithTransportProtocol(transport.TTHeaderFramed),
		client.WithMetaHandler(transmeta.ClientTTHeaderHandler),
	)
	if err != nil {
		panic(err)
	}
	defer func() {
		if closer, ok := cli.(io.Closer); ok {
			_ = closer.Close()
		}
	}()

	ctx := context.Background()

	// ✅ 关键：设置目标服务名和服务方法（用于路由）
	ctx = metainfo.WithValue(ctx, "uuid", "123545")

	res, err := cli.TMethod(ctx, &service.TestRequest{Msg: "hello", S: &common.TestStruct{
		SBoolReq:      true,
		SListString:   []string{"hello"},
		SSetI16:       []int16{1},
		SMapI32String: map[int32]string{1: "hello"},
	}})
	if err != nil {
		panic(err)
	}
	fmt.Println(res.Msg)
}
