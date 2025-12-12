package main

import (
	"context"
	"fmt"
	service "kit/kitex_gen/kit/service"
)

// TestServiceImpl implements the last service interface defined in the IDL.
type TestServiceImpl struct{}

// TMethod implements the TestServiceImpl interface.
func (s *TestServiceImpl) TMethod(ctx context.Context, req *service.TestRequest) (resp *service.TestResponse, err error) {
	// TODO: Your code here...
	fmt.Println("收到req", req.Msg)
	return &service.TestResponse{Msg: "hello " + req.Msg, S: req.S}, nil
}
