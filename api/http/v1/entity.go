package v1

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// 全局常量，用于设置错误码
const (
	CodeSuccess           ErrCode = 0     // http请求成功
	CodeBodyBindErr       ErrCode = 10001 // 参数绑定错误
	CodeParamErr          ErrCode = 10002 // 请求参数不合法
	CodeRegisterErr       ErrCode = 10003 // 注册错误
	CodeLoginErr          ErrCode = 10003 // 登录错误
	CodeLogoutErr         ErrCode = 10004 // 登出错误
	CodeGetUserInfoErr    ErrCode = 10005 // 获取用户信息错误
	CodeUpdateUserInfoErr ErrCode = 10006 // 更新用户信息错误
)

type (
	DebugType int // debug类型
	ErrCode   int // 错误码
)

// HttpResponse http独立请求返回结构体
type HttpResponse struct {
	Code ErrCode     `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

// ResponseWithError http请求返回处理函数
func (rsp *HttpResponse) ResponseWithError(c *gin.Context, code ErrCode, msg string) {
	// 将错误信息填充到 HttpResponse 结构体中的 Code 和 Msg 字段
	rsp.Code = code
	rsp.Msg = msg

	// 将该结构体通过 JSON 格式返回给客户端
	// 并指定返回的 HTTP 状态码为 500（http.StatusInternalServerError）
	c.JSON(http.StatusInternalServerError, rsp)
}

func (rsp *HttpResponse) ResponseSuccess(c *gin.Context) {
	// 将成功的信息填充到 HttpResponse 结构体中的 Code 和 Msg 字段
	rsp.Code = CodeSuccess
	rsp.Msg = "success"

	// 将该结构体通过 JSON 格式返回给客户端
	// 并指定返回的 HTTP 状态码为 200（http.StatusOK）
	c.JSON(http.StatusOK, rsp)
}

// 这个返回函数给客户端多返回了一个 data 也就是实际的数据（从缓存中获取的用户信息）
func (rsp *HttpResponse) ResponseWithData(c *gin.Context, data interface{}) {
	rsp.Code = CodeSuccess
	rsp.Msg = "success"
	rsp.Data = data
	c.JSON(http.StatusOK, rsp)
}
