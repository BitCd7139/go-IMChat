package chat

import (
	"IMChat/internal/config"
	"IMChat/internal/dto/request"
	"strconv"

	"IMChat/pkg/constants"
	"IMChat/pkg/zlog"
	"context"
	"encoding/json"
	"net/http"
	"time"

	mykafka "IMChat/internal/service/kafka"
	myredis "IMChat/internal/service/redis"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/segmentio/kafka-go"
	//"github.com/segmentio/kafka-go"
)

type MessageBack struct {
	Message []byte
	Uuid    string
}

type Client struct {
	Conn          *websocket.Conn
	Uuid          string
	SendTo        chan []byte
	SendBack      chan *MessageBack
	LastHeartbeat time.Time
	ServerId      string
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  2048,
	WriteBufferSize: 2048,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func (c *Client) Read() {
	defer func() {
		ClientLogout(c.Uuid)
	}()

	for {
		err := c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		if err != nil {
			// 没有心跳
			zlog.Error("设置读超时失败: " + err.Error())
			return
		}

		_, jsonMessage, err := c.Conn.ReadMessage()
		if err != nil {
			zlog.Error(err.Error())
			return
		}

		// 处理心跳
		if string(jsonMessage) == "ping" {
			c.LastHeartbeat = time.Now()
			if config.GetConfig().KafkaConfig.MessageMode == "redis_pubsub" && c.Uuid != "" {
				_ = myredis.RefreshUserPresenceTTL(c.Uuid, userPresenceTTL)
			}
			if err := c.Conn.WriteMessage(websocket.TextMessage, []byte("pong")); err != nil {
				zlog.Error(err.Error())
				return
			}
			continue
		}

		// 处理业务消息
		var message request.ChatMessageRequest
		if err := json.Unmarshal(jsonMessage, &message); err != nil {
			zlog.Error("JSON解析失败: " + err.Error())
			continue
		}

		// 把消息丢给“消息中心”（Kafka/Redis PubSub）
		err = Processor.Process(c, jsonMessage)
		if err != nil {
			zlog.Error("消息处理失败: " + err.Error())
			continue
		}
	}
}

func (c *Client) Write() {
	zlog.Debug("ws write goroutine begin")

	defer func() {
		zlog.Debug("ws write goroutine exit")
		// 确保退出时关闭连接
		_ = c.Conn.Close()
	}()

	for {
		select {
		case messageBack, ok := <-c.SendBack:
			// 1. 处理 Channel 被关闭的情况 (优雅退出)
			if !ok {
				// 发送标准的 WebSocket Close 帧
				_ = c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// 2. 设置写入超时时间
			if err := c.Conn.SetWriteDeadline(time.Now().Add(5 * time.Second)); err != nil {
				zlog.Error("设置写超时失败: " + err.Error())
				return // 此时 return 会触发 defer 里的 Close
			}

			// 3. 执行物理写入
			if err := c.Conn.WriteMessage(websocket.TextMessage, messageBack.Message); err != nil {
				zlog.Error("消息发送失败: " + err.Error())
				return // 发送失败通常意味着连接已死，直接退出
			}

			// 4. 【分布式改造点】处理消息状态更新，抛给异步任务或 Kafka
			c.handleMessageAck(messageBack.Uuid)
		}
	}
}

// 独立出一个方法处理已送达的状态更新
func (c *Client) handleMessageAck(uuid string) {
	// 方案A (强烈推荐分布式用法)：把 ACK 扔给 Kafka
	if Processor.Mode == "kafka" {
		ackMsg := map[string]string{"uuid": uuid, "status": "sent"}
		ackBytes, _ := json.Marshal(ackMsg)
		err := mykafka.KafkaService.AckWriter.WriteMessages(context.Background(), kafka.Message{
			Key:   []byte(uuid),
			Value: ackBytes,
		})
		if err != nil {
			zlog.Error(err.Error())
			return
		}
	} else {
		select {
		case Processor.LocalServer.MsgAckChan <- uuid:
			// 成功投递给后台 DB 更新 Worker
		default:
			zlog.Warn("DB更新Worker繁忙，ACK可能延迟")
			// 可以选择丢弃或开启新的备用逻辑
		}
	}
}

func NewClientInit(c *gin.Context, clientId string) {
	kafkaClient := config.GetConfig().KafkaConfig
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		zlog.Error(err.Error())
	}

	client := &Client{
		Conn:     conn,
		Uuid:     clientId,
		SendTo:   make(chan []byte, constants.CHANNEL_SIZE),
		SendBack: make(chan *MessageBack, constants.CHANNEL_SIZE),
		ServerId: strconv.Itoa(config.GetConfig().ServerId),
	}
	if kafkaClient.MessageMode == "channel" || kafkaClient.MessageMode == "redis_pubsub" {
		Processor.LocalServer.SendClientToLogin(client)
	}

	go client.Read()
	go client.Write()
	zlog.Info("ws连接成功: " + clientId + "\n")
}

func ClientLogout(clientId string) (string, int) {
	kafkaConfig := config.GetConfig().KafkaConfig
	client := Processor.LocalServer.GetClient(clientId)
	if client != nil {
		if kafkaConfig.MessageMode == "channel" || kafkaConfig.MessageMode == "redis_pubsub" {
			Processor.LocalServer.SendClientToLogout(client)
		}

		if err := client.Conn.Close(); err != nil {
			zlog.Error(err.Error())
			return constants.SYSTEM_ERROR, -1
		}
		close(client.SendBack)
		close(client.SendTo)
	}
	return "退出成功", 0
}

func (s *Server) SendClientToLogin(client *Client) {
	s.mutex.Lock()
	s.Login <- client
	s.mutex.Unlock()
}

func (s *Server) SendClientToLogout(client *Client) {
	s.mutex.Lock()
	s.Logout <- client
	s.mutex.Unlock()
}

func (s *Server) SendMessageToTransmit(message []byte) {
	s.mutex.Lock()
	s.Transmit <- message
	s.mutex.Unlock()
}

func (s *Server) RemoveClient(uuid string) {
	s.mutex.Lock()
	delete(s.clients, uuid)
	s.mutex.Unlock()
}
