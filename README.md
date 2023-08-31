# GoUser



## 介绍
项目介绍：GoUser 是基于Go开发的支持账号注册，账号登陆等用户操作的用户管理系统



## 项目部署

直接拉代码运行即可，暂时只支持本地运行，不支持远程访问，

具体部署细节暂无（需要链接一个 MySQL 和一个 Redis 数据库）

本地访问：[localhost:8080/static/register.html](http://localhost:8080/static/register.html)

 

## 项目学习目标

通过这个项目学习 golang gin，gorm 框架的使用，Redis 缓存的应用。



## 项目文档

# 1. 实现 Ping 健康检查操作

## 1、代码实现部分

拿到一个项目的源码，我一般会从 main 函数，也就是程序开始的位置开始阅读，跟随整个代码的运行逻辑，或者说运行周期，然后跟着读完整个流程，抓住整个项目的核心代码，抓住他的核心模块。（ cmd\main.go ）

```Go
package main

import (
   "gouse/config"
   "gouse/internal/router"
)

func Init() {
   // 调用了 config 包中的 InitConfig 函数，用于初始化日志信息。
   config.InitConfig()
}

func main() {
   Init()                      // 初始化日志信息
   router.InitRouterAndServe() // 初始化路由并启动 HTTP服务器。
}
```

这个就是 main 方法的实现，其实就只有两行：

1）Init() 是初始化日志信息，实现还是挺复杂的，我就直接先跳过这个具体的实现逻辑；

2）然后是 router.InitRouterAndServe()，这个是初始化路由并启动 HTTP服务器，那我们就跟着这个脉络继续往下看。( internal\router\router.go )

```Go
package router

import (
   "github.com/gin-gonic/gin"
   log "github.com/sirupsen/logrus"
   api "gouse/api/http/v1"
   "gouse/config"
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

   // 启动 server 的操作
   // 获取全局配置文件中的应用程序端口号（到时候会去配置文件看一眼）
   port := config.GetGlobalConf().AppConfig.Port

   // r.Run 启动了HTTP服务器，并监听指定的端口号
   // strconv.Itoa(port)) 将整数类型的端口号转换为字符串类型
   // 所以这段代码就是 r.Run(":8080") (假设端口号是 8080)
   if err := r.Run(":" + strconv.Itoa(port)); err != nil {
      log.Error("start server err:" + err.Error())
   }
}
```

这里就是启动 HTTP服务的具体逻辑，

1）首先是根据环境设定运行模式（这里我待会马上讲这个实现）；

2）然后是创建一个默认的 gin 路由实例 r，用于处理请求和路由；

3）接下来是我们要实现的处理主逻辑（这里其实我们就能知道，这一个函数的内容，就是整个项目的核心代码），这里我为了让我们的项目能尽快先跑起来，就先只实现一个 Ping 方法（待会我就讲这个 Ping 方法的实现）；

4）然后就是根据全局配置文件信息拿到端口号（之后会讲全局服务配置信息的实现）；

5）最后是根据我们刚刚拿到的端口号启动 HTTP服务器，并监听指定的端口号。

**先来讲根据环境设定的运行模式的实现。( internal\router\router.go )**

```Go
// setAppRunMode 设置运行模式
func setAppRunMode() {
   // 根据我们设置的配置信息，设置代码的运行模式（这里是设置成 release 模式）
   if config.GetGlobalConf().AppConfig.RunMode == "release" {
      gin.SetMode(gin.ReleaseMode)
   }
}
```

根据我们的全局服务配置信息来确定代码的运行模式，好我们就从这里出发，来实现全局服务配置信息，这里我们需要的是这个服务的运行模式。

**全局服务配置信息**

首先是 main 方法里的日志信息，我们暂时跳过。( config\config.go )

```Go
// InitConfig 初始化日志
func InitConfig() {

}
```

然后是获取全局服务信息的方法。( config\config.go )

```Go
package config

import (
   log "github.com/sirupsen/logrus"

   "sync"

   "github.com/spf13/viper"
)

// 这两个变量都没有被赋初值，它们的零值将会被使用，
// 即 config 的零值为 GlobalConfig 结构体的零值，once 的零值为默认的初始状态。
var (
   config GlobalConfig // 全局业务配置文件
   once   sync.Once    // sync.Once 是一个并发安全的标志，用于确保某个函数只会被执行一次
)

// GetGlobalConf 获取全局配置文件
func GetGlobalConf() *GlobalConfig {
   // 通过 sync.Once 包的 Do 方法，确保 readConf 函数只会被执行一次
   // 我就直接理解成一个线程安全的单例设计模式了
   once.Do(readConf)
   return &config
}
```

首先我们需要一个全局的变量 config GlobalConfig，我们这里是采取了一个单例的设计模式来初始化我们的全局服务信息，这个时候我们在欠缺了两个重要的东西，一个是这个结构体类型的全局变量的结构体的具体实现，一个是 readConf 函数的实现。

首先是这个全局变量。( config\config.go )

```Go
// GlobalConfig 业务配置结构体
type GlobalConfig struct {
   AppConfig AppConf `yaml:"app" mapstructure:"app"` // 服务配置
}
```

它里面得有我们需要的信息对吧，所以我们就在里面存上服务配置信息的结构体，接着就是实现服务配置的结构体了。( config\config.go )

```Go
// 补充知识：标签
// yaml:"app_name"：这是一个用于 yaml 包的标签。
// 它告诉编译器在解析一个 YAML 配置文件时，将字段名 app_name 映射到结构体字段 AppName 上。
// 也就是说，当从 YAML 配置文件读取值时，将使用 app_name 作为字段的键，将对应的值赋给 AppName 字段。
//
// mapstructure:"app_name"：这是一个用于 mapstructure 包的标签。
// mapstructure 提供了一种方式将一个 map 或 struct 赋值给另一个 struct 的字段，并且可以对字段名进行自定义映射。
// 在这个标签中，app_name 被用作字段名的映射规则，将对应的值赋给 AppName 字段。
//
// 也就是说，设置这两个标签可以使得 AppName 字段支持两种不同的赋值方式:
//
//当使用 yaml 解析配置文件时，解析器会将配置文件中的 app_name 键对应的值赋给 AppName 字段。
//当使用 mapstructure 将一个 map 或 struct 赋值给结构体时，如果该 map 或 struct 中存在 app_name 键，那么其对应的值也会被赋给 AppName 字段。

// AppConf 服务配置
// 这里我就简单记一下，服务配置包含：业务名称，版本号，端口号，运行模式
// 其实并不是这个服务配置需要这几项，主要是你需要用到的信息，以及你需要在配置文件 app.yml 里面写好
type AppConf struct {
   AppName string `yaml:"app_name" mapstructure:"app_name"` // 业务名
   Version string `yaml:"version" mapstructure:"version"`   // 版本号
   Port    int    `yaml:"port" mapstructure:"port"`         // 端口号
   RunMode string `yaml:"run_mode" mapstructure:"run_mode"` // 运行模式
}
```

这里的服务配置包含了 4 个信息，根据之前写的程序可以知道，我们比较急迫需要的其实就是端口号和运行模式，所以我们肯定是需要这两个字段，业务名和版本号源码中，所以我也加上了。那接下来就是读取解析这个配置信息了。( config\config.go )

```Go
// 读取配置信息
func readConf() {
   // 使用 viper 包设置配置文件的名称为 "app"，格式为 YAML
   // 我们前面通过标签的设置支持了两种格式，其中一种就是 YAML 格式（也就是这里使用的格式）
   viper.SetConfigName("app")
   viper.SetConfigType("yml")

   // 然后添加配置文件的搜索路径，包括当前目录，当前目录下的 "conf" 目录，以及上级目录下的 "conf" 目录。
   viper.AddConfigPath(".")
   viper.AddConfigPath("./conf")
   viper.AddConfigPath("../conf")

   // 读取配置文件
   err := viper.ReadInConfig()
   if err != nil {
      panic("read config file err:" + err.Error())
   }

   // 将读取到的配置信息解析为 config 变量对应的结构体类型
   err = viper.Unmarshal(&config)
   if err != nil {
      panic("config file unmarshal err:" + err.Error())
   }

   // 通过日志输出 config 变量的内容，以便于确认配置文件是否被正确读取和解析
   log.Infof("config === %+v", config)
}
```

这里我们需要在 conf 包下实现一个配置文件，包含 app 的配置信息。( conf\app.yml )

```YAML
# 服务的配置信息
app:
  app_name: "gouse" # 应用名称
  version: "v1.0.1" # 版本
  port: 8080        # 服务启用端口
  run_mode: release # 可选dev、release模式
```

这样，我们的全局服务信息就实现完成了，现在我们就能直接通过调用这个 GetGlobalConf() 函数，就能取到我们实现的四个服务配置信息，就那那个运行模式举例：config.GetGlobalConf().AppConfig.RunMode 这样调用就能轻松取到~

**接下来我们重新返回到项目代码的核心部分，下一步就是实现 Ping 方法了。**( api\http\v1\api.go )

```Go
package v1

import (
   "encoding/json"
   "fmt"
   "github.com/gin-gonic/gin"
   "gouse/config"
   "net/http"
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
```

这样我们能先把项目程序跑起来了~

## 2、知识学习部分

重点知识涵盖：gin 框架的使用，如何格式化成 JSON 子串，如何格式化成字符串，golang 的单例实现方法，golang 标签相关知识，YAML 相关，viper 包的使用实例，logrus 库的使用（一个日志库）。

### 1）gin 框架的使用

```Go
// 创建了一个默认的 gin 路由实例 r，用于处理请求和路由
r := gin.Default()
// 当接收到 /ping GET 请求时，调用 api.Ping 函数来处理请求
r.GET("ping", api.Ping)
// r.Run 启动了 HTTP服务器，并监听指定的端口号
// strconv.Itoa(port)) 将整数类型的端口号转换为字符串类型
// 所以这段代码就是 r.Run(":8080") (假设端口号是 8080)
if err := r.Run(":" + strconv.Itoa(port)); err != nil {
   log.Error("start server err:" + err.Error())
}
// setAppRunMode 设置运行模式
func setAppRunMode() {
   // 根据我们设置的配置信息，设置代码的运行模式（这里是设置成 release 模式）
   if config.GetGlobalConf().AppConfig.RunMode == "release" {
      gin.SetMode(gin.ReleaseMode)
   }
}
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
// 使用 gin.Context 的 String 方法，
// 将 HTTP 状态码设置为 200（OK），并将 appInfo 字符串作为响应内容发送给客户端。
// 所以这段代码的逻辑就是，客户端 Ping 服务器，服务器返回这个响应内容 appInfo 给客户端
c.String(http.StatusOK, appInfo)
```

### 2）格式化成 JSON 子串

```Go
// 将 appConfig 这个全局配置对象转换成格式化的 JSON 字符串
// "" 是 JSON 字符串中的字段间的分隔符，"  " 是每一级缩进的字符串。
confInfo, _ := json.MarshalIndent(appConfig, "", "  ")
```

### 3）如何格式化成字符串

```Go
// 将配置信息和应用程序信息格式化为一个字符串。(这是一个格式化操作)
appInfo := fmt.Sprintf("app_name: %s\nversion: %s\n\n%s"
, appConfig.AppName, appConfig.Version, string(confInfo))
```

### 4）golang 的单例实现方法

```Go
// 这两个变量都没有被赋初值，它们的零值将会被使用，
// 即 config 的零值为 GlobalConfig 结构体的零值，once 的零值为默认的初始状态。
var (
   config GlobalConfig // 全局业务配置文件
   once   sync.Once    // sync.Once 是一个并发安全的标志，用于确保某个函数只会被执行一次
)

// GetGlobalConf 获取全局配置文件
func GetGlobalConf() *GlobalConfig {
   // 通过 sync.Once 包的 Do 方法，确保 readConf 函数只会被执行一次
   // 我就直接理解成一个线程安全的单例设计模式了
   once.Do(readConf)
   return &config
}
```

### 5）golang 的标签

```Go
// 补充知识：标签
// yaml:"app_name"：这是一个用于 yaml 包的标签。
// 它告诉编译器在解析一个 YAML 配置文件时，将字段名 app_name 映射到结构体字段 AppName 上。
// 也就是说，当从 YAML 配置文件读取值时，将使用 app_name 作为字段的键，将对应的值赋给 AppName 字段。
//
// mapstructure:"app_name"：这是一个用于 mapstructure 包的标签。
// mapstructure 提供了一种方式将一个 map 或 struct 赋值给另一个 struct 的字段，并且可以对字段名进行自定义映射。
// 在这个标签中，app_name 被用作字段名的映射规则，将对应的值赋给 AppName 字段。
//
// 也就是说，设置这两个标签可以使得 AppName 字段支持两种不同的赋值方式:
//
//当使用 yaml 解析配置文件时，解析器会将配置文件中的 app_name 键对应的值赋给 AppName 字段。
//当使用 mapstructure 将一个 map 或 struct 赋值给结构体时，如果该 map 或 struct 中存在 app_name 键，那么其对应的值也会被赋给 AppName 字段。

// AppConf 服务配置
// 这里我就简单记一下，服务配置包含：业务名称，版本号，端口号，运行模式
// 其实并不是这个服务配置需要这几项，主要是你需要用到的信息，以及你需要在配置文件 app.yml 里面写好
type AppConf struct {
   AppName string `yaml:"app_name" mapstructure:"app_name"` // 业务名
   Version string `yaml:"version" mapstructure:"version"`   // 版本号
   Port    int    `yaml:"port" mapstructure:"port"`         // 端口号
   RunMode string `yaml:"run_mode" mapstructure:"run_mode"` // 运行模式
}
```

### 6）YAML 配置相关

```Go
// GlobalConfig 业务配置结构体
type GlobalConfig struct {
   AppConfig AppConf `yaml:"app" mapstructure:"app"` // 服务配置
}
# 服务的配置信息
app:
  app_name: "gouse" # 应用名称
  version: "v1.0.1" # 版本
  port: 8080        # 服务启用端口
  run_mode: release # 可选dev、release模式
```

### 7）viper 框架的使用

```Go
// 读取配置信息
func readConf() {
   // 使用 viper 包设置配置文件的名称为 "app"，格式为 YAML
   // 我们前面通过标签的设置支持了两种格式，其中一种就是 YAML 格式（也就是这里使用的格式）
   viper.SetConfigName("app")
   viper.SetConfigType("yml")

   // 然后添加配置文件的搜索路径，
   // 包括当前目录，当前目录下的 "config" 目录，以及上级目录下的 "config" 目录。
   viper.AddConfigPath(".")
   viper.AddConfigPath("./conf")
   viper.AddConfigPath("../conf")

   // 读取配置文件
   err := viper.ReadInConfig()
   if err != nil {
      panic("read config file err:" + err.Error())
   }

   // 将读取到的配置信息解析为 config 变量对应的结构体类型
   err = viper.Unmarshal(&config)
   if err != nil {
      panic("config file unmarshal err:" + err.Error())
   }

   // 通过日志输出 config 变量的内容，以便于确认配置文件是否被正确读取和解析
   log.Infof("config === %+v", config)
}
```

# 2. 实现用户注册

## 1、代码实现部分

**代码的核心部分其实就是在 InitRouterAndServe() 方法里添加上这一句代码。**( internal\router\router.go )

```Go
// 用户注册
r.POST("/user/register", api.Register)
```

第一个路由参数为什么填的是 "/user/register" ，我去看了一下前端的代码，虽然看不懂，但是大概知道应该是前端设定好了路由，所以填了这个。我把前端的这段代码拉取下来了（总之把源码中的前端部分 CV 过来就行）

```JavaScript
$.ajax({
  type: "POST",
  dataType: "json",
  url: urlPrefix + '/user/register',
  contentType: "application/json",
  data: JSON.stringify({
    "user_name": username.value,
    "pass_word": passwd.value,
    "age": parseInt(age.value),
    "gender": gender.value,
    "nick_name": nickname.value,
  }),
```

应该就是这部分设定好的路径。

然后是第二个参数，实现 api.Register 方法，实现注册的逻辑即可。

我们从这段源码慢慢分析。( api\http\v1\api.go )

```Go
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
```

**首先是设置一个存储注册请求字段的结构体。**( internel\service\entity.go )

```Go
package service

// RegisterRequest 注册请求
type RegisterRequest struct {
   UserName string `json:"user_name"`
   Password string `json:"pass_word"`
   Age      int    `json:"age"`
   Gender   string `json:"gender"`
   NickName string `json:"nick_name"`
}
```

这样 req 变量就获取到这样一个结构体指针了

**然后是设置一个结构体存储 http的请求的返回信息。**( api\http\v1\entity.go )

```Go
// HttpResponse http独立请求返回结构体
type HttpResponse struct {
   Code ErrCode     `json:"code"`
   Msg  string      `json:"msg"`
   Data interface{} `json:"data"`
}
```

这里我们顺便把错误码也给定义好。( api\http\v1\entity.go )

```Go
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
```

**然后是解析 JSON 格式的请求**，这个应该就是前后端交互的部分，我们去处理前端发来的信息，存进 req 中。

接着是差错处理，我们顺带实现了 ResponseWithError 方法处理错误的 http请求的返回信息。

( api\http\v1\entity.go )

```Go
// ResponseWithError http请求返回处理函数
func (rsp *HttpResponse) ResponseWithError(c *gin.Context, code ErrCode, msg string) {
   // 将错误信息填充到 HttpResponse 结构体中的 Code 和 Msg 字段
   rsp.Code = code
   rsp.Msg = msg

   // 将该结构体通过 JSON 格式返回给客户端
   // 并指定返回的 HTTP 状态码为 500（http.StatusInternalServerError）
   c.JSON(http.StatusInternalServerError, rsp)
}
```

**如果解析没有错误，就调用这里的 service.Register 服务函数处理注册业务逻辑，如果错误就处理错误。**

**最后调用 rsp.ResponseSuccess 方法返回一个表示成功的 HTTP 响应给客户端。**

( api\http\v1\entity.go )

```Go
func (rsp *HttpResponse) ResponseSuccess(c *gin.Context) {
   // 将成功的信息填充到 HttpResponse 结构体中的 Code 和 Msg 字段
   rsp.Code = CodeSuccess
   rsp.Msg = "success"

   // 将该结构体通过 JSON 格式返回给客户端
   //  并指定返回的 HTTP 状态码为 200（http.StatusOK）
   c.JSON(http.StatusOK, rsp)
}
```

**现在的代码的核心就在 service.Register 方法的实现上了**，

这里我再次把整体的一个实现先贴出来，慢慢讲解。( internel\service\user.go )

```Go
package service

import (
   "fmt"
   log "github.com/sirupsen/logrus"
   "gouse/internal/dao"
   "gouse/internal/model"
   "gouse/pkg/constant"
   "gouse/utils"
)

// Register 用户注册
func Register(req *RegisterRequest) error {
   // 对接收到的请求参数 req 进行检查，
   // 用户名，密码不能为空，年龄不能 <= 0 岁，判断性别是否输入正确（暂时只支持男和女）
   if req.UserName == "" || req.Password == "" || req.Age <= 0 || !utils.Contains([]string{constant.GenderMale, constant.GenderFeMale}, req.Gender) {
      log.Errorf("register param invalid")
      return fmt.Errorf("register param invalid")
   }

   // 调用 dao.GetUserByName(req.UserName) 方法，根据用户名查询数据库，判断用户是否已经存在
   existedUser, err := dao.GetUserByName(req.UserName)
   // 如果查询时出现错误:
   if err != nil {
      log.Errorf("Register|%v", err)
      return fmt.Errorf("register|%v", err)
   }
   // 如果该用户已经存在:
   if existedUser != nil {
      log.Errorf("用户已经注册,user_name==%s", req.UserName)
      return fmt.Errorf("用户已经注册，不能重复注册")
   }

   // 创建一个用户对象，包含相应的属性
   user := &model.User{
      Name:     req.UserName,
      Age:      req.Age,
      Gender:   req.Gender,
      PassWord: req.Password,
      NickName: req.NickName,

      CreateModel: model.CreateModel{
         Creator: req.UserName,
      },
      ModifyModel: model.ModifyModel{
         Modifier: req.UserName,
      },
   }

   // 打印日志信息
   log.Infof("user ====== %+v", user)

   // 将新的用户对象存储到数据库中，
   // 如果在存储过程中出现错误，会将错误信息赋值给变量 err
   if err := dao.CreateUser(user); err != nil {
      log.Errorf("Register|%v", err)
      return fmt.Errorf("register|%v", err)
   }

   // 注册成功，返回 nil
   return nil
}
```

**首先是检查我们从前端那里接收到的请求参数**（也就是我们填写在表单上的数据），看看是否符合要求，因为代码比较长，我就还是把一点点源代码放出来分析。说的就是这一段代码：

```Go
// 对接收到的请求参数 req 进行检查，
// 用户名，密码不能为空，年龄不能 <= 0 岁，判断性别是否输入正确（暂时只支持男和女）
if req.UserName == "" || req.Password == "" || req.Age <= 0 || !utils.Contains([]string{constant.GenderMale, constant.GenderFeMale}, req.Gender) {
   log.Errorf("register param invalid")
   return fmt.Errorf("register param invalid")
}
```

这里我们封装了 utils.Contains 还有一些全局变量来辅助我们做判断。

( utils\utils.go )

```Go
// 判断字符串 tg 是否能在字符串类型的切片 source 中找到
func Contains(source []string, tg string) bool {
   // 使用 range 关键字获取每个元素的索引和值，
   // 并分别赋值给变量 _（省略符号）和变量 s
   for _, s := range source {
      if s == tg {
         return true
      }
   }
   return false
}
```

( pkg\constant\const.go )

```Go
const (
   GenderMale   = "male"
   GenderFeMale = "female"
)
```

**接着是调用 dao.GetUserByName() 方法，根据用户名查询数据库，判断用户是否已经存在**

```Go
existedUser, err := dao.GetUserByName(req.UserName)
```

我们来看看这个方法的具体实现是什么样的

```Go
// GetUserByName 根据姓名获取用户
func GetUserByName(name string) (*model.User, error) {
   // user 变量存储查询结果
   user := &model.User{}

   // 使用 utils.GetDB() 方法获取一个 *gorm.DB 类型的数据库连接，
   // 通过 Model 方法指定要操作的数据模型为 model.User，
   // 使用 Where 方法指定查询条件为 name=?，并且将查询结果存储到 user 中
   if err := utils.GetDB().Model(model.User{}).Where("name=?", name).First(user).Error; err != nil {
      // 表示根据姓名未找到对应的用户
      if err.Error() == gorm.ErrRecordNotFound.Error() {
         return nil, nil
      }
      log.Errorf("GetUserByName fail:%v", err)
      return nil, fmt.Errorf("GetUserByName fail:%v", err) // 这里报错
   }
   return user, nil
}
```

**首先我们得有一个结构体来存储用户的相关信息**

```Go
user := &model.User{}
```

这里我们正在使用的是 golang 的 gorm 框架，用来和数据交互的，这个框架该怎么使用的呢？

( internel\model\model.go )

```Go
import "time"

// 补充知识：gorm 标签
// gorm 标签被用于定义数据库表的列名和列属性，
// 以便 gorm 库可以根据这些标签来自动生成数据库表结构以及执行相关的查询操作。
//
// 以下面的代码为例：
//
// 1）gorm:"type:varchar(100);not null;default ''"：
// 该标签定义了数据库表的列属性。
// type:varchar(100) 表示该字段的数据类型为 varchar，长度为 100，
// not null 表示该字段不能为空；default '' 表示该字段的默认值为空字符串。
//
// 2）gorm:"autoCreateTime"：
// 该标签用于指定在创建记录时自动生成时间。
// autoCreateTime 是 gorm 提供的一个特殊标签，用于指示在创建记录时自动设置该字段的值为当前时间。
//
// 3）gorm:"autoUpdateTime"：
// 该标签用于指定在更新记录时自动生成时间。
// autoUpdateTime 是 gorm 提供的一个特殊标签，用于指示在更新记录时自动设置该字段的值为当前时间。
//
// 4）gorm:"column:xxx"：
// 该标签用于指定数据库表列的名称。
// column:xxx 中的 xxx 表示该字段在数据库表中的实际列名。

// CreateModel 内嵌model
type CreateModel struct {
   Creator    string    `gorm:"type:varchar(100);not null;default ''"`
   CreateTime time.Time `gorm:"autoCreateTime"` // 在创建记录时自动生成时间
}

// ModifyModel 内嵌model
type ModifyModel struct {
   Modifier   string    `gorm:"type:varchar(100);not null;default ''"`
   ModifyTime time.Time `gorm:"autoUpdateTime"` // 在更新记录时自动生成时间
}

// User 用户
type User struct {
   CreateModel
   ModifyModel
   ID       int    `gorm:"column:id"`       // ID
   Name     string `gorm:"column:name"`     // 姓名
   Gender   string `gorm:"column:gender"`   // 性别
   Age      int    `gorm:"column:age"`      // 年龄
   PassWord string `gorm:"column:password"` // 密码
   NickName string `gorm:"column:nickname"` // 昵称
}
```

这里我们需要中途补充一点，就是关于数据库表设计 ( 在我们使用的数据库中，建一个表 )，为什要在这里提这个呢？因为我们前面代码结构体字段的设计就是根据数据库来实现的。

```Go
use camps_user;
create table if not exists users(
   `id` int not null auto_increment,
   `name` varchar(100) not null,
   `age` int not null,
   `gender` varchar(30) not null,
   `password` varchar(255) not null default '',
   `nickname` varchar(100) not null default '',
   `head_url` varchar(1024) not null default '',
   `create_time` timestamp null default current_timestamp comment '创建时间',
   `creator` varchar(100) not null default '',
   `modify_time` timestamp null default current_timestamp on update current_timestamp comment '最后一次修改时间',
   `modifier` varchar(100) not null default '',
   primary key ( id )
);
```

**然后就是根据数据库的连接调用一个查询方法，根据名字查询是否存在这样一个用户**

```Go
if err := utils.GetDB().Model(model.User{}).Where("name=?", name).First(user).Error; err != nil {
   // 表示根据姓名未找到对应的用户
   if err.Error() == gorm.ErrRecordNotFound.Error() {
      return nil, nil
   }
   log.Errorf("GetUserByName fail:%v", err)
   return nil, fmt.Errorf("GetUserByName fail:%v", err) // 这里报错
}
return user, nil
```

但是我们现在又需要去做一件事情，就是去获取数据库的连接，实现 utils.GetDB() 方法

**获取数据库的连接** (utils\db.go)

```Go
var (
   db     *gorm.DB // 全局变量 db，用于保存数据库连接
   dbOnce sync.Once
)

// openDB 连接db
func openDB() {
   // 从全局配置中获取 MySQL 配置信息 mysqlConf
   mysqlConf := config.GetGlobalConf().DbConfig

   // 使用 fmt.Sprintf 格式化连接参数 connArgs，拼接用户名、密码、主机、端口和数据库名称等连接数据库所需的信息。
   connArgs := fmt.Sprintf("%s:%s@(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local", mysqlConf.User,
      mysqlConf.Password, mysqlConf.Host, mysqlConf.Port, mysqlConf.Dbname)
   log.Info("mdb addr:" + connArgs) // 打个日志，输出拼接后的信息

   var err error
   // 调用 gorm.Open() 方法打开与 MySQL 数据库的连接，并将连接结果赋值给全局变量 db（gorm 库的使用）
   db, err = gorm.Open(mysql.Open(connArgs), &gorm.Config{})
   if err != nil {
      panic("failed to connect database")
   }

   // 通过 db.DB() 方法获取 *sql.DB 类型的数据库连接对象 sqlDB
   sqlDB, err := db.DB()
   if err != nil {
      panic("fetch db connection err:" + err.Error())
   }

   sqlDB.SetMaxIdleConns(mysqlConf.MaxIdleConn)                                        //设置最大空闲连接
   sqlDB.SetMaxOpenConns(mysqlConf.MaxOpenConn)                                        //设置最大打开的连接
   sqlDB.SetConnMaxLifetime(time.Duration(mysqlConf.MaxIdleTime * int64(time.Second))) //设置空闲时间为(s)
}

// GetDB 获取数据库连接
func GetDB() *gorm.DB {
   dbOnce.Do(openDB)
   return db
}
```

获取数据库连接的单例是很好实现，但是，我们还需要做一件事情，先跟着 openDB 方法的实现逻辑走一走吧。

首先就是从全局配置中获取 MySQL 配置信息 mysqlConf，那我们又得去更新一下我们的全局配置了

**在全局配置中添加数据库相关配置信息** ( config\config.go )

```Go
// DbConf 数据库配置结构
// 补充：
// 空闲连接是指：处于连接池中且当前没有被使用的数据库连接
// 连接最大空闲时间是指：连接可以保持空闲的最长时间
type DbConf struct {
   Host        string `yaml:"host" mapstructure:"host"`                   // db主机地址
   Port        string `yaml:"port" mapstructure:"port"`                   // db端口
   User        string `yaml:"user" mapstructure:"user"`                   // 用户名
   Password    string `yaml:"password" mapstructure:"password"`           // 密码
   Dbname      string `yaml:"dbname" mapstructure:"dbname"`               // db名
   MaxIdleConn int    `yaml:"max_idle_conn" mapstructure:"max_idle_conn"` // 最大空闲连接数
   MaxOpenConn int    `yaml:"max_open_conn" mapstructure:"max_open_conn"` // 最大打开的连接数
   MaxIdleTime int64  `yaml:"max_idle_time" mapstructure:"max_idle_time"` // 连接最大空闲时间
}
```

我们的全局业务配置结构体也需要新增数据库配置信息字段

```Go
// GlobalConfig 业务配置结构体
type GlobalConfig struct {
   AppConfig AppConf `yaml:"app" mapstructure:"app"` // 服务配置
   DbConfig  DbConf  `yaml:"db" mapstructure:"db"`   // 数据库配置
}
```

不要忘了配置文件，这样我们的配置信息就完成了

```YAML
# 数据库的配置
db:
  host: "0.0.0.0"     # host
  port: 8086          # port
  user: "root"        # user
  password: "123456"  # password
  dbname: "camps_user"    # dbname
  max_idle_conn: 5    # 最大空闲连接数
  max_open_conn: 20   # 最大连接数
  max_idle_time: 300  # 最大空闲时间
```

**然后使用 fmt.Sprintf 格式化连接参数 connArgs，拼接用户名、密码、主机、端口和数据库名称等连接数据库所需的信息，为连接数据库做准备工作。**

```Go
// 使用 fmt.Sprintf 格式化连接参数 connArgs，拼接用户名、密码、主机、端口和数据库名称等连接数据库所需的信息。
connArgs := fmt.Sprintf("%s:%s@(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local", mysqlConf.User,
   mysqlConf.Password, mysqlConf.Host, mysqlConf.Port, mysqlConf.Dbname)
log.Info("mdb addr:" + connArgs) // 打个日志，输出拼接后的信息
```

**接着获取数据库的连接**

```Go
var err error
// 调用 gorm.Open() 方法打开与 MySQL 数据库的连接，并将连接结果赋值给全局变量 db（gorm 库的使用）
db, err = gorm.Open(mysql.Open(connArgs), &gorm.Config{})
if err != nil {
   panic("failed to connect database")
}
```

**获取数据库连接对象**

```Go
// 通过 db.DB() 方法获取 *sql.DB 类型的数据库连接对象 sqlDB
sqlDB, err := db.DB()
if err != nil {
   panic("fetch db connection err:" + err.Error())
}
```

**配置好数据库**

```Go
sqlDB.SetMaxIdleConns(mysqlConf.MaxIdleConn)                                        //设置最大空闲连接
sqlDB.SetMaxOpenConns(mysqlConf.MaxOpenConn)                                        //设置最大打开的连接
sqlDB.SetConnMaxLifetime(time.Duration(mysqlConf.MaxIdleTime * int64(time.Second))) //设置空闲时间为(s)
```

这样我们就成功获取了数据库的连接对象，也完成了查询的操作，让我们再次回到用户注册的逻辑实现。

**接下来就是对查询结果的处理**

如果查询过程出错就报错误信息，这里提两句，对应的数据库表如果没有创建，就会查询出错

如果用户存在，自然也不用创建，就报用户已经存在的错误

```Go
// 如果查询时出现错误:
if err != nil {
   log.Errorf("Register|%v", err)
   return fmt.Errorf("register|%v", err)
}
// 如果该用户已经存在:
if existedUser != nil {
   log.Errorf("用户已经注册,user_name==%s", req.UserName)
   return fmt.Errorf("用户已经注册，不能重复注册")
}
```

**然后就是根据前端传来的数据，创建一个用户的结构体对象**

```Go
// 创建一个用户对象，包含相应的属性
user := &model.User{
   Name:     req.UserName,
   Age:      req.Age,
   Gender:   req.Gender,
   PassWord: req.Password,
   NickName: req.NickName,

   CreateModel: model.CreateModel{
      Creator: req.UserName,
   },
   ModifyModel: model.ModifyModel{
      Modifier: req.UserName,
   },
}
```

**最后就是将创建好的用户对象存储进数据库中**

```Go
// 将新的用户对象存储到数据库中，
// 如果在存储过程中出现错误，会将错误信息赋值给变量 err
if err := dao.CreateUser(user); err != nil {
   log.Errorf("Register|%v", err)
   return fmt.Errorf("register|%v", err)
}
```

这里我们还需要实现一个 CreateUser 方法，就是将该用户对象存储进数据库中

**实现 CreateUser 方法，将该用户对象存储进数据库中** ( internel\dao\user.go )  

```Go
// CreateUser 创建一个用户
func CreateUser(user *model.User) error {
   // 用 Create 方法创建数据库
   if err := utils.GetDB().Model(&model.User{}).Create(user).Error; err != nil {
      log.Errorf("CreateUser fail: %v", err)
      return fmt.Errorf("CreateUser fail: %v", err)
   }
   log.Infof("insert success")
   return nil
}
```

这里其实就是调用数据库连接对象的 Create 方法。

这样我们就成功实现了 service.Register() 方法，实现了用户注册的具体逻辑了。

我们这个时候就能把程序跑起来，然后尝试注册一个用户了~

## 2、知识学习部分

重点知识涵盖：gin 框架中用于前后端交互的方法，错误码的设置，golang 的差错处理，gorm 框架操作MySQL，MySQL的配置以及库表的设计，gorm 框架连接MySQL（单例获取MySQL连接对象）

### 1）gin 框架中的前后端交互

```Go
// ShouldBindJSON(&req) 解析 JSON 格式的请求体，并将解析结果存储在 req 变量中。
// 如果解析过程中发生错误，错误信息将存储在 err 变量中
err := c.ShouldBindJSON(&req)
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
   //  并指定返回的 HTTP 状态码为 200（http.StatusOK）
   c.JSON(http.StatusOK, rsp)
}
```

### 2）错误码的设置

```Go
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
```

### 3）golang 的差错处理

```Go
// 调用 dao.GetUserByName(req.UserName) 方法，根据用户名查询数据库，判断用户是否已经存在
existedUser, err := dao.GetUserByName(req.UserName)
// 如果查询时出现错误:
if err != nil {
   log.Errorf("Register|%v", err)
   return fmt.Errorf("register|%v", err)
}
// 如果该用户已经存在:
if existedUser != nil {
   log.Errorf("用户已经注册,user_name==%s", req.UserName)
   return fmt.Errorf("用户已经注册，不能重复注册")
}
```

### 4）gorm 框架的 Where 方法

```Go
// GetUserByName 根据姓名获取用户
func GetUserByName(name string) (*model.User, error) {
   // user 变量存储查询结果
   user := &model.User{}

   // 使用 utils.GetDB() 方法获取一个 *gorm.DB 类型的数据库连接，
   // 通过 Model 方法指定要操作的数据模型为 model.User，
   // 使用 Where 方法指定查询条件为 name=?，并且将查询结果存储到 user 中
   if err := utils.GetDB().Model(model.User{}).Where("name=?", name).First(user).Error; err != nil {
      // 表示根据姓名未找到对应的用户
      if err.Error() == gorm.ErrRecordNotFound.Error() {
         return nil, nil
      }
      log.Errorf("GetUserByName fail:%v", err)
      return nil, fmt.Errorf("GetUserByName fail:%v", err) // 这里报错
   }
   return user, nil
}
```

### 5）MySQL配置与库表设计

```Go
// DbConf 数据库配置结构
// 补充：
// 空闲连接是指：处于连接池中且当前没有被使用的数据库连接
// 连接最大空闲时间是指：连接可以保持空闲的最长时间
type DbConf struct {
   Host        string `yaml:"host" mapstructure:"host"`                   // db主机地址
   Port        string `yaml:"port" mapstructure:"port"`                   // db端口
   User        string `yaml:"user" mapstructure:"user"`                   // 用户名
   Password    string `yaml:"password" mapstructure:"password"`           // 密码
   Dbname      string `yaml:"dbname" mapstructure:"dbname"`               // db名
   MaxIdleConn int    `yaml:"max_idle_conn" mapstructure:"max_idle_conn"` // 最大空闲连接数
   MaxOpenConn int    `yaml:"max_open_conn" mapstructure:"max_open_conn"` // 最大打开的连接数
   MaxIdleTime int64  `yaml:"max_idle_time" mapstructure:"max_idle_time"` // 连接最大空闲时间
}

// GlobalConfig 业务配置结构体
type GlobalConfig struct {
   AppConfig AppConf `yaml:"app" mapstructure:"app"` // 服务配置
   DbConfig  DbConf  `yaml:"db" mapstructure:"db"`   // 数据库配置
}
# 数据库的配置
db:
  host: "0.0.0.0"     # host
  port: 8086          # port
  user: "root"        # user
  password: "123456"  # password
  dbname: "camps_user"    # dbname
  max_idle_conn: 5    # 最大空闲连接数
  max_open_conn: 20   # 最大连接数
  max_idle_time: 300  # 最大空闲时间
use camps_user;
create table if not exists users(
   `id` int not null auto_increment,
   `name` varchar(100) not null,
   `age` int not null,
   `gender` varchar(30) not null,
   `password` varchar(255) not null default '',
   `nickname` varchar(100) not null default '',
   `head_url` varchar(1024) not null default '',
   `create_time` timestamp null default current_timestamp comment '创建时间',
   `creator` varchar(100) not null default '',
   `modify_time` timestamp null default current_timestamp on update current_timestamp comment '最后一次修改时间',
   `modifier` varchar(100) not null default '',
   primary key ( id )
);
```

### 6）gorm 框架连接 MySQL

```Go
var (
   db     *gorm.DB // 全局变量 db，用于保存数据库连接
   dbOnce sync.Once
)

// openDB 连接db
func openDB() {
   // 从全局配置中获取 MySQL 配置信息 mysqlConf
   mysqlConf := config.GetGlobalConf().DbConfig

   // 使用 fmt.Sprintf 格式化连接参数 connArgs，拼接用户名、密码、主机、端口和数据库名称等连接数据库所需的信息。
   connArgs := fmt.Sprintf("%s:%s@(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local", mysqlConf.User,
      mysqlConf.Password, mysqlConf.Host, mysqlConf.Port, mysqlConf.Dbname)
   log.Info("mdb addr:" + connArgs) // 打个日志，输出拼接后的信息

   var err error
   // 调用 gorm.Open() 方法打开与 MySQL 数据库的连接，并将连接结果赋值给全局变量 db（gorm 库的使用）
   db, err = gorm.Open(mysql.Open(connArgs), &gorm.Config{})
   if err != nil {
      panic("failed to connect database")
   }

   // 通过 db.DB() 方法获取 *sql.DB 类型的数据库连接对象 sqlDB
   sqlDB, err := db.DB()
   if err != nil {
      panic("fetch db connection err:" + err.Error())
   }

   sqlDB.SetMaxIdleConns(mysqlConf.MaxIdleConn)                                        //设置最大空闲连接
   sqlDB.SetMaxOpenConns(mysqlConf.MaxOpenConn)                                        //设置最大打开的连接
   sqlDB.SetConnMaxLifetime(time.Duration(mysqlConf.MaxIdleTime * int64(time.Second))) //设置空闲时间为(s)
}

// GetDB 获取数据库连接
func GetDB() *gorm.DB {
   dbOnce.Do(openDB)
   return db
}
```

### 7）gorm 框架的 Create 方法

```Go
// 将新的用户对象存储到数据库中，
// 如果在存储过程中出现错误，会将错误信息赋值给变量 err
if err := dao.CreateUser(user); err != nil {
   log.Errorf("Register|%v", err)
   return fmt.Errorf("register|%v", err)
}
// CreateUser 创建一个用户
func CreateUser(user *model.User) error {
   // 用 Create 方法创建数据库
   if err := utils.GetDB().Model(&model.User{}).Create(user).Error; err != nil {
      log.Errorf("CreateUser fail: %v", err)
      return fmt.Errorf("CreateUser fail: %v", err)
   }
   log.Infof("insert success")
   return nil
}
```

# 3. 实现用户登录

## 代码实现部分

**代码的核心部分其实就是在 InitRouterAndServe() 方法里添加上这一句代码。**( internal\router\router.go )

```Go
// 用户登录
r.POST("/user/login", api.Login)
```

这里有个我不太明白的地方，是关于设置静态文件的路由和图片的路由，这里我就先直接从源码 CV 过来了

```Go
// 设置静态文件的路由，这里将 /static/ 映射到 ./web/static/ 目录，即 /static/ 为静态文件资源的访问路径。
r.Static("/static/", "./web/static/")

// 设置上传图片文件的路由，将 /upload/images/ 映射到 ./web/upload/images/ 目录，即 /upload/images/ 为已上传图片文件的访问路径。
r.Static("/upload/images/", "./web/upload/images/")
```

r.POST("/user/login", api.Login)，其实我们需要实现的逻辑的就是 api.Login() 函数

来看这个函数的具体实现：（ api\http\v1\api.go ）

```Go
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
```

我们还是一点点慢慢分析。

**首先是登录请求的信息和 HTTP请求的响应信息**

```Go
// req 存储登录请求的信息
req := &service.LoginRequest{}

// rsp 存储 HTTP请求的响应信息
rsp := &HttpResponse{}
```

这里我们就设定一下存储登录请求的结构体字段就好了，HTTP请求的响应我们在用户注册的时候已经完成了 

（ internel\service\entity.go ）

```Go
// LoginRequest 登陆请求
type LoginRequest struct {
   UserName string `json:"user_name"`
   PassWord string `json:"pass_word"`
}
```

**接着是解析 JSON 格式的请求体**，这个也是老操作了（ api\http\v1\api.go ）

```Go
// ShouldBindJSON(&req) 解析 JSON 格式的请求体，并将解析结果存储在 req 变量中。
// 如果解析过程中发生错误，错误信息将存储在 err 变量中
err := c.ShouldBindJSON(&req)
if err != nil {
   log.Errorf("request json err %v", err)
   rsp.ResponseWithError(c, CodeBodyBindErr, err.Error())
   return
}
```

**生成一个唯一的 uuid 并存入上下文**，用于区分不同的用户

```Go
// 生成一个唯一的 uuid
// 这里使用用户名和当前时间拼接后进行 MD5 哈希算法生成
uuid := utils.Md5String(req.UserName + time.Now().GoString())

// 将生成的 uuid 存入上下文（context）中，以便后续使用
ctx := context.WithValue(context.Background(), "uuid", uuid)

// 输出登录的开始日志，记录用户名和密码
log.Infof("loggin start,user:%s, password:%s", req.UserName, req.PassWord)
```

这里我们需要实现一个 MD5 的哈希算法来生成这个 uuid（ utils\utils.go ）

```Go
// 用哈希算法生成一个唯一的 uuid
func Md5String(s string) string {
   // 创建一个 md5 实例
   h := md5.New()

   // 通过调用 h.Write([]byte(s)) 将字符串 s 的字节表示写入到 md5 实例中
   h.Write([]byte(s))

   // 调用 h.Sum(nil) 完成 MD5 的计算并返回结果
   // Sum() 方法接受一个切片参数，用于存储 MD5 值
   // 由于我们在这里不需要对 MD5 值进行其他操作，所以将参数设置为 nil
   // 再将计算得到的 MD5 值以十六进制的形式表示，并由 hex.EncodeToString() 转换为字符串
   // 注意：
   // 这段代码的作用是对给定的字符串进行 MD5 计算，并将计算结果以字符串的形式返回
   // 通常，MD5 值用于验证数据的完整性和唯一性，但请注意，MD5 算法已经被认为不再安全，不适用于密码等敏感数据的加密
   str := hex.EncodeToString(h.Sum(nil))
   return str
}
```

**然后调用 service.Login 函数（检查用户名和密码是否正确，实现登录逻辑）**

```Go
// 调用 service.Login 函数（检查用户名和密码是否正确）
// 如果登录失败，将返回错误信息，并使用 rsp 对象构建错误响应
session, err := service.Login(ctx, req)
if err != nil {
   rsp.ResponseWithError(c, CodeLoginErr, err.Error())
   return
}
```

这里我们就需要去实现 service.Login 函数的具体逻辑，也就是我们代码逻辑实现的核心

**service.Login 函数的具体实现**（ internel\service\user.go ）

```Go
// Login 用户登陆
func Login(ctx context.Context, req *LoginRequest) (string, error) {
   // 从上下文对象中获取请求的唯一标识符(uuid)，并使用 log.Debugf 打印日志表明有用户访问登录功能
   uuid := ctx.Value(constant.ReqUuid)
   log.Debugf(" %s| Login access from:%s,@,%s", uuid, req.UserName, req.PassWord)

   // 调用 getUserInfo 函数，根据 req.UserName 请求中的用户名获取用户信息（user）
   user, err := getUserInfo(req.UserName)
   if err != nil {
      log.Errorf("Login|%v", err)
      return "", fmt.Errorf("login|%v", err)
   }

   // 用户存在，比对输入的密码和用户密码是否一致
   if req.PassWord != user.PassWord {
      log.Errorf("Login|password err: req.password=%s|user.password=%s", req.PassWord, user.PassWord)
      return "", fmt.Errorf("password is not correct")
   }

   // 如果密码匹配成功，调用 utils.GenerateSession 函数生成一个新的 session 字符串
   session := utils.GenerateSession(user.Name)

   // 并调用 cache.SetSessionInfo 函数将用户信息和 session 存储到缓存中
   err = cache.SetSessionInfo(user, session)
   if err != nil {
      log.Errorf(" Login|Failed to SetSessionInfo, uuid=%s|user_name=%s|session=%s|err=%v", uuid, user.Name, session, err)
      return "", fmt.Errorf("login|SetSessionInfo fail:%v", err)
   }

   // 最后，使用 log.Infof 打印登录成功的日志，并返回生成的 session 字符串作为登录成功的标识
   log.Infof("Login successfully, %s@%s with redis_session session_%s", req.UserName, req.PassWord, session)
   return session, nil
}
```

我们再来一点一点分析这个函数。

**首先根据 uuid 打个日志**

```Go
// 从上下文对象中获取请求的唯一标识符(uuid)，并使用 log.Debugf 打印日志表明有用户访问登录功能
uuid := ctx.Value(constant.ReqUuid)
log.Debugf(" %s| Login access from:%s,@,%s", uuid, req.UserName, req.PassWord)
```

**接着调用 getUserInfo 函数，根据 req.UserName 请求中的用户名获取用户信息**（ internel\service\user.go ） 

```Go
// 调用 getUserInfo 函数，根据 req.UserName 请求中的用户名获取用户信息（user）
user, err := getUserInfo(req.UserName)
if err != nil {
   log.Errorf("Login|%v", err)
   return "", fmt.Errorf("login|%v", err)
}
```

**那现在我们就需要实现 getUserInfo 函数，根据用户名获取用户信息**

```Go
// 根据 req.UserName 请求中的用户名获取用户信息
func getUserInfo(userName string) (*model.User, error) {
   // 通过用户名从缓存中获取用户信息，如果找到就直接返回
   user, err := cache.GetUserInfoFromCache(userName)
   if err == nil && user.Name == userName {
      log.Infof("cache_user ======= %v", user)
      return user, nil
   }

   // 通过用户名从数据库取用户信息
   user, err = dao.GetUserByName(userName)
   // 查询过程出现错误
   if err != nil {
      return user, err
   }
   // 查询不到，用户不存在
   if user == nil {
      return nil, fmt.Errorf("用户尚未注册")
   }
   log.Infof("user === %+v", user)

   // 将用户信息存入缓存，这样下一次就能直接从缓存取信息
   err = cache.SetUserCacheInfo(user)
   // 如果存入缓存出错就打印错误日志
   if err != nil {
      log.Error("cache userinfo failed for user:", user.Name, " with err:", err.Error())
   }

   // 存入成功，打印成功日志，最后返回用户信息
   log.Infof("getUserInfo successfully, with key userinfo_%s", user.Name)
   return user, nil
}
```

那就继续一步步分析这个函数的实现

**首先就是根据用户名从 Redis 缓存中获取用户信息**

```Go
// 通过用户名从缓存中获取用户信息，如果找到就直接返回
user, err := cache.GetUserInfoFromCache(userName)
if err == nil && user.Name == userName {
   log.Infof("cache_user ======= %v", user)
   return user, nil
}
```

那这里我们就需要实现 cache.GetUserInfoFromCache() 函数

**cache.GetUserInfoFromCache() 函数的实现** （ internel\cache\cache.go ） 

```Go
// 从 Redis 缓存中获取用户信息
func GetUserInfoFromCache(username string) (*model.User, error) {
   // 用全局常量 constant.UserInfoPrefix + 用户名拼装一个 Redis 的 key（redisKey）
   redisKey := constant.UserInfoPrefix + username

   // 知识补充：Get 方法的参数是一个 context.Background()，
   // 它用于创建一个空的上下文。在这里，我们使用一个空的上下文作为参数，表示不对操作设置任何超时时间或取消信号。

   // 调用 utils.GetRedisCli() 获取一个 Redis 客户端实例
   // 使用 Get 方法从 Redis 中根据 redisKey 获取相应的值，并将结果赋给变量 val
   val, err := utils.GetRedisCli().Get(context.Background(), redisKey).Result()
   if err != nil {
      return nil, err
   }

   // 创建一个 model.User 对象的指针 user，用于存储从缓存中获取到的用户信息
   user := &model.User{}

   // json.Unmarshal 函数接受两个参数：
   // 第一个参数是一个字节切片，通过 []byte(val) 将获取到的字符串值转换为字节切片
   // 第二个参数是一个指向 model.User 对象的指针，函数会将解析后的值填充到 user 变量所指向的内存地址中
   // 该函数在此处的作用：
   // 使用 json.Unmarshal 函数，将获取到的缓存值 val（类型为字符串）反序列化为一个 model.User 对象
   err = json.Unmarshal([]byte(val), user)
   return user, err
}
```

我们再慢慢看这个函数的实现逻辑

**首先是拼接一个 rediskey，然后用这个 key 去 Redis 中查找用户信息**

```Go
// 用全局常量 constant.UserInfoPrefix + 用户名拼装一个 Redis 的 key（redisKey）
redisKey := constant.UserInfoPrefix + username

// 知识补充：Get 方法的参数是一个 context.Background()，
// 它用于创建一个空的上下文。在这里，我们使用一个空的上下文作为参数，表示不对操作设置任何超时时间或取消信号。

// 调用 utils.GetRedisCli() 获取一个 Redis 客户端实例
// 使用 Get 方法从 Redis 中根据 redisKey 获取相应的值，并将结果赋给变量 val
val, err := utils.GetRedisCli().Get(context.Background(), redisKey).Result()
if err != nil {
   return nil, err
}
```

这里的重点是我们需要一个 Redis 客户端的实例对象，所以需要实现连接 Redis 数据库的逻辑

**连接 Redis 的实现** （ utils\redis.go ）

```Go
package utils

import (
   "fmt"
   "github.com/redis/go-redis/v9"
   log "github.com/sirupsen/logrus"
   "golang.org/x/net/context"
   "gouse/config"
   "sync"
)

var (
   redisConn *redis.Client // Redis 数据库连接对象
   redisOnce sync.Once
)

// openDB 连接db
func initRedis() {
   // 从全局配置信息中获取 Redis 数据库的配置信息
   redisConfig := config.GetGlobalConf().RedisConfig
   log.Infof("redisConfig=======%+v", redisConfig)

   // 使用 fmt.Sprintf() 构建出 Redis 主机地址 addr，格式为 主机地址:端口号
   addr := fmt.Sprintf("%s:%d", redisConfig.Host, redisConfig.Port)

   // 使用 redis.NewClient() 方法创建一个客户端连接对象
   // 传递一个 redis.Options 结构体作为参数，该结构体包含了 Redis 的连接和配置信息
   redisConn = redis.NewClient(&redis.Options{
      Addr:     addr,                 // Redis 服务器的主机地址和端口号
      Password: redisConfig.PassWord, // Redis 的密码
      DB:       redisConfig.DB,       // Redis 的数据库编号
      PoolSize: redisConfig.PoolSize, // Redis 连接池的大小
   })
   if redisConn == nil {
      panic("failed to call redis.NewClient")
   }

   // redisConn.Set() 方法参数详解：
   // 第一个参数是一个上下文对象，可以使用 context.Background() 创建一个空的上下文对象
   // 第二个参数是要设置的键名，这里是 "abc"
   // 第三个参数是要设置的键值，这里是 100
   // 第四个参数是过期时间，以秒为单位，这里是设置为 60 秒

   // 这段代码的目的是将键 "abc" 的值设置为 100，并在 60 秒后过期
   // 通过 res 和 err 的返回值，可以判断这个设置操作是否成功
   res, err := redisConn.Set(context.Background(), "abc", 100, 60).Result()
   log.Infof("res=======%v,err======%v", res, err)

   // 调用 redisConn.Ping() 方法来测试与 Redis 的连接是否正常
   _, err = redisConn.Ping(context.Background()).Result()
   if err != nil {
      panic("Failed to ping redis, err:%s")
   }
}

// 关闭 Redis 数据库
func CloseRedis() {
   redisConn.Close()
}

// GetRedisCli 获取数据库连接
func GetRedisCli() *redis.Client {
   redisOnce.Do(initRedis)
   return redisConn
}
```

跟连接 MySQL 数据库类似，这里也是实现了一个单例的连接逻辑，这里我们就来讲他的核心逻辑实现 initRedis() 

**首先是要从全局配置信息里获取 Redis 数据库的信息，所以我们又得更新一下我们的全局配置了**

（ config\config.go ）

```Go
// RedisConf Redis 配置
type RedisConf struct {
   Host     string `yaml:"rhost" mapstructure:"rhost"`       // 主机地址
   Port     int    `yaml:"rport" mapstructure:"rport"`       // 端口
   DB       int    `yaml:"rdb" mapstructure:"rdb"`           // 编号
   PassWord string `yaml:"passwd" mapstructure:"passwd"`     // 密码
   PoolSize int    `yaml:"poolsize" mapstructure:"poolsize"` // 连接池大小
}

// 缓存配置
type Cache struct {
   SessionExpired int `yaml:"session_expired" mapstructure:"session_expired"` // 会话缓存过期时间
   UserExpired    int `yaml:"user_expired" mapstructure:"user_expired"`       // 用户缓存过期时间
}

// GlobalConfig 业务配置结构体
type GlobalConfig struct {
   AppConfig   AppConf   `yaml:"app" mapstructure:"app"`     // 服务配置
   DbConfig    DbConf    `yaml:"db" mapstructure:"db"`       // 数据库配置
   RedisConfig RedisConf `yaml:"redis" mapstructure:"redis"` // redis 配置
   Cache       Cache     `yaml:"cache" mapstructure:"cache"` // cache 配置
}
```

**然后是配置文件** （ conf\app.yml ）

```YAML
# Redis 配置
redis:
  rhost: "0.0.0.0"
  rport: 8089
  rdb: 0
  passwd: ''
  poolsize: 100

# 缓存配置
cache:
  session_expired: 7200 # second
  user_expired: 300  # second
```

补充好了全局配置信息，我们继续实现与 Redis 数据库建立连接的代码

**接着就是根据构建的 Redis 主机地址等信息调用 redis.NewClient() 方法创建一个客户端连接对象**

（ utils\redis.go ）

```Go
// 使用 fmt.Sprintf() 构建出 Redis 主机地址 addr，格式为 主机地址:端口号
addr := fmt.Sprintf("%s:%d", redisConfig.Host, redisConfig.Port)

// 使用 redis.NewClient() 方法创建一个客户端连接对象
// 传递一个 redis.Options 结构体作为参数，该结构体包含了 Redis 的连接和配置信息
redisConn = redis.NewClient(&redis.Options{
   Addr:     addr,                 // Redis 服务器的主机地址和端口号
   Password: redisConfig.PassWord, // Redis 的密码
   DB:       redisConfig.DB,       // Redis 的数据库编号
   PoolSize: redisConfig.PoolSize, // Redis 连接池的大小
})
if redisConn == nil {
   panic("failed to call redis.NewClient")
}
```

**最后就是验证 Redis 数据库是否正常连接**

```Go
// redisConn.Set() 方法参数详解：
// 第一个参数是一个上下文对象，可以使用 context.Background() 创建一个空的上下文对象
// 第二个参数是要设置的键名，这里是 "abc"
// 第三个参数是要设置的键值，这里是 100
// 第四个参数是过期时间，以秒为单位，这里是设置为 60 秒

// 这段代码的目的是将键 "abc" 的值设置为 100，并在 60 秒后过期
// 通过 res 和 err 的返回值，可以判断这个设置操作是否成功
res, err := redisConn.Set(context.Background(), "abc", 100, 60).Result()
log.Infof("res=======%v,err======%v", res, err)

// 调用 redisConn.Ping() 方法来测试与 Redis 的连接是否正常
_, err = redisConn.Ping(context.Background()).Result()
if err != nil {
   panic("Failed to ping redis, err:%s")
}
```

完成了获取 Redis 客户端示例的逻辑，我们继续实现从 Redis 中获取用户信息的代码

**接着是将取出的缓存值通过反序列化，变成我们用于存放用户数据的 User 对象** （ internel\cache\cache.go ）

```Go
// 创建一个 model.User 对象的指针 user，用于存储从缓存中获取到的用户信息
user := &model.User{}

// json.Unmarshal 函数接受两个参数：
// 第一个参数是一个字节切片，通过 []byte(val) 将获取到的字符串值转换为字节切片
// 第二个参数是一个指向 model.User 对象的指针，函数会将解析后的值填充到 user 变量所指向的内存地址中
// 该函数在此处的作用：
// 使用 json.Unmarshal 函数，将获取到的缓存值 val（类型为字符串）反序列化为一个 model.User 对象
err = json.Unmarshal([]byte(val), user)
return user, err
```

完成了从 Redis 缓存获取用户信息的逻辑，我们继续编写获取用户信息的逻辑

如果我们从 Redis 缓存中已经获取到信息了，那就直接返回，如果没能拿到，那

**接下来就是从数据库中获取用户信息** （ internel\service\user.go ）  

```Go
// 通过用户名从数据库取用户信息
user, err = dao.GetUserByName(userName)
// 查询过程出现错误
if err != nil {
   return user, err
}
// 查询不到，用户不存在
if user == nil {
   return nil, fmt.Errorf("用户尚未注册")
}
log.Infof("user === %+v", user)
```

**最后就是将用户信息存入缓存中**，这样我们下一次取的效率就提升了 （ internel\service\user.go ）

```Go
// 将用户信息存入缓存，这样下一次就能直接从缓存取信息
err = cache.SetUserCacheInfo(user)
// 如果存入缓存出错就打印错误日志
if err != nil {
   log.Error("cache userinfo failed for user:", user.Name, " with err:", err.Error())
}

// 存入成功，打印成功日志，最后返回用户信息
log.Infof("getUserInfo successfully, with key userinfo_%s", user.Name)
return user, nil
```

**这里我们就要实现将用户信息存入缓存了逻辑** ( internel\cache\cache.go )

```Go
// 将用户信息存入 Redis 缓存
func SetUserCacheInfo(user *model.User) error {
   // 用全局常量 constant.UserInfoPrefix + 用户名拼装一个 Redis 的 key（redisKey）
   redisKey := constant.UserInfoPrefix + user.Name

   // json.Marshal() 将用户对象转换为 JSON 字符串表示，并将其存储在变量 val 中
   val, err := json.Marshal(user)
   if err != nil {
      return err
   }

   // 从配置中获取用户缓存过期时间，并将其转换为 time.Duration 类型
   // 通过将过期时间与 time.Second 相乘，得到过期时间的秒数
   expired := time.Second * time.Duration(config.GetGlobalConf().Cache.UserExpired)

   // 最后，执行 utils.GetRedisCli().Set() 方法将用户信息存入 Redis 中，该方法接受四个参数：
   // 第一个参数是一个上下文对象，可以使用 context.Background() 创建一个空的上下文对象
   // 第二个参数是要设置的键名，这里是 redisKey
   // 第三个参数是要设置的键值，这里是 val，即用户信息的 JSON 字符串表示
   // 第四个参数是过期时间，以秒为单位，这里是通过过期时间的秒数计算得到
   _, err = utils.GetRedisCli().Set(context.Background(), redisKey, val, expired*time.Second).Result()
   return err
}
```

当我们成功获取到用户信息了，那就继续实现用户登录的相关逻辑

**接着是比对密码是否正确，如果正确就生成一个 session** （ internel\service\user.go ）

```Go
// 用户存在，比对输入的密码和用户密码是否一致
if req.PassWord != user.PassWord {
   log.Errorf("Login|password err: req.password=%s|user.password=%s", req.PassWord, user.PassWord)
   return "", fmt.Errorf("password is not correct")
}

// 如果密码匹配成功，调用 utils.GenerateSession 函数生成一个新的 session 字符串
session := utils.GenerateSession(user.Name)
```

这里我们是实现了一个生成 session 的逻辑 （ utils\utils.go ）

```Go
// 生成一个新的 session 字符串
func GenerateSession(userName string) string {
   // 使用 fmt.Sprintf() 函数将用户名和字符串 "session" 拼接为一个新的字符串
   // 将拼接得到的新字符串作为参数传递给 Md5String() 函数，通过哈希算法生成 MD5 值的字符串表示
   return Md5String(fmt.Sprintf("%s:%s", userName, "session"))
}
```

**最后是将用户信息和 session 存进缓存里**

```Go
// 并调用 cache.SetSessionInfo 函数将用户信息和 session 存储到缓存中
err = cache.SetSessionInfo(user, session)
if err != nil {
   log.Errorf(" Login|Failed to SetSessionInfo, uuid=%s|user_name=%s|session=%s|err=%v", uuid, user.Name, session, err)
   return "", fmt.Errorf("login|SetSessionInfo fail:%v", err)
}

// 最后，使用 log.Infof 打印登录成功的日志，并返回生成的 session 字符串作为登录成功的标识
log.Infof("Login successfully, %s@%s with redis_session session_%s", req.UserName, req.PassWord, session)
return session, nil
```

这里我们也要实现一个将会话信息存入缓存的逻辑 ( internel\cache\cache.go )

```Go
// 将用户信息与会话字符串存入 Redis 缓存中
func SetSessionInfo(user *model.User, session string) error {
   // 用全局常量 constant.UserInfoPrefix + 会话字符串拼装一个 Redis 的 key（redisKey）
   redisKey := constant.SessionKeyPrefix + session

   // json.Marshal() 将用户对象转换为 JSON 字符串表示，并将其存储在变量 val 中
   val, err := json.Marshal(&user)
   if err != nil {
      return err
   }

   // 从配置中获取会话缓存过期时间，并将其转换为 time.Duration 类型
   // 通过将过期时间与 time.Second 相乘，得到过期时间的秒数
   expired := time.Second * time.Duration(config.GetGlobalConf().Cache.SessionExpired)

   // 最后，执行 utils.GetRedisCli().Set() 方法将用户信息存入 Redis 中
   _, err = utils.GetRedisCli().Set(context.Background(), redisKey, val, expired*time.Second).Result()
   return err
}
```

我们完成用户登录逻辑的具体实现之后，

**最后一步就是设置 cookie 了** （ api\http\v1\api.go ）

```Go
// 登陆成功，使用 c.SetCookie 函数设置一个名为 constant.SessionKey 的 Cookie。下面是函数参数介绍：
// SessionKey 是 cookie 的名称; session 是登录成功生成的 session 值; CookieExpire 决定了 cookie 的有效期
// “/” 是 cookie 的路径（表示该 Cookie 对所有路径都有效）; "" 是 cookie 的域名（空字符串表示该 Cookie 对所有域名都有效）
// false 是指定该 Cookie 只能通过 HTTP 协议传输，不能通过 JavaScript 访问; true 是指定该 Cookie 在安全的 HTTPS 连接中也会被传输
c.SetCookie(constant.SessionKey, session, constant.CookieExpire, "/", "", false, true)

// 如果注册逻辑执行成功，
// 调用 rsp.ResponseSuccess 方法返回一个表示成功的 HTTP 响应给客户端
rsp.ResponseSuccess(c)
```

（备注：有一些全局常量的实现遗漏了，到时候缺什么根据源码添加上去即可）

这样我们总算是完成了用户登录逻辑的实现~

## 知识学习部分

重点知识涵盖：MD5 哈希算法的应用，context 包的应用，go-redis 框架的应用，Redis 数据库配置，结构体对象的序列化和 JSON 子串的反序列化，字符串拼接操作，gin 框架设置 Cookie。

### 1）MD5 哈希算法的应用

```Go
// 生成一个唯一的 uuid
// 这里使用用户名和当前时间拼接后进行 MD5 哈希算法生成
uuid := utils.Md5String(req.UserName + time.Now().GoString())
// 用哈希算法生成一个唯一的 uuid
func Md5String(s string) string {
   // 创建一个 md5 实例
   h := md5.New()

   // 通过调用 h.Write([]byte(s)) 将字符串 s 的字节表示写入到 md5 实例中
   h.Write([]byte(s))

   // 调用 h.Sum(nil) 完成 MD5 的计算并返回结果
   // Sum() 方法接受一个切片参数，用于存储 MD5 值
   // 由于我们在这里不需要对 MD5 值进行其他操作，所以将参数设置为 nil
   // 再将计算得到的 MD5 值以十六进制的形式表示，并由 hex.EncodeToString() 转换为字符串
   // 注意：
   // 这段代码的作用是对给定的字符串进行 MD5 计算，并将计算结果以字符串的形式返回
   // 通常，MD5 值用于验证数据的完整性和唯一性，但请注意，MD5 算法已经被认为不再安全，不适用于密码等敏感数据的加密
   str := hex.EncodeToString(h.Sum(nil))
   return str
}
```

### 2）context 包的应用

```Go
// 将生成的 uuid 存入上下文（context）中，以便后续使用
ctx := context.WithValue(context.Background(), "uuid", uuid)
// 从上下文对象中获取请求的唯一标识符(uuid)，并使用 log.Debugf 打印日志表明有用户访问登录功能
uuid := ctx.Value(constant.ReqUuid)
log.Debugf(" %s| Login access from:%s,@,%s", uuid, req.UserName, req.PassWord)
```

### 3）go-redis 框架的 Get 方法

```Go
// 调用 utils.GetRedisCli() 获取一个 Redis 客户端实例
// 使用 Get 方法从 Redis 中根据 redisKey 获取相应的值，并将结果赋给变量 val
val, err := utils.GetRedisCli().Get(context.Background(), redisKey).Result()
if err != nil {
   return nil, err
}
```

### 4）go-redis 框架连接 Redis

```Go
var (
   redisConn *redis.Client // Redis 数据库连接对象
   redisOnce sync.Once
)

// openDB 连接db
func initRedis() {
   // 从全局配置信息中获取 Redis 数据库的配置信息
   redisConfig := config.GetGlobalConf().RedisConfig
   log.Infof("redisConfig=======%+v", redisConfig)

   // 使用 fmt.Sprintf() 构建出 Redis 主机地址 addr，格式为 主机地址:端口号
   addr := fmt.Sprintf("%s:%d", redisConfig.Host, redisConfig.Port)

   // 使用 redis.NewClient() 方法创建一个客户端连接对象
   // 传递一个 redis.Options 结构体作为参数，该结构体包含了 Redis 的连接和配置信息
   redisConn = redis.NewClient(&redis.Options{
      Addr:     addr,                 // Redis 服务器的主机地址和端口号
      Password: redisConfig.PassWord, // Redis 的密码
      DB:       redisConfig.DB,       // Redis 的数据库编号
      PoolSize: redisConfig.PoolSize, // Redis 连接池的大小
   })
   if redisConn == nil {
      panic("failed to call redis.NewClient")
   }

   // redisConn.Set() 方法参数详解：
   // 第一个参数是一个上下文对象，可以使用 context.Background() 创建一个空的上下文对象
   // 第二个参数是要设置的键名，这里是 "abc"
   // 第三个参数是要设置的键值，这里是 100
   // 第四个参数是过期时间，以秒为单位，这里是设置为 60 秒

   // 这段代码的目的是将键 "abc" 的值设置为 100，并在 60 秒后过期
   // 通过 res 和 err 的返回值，可以判断这个设置操作是否成功
   res, err := redisConn.Set(context.Background(), "abc", 100, 60).Result()
   log.Infof("res=======%v,err======%v", res, err)

   // 调用 redisConn.Ping() 方法来测试与 Redis 的连接是否正常
   _, err = redisConn.Ping(context.Background()).Result()
   if err != nil {
      panic("Failed to ping redis, err:%s")
   }
}

// 关闭 Redis 数据库
func CloseRedis() {
   redisConn.Close()
}

// GetRedisCli 获取数据库连接
func GetRedisCli() *redis.Client {
   redisOnce.Do(initRedis)
   return redisConn
}
```

### 5）go-redis 框架的 Set 方法

```Go
// 这段代码的目的是将键 "abc" 的值设置为 100，并在 60 秒后过期
// 通过 res 和 err 的返回值，可以判断这个设置操作是否成功
res, err := redisConn.Set(context.Background(), "abc", 100, 60).Result()
log.Infof("res=======%v,err======%v", res, err)
// 将用户信息存入 Redis 缓存
func SetUserCacheInfo(user *model.User) error {
   // 用全局常量 constant.UserInfoPrefix + 用户名拼装一个 Redis 的 key（redisKey）
   redisKey := constant.UserInfoPrefix + user.Name

   // json.Marshal() 将用户对象转换为 JSON 字符串表示，并将其存储在变量 val 中
   val, err := json.Marshal(user)
   if err != nil {
      return err
   }

   // 从配置中获取用户缓存过期时间，并将其转换为 time.Duration 类型
   // 通过将过期时间与 time.Second 相乘，得到过期时间的秒数
   expired := time.Second * time.Duration(config.GetGlobalConf().Cache.UserExpired)

   // 最后，执行 utils.GetRedisCli().Set() 方法将用户信息存入 Redis 中，该方法接受四个参数：
   // 第一个参数是一个上下文对象，可以使用 context.Background() 创建一个空的上下文对象
   // 第二个参数是要设置的键名，这里是 redisKey
   // 第三个参数是要设置的键值，这里是 val，即用户信息的 JSON 字符串表示
   // 第四个参数是过期时间，以秒为单位，这里是通过过期时间的秒数计算得到
   _, err = utils.GetRedisCli().Set(context.Background(), redisKey, val, expired*time.Second).Result()
   return err
}
// 将用户信息与会话字符串存入 Redis 缓存中
func SetSessionInfo(user *model.User, session string) error {
   // 用全局常量 constant.UserInfoPrefix + 会话字符串拼装一个 Redis 的 key（redisKey）
   redisKey := constant.SessionKeyPrefix + session

   // json.Marshal() 将用户对象转换为 JSON 字符串表示，并将其存储在变量 val 中
   val, err := json.Marshal(&user)
   if err != nil {
      return err
   }

   // 从配置中获取会话缓存过期时间，并将其转换为 time.Duration 类型
   // 通过将过期时间与 time.Second 相乘，得到过期时间的秒数
   expired := time.Second * time.Duration(config.GetGlobalConf().Cache.SessionExpired)

   // 最后，执行 utils.GetRedisCli().Set() 方法将用户信息存入 Redis 中
   _, err = utils.GetRedisCli().Set(context.Background(), redisKey, val, expired*time.Second).Result()
   return err
}
```

### 6）go-redis 框架的 Ping 方法

```Go
// 调用 redisConn.Ping() 方法来测试与 Redis 的连接是否正常
_, err = redisConn.Ping(context.Background()).Result()
if err != nil {
   panic("Failed to ping redis, err:%s")
}
```

### 7）Redis 数据库与缓存配置

```Go
// RedisConf Redis 配置
type RedisConf struct {
   Host     string `yaml:"rhost" mapstructure:"rhost"`       // 主机地址
   Port     int    `yaml:"rport" mapstructure:"rport"`       // 端口
   DB       int    `yaml:"rdb" mapstructure:"rdb"`           // 编号
   PassWord string `yaml:"passwd" mapstructure:"passwd"`     // 密码
   PoolSize int    `yaml:"poolsize" mapstructure:"poolsize"` // 连接池大小
}

// 缓存配置
type Cache struct {
   SessionExpired int `yaml:"session_expired" mapstructure:"session_expired"` // 会话缓存过期时间
   UserExpired    int `yaml:"user_expired" mapstructure:"user_expired"`       // 用户缓存过期时间
}

// GlobalConfig 业务配置结构体
type GlobalConfig struct {
   AppConfig   AppConf   `yaml:"app" mapstructure:"app"`     // 服务配置
   DbConfig    DbConf    `yaml:"db" mapstructure:"db"`       // 数据库配置
   RedisConfig RedisConf `yaml:"redis" mapstructure:"redis"` // redis 配置
   Cache       Cache     `yaml:"cache" mapstructure:"cache"` // cache 配置
# Redis 配置
redis:
  rhost: "0.0.0.0"
  rport: 8089
  rdb: 0
  passwd: ''
  poolsize: 100

# 缓存配置
cache:
  session_expired: 7200 # second
  user_expired: 300  # second
```

### 8）结构体对象的序列化

```Go
// json.Marshal() 将用户对象转换为 JSON 字符串表示，并将其存储在变量 val 中
val, err := json.Marshal(&user)
if err != nil {
   return err
}
```

### 9）JSON 子串的反序列化

```Go
// 创建一个 model.User 对象的指针 user，用于存储从缓存中获取到的用户信息
user := &model.User{}

// json.Unmarshal 函数接受两个参数：
// 第一个参数是一个字节切片，通过 []byte(val) 将获取到的字符串值转换为字节切片
// 第二个参数是一个指向 model.User 对象的指针，函数会将解析后的值填充到 user 变量所指向的内存地址中
// 该函数在此处的作用：
// 使用 json.Unmarshal 函数，将获取到的缓存值 val（类型为字符串）反序列化为一个 model.User 对象
err = json.Unmarshal([]byte(val), user)
return user, err
```

### 10）字符串拼接操作

```Go
// 使用 fmt.Sprintf() 函数将用户名和字符串 "session" 拼接为一个新的字符串
// 将拼接得到的新字符串作为参数传递给 Md5String() 函数，通过哈希算法生成 MD5 值的字符串表示
return Md5String(fmt.Sprintf("%s:%s", userName, "session"))
```

### 11）gin 框架设置 Cookie

```Go
const (
   SessionKey   = "user_session"
   CookieExpire = 3600
)
// 登陆成功，使用 c.SetCookie 函数设置一个名为 constant.SessionKey 的 Cookie。下面是函数参数介绍：
// SessionKey 是 cookie 的名称; session 是登录成功生成的 session 值; CookieExpire 决定了 cookie 的有效期
// “/” 是 cookie 的路径（表示该 Cookie 对所有路径都有效）; "" 是 cookie 的域名（空字符串表示该 Cookie 对所有域名都有效）
// false 是指定该 Cookie 只能通过 HTTP 协议传输，不能通过 JavaScript 访问; true 是指定该 Cookie 在安全的 HTTPS 连接中也会被传输
c.SetCookie(constant.SessionKey, session, constant.CookieExpire, "/", "", false, true)
```

# 4. 实现用户登出

## 1、代码实现部分

**代码的核心部分其实就是在 InitRouterAndServe() 方法里添加上这一句代码。**( internal\router\router.go )

```Go
// 用户登出
r.POST("/user/logout", AuthMiddleWare(), api.Logout)
```

然后就是实现这个登出操作了，第一个要实现的就是 AuthMiddleWare() 函数

```Go
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
```

这个中间件函数的作用就是检查 Cookie 值，说人话就是检查现在的登录状态，如果不是登录状态就终止请求

**接着就是登出操作的具体实现 api.Logout** （ api\http\v1\api.go ）

```Go
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
```

这里我们需要实现登出请求结构体 （ internel\service\entity.go ）

```Go
// LogoutRequest 登出请求
type LogoutRequest struct {
   UserName string `json:"user_name"`
}
```

**接着就是实现 Logout() 登出操作的具体逻辑**

```Go
   // 实现 Logout() 登出操作的具体逻辑
   if err := service.Logout(ctx, req); err != nil {
      rsp.ResponseWithError(c, CodeLogoutErr, err.Error())
      return
   }
```

具体实现：（ internel\service\user.go ）

```Go
// Logout 退出登陆
func Logout(ctx context.Context, req *LogoutRequest) error {
   // 从上下文中获取请求的唯一标识 uuid 和会话标识 session，以及解析请求的 UserName
   uuid := ctx.Value(constant.ReqUuid)
   session := ctx.Value(constant.SessionKey).(string)
   log.Infof("%s|Logout access from,user_name=%s|session=%s", uuid, req.UserName, session)

   // 从缓存中获取会话信息，用于验证用户是否处于登录状态
   _, err := cache.GetSessionInfo(session)
   if err != nil {
      log.Errorf("%s|Failed to get with session=%s|err =%v", uuid, session, err)
      return fmt.Errorf("Logout|GetSessionInfo err:%v", err)
   }

   // 从缓存中删除会话信息，表示用户已退出登录
   err = cache.DelSessionInfo(session)
   if err != nil {
      log.Errorf("%s|Failed to delSessionInfo :%s", uuid, session)
      return fmt.Errorf("del session err:%v", err)
   }
   log.Infof("%s|Success to delSessionInfo :%s", uuid, session)
   return nil
}
```

这里有两个我们需要实现的方法

**首先是 GetSessionInfo() 从缓存中获取会话信息**

```Go
// 从缓存中获取会话信息，用于验证用户是否处于登录状态
_, err := cache.GetSessionInfo(session)
if err != nil {
   log.Errorf("%s|Failed to get with session=%s|err =%v", uuid, session, err)
   return fmt.Errorf("Logout|GetSessionInfo err:%v", err)
}
```

具体实现：( internel\cache\cache.go )

```Go
// 从缓存中获取会话信息
func GetSessionInfo(session string) (*model.User, error) {
   // 构建 Redis 中存储会话信息的键 redisKey
   redisKey := constant.SessionKeyPrefix + session

   //  获取 Redis 的客户端连接实例，并使用 Get() 方法从 Redis 中获取与 redisKey 对应的值
   val, err := utils.GetRedisCli().Get(context.Background(), redisKey).Result()
   if err != nil {
      return nil, err
   }

   // 将获取到的会话信息值 val 转换为 model.User 的结构体指针类型
   // 并使用 json.Unmarshal() 方法将 val 反序列化为 model.User 结构体
   user := &model.User{}
   err = json.Unmarshal([]byte(val), &user)
   return user, err
}
```

**然后就是从缓存中删除会话信息**

```Go
// 从缓存中删除会话信息，表示用户已退出登录
err = cache.DelSessionInfo(session)
if err != nil {
   log.Errorf("%s|Failed to delSessionInfo :%s", uuid, session)
   return fmt.Errorf("del session err:%v", err)
}
```

具体实现：( internel\cache\cache.go )

```Go
// 删除缓存中的会话信息
func DelSessionInfo(session string) error {
   // 构建 Redis 中存储会话信息的键 redisKey
   redisKey := constant.SessionKeyPrefix + session

   // Del() 方法返回删除的键的数量和可能的错误信息
   _, err := utils.GetRedisCli().Del(context.Background(), redisKey).Result()
   return err
}
```

这样我们就实现了用户的登出操作了

## 2、知识学习部分

重点知识包涵盖：gin 框架的应用，go-redis 框架的应用。

### 1）gin 框架 POST 方法

```Go
// 用户登出 // 这里设置了一个中间件函数
r.POST("/user/logout", AuthMiddleWare(), api.Logout)
// 这是一个用于对请求进行身份验证的中间件函数
// 补充知识：gin.HandlerFunc的参数为 *gin.Context
func AuthMiddleWare() gin.HandlerFunc {
   // 这里我把具体实现删了，重点看返回的框架形式
   return func(c *gin.Context) {
      return
   }
}
```

### 2）gin 框架 Next 方法

```Go
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
```

### 3）gin 框架的 JSON 方法

```Go
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
```

### 4）gin 框架的 Abort 方法

```Go
// c.Abort() 是一个用于终止请求的函数，它可以停止请求链的继续处理，
// 确保本次请求不再继续向后执行其他的中间件或请求处理函数
c.Abort()
```

### 5）gin 框架的 SetCookie 方法

他能设置 Cookie：（登录操作）

```Go
// 登陆成功，使用 c.SetCookie 函数设置一个名为 constant.SessionKey 的 Cookie。下面是函数参数介绍：
// SessionKey 是 cookie 的名称; session 是登录成功生成的 session 值; CookieExpire 决定了 cookie 的有效期
// “/” 是 cookie 的路径（表示该 Cookie 对所有路径都有效）; "" 是 cookie 的域名（空字符串表示该 Cookie 对所有域名都有效）
// false 是指定该 Cookie 只能通过 HTTP 协议传输，不能通过 JavaScript 访问 true 是指定该 Cookie 在安全的 HTTPS 连接中也会被传输
c.SetCookie(constant.SessionKey, session, constant.CookieExpire, "/", "", false, true)
```

也能取消 Cookie：（登出操作）

```Go
// 设置一个过期时间为负值的 Cookie，实现删除客户端浏览器中存储的会话标识的目的，即实现用户的登出操作
c.SetCookie(constant.SessionKey, session, -1, "/", "", false, true)
```

### 6）go-redis 框架的 Del 方法

```Go
// 删除缓存中的会话信息
func DelSessionInfo(session string) error {
   // 构建 Redis 中存储会话信息的键 redisKey
   redisKey := constant.SessionKeyPrefix + session

   // Del() 方法返回删除的键的数量和可能的错误信息
   _, err := utils.GetRedisCli().Del(context.Background(), redisKey).Result()
   return err
}
```

# 5. 实现获取用户信息

## 1、代码实现部分

这里我就直入主题了 （ internel\router\router.go ）

```Go
// 获取用户信息
r.GET("/user/get_user_info", AuthMiddleWare(), api.GetUserInfo)
```

老样子，我们来实现这个 api.GetUserInfo 方法

**api.GetUserInfo 方法的实现** （ api\http\v1\api.go ）

```Go
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
```

顺着逻辑往下走，我们第一个需要添加的是 GetUserInfoRequest 的全局请求结构体 

（ internel\service\entity.go ）

```Go
// GetUserInfoRequest 获取用户信息请求
type GetUserInfoRequest struct {
   UserName string `json:"user_name"`
}
```

**然后就是从缓存中获取用户信息 service.GetUserInfo() 的实现** （ internel\service\user.go ）

```Go
// 从缓存中获取用户信息，只能在用户登陆的情况下使用
func GetUserInfo(ctx context.Context, req *GetUserInfoRequest) (*GetUserInfoResponse, error) {
   // 从上下文取到 uuid 和 session 信息
   uuid := ctx.Value(constant.ReqUuid)
   session := ctx.Value(constant.SessionKey).(string)
   log.Infof("%s|GetUserInfo access from,user_name=%s|session=%s", uuid, req.UserName, session)

   // uuid 和 session 不能为空，也就是需要处于登录状态
   if session == "" || req.UserName == "" {
      return nil, fmt.Errorf("GetUserInfo|request params invalid")
   }

   // 根据 session 从缓存中获取用户信息
   user, err := cache.GetSessionInfo(session)
   if err != nil {
      log.Errorf("%s|Failed to get with session=%s|err =%v", uuid, session, err)
      return nil, fmt.Errorf("getUserInfo|GetSessionInfo err:%v", err)
   }

   // 验证获取到的用户信息和请求的用户名是否相同
   if user.Name != req.UserName {
      log.Errorf("%s|session info not match with username=%s", uuid, req.UserName)
   }

   log.Infof("%s|Succ to GetUserInfo|user_name=%s|session=%s", uuid, req.UserName, session)

   // 填好用户信息的返回结构，然后返回
   return &GetUserInfoResponse{
      UserName: user.Name,
      Age:      user.Age,
      Gender:   user.Gender,
      PassWord: user.PassWord,
      NickName: user.NickName,
   }, nil
}
```

这里需要添加的是一个获取用户信息返回结构对象 （ internel\service\entity.go ）

```Go
// GetUserInfoResponse 获取用户信息返回结构
type GetUserInfoResponse struct {
   UserName string `json:"user_name"`
   Age      int    `json:"age"`
   Gender   string `json:"gender"`
   PassWord string `json:"pass_word"`
   NickName string `json:"nick_name"`
}
```

**然后就是通过 session 从缓存中获取用户信息具体实现** （ internel\cache\cache.go ）

```Go
// 根据 session 从缓存中获取用户信息
func GetSessionInfo(session string) (*model.User, error) {
   // 构建 Redis 中存储会话信息的键 redisKey
   redisKey := constant.SessionKeyPrefix + session

   //  获取 Redis 的客户端连接实例，并使用 Get() 方法从 Redis 中获取与 redisKey 对应的值
   val, err := utils.GetRedisCli().Get(context.Background(), redisKey).Result()
   if err != nil {
      return nil, err
   }

   // 将获取到的会话信息值 val 转换为 model.User 的结构体指针类型
   // 并使用 json.Unmarshal() 方法将 val 反序列化为 model.User 结构体
   user := &model.User{}
   err = json.Unmarshal([]byte(val), &user)
   return user, err
}
```

最后要将用户信息返回给客户端 （ api\http\v1\api.go ） 

```Go
// 这个返回函数给客户端多返回了一个 data 也就是实际的数据（从缓存中获取的用户信息）
rsp.ResponseWithData(c, userInfo)
```

（ api\http\v1\entity ）

```Go
// 这个返回函数给客户端多返回了一个 data 也就是实际的数据（从缓存中获取的用户信息）
func (rsp *HttpResponse) ResponseWithData(c *gin.Context, data interface{}) {
   rsp.Code = CodeSuccess
   rsp.Msg = "success"
   rsp.Data = data
   c.JSON(http.StatusOK, rsp)
}
```

## 2、知识学习部分

重点知识包含：gin 框架的使用，其他基本上没什么新东西

### 1）gin 框架的 Query 方法

```Go
// 从 HTTP 请求的查询参数中获取用户名 userName
userName := c.Query("username")
```

这个是我们需要提前配置好了路由，通过发送一个HTTP GET请求到注册用户的URL，并在查询参数中提供"username"参数的值，才能使用查询参数查询到我们需要的值。

# 6. 实现更新用户信息

## 1、代码实现部分

这里我实现的是修改用户昵称（ internel\router\router.go ）

```Go
// 更新用户信息
r.POST("/user/update_nick_name", AuthMiddleWare(), api.UpdateNickName)
```

**api.UpdateNickName 函数的实现**（ api\http\v1\api.go ）

```Go
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
```

 **首先是创建存放修改用户信息返回的结构体对象** （internel\sevice\entity.go）

```Go
// UpdateNickNameRequest 修改用户信息返回结构
type UpdateNickNameRequest struct {
   UserName    string `json:"user_name"`
   NewNickName string `json:"new_nick_name"`
}
```

然后下一个重点就是更改用户信息的具体实现了

```Go
// 更改用户信息
if err := service.UpdateUserNickName(ctx, req); err != nil {
   rsp.ResponseWithError(c, CodeUpdateUserInfoErr, err.Error())
   return
}
```

（ internel\service\user.go ）

```Go
// 更改用户信息
func UpdateUserNickName(ctx context.Context, req *UpdateNickNameRequest) error {
   // 从上下文取出 uuid 和 session
   uuid := ctx.Value(constant.ReqUuid)
   session := ctx.Value(constant.SessionKey).(string)
   log.Infof("%s|UpdateUserNickName access from,user_name=%s|session=%s", uuid, req.UserName, session)
   log.Infof("UpdateUserNickName|req==%v", req)

   // session 和 用户名不能为空，保证处于登录状态
   if session == "" || req.UserName == "" {
      return fmt.Errorf("UpdateUserNickName|request params invalid")
   }

   // 从缓存中获取用户信息
   user, err := cache.GetSessionInfo(session)
   if err != nil {
      log.Errorf("%s|Failed to get with session=%s|err =%v", uuid, session, err)
      return fmt.Errorf("UpdateUserNickName|GetSessionInfo err:%v", err)
   }

   // 验证获取到的用户信息和请求的用户名是否相同
   if user.Name != req.UserName {
      log.Errorf("UpdateUserNickName|%s|session info not match with username=%s", uuid, req.UserName)
   }

   // 根据请求中的 NickName 构建一个新的用户信息对象
   updateUser := &model.User{
      NickName: req.NewNickName,
   }

   // 返回这个用户信息更新函数的结果
   return updateUserInfo(updateUser, req.UserName, session)
}
```

**这里重点要实现的就是 updateUserInfo() 这个返回函数** （ internel\service\user.go ）

```Go
// 补充说明：
// 我们设置了两个存入缓存的逻辑，或者说设置了两种缓存键
// 一种是通过用户名从缓存取户信息，一种是通过 session 从缓存获取用户信息
//
// 修改昵称的逻辑
func updateUserInfo(user *model.User, userName, session string) error {
   // 更新数据库的用户信息，返回的是被更新的行数
   affectedRows := dao.UpdateUserInfo(userName, user)

   // 如果affectedRows等于1，表示数据库更新成功
   if affectedRows == 1 {
      // 通过userName从数据库中获取更新后的用户信息
      user, err := dao.GetUserByName(userName)
      if err == nil {
         // 再将新的用户信息更新到缓存
         cache.UpdateCachedUserInfo(user)

         // 再把将用户信息与会话字符串存入 Redis 缓存中
         if session != "" {
            err = cache.SetSessionInfo(user, session)

            // 如果出错就删除缓存中的会话信息
            if err != nil {
               log.Error("update session failed:", err.Error())
               cache.DelSessionInfo(session)
            }
         }
      } else {
         log.Error("Failed to get dbUserInfo for cache, username=%s with err:", userName, err.Error())
      }
   }
   return nil
}
```

这个函数的具体逻辑主要是三个部分

**首先是 dao.UpdateUserInfo() ，更新数据库的用户信息**  （ internel\dao\user.go ） 

```Go
// UpdateUserInfo 更新用户信息
func UpdateUserInfo(userName string, user *model.User) int64 {
   // Updates方法用于更新满足条件的记录，参数user包含新的用户信息，RowsAffected返回被影响的行数
   return utils.GetDB().Model(&model.User{}).Where("`name` = ?", userName).Updates(user).RowsAffected
}
```

**然后是 UpdateCachedUserInfo()，更新用户信息到缓存** （ internel\cache\cache.go ）

```Go
// 再将新的用户信息更新到缓存
func UpdateCachedUserInfo(user *model.User) error {
   // 将用户信息存入缓存
   err := SetUserCacheInfo(user)

   // 如果将用户信息存入缓存时发生了问题就把对应的缓存键删了
   if err != nil {
      redisKey := constant.UserInfoPrefix + user.Name
      utils.GetRedisCli().Del(context.Background(), redisKey).Result()
   }
   return err
}
```

**最后是 DelSessionInfo()， 删除缓存中的会话信息** （ internel\cache\cache.go ）

```Go
// 删除缓存中的会话信息
func DelSessionInfo(session string) error {
   // 构建 Redis 中存储会话信息的键 redisKey
   redisKey := constant.SessionKeyPrefix + session

   // Del() 方法返回删除的键的数量和可能的错误信息
   _, err := utils.GetRedisCli().Del(context.Background(), redisKey).Result()
   return err
}
```

## 2、知识学习部分

重点知识涵盖：gorm 框架的应用

### 1）gorm 框架的 Updates 方法

```Go
// UpdateUserInfo 更新用户信息
func UpdateUserInfo(userName string, user *model.User) int64 {
   // Updates方法用于更新满足条件的记录，参数user包含新的用户信息，RowsAffected返回被影响的行数
   return utils.GetDB().Model(&model.User{}).Where("`name` = ?", userName).Updates(user).RowsAffected
}
```
