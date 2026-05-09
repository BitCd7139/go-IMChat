package message_chan

import (
	"encoding/json"

	"IMChat/internal/dto/request"
	"IMChat/internal/dto/response"
	"IMChat/pkg/zlog"
)

func handleAV(req *request.ChatMessageRequest, deliver Deliver, norm NormalizePath) error {
	var avData request.AVData
	if err := json.Unmarshal([]byte(req.AVdata), &avData); err != nil {
		zlog.Error(err.Error())
	}
	m := avEntity(req)
	if avData.MessageId == "PROXY" && (avData.Type == "start_call" || avData.Type == "receive_call" || avData.Type == "reject_call") {
		m.SendAvatar = norm(m.SendAvatar)
		createMessage(&m)
	}
	if len(req.ReceiveId) == 0 || req.ReceiveId[0] != 'U' {
		return nil
	}
	row := response.AVMessageResponse{
		SendId:     m.SendId,
		SendName:   m.SendName,
		SendAvatar: m.SendAvatar,
		ReceiveId:  m.ReceiveId,
		Type:       m.Type,
		Content:    m.Content,
		Url:        m.Url,
		FileSize:   m.FileSize,
		FileName:   m.FileName,
		FileType:   m.FileType,
		CreatedAt:  m.CreatedAt.Format("2006-01-02 15:04:05"),
		AVdata:     m.AVdata,
	}
	payload, err := json.Marshal(row)
	if err != nil {
		zlog.Error(err.Error())
		return err
	}
	deliver(m.ReceiveId, payload, m.Uuid)
	return nil
}
