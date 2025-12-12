namespace go kit.common

// typedef
typedef i32 TestInteger

// Enum
enum TestEnum {
    Enum1 = 1,
    Enum2,
    Enum3 = 10,
}

// Constant
const i32 TestIntConstant = 1234;

// Struct
struct TestStruct {
    1: bool sBool
    2: required bool sBoolReq
    3: optional bool sBoolOpt
    4: list<string> sListString
    5: set<i16> sSetI16
    6: map<i32,string> sMapI32String
}
