package db

import (
	"fmt"
	"os"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type User struct {
	Username        string `gorm:"comment:'用户名'"`
	Email           string `gorm:"comment:'邮箱'"`
	Password        string `gorm:"comment:'密码'"`
	Status          string `gorm:"comment:'Account status,禁用: disabled, 启用(默认) enabled'"` // 0:正常 1:禁用
	IdType          string `gorm:"comment:'账号类型: default-默认账号, wework-企业微信, feishu-飞书, apple-苹果账号, system-系统账号'"`
	ConversationId  string `gorm:"comment:'当前会话的conversation_id'"`
	ParentMessageId string `gorm:"comment:'最后一次请求的消息ID(下一次请求的parent_message_id)'"`
	gorm.Model
}

type WeWorkAccessToken struct {
	AccessToken string `gorm:"comment:'AccessToken'"`
	gorm.Model
}

var DB *gorm.DB

func Database() {

	db, err := gorm.Open(sqlite.Open("data/data.db"), &gorm.Config{})

	if err != nil {
		fmt.Println("连接数据库不成功")
		os.Exit(1)
	}
	db.Logger.LogMode(1) //开启日志
	//设置连接池
	DB = db

	// 迁移 schema
	DB.AutoMigrate(&User{})
	DB.AutoMigrate(&WeWorkAccessToken{})
}
