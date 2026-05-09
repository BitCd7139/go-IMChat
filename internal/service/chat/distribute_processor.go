package chat

import (
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

func (dp *DistributedProcessor) Process(fromClient *Client, message []byte) error {
	var chatMsg model.Message
	if err := json.Unmarshal(message, &chatMsg); err != nil {
		return errors.New("invalid message format")
	}

	if dp.Mode == "kafka" {
		// 使用 RoomID 或 ToUser 作为 Key，确保相关消息进入同一分区（保证顺序）
		routingKey := chatMsg.ReceiveId
		if routingKey == "" {
			routingKey = chatMsg.ReceiveId
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

	} else {
		// 【单机 fallback 模式】 尝试非阻塞写入全局 Transmit
		select {
		case dp.Transmit <- message:
			// 成功写入全局转发 channel
		default:
			// 全局 channel 满了，尝试写入用户自己的缓存 channel
			select {
			case fromClient.SendTo <- message:
				// 成功写入用户缓存 channel
			default:
				// 都满了，直接返回错误给客户端
				errMsg := []byte("系统繁忙，消息发送失败，请稍后重试")
				fromClient.Conn.WriteMessage(websocket.TextMessage, errMsg)
				return errors.New("server is too busy, channel full")
			}
		}
		return nil
	}
}

func (dp *DistributedProcessor) Distribute(message []byte) error {
	// 1. 解析消息，知道这条消息要发给谁
	var chatMsg model.Message
	if err := json.Unmarshal(message, &chatMsg); err != nil {
		zlog.Error("反序列化消息失败: " + err.Error())
		return err
	}

	// 2. 根据目标查找本地连接
	// 如果是单聊：
	if chatMsg.ReceiveId != "" {
		targetClient := dp.LocalServer.GetClient(chatMsg.ReceiveId)
		if targetClient == nil {
			// 目标用户不在此节点，直接返回 nil 即可，连着目标用户的那个节点的 Consumer 会处理它的
			return nil
		}

		// 找到用户了，推送消息
		return dp.sendToClient(targetClient, message)
	}

	// TODO:群聊逻辑

	return errors.New("unknown message routing type")
}

func (dp *DistributedProcessor) sendToClient(client *Client, message []byte) error {
	select {
	case client.SendTo <- message:
		return nil
	default:
		// 客户端本地 channel 已满（可能是网络慢，或者恶意刷屏）
		// 策略：可以踢掉该用户，或者丢弃消息
		zlog.Error("客户端接收通道已满，丢弃消息")
		// client.Conn.Close() // 可选：踢下线
		return errors.New("client channel full")
	}
}
