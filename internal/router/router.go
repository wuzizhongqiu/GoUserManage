package router

import (
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	api "gouse/api/http/v1"
	"gouse/config"
	"gouse/pkg/constant"
	"net/http"
	"strconv"
)

// InitRouterAndServe 路由配置、启动服务
func InitRouterAndServe() {
	// 根据不同的环境（例如：开发环境、测试环境、生产环境）设置不同的模式。
	setAppRunMode()

	// 创建了一个默认的 gin 路由实例 r，用于处理请求和路由。
	r := gin.Default()

	// 下面是注册路由处理函数，包括用户注册、用户登录、用户登出、获取用户信息、更新用户信息等。
	// 当接收到 /ping GET 请求时，调用 api.Ping 函数来处理请求。(健康检查)
	r.GET("ping", api.Ping)

	// 用户注册
	r.POST("/user/register", api.Register)

	// 用户登录
	r.POST("/user/login", api.Login)

	// 用户登出
	r.POST("/user/logout", AuthMiddleWare(), api.Logout)

	// 获取用户信息
	r.GET("/user/get_user_info", AuthMiddleWare(), api.GetUserInfo)

	// 更新用户信息
	r.POST("/user/update_nick_name", AuthMiddleWare(), api.UpdateNickName)

	// 设置静态文件的路由，这里将 /static/ 映射到 ./web/static/ 目录，即 /static/ 为静态文件资源的访问路径。
	r.Static("/static/", "./web/static/")

	// 设置上传图片文件的路由，将 /upload/images/ 映射到 ./web/upload/images/ 目录，即 /upload/images/ 为已上传图片文件的访问路径。
	r.Static("/upload/images/", "./web/upload/images/")

	// 启动 server 的操作
	// 获取全局配置文件中的应用程序端口号（到时候会去配置文件看一眼）
	port := config.GetGlobalConf().AppConfig.Port

	// r.Run 启动了 HTTP服务器，并监听指定的端口号
	// strconv.Itoa(port)) 将整数类型的端口号转换为字符串类型
	// 所以这段代码就是 r.Run(":8080") (假设端口号是 8080)
	if err := r.Run(":" + strconv.Itoa(port)); err != nil {
		log.Error("start server err:" + err.Error())
	}
}

// setAppRunMode 设置运行模式
func setAppRunMode() {
	// 根据我们设置的配置信息，设置代码的运行模式（这里是设置成 release 模式）
	if config.GetGlobalConf().AppConfig.RunMode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
}

// 这是一个用于对请求进行身份验证的中间件函数
// 补充知识：gin.HandlerFunc的参数为 *gin.Context
func AuthMiddleWare() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 补充：c.Cookie(）就是根据请求的 SessionKey 获取 Cookie 值
		// 获取请求中的名为 SessionKey 的 cookie 值，并赋值给 session 变量
		if session, err := c.Cookie(constant.SessionKey); err == nil {
			// 如果没有出现错误且 session 不为空，说明存在有效的 session
			// 则调用 c.Next() 继续处理后续的请求处理函数，即允许通过该中间件
			// 补充知识：c.Next() 的作用是能将多个中间件串联起来调用
			if session != "" {
				c.Next()
				return
			}
		}

		// 补充知识：
		// JSON 方法是 gin.Context 的方法，用于设置响应的内容为 JSON 格式
		// 它接收两个参数：HTTP 状态码和要发送的 JSON 数据
		// 1）http.StatusUnauthorized 是一个 net/http 包提供的常量，表示 HTTP 状态码 401（未授权）
		// 2）gin.H{"error": "err"} 是一个 map 类型的数据，表示要发送的 JSON 数据
		// 发送的 JSON 数据中包含一个名为 error 的字段，其值为 "err"
		// 最终 Gin 框架会将 HTTP 状态码设置为 401（未授权），并将 JSON 数据发送给客户端作为响应
		// 作用：
		// 如果没有找到或者 session 为空，则返回一个未授权的错误响应
		c.JSON(http.StatusUnauthorized, gin.H{"error": "err"})

		// c.Abort() 是一个用于终止请求的函数，它可以停止请求链的继续处理，
		// 确保本次请求不再继续向后执行其他的中间件或请求处理函数
		c.Abort()
		return
	}
}
