package main

import (
	"github.com/cloudwego/kitex/pkg/registry"
	etcd "github.com/kitex-contrib/registry-etcd"
	service "kit/kitex_gen/kit/service/testservice"
	"log"
	"net"
)

func main() {
	r, err := etcd.NewEtcdRegistry([]string{"127.0.0.1:2379"}) // r should not be reused.
	if err != nil {
		log.Fatal(err)
	}

	addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:8888")
	if err != nil {
		log.Fatal(err)
	}

	if err := r.Register(&registry.Info{
		ServiceName: "TestService",
		Addr:        addr,
	}); err != nil {
		log.Fatal(err)
	}

	svr := service.NewServer(new(TestServiceImpl))

	err = svr.Run()

	if err != nil {
		log.Println(err.Error())
	}
}
