package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/thlz998/openai-bot/bootstrap"
	"github.com/thlz998/openai-bot/config"
	"github.com/thlz998/openai-bot/db"
)

func main() {
	err := config.LoadConfig()
	db.Database()
	if err != nil {
		log.Warn("没有找到配置文件，尝试读取环境变量")
	}
	bootstrap.StartWeWorkBot()
}
