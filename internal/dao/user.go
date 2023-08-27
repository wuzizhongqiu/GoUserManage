package dao

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gouse/internal/model"
	"gouse/utils"
)

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

// UpdateUserInfo 更新用户信息
func UpdateUserInfo(userName string, user *model.User) int64 {
	// Updates方法用于更新满足条件的记录，参数user包含新的用户信息，RowsAffected返回被影响的行数
	return utils.GetDB().Model(&model.User{}).Where("`name` = ?", userName).Updates(user).RowsAffected
}
