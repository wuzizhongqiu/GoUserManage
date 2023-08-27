package service

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"gouse/internal/cache"
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
