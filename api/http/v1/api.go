package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"gouse/config"
	"gouse/internal/service"
	"gouse/pkg/constant"
	"gouse/utils"
	"net/http"
	"time"
)

// Ping 健康检查
// Ping() 函数的参数 c 是一个 gin.Context 上下文对象，代表了一个 HTTP 请求的上下文。
// 通过该对象，我们可以获取请求的信息、设置响应的内容等操作。
func Ping(c *gin.Context) {
	// 获取这个包含了应用程序的配置信息的全局配置对象，app 服务配置信息
	appConfig := config.GetGlobalConf().AppConfig

	// 将 appConfig 这个全局配置对象转换成格式化的 JSON 字符串
	// "" 是 JSON 字符串中的字段间的分隔符，"  " 是每一级缩进的字符串。
	confInfo, _ := json.MarshalIndent(appConfig, "", "  ")

	// 将配置信息和应用程序信息格式化为一个字符串。(这是一个格式化操作)
	appInfo := fmt.Sprintf("app_name: %s\nversion: %s\n\n%s", appConfig.AppName, appConfig.Version,
		string(confInfo))

	// 使用 gin.Context 的 String 方法，
	// 将 HTTP 状态码设置为 200（OK），并将 appInfo 字符串作为响应内容发送给客户端。
	// 所以这段代码的逻辑就是，客户端 Ping 服务器，服务器返回这个响应内容 appInfo 给客户端
	c.String(http.StatusOK, appInfo)
}

// Register 注册
func Register(c *gin.Context) {
	// req 变量用于存储注册请求信息（RegisterRequest 是一个结构体类型）
	req := &service.RegisterRequest{}

	// rsp 变量用于存储 http请求的响应信息
	rsp := &HttpResponse{}

	// ShouldBindJSON(&req) 解析 JSON 格式的请求体，并将解析结果存储在 req 变量中。
	// 如果解析过程中发生错误，错误信息将存储在 err 变量中
	err := c.ShouldBindJSON(&req)

	// 如果解析请求参数时出现错误，将错误信息打印到日志中，
	// 并通过 rsp.ResponseWithError 方法返回带有错误信息的 HTTP 响应。
	if err != nil {
		log.Errorf("request json err %v", err)
		rsp.ResponseWithError(c, CodeBodyBindErr, err.Error())
		return
	}

	// 如果没有解析错误，则调用名为 Register 的服务函数处理注册业务逻辑。
	// 如果处理过程中发生错误，将错误信息通过 rsp.ResponseWithError 方法返回给客户端。
	if err := service.Register(req); err != nil {
		rsp.ResponseWithError(c, CodeRegisterErr, err.Error())
		return
	}

	// 如果注册逻辑执行成功，
	// 调用 rsp.ResponseSuccess 方法返回一个表示成功的 HTTP 响应给客户端
	rsp.ResponseSuccess(c)
}

// Login 登录
func Login(c *gin.Context) {
	// req 存储登录请求的信息
	req := &service.LoginRequest{}

	// rsp 存储 HTTP请求的响应信息
	rsp := &HttpResponse{}

	// ShouldBindJSON(&req) 解析 JSON 格式的请求体，并将解析结果存储在 req 变量中。
	// 如果解析过程中发生错误，错误信息将存储在 err 变量中
	err := c.ShouldBindJSON(&req)
	if err != nil {
		log.Errorf("request json err %v", err)
		rsp.ResponseWithError(c, CodeBodyBindErr, err.Error())
		return
	}

	// 生成一个唯一的 uuid
	// 这里使用用户名和当前时间拼接后进行 MD5 哈希算法生成
	uuid := utils.Md5String(req.UserName + time.Now().GoString())

	// 将生成的 uuid 存入上下文（context）中，以便后续使用
	ctx := context.WithValue(context.Background(), "uuid", uuid)

	// 输出登录的开始日志，记录用户名和密码
	log.Infof("loggin start,user:%s, password:%s", req.UserName, req.PassWord)

	// 调用 service.Login 函数（检查用户名和密码是否正确）
	// 如果登录失败，将返回错误信息，并使用 rsp 对象构建错误响应
	session, err := service.Login(ctx, req)
	if err != nil {
		rsp.ResponseWithError(c, CodeLoginErr, err.Error())
		return
	}

	// 登陆成功，使用 c.SetCookie 函数设置一个名为 constant.SessionKey 的 Cookie。下面是函数参数介绍：
	// SessionKey 是 cookie 的名称; session 是登录成功生成的 session 值; CookieExpire 决定了 cookie 的有效期
	// “/” 是 cookie 的路径（表示该 Cookie 对所有路径都有效）; "" 是 cookie 的域名（空字符串表示该 Cookie 对所有域名都有效）
	// false 是指定该 Cookie 只能通过 HTTP 协议传输，不能通过 JavaScript 访问; true 是指定该 Cookie 在安全的 HTTPS 连接中也会被传输
	c.SetCookie(constant.SessionKey, session, constant.CookieExpire, "/", "", false, true)

	// 如果注册逻辑执行成功，
	// 调用 rsp.ResponseSuccess 方法返回一个表示成功的 HTTP 响应给客户端
	rsp.ResponseSuccess(c)
}

