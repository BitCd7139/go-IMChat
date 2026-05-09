package message_chan

import (
	"encoding/json"
	"errors"
	"time"

	"IMChat/internal/dto/response"
	myredis "IMChat/internal/service/redis"
	"IMChat/pkg/constants"
	"IMChat/pkg/zlog"

	"github.com/go-redis/redis/v8"
)

func appendUserMessageListIfCached(sendID, recvID string, row response.GetMessageListResponse) {
	key := "message_list_" + sendID + "_" + recvID
	rspString, err := myredis.GetKeyNilError(key)
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			zlog.Error(err.Error())
		}
		return
	}
	var rsp []response.GetMessageListResponse
	if err := json.Unmarshal([]byte(rspString), &rsp); err != nil {
		zlog.Error(err.Error())
		return
	}
	rsp = append(rsp, row)
	rspByte, err := json.Marshal(rsp)
	if err != nil {
		zlog.Error(err.Error())
		return
	}
	if err := myredis.SetKeyEx(key, string(rspByte), time.Minute*constants.REDIS_TIMEOUT); err != nil {
		zlog.Error(err.Error())
	}
}

func appendGroupMessageListIfCached(groupID string, row response.GetGroupMessageListResponse) {
	key := "group_messagelist_" + groupID
	rspString, err := myredis.GetKeyNilError(key)
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			zlog.Error(err.Error())
		}
		return
	}
	var rsp []response.GetGroupMessageListResponse
	if err := json.Unmarshal([]byte(rspString), &rsp); err != nil {
		zlog.Error(err.Error())
		return
	}
	rsp = append(rsp, row)
	rspByte, err := json.Marshal(rsp)
	if err != nil {
		zlog.Error(err.Error())
		return
	}
	if err := myredis.SetKeyEx(key, string(rspByte), time.Minute*constants.REDIS_TIMEOUT); err != nil {
		zlog.Error(err.Error())
	}
}
