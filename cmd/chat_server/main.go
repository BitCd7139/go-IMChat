package main

import (
	"IMChat/internal/config"
	"IMChat/internal/server/http_server"
	"IMChat/internal/service/chat"
	myredis "IMChat/internal/service/redis"

	"IMChat/pkg/zlog"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	conf := config.GetConfig()
	host := conf.MainConfig.Host
	port := conf.MainConfig.Port
	messageMode := conf.KafkaConfig.MessageMode

	chatLoopStarted := false
	if messageMode == "channel" || messageMode == "redis_pubsub" {
		go chat.ChatServer.Start()
		chatLoopStarted = true
	}
	if messageMode == "redis_pubsub" {
		chat.StartPubSubGateway()
	}

	go func() {
		if err := http_server.GE.RunTLS(fmt.Sprintf("%s:%d", host, port), "pkg/ssl/localhost+2.pem", "pkg/ssl/localhost+2-key.pem"); err != nil {
			zlog.Fatal("server running fault")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit

	if messageMode == "redis_pubsub" {
		chat.StopPubSubGateway()
	}

	if chatLoopStarted {
		chat.ChatServer.Close()
	}

	zlog.Info("关闭服务器...")

	if err := myredis.DeleteAllRedisKeys(); err != nil {
		zlog.Error(err.Error())
	} else {
		zlog.Info("所有Redis键已删除")
	}

	zlog.Info("服务器已关闭")
}
