package chat

import (
	"IMChat/internal/service/message_chan"
)

// HandleChatMessage delegates to msgpipeline (persist, Redis list cache, routing) and uses this Server for delivery.
func (s *Server) HandleChatMessage(data []byte) error {
	return message_chan.Handle(data, s.deliverOutbound, normalizePath)
}

func (s *Server) deliverOutbound(userID string, wsPayload []byte, msgUUID string) {
	s.sendMessageBackToUser(userID, &MessageBack{Message: wsPayload, Uuid: msgUUID})
}
