package v1

import (
	"IMChat/internal/dto/request"
	"IMChat/internal/service/chat"
	"IMChat/pkg/constants"
	"IMChat/pkg/zlog"
	"net/http"

	"github.com/gin-gonic/gin"
)

func WsLogin(c *gin.Context) {
	clientId := c.Query("client_id")
	if clientId == "" {
		zlog.Error("client_id is empty")
		JsonBack(c, "client_id is empty", http.StatusBadRequest, nil)
		return
	}
	chat.NewClientInit(c, clientId)
}

func WsLogout(c *gin.Context) {
	var req request.WsLogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		zlog.Error("WsLogout error: " + err.Error())
		JsonBack(c, constants.SYSTEM_ERROR, http.StatusInternalServerError, nil)
		return
	}

	message, ret := chat.ClientLogout(req.OwnerId)
	if ret != http.StatusOK {
		zlog.Error("WsLogout error: " + message)
		JsonBack(c, constants.SYSTEM_ERROR, http.StatusInternalServerError, nil)
		return
	}
	JsonBack(c, message, ret, nil)
}
