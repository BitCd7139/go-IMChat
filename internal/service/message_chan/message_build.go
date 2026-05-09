package message_chan

import (
	"fmt"
	"time"

	"IMChat/internal/dto/request"
	"IMChat/internal/model"
	"IMChat/pkg/enum/message/message_status_enum"
	"IMChat/pkg/util/random"
)

func newMessageUUID() string {
	return fmt.Sprintf("M%s", random.GetNowAndLenRandomString(11))
}

func textEntity(req *request.ChatMessageRequest, norm NormalizePath) model.Message {
	m := model.Message{
		Uuid:       newMessageUUID(),
		SessionId:  req.SessionId,
		Type:       req.Type,
		Content:    req.Content,
		Url:        "",
		SendId:     req.SendId,
		SendName:   req.SendName,
		SendAvatar: norm(req.SendAvatar),
		ReceiveId:  req.ReceiveId,
		FileSize:   "0B",
		FileType:   "",
		FileName:   "",
		Status:     message_status_enum.Unsent,
		CreatedAt:  time.Now(),
		AVdata:     "",
	}
	return m
}

func fileEntity(req *request.ChatMessageRequest, norm NormalizePath) model.Message {
	m := model.Message{
		Uuid:       newMessageUUID(),
		SessionId:  req.SessionId,
		Type:       req.Type,
		Content:    "",
		Url:        req.Url,
		SendId:     req.SendId,
		SendName:   req.SendName,
		SendAvatar: norm(req.SendAvatar),
		ReceiveId:  req.ReceiveId,
		FileSize:   req.FileSize,
		FileType:   req.FileType,
		FileName:   req.FileName,
		Status:     message_status_enum.Unsent,
		CreatedAt:  time.Now(),
		AVdata:     "",
	}
	return m
}

func avEntity(req *request.ChatMessageRequest) model.Message {
	return model.Message{
		Uuid:       newMessageUUID(),
		SessionId:  req.SessionId,
		Type:       req.Type,
		Content:    "",
		Url:        "",
		SendId:     req.SendId,
		SendName:   req.SendName,
		SendAvatar: req.SendAvatar,
		ReceiveId:  req.ReceiveId,
		FileSize:   "",
		FileType:   "",
		FileName:   "",
		Status:     message_status_enum.Unsent,
		CreatedAt:  time.Now(),
		AVdata:     req.AVdata,
	}
}
