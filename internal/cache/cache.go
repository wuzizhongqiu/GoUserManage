package cache

import (
	"encoding/json"
	"golang.org/x/net/context"
	"gouse/config"
	"gouse/internal/model"
	"gouse/pkg/constant"
	"gouse/utils"
	"time"
)

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
	// 使用 json.Unmarshal 函数，将获取到的缓存值JSON类型字串 val反序列化为一个 model.User 对象
	err = json.Unmarshal([]byte(val), user)
	return user, err
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

// 删除缓存中的会话信息
func DelSessionInfo(session string) error {
	// 构建 Redis 中存储会话信息的键 redisKey
	redisKey := constant.SessionKeyPrefix + session

	// Del() 方法返回删除的键的数量和可能的错误信息
	_, err := utils.GetRedisCli().Del(context.Background(), redisKey).Result()
	return err
}
