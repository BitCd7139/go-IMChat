package chat

import (
	"IMChat/pkg/constants"
	"IMChat/pkg/zlog"
	//"encoding/json"
	//"errors"
	"fmt"
	//"github.com/go-redis/redis/v8"
	//"github.com/gorilla/websocket"
	"log"
	//"reflect"
	"strings"
	"sync"
	//"time"
)

type Server struct {
	Clients  map[string]*Client
	mutex    *sync.Mutex
	Transmit chan []byte
	Login    chan *Client
	Logout   chan *Client
}

var ChatServer *Server

func init() {
	if ChatServer == nil {
		ChatServer = &Server{
			mutex:    new(sync.Mutex),
			Transmit: make(chan []byte, constants.CHANNEL_SIZE),
			Login:    make(chan *Client, constants.CHANNEL_SIZE),
			Logout:   make(chan *Client, constants.CHANNEL_SIZE),
		}
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
	defer func() {
		close(s.Transmit)
		close(s.Login)
		close(s.Logout)
	}()
	for {
		select {
		case client := <-s.Login:
			{
				s.mutex.Lock()
				s.Clients[client.Uuid] = client
				s.mutex.Unlock()

				go client.Write()
				zlog.Debug(fmt.Sprintf("欢迎新用户 %s 加入聊天室", client.Uuid))

				welcomeMsg := &MessageBack{
					Message: []byte("欢迎来到聊天室"),
					Uuid:    "SYSTEM_MSG",
				}
				client.SendBack <- welcomeMsg
			}
			//TODO Logout & Transmit
		case client := <-s.Logout:
			{
				s.mutex.Lock()
				delete(s.Clients, client.Uuid)
				s.mutex.Unlock()

				quitMsg := &MessageBack{
					Message: []byte("您已退出聊天室"),
					Uuid:    "SYSTEM_MSG",
				}
				client.SendBack <- quitMsg
			}

		}
	}
}

func (s *Server) Close() {
	close(s.Transmit)
	close(s.Login)
	close(s.Logout)
}
