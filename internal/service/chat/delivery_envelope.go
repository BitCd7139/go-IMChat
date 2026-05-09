package chat

// chatDeliveryEnvelope is published to Redis channel chat:node:{serverId}.
type chatDeliveryEnvelope struct {
	TargetUserID string `json:"target_user_id"`
	Uuid         string `json:"uuid"`
	Payload      []byte `json:"payload"`
}
