package chat

import (
	"IMChat/internal/config"
	"IMChat/internal/dao"
	"IMChat/internal/model"
	myredis "IMChat/internal/service/redis"
	"IMChat/pkg/constants"
	"IMChat/pkg/enum/message/message_status_enum"
	"IMChat/pkg/zlog"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// userPresenceTTL matches WebSocket read deadline window (see Client.Read).
const userPresenceTTL = 90 * time.Second

type Server struct {
	clients    map[string]*Client
	mutex      *sync.RWMutex
	Transmit   chan []byte
	Login      chan *Client
	Logout     chan *Client
	MsgAckChan chan string
}

var ChatServer *Server

func NewServer() *Server {
	return &Server{
		mutex:    &sync.RWMutex{},
		clients:  make(map[string]*Client),
		Transmit: make(chan []byte, constants.CHANNEL_SIZE),
		Login:    make(chan *Client, constants.CHANNEL_SIZE),
		Logout:   make(chan *Client, constants.CHANNEL_SIZE),
	}
}

func normalizePath(path string) string {
	if path == "https://cube.elemecdn.com/0/88/03b0d39583f48206768a7534e55bcpng.png" {
		return path
	}
	staticIndex := strings.Index(path, "/static/")
	if staticIndex == -1 {
		log.Println(path)
		zlog.Error("路径不合法")
	}
	return path[staticIndex:]
}

func (s *Server) Start() {
	go s.StartDBWorkers(5)
	defer func() {
		s.Close()
	}()
	for {
		select {
		case client := <-s.Login:
			{
				s.mutex.Lock()
				fmt.Println(client)
				s.clients[client.Uuid] = client
				s.mutex.Unlock()

				if config.GetConfig().KafkaConfig.MessageMode == "redis_pubsub" {
					if err := myredis.SetUserPresence(client.Uuid, client.ServerId, userPresenceTTL); err != nil {
						zlog.Error("SetUserPresence failed: " + err.Error())
					}
				}

				go client.Write()
				zlog.Debug(fmt.Sprintf("欢迎新用户 %s 加入聊天室", client.Uuid))
				err := client.Conn.WriteMessage(websocket.TextMessage, []byte("欢迎来到聊天服务器"))
				if err != nil {
					zlog.Error(err.Error())
				}

				welcomeMsg := &MessageBack{
					Message: []byte("欢迎来到聊天室"),
					Uuid:    "SYSTEM_MSG",
				}
				client.SendBack <- welcomeMsg
			}
			//TODO Logout & Transmit
		case client := <-s.Logout:
			{
				if config.GetConfig().KafkaConfig.MessageMode == "redis_pubsub" {
					_ = myredis.DelUserPresence(client.Uuid)
				}
				s.mutex.Lock()
				delete(s.clients, client.Uuid)
				s.mutex.Unlock()
				zlog.Debug(fmt.Sprintf("用户 %s 已退出聊天室", client.Uuid))
				if err := client.Conn.WriteMessage(websocket.TextMessage, []byte("已退出登录")); err != nil {
					zlog.Error(err.Error())
				}

				quitMsg := &MessageBack{
					Message: []byte("您已退出聊天室"),
					Uuid:    "SYSTEM_MSG",
				}
				client.SendBack <- quitMsg
			}
		case data := <-s.Transmit:
			if err := s.HandleChatMessage(data); err != nil {
				zlog.Error(err.Error())
			}

		}
	}
}

func (s *Server) Close() {
	close(s.Transmit)
	close(s.Login)
	close(s.Logout)
	close(s.MsgAckChan)
}

func (s *Server) GetClient(clientId string) *Client {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.clients[clientId]
}

func (s *Server) SetClient(clientId string, client *Client) {
	s.mutex.Lock()
	s.clients[clientId] = client
	s.mutex.Unlock()
}

func (s *Server) StartDBWorkers(workerCount int) {
	// 初始化带缓冲的 channel，容量根据业务量定
	s.MsgAckChan = make(chan string, 10000)

	// 启动固定数量的协程专门写数据库
	for i := 0; i < workerCount; i++ {
		go func(id int) {
			zlog.Info(fmt.Sprintf("DB Worker %d 启动", id))
			for uuid := range s.MsgAckChan {
				// 这里的代码就是你原来的 DB 更新逻辑
				if res := dao.GormDB.Model(&model.Message{}).
					Where("uuid = ?", uuid).
					Update("status", message_status_enum.Sent); res.Error != nil {
					zlog.Error("更新消息状态失败: " + res.Error.Error())
				}
			}
		}(i)
	}
}
