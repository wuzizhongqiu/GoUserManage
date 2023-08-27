package utils

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gouse/config"
	"sync"
	"time"
)

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
