package main

import (
	"gitee.com/cloudtrek/chatgptbot/bootstrap"
	"gitee.com/cloudtrek/chatgptbot/config"
	log "github.com/sirupsen/logrus"
)

func main() {
	err := config.LoadConfig()
	if err != nil {
		log.Warn("没有找到配置文件，尝试读取环境变量")
	}
	bootstrap.StartWeWorkBot()
}
