package chat

import (
	"IMChat/internal/config"
	"IMChat/internal/dao"
	//"IMChat/internal/dto/request"
	"IMChat/internal/model"
	"IMChat/pkg/constants"
	"IMChat/pkg/enum/message/message_status_enum"
	"IMChat/pkg/zlog"
	"context"
	//"encoding/json"
	//"log"
	"net/http"
	//"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	//"github.com/segmentio/kafka-go"
)

type MessageBack struct {
	Message []byte
	Uuid    string
}

type Client struct {
	Conn     *websocket.Conn
	Uuid     string
	SendTo   chan []byte
	SendBack chan *MessageBack
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  2048,
	WriteBufferSize: 2048,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

var ctx = context.Background()

var messageMode = config.GetConfig().KafkaConfig.MessageMode

func (c *Client) Read() {
	zlog.Debug("ws read goroutine begin")
	for {
		//_, jsonMessage, err := c.Conn.ReadMessage()
		//if err != nil {
		//	zlog.Error(err.Error())
		//	return
		//} else {
		//	var message = request.ChatMessageRequest{}
		//	if err := json.Unmarshal(jsonMessage, &message); err != nil {
		//		zlog.Error(err.Error())
		//	}
		//	log.Println("接受到消息为: ", jsonMessage)
		//	if messageMode == "channel" {
		//		// 如果server的转发channel没满，先把sendto中的给transmit
		//		for len(ChatServer.Transmit) < constants.CHANNEL_SIZE && len(c.SendTo) > 0 {
		//			sendToMessage := <-c.SendTo
		//			ChatServer.SendMessageToTransmit(sendToMessage)
		//		}
		//		// 如果server没满，sendto空了，直接给server的transmit
		//		if len(ChatServer.Transmit) < constants.CHANNEL_SIZE {
		//			ChatServer.SendMessageToTransmit(jsonMessage)
		//		} else if len(c.SendTo) < constants.CHANNEL_SIZE {
		//			// 如果server满了，直接塞sendto
		//			c.SendTo <- jsonMessage
		//		} else {
		//			// 否则考虑加宽channel size，或者使用kafka
		//			if err := c.Conn.WriteMessage(websocket.TextMessage, []byte("由于目前同一时间过多用户发送消息，消息发送失败，请稍后重试")); err != nil {
		//				zlog.Error(err.Error())
		//			}
		//		}
		//	} else {
		//		if err := myKafka.KafkaService.ChatWriter.WriteMessages(ctx, kafka.Message{
		//			Key:   []byte(strconv.Itoa(config.GetConfig().KafkaConfig.Partition)),
		//			Value: jsonMessage,
		//		}); err != nil {
		//			zlog.Error(err.Error())
		//		}
		//		zlog.Info("已发送消息：" + string(jsonMessage))
		//	}
		//}
	}
}

func (c *Client) Write() {
	zlog.Debug("ws write goroutine begin")
	defer func() {
		err := c.Conn.Close()
		if err != nil {
			zlog.Error(err.Error())
			return
		}
	}()

	for messageBack := range c.SendBack {
		err := c.Conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
		if err != nil {
			zlog.Error(err.Error())
			return
		}

		err = c.Conn.WriteMessage(websocket.TextMessage, messageBack.Message)
		if err != nil {
			zlog.Error(err.Error())
			return
		}

		go func(uuid string) {
			if res := dao.GormDB.Model(&model.Message{}).Where("uuid = ?", uuid).Update("status", message_status_enum.Sent); res.Error != nil {
				zlog.Error(res.Error.Error())
			}
		}(messageBack.Uuid)
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
	}
	if kafkaClient.MessageMode == "channel" {
		//
	} else {
		//kafkaClient
	}
	go client.Read()
	go client.Write()
	zlog.Info("ws连接成功: " + clientId + "\n")
}

func ClientLogout(clientId string) (string, int) {
	kafkaConfig := config.GetConfig().KafkaConfig
	client := ChatServer.Clients[clientId]
	if client != nil {
		if kafkaConfig.MessageMode == "channel" {
			//ChatServer.SendClientToLogout(client)
		} else {
			//KafkaChatServer.SendClientToLogout(client)
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
