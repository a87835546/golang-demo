package consts

/**
 * @author 大菠萝
 * @description //TODO go中的常量定义方式，通过构建两组常量达到key/value键值对的效果 达到java中枚举的效果
 * @date 5:27 pm 9/7/22
 * @param
 * @return
 **/
const (
	Success         = int32(10000)
	SystemErr       = int32(10001)
	ParamErr        = int32(10002)
	LessParam       = int32(10003)
	MethodErr       = int32(10004)
	TokenErr        = int32(10005)
	TokenEmpty      = int32(10006)
	LoginErr        = int32(10007)
	RegisterErr     = int32(10008)
	OperationPwdErr = int32(10009)
	PasswordErr     = int32(10010)
)

// MessageMap TODO 对应个 JSON 文档最好
var MessageMap = map[int32]string{
	SystemErr:       "系统内部错误",
	ParamErr:        "参数错误",
	MethodErr:       "方法错误",
	LessParam:       "缺少参数",
	Success:         "请求成功",
	TokenErr:        "鉴权错误",
	TokenEmpty:      "空鉴权",
	LoginErr:        "用户名或密码错误",
	RegisterErr:     "用户当日注册频率过高(ip)",
	OperationPwdErr: "操作密码错误",
	PasswordErr:     "密码两次不一致",
}
