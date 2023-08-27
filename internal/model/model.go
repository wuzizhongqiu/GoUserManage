package model

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