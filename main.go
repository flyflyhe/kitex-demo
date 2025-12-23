package main

import (
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
	etcd "github.com/kitex-contrib/registry-etcd"
	service "kit/kitex_gen/kit/service/testservice"
	"log"
	"net"
	"sync"
)

func main() {
	wg := sync.WaitGroup{}
	wg.Add(2)
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
			server.WithRegistry(r), // 注册到 etcd
			server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{
				ServiceName: "TestService", // 服务名必须与客户端一致
			}))

		err = svr.Run()

		if err != nil {
			log.Println(err.Error())
		}
	}()

	wg.Wait()
}
