package chat

import (
	"IMChat/internal/config"
	myredis "IMChat/internal/service/redis"
	"IMChat/pkg/zlog"
	"context"
	"encoding/json"
	"errors"

	"github.com/go-redis/redis/v8"
)

// sendMessageBackToUser delivers to a local WebSocket client, or publishes to the peer node's Redis inbox.
func (s *Server) sendMessageBackToUser(userID string, mb *MessageBack) {
	if userID == "" || mb == nil {
		return
	}
	s.mutex.RLock()
	cl := s.clients[userID]
	s.mutex.RUnlock()
	if cl != nil {
		select {
		case cl.SendBack <- mb:
		default:
			zlog.Warn("SendBack channel full for user " + userID)
		}
		return
	}
	if config.GetConfig().KafkaConfig.MessageMode != "redis_pubsub" {
		return
	}
	serverID, err := myredis.GetUserPresenceServerID(userID)
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			zlog.Error(err.Error())
		}
		return
	}
	if serverID == "" {
		return
	}
	env := chatDeliveryEnvelope{
		TargetUserID: userID,
		Uuid:         mb.Uuid,
		Payload:      append([]byte(nil), mb.Message...),
	}
	b, err := json.Marshal(&env)
	if err != nil {
		zlog.Error(err.Error())
		return
	}
	chName := myredis.NodeInboxChannel(serverID)
	if err := myredis.PublishCtx(context.Background(), chName, string(b)); err != nil {
		zlog.Error("Redis publish failed: " + err.Error())
	}
}
