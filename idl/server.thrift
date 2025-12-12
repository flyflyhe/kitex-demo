namespace go kit.service

include "common.thrift"

struct TestRequest {
   1: string msg
   2: common.TestStruct s
}

struct TestResponse {
   1: string msg
   2: common.TestStruct s
}

service TestService {
   TestResponse tMethod(1: TestRequest req)
}
