package main

import (
	"context"
	"encoding/json"
	"fmt"
	service "kit/kitex_gen/kit/service"
)

// TestServiceImpl implements the last service interface defined in the IDL.
type TestServiceImpl struct{}

// TMethod implements the TestServiceImpl interface.
func (s *TestServiceImpl) TMethod(ctx context.Context, req *service.TestRequest) (resp *service.TestResponse, err error) {
	reqJ, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	fmt.Println("收到req", string(reqJ))
	return &service.TestResponse{Msg: "hello", S: req.S}, nil
}
