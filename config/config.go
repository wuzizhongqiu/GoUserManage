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

// GetGlobalConf 获取全局配置文件
func GetGlobalConf() *GlobalConfig {
	// 通过 sync.Once 包的 Do 方法，确保 readConf 函数只会被执行一次
	// 我就直接理解成一个线程安全的单例设计模式了
	once.Do(readConf)
	return &config
}

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

// InitConfig 初始化日志
func InitConfig() {

}
