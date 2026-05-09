package message_chan

import (
	"encoding/json"

	"IMChat/internal/dto/request"
	"IMChat/pkg/enum/message/message_type_enum"
)

// Handle parses the wire payload and runs the appropriate pipeline (persist + cache + deliver).
func Handle(data []byte, deliver Deliver, norm NormalizePath) error {
	var req request.ChatMessageRequest
	if err := json.Unmarshal(data, &req); err != nil {
		return err
	}
	switch req.Type {
	case message_type_enum.Text:
		return handleText(&req, deliver, norm)
	case message_type_enum.File:
		return handleFile(&req, deliver, norm)
	case message_type_enum.AudioOrVideo:
		return handleAV(&req, deliver, norm)
	default:
		return nil
	}
}
