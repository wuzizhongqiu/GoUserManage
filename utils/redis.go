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
