package sensfilter

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"time"
)

type DatabaseConf struct {
	DSN       string `json:"dataSourceName" default:"username:password@tcp(127.0.0.1:3306)/db_name?charset=utf8mb4&parseTime=True&loc=Local"`
	TableName string `json:"tableName" default:"sensitive_word"`
}

type SensitiveWord struct {
	ID        uint32    `gorm:"primaryKey;autoIncrement"`
	Word      string    `gorm:"not null;type:varchar(128);uniqueIndex"`
	CreatedAt time.Time `gorm:"column:created_at;type:TIMESTAMP;default:CURRENT_TIMESTAMP"`
}

func CreateMySQLTable(conf *DatabaseConf) error {
	// 连接数据库
	db, err := gorm.Open(mysql.Open(conf.DSN), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	// 修改表名，定义表结构并创建表
	err = db.Table(conf.TableName).AutoMigrate(&SensitiveWord{})
	if err != nil {
		panic(err)
	}
	return nil
}
