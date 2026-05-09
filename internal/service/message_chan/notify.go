package message_chan

import (
	"encoding/json"

	"IMChat/internal/dto/request"
	"IMChat/internal/dto/response"
	"IMChat/internal/model"
	"IMChat/pkg/zlog"
)

func buildUserListRow(m model.Message, req *request.ChatMessageRequest) response.GetMessageListResponse {
	return response.GetMessageListResponse{
		SendId:     m.SendId,
		SendName:   m.SendName,
		SendAvatar: req.SendAvatar,
		ReceiveId:  m.ReceiveId,
		Type:       m.Type,
		Content:    m.Content,
		Url:        m.Url,
		FileSize:   m.FileSize,
		FileName:   m.FileName,
		FileType:   m.FileType,
		CreatedAt:  m.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

func buildGroupListRow(m model.Message, req *request.ChatMessageRequest) response.GetGroupMessageListResponse {
	return response.GetGroupMessageListResponse{
		SendId:     m.SendId,
		SendName:   m.SendName,
		SendAvatar: req.SendAvatar,
		ReceiveId:  m.ReceiveId,
		Type:       m.Type,
		Content:    m.Content,
		Url:        m.Url,
		FileSize:   m.FileSize,
		FileName:   m.FileName,
		FileType:   m.FileType,
		CreatedAt:  m.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

func deliverUserPair(m model.Message, req *request.ChatMessageRequest, deliver Deliver) error {
	row := buildUserListRow(m, req)
	payload, err := json.Marshal(row)
	if err != nil {
		zlog.Error(err.Error())
		return err
	}
	deliver(m.ReceiveId, payload, m.Uuid)
	deliver(m.SendId, payload, m.Uuid)
	appendUserMessageListIfCached(m.SendId, m.ReceiveId, row)
	return nil
}

func deliverGroup(m model.Message, req *request.ChatMessageRequest, deliver Deliver) error {
	row := buildGroupListRow(m, req)
	payload, err := json.Marshal(row)
	if err != nil {
		zlog.Error(err.Error())
		return err
	}
	members, err := groupMemberUUIDs(m.ReceiveId)
	if err != nil {
		return err
	}
	for _, member := range members {
		deliver(member, payload, m.Uuid)
	}
	appendGroupMessageListIfCached(m.ReceiveId, row)
	return nil
}

func routeTextOrFile(m model.Message, req *request.ChatMessageRequest, deliver Deliver) error {
	if len(m.ReceiveId) == 0 {
		return nil
	}
	switch m.ReceiveId[0] {
	case 'U':
		return deliverUserPair(m, req, deliver)
	case 'G':
		return deliverGroup(m, req, deliver)
	}
	return nil
}
