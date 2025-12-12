package main

import (
	"context"
	"fmt"
	"github.com/bytedance/gopkg/lang/dirtmake"
	"github.com/cloudwego/frugal"
	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/client/genericclient"
	"github.com/cloudwego/kitex/pkg/generic"
	"github.com/cloudwego/kitex/pkg/transmeta"
	"github.com/cloudwego/kitex/transport"
	"kit/kitex_gen/kit/common"
	"kit/kitex_gen/kit/service"
	"net"
)

func main() {
	addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:8080")
	if err != nil {
		panic(err)
	}
	cli, err := genericclient.NewClient("hello", generic.BinaryThriftGenericV2("TestService"),
		client.WithHostPorts(addr.String()),
		client.WithTransportProtocol(transport.TTHeader|transport.TTHeaderStreaming),
		client.WithMetaHandler(transmeta.ClientTTHeaderHandler),
		client.WithMetaHandler(transmeta.ClientHTTP2Handler))
	if err != nil {
		panic(err)
	}

	args := &service.TestRequest{Msg: "hello", S: &common.TestStruct{
		SBoolReq:      true,
		SListString:   []string{"hello"},
		SSetI16:       []int16{1},
		SMapI32String: map[int32]string{1: "hello"},
	}}

	size := frugal.EncodedSize(args)
	buf := dirtmake.Bytes(size, size)
	_, _ = frugal.EncodeObject(buf, nil, args)

	ctx := context.WithValue(context.Background(), "uuid", "123456")
	res, err := cli.GenericCall(ctx, "tMethod", buf)
	if err != nil {
		panic(err)
	}
	fmt.Println(res)
}
