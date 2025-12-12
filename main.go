package main

import (
	service "kit/kitex_gen/kit/service/testservice"
	"log"
)

func main() {
	svr := service.NewServer(new(TestServiceImpl))

	err := svr.Run()

	if err != nil {
		log.Println(err.Error())
	}
}
