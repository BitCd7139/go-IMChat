package message_chan

import (
	"IMChat/internal/dao"
	"IMChat/internal/model"
	"IMChat/pkg/zlog"
)

func createMessage(m *model.Message) {
	if res := dao.GormDB.Create(m); res.Error != nil {
		zlog.Error(res.Error.Error())
	}
}
