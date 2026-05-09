package message_chan

import (
	"encoding/json"

	"IMChat/internal/dao"
	"IMChat/internal/model"
	"IMChat/pkg/zlog"
)

func groupMemberUUIDs(groupUUID string) ([]string, error) {
	var group model.GroupInfo
	if res := dao.GormDB.Where("uuid = ?", groupUUID).First(&group); res.Error != nil {
		zlog.Error(res.Error.Error())
		return []string{}, nil
	}
	var members []string
	if err := json.Unmarshal(group.Members, &members); err != nil {
		zlog.Error(err.Error())
		return nil, err
	}
	return members, nil
}
