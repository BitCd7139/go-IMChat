package chat

import (
	"IMChat/internal/config"
	myredis "IMChat/internal/service/redis"
	"IMChat/pkg/zlog"
	"context"
	"encoding/json"
	"strconv"
	"sync"
)

// PubSubGateway subscribes to this node's Redis inbox and forwards payloads to local WebSocket clients.
type PubSubGateway struct {
	mu      sync.Mutex
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	started bool
}

var pubSubGateway = &PubSubGateway{}

// StartPubSubGateway starts the Redis subscriber when messageMode is redis_pubsub.
func StartPubSubGateway() {
	if config.GetConfig().KafkaConfig.MessageMode != "redis_pubsub" {
		return
	}

	pubSubGateway.mu.Lock()
	if pubSubGateway.started {
		pubSubGateway.mu.Unlock()
		return
	}
	pubSubGateway.started = true
	ctx, cancel := context.WithCancel(context.Background())
	pubSubGateway.cancel = cancel
	pubSubGateway.mu.Unlock()

	sid := strconv.Itoa(config.GetConfig().MainConfig.ServerId)
	chName := myredis.NodeInboxChannel(sid)
	pubsub := myredis.Subscribe(ctx, chName)

	pubSubGateway.wg.Add(1)
	go func() {
		defer pubSubGateway.wg.Done()
		defer pubsub.Close()
		msgCh := pubsub.Channel()
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-msgCh:
				if !ok {
					return
				}
				if msg == nil {
					continue
				}
				deliverPubSubPayload([]byte(msg.Payload))
			}
		}
	}()
}

func deliverPubSubPayload(payload []byte) {
	var env chatDeliveryEnvelope
	if err := json.Unmarshal(payload, &env); err != nil {
		zlog.Error("pubsub envelope: " + err.Error())
		return
	}
	mb := &MessageBack{Message: env.Payload, Uuid: env.Uuid}
	if Processor == nil || Processor.LocalServer == nil {
		return
	}
	if cl := Processor.LocalServer.GetClient(env.TargetUserID); cl != nil {
		select {
		case cl.SendBack <- mb:
		default:
			zlog.Warn("gateway SendBack full for " + env.TargetUserID)
		}
	}
}

// StopPubSubGateway cancels the subscriber and waits for the receive loop to exit.
func StopPubSubGateway() {
	pubSubGateway.mu.Lock()
	if !pubSubGateway.started {
		pubSubGateway.mu.Unlock()
		return
	}
	cancel := pubSubGateway.cancel
	pubSubGateway.mu.Unlock()

	if cancel != nil {
		cancel()
	}
	pubSubGateway.wg.Wait()

	pubSubGateway.mu.Lock()
	pubSubGateway.started = false
	pubSubGateway.cancel = nil
	pubSubGateway.mu.Unlock()
}