// Logout 登出
func Logout(c *gin.Context) {
	// 获取请求中的名为 SessionKey 的 cookie 值，并赋值给 session 变量
	session, _ := c.Cookie(constant.SessionKey)

	// 创建一个 SessionKey 和 session 的键值对到上下文中
	ctx := context.WithValue(context.Background(), constant.SessionKey, session)

	// req 存放登出请求的结构体对象指针
	req := &service.LogoutRequest{}

	// rsp 存储 HTTP请求的响应信息
	rsp := &HttpResponse{}

	// ShouldBindJSON(&req) 解析 JSON 格式的请求体，并将解析结果存储在 req 变量中
	err := c.ShouldBindJSON(req)
	if err != nil {
		log.Errorf("bind get logout request json err %v", err)
		rsp.ResponseWithError(c, CodeBodyBindErr, err.Error())
		return
	}

	// 生成一个唯一的 uuid
	uuid := utils.Md5String(req.UserName + time.Now().GoString())

	// 更新上下文对象，把 uuid 存储进去（这样上下文就存在两个键值对了）
	ctx = context.WithValue(ctx, "uuid", uuid)

	// 实现 Logout() 登出操作的具体逻辑
	if err := service.Logout(ctx, req); err != nil {
		rsp.ResponseWithError(c, CodeLogoutErr, err.Error())
		return
	}

	// 设置一个过期时间为负值的 Cookie，实现删除客户端浏览器中存储的会话标识的目的，即实现用户的登出操作
	c.SetCookie(constant.SessionKey, session, -1, "/", "", false, true)
	rsp.ResponseSuccess(c)
}

// GetUserInfo 获取用户信息
func GetUserInfo(c *gin.Context) {
	// 从 HTTP 请求的查询参数中获取用户名 userName
	userName := c.Query("username")

	// 从 HTTP 请求中获取会话标识 session
	session, _ := c.Cookie(constant.SessionKey)

	// 创建了个上下文，往里面存了个键值对
	ctx := context.WithValue(context.Background(), constant.SessionKey, session)

	// req 存的是获取用户请求的结构体对象，顺带初始化了 UserName 字段
	req := &service.GetUserInfoRequest{
		UserName: userName,
	}

	// rsp 存储 HTTP请求的响应信息
	rsp := &HttpResponse{}

	// 生成一个唯一的 uuid
	uuid := utils.Md5String(req.UserName + time.Now().GoString())

	// 把 uuid 也存进上下文
	ctx = context.WithValue(ctx, "uuid", uuid)

	// 从缓存中获取用户信息
	userInfo, err := service.GetUserInfo(ctx, req)
	if err != nil {
		rsp.ResponseWithError(c, CodeGetUserInfoErr, err.Error())
		return
	}

	// 这个返回函数给客户端多返回了一个 data 也就是实际的数据（从缓存中获取的用户信息）
	rsp.ResponseWithData(c, userInfo)
}

// UpdateNickName 更新用户昵称
func UpdateNickName(c *gin.Context) {
	// 存放修改用户信息返回的结构体对象
	req := &service.UpdateNickNameRequest{}

	// rsp 存储 HTTP请求的响应信息
	rsp := &HttpResponse{}

	// 解析 JSON 格式的请求体，并将解析结果存储在 req 变量中
	err := c.ShouldBindJSON(req)
	if err != nil {
		log.Errorf("bind update user info request json err %v", err)
		rsp.ResponseWithError(c, CodeBodyBindErr, err.Error())
		return
	}

	// 获取 Cookie 值
	session, _ := c.Cookie(constant.SessionKey)
	log.Infof("UpdateNickName|session=%s", session)

	// 创建了个上下文，往里面存了个键值对
	ctx := context.WithValue(context.Background(), constant.SessionKey, session)

	// 生成唯一的 uuid，也存进上下文中
	uuid := utils.Md5String(req.UserName + time.Now().GoString())
	ctx = context.WithValue(ctx, "uuid", uuid)

	// 更改用户信息
	if err := service.UpdateUserNickName(ctx, req); err != nil {
		rsp.ResponseWithError(c, CodeUpdateUserInfoErr, err.Error())
		return
	}
	rsp.ResponseSuccess(c)
}
