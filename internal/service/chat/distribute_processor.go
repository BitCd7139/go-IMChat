package chat

import (
	"IMChat/internal/config"
	"IMChat/internal/dto/request"
	"IMChat/internal/model"
	"IMChat/pkg/zlog"
	"context"
	"encoding/json"
	"errors"

	"github.com/gorilla/websocket"
	"github.com/segmentio/kafka-go"
)

type DistributedProcessor struct {
	Mode        string
	Transmit    chan []byte
	KafkaWriter *kafka.Writer
	LocalServer *Server
}

var Processor *DistributedProcessor

func init() {
	conf := config.GetConfig()
	ChatServer = NewServer()
	Processor = &DistributedProcessor{
		Mode:        conf.KafkaConfig.MessageMode,
		Transmit:    ChatServer.Transmit,
		KafkaWriter: &kafka.Writer{},
		LocalServer: ChatServer,
	}
}

func (dp *DistributedProcessor) Process(fromClient *Client, message []byte) error {
	var req request.ChatMessageRequest
	if err := json.Unmarshal(message, &req); err != nil {
		return errors.New("invalid message format")
	}

	if dp.Mode == "redis_pubsub" {
		return dp.LocalServer.HandleChatMessage(message)
	}

	if dp.Mode == "kafka" {
		routingKey := req.ReceiveId
		if routingKey == "" {
			routingKey = req.SendId
		}

		err := dp.KafkaWriter.WriteMessages(context.Background(), kafka.Message{
			Key:   []byte(routingKey),
			Value: message,
		})
		if err != nil {
			zlog.Error("Kafka写入失败: " + err.Error())
			return err
		}
		zlog.Info("消息已发布到Kafka: " + string(message))
		return nil
	}

	// 单机 channel 模式：非阻塞写入与 LocalServer 共用的 Transmit
	select {
	case dp.Transmit <- message:
	default:
		select {
		case fromClient.SendTo <- message:
		default:
			errMsg := []byte("系统繁忙，消息发送失败，请稍后重试")
			_ = fromClient.Conn.WriteMessage(websocket.TextMessage, errMsg)
			return errors.New("server is too busy, channel full")
		}
	}
	return nil
}

func (dp *DistributedProcessor) Distribute(message []byte) error {
	var chatMsg model.Message
	if err := json.Unmarshal(message, &chatMsg); err != nil {
		zlog.Error("反序列化消息失败: " + err.Error())
		return err
	}

	if chatMsg.ReceiveId != "" {
		targetClient := dp.LocalServer.GetClient(chatMsg.ReceiveId)
		if targetClient == nil {
			return nil
		}
		return dp.sendToClient(targetClient, message)
	}

	return errors.New("unknown message routing type")
}

func (dp *DistributedProcessor) sendToClient(client *Client, message []byte) error {
	select {
	case client.SendTo <- message:
		return nil
	default:
		zlog.Error("客户端接收通道已满，丢弃消息")
		return errors.New("client channel full")
	}
}
