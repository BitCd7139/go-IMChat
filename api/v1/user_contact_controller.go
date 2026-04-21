package v1

import (
	"IMChat/internal/dto/request"
	"IMChat/internal/service/gorm"
	"IMChat/pkg/constants"
	"IMChat/pkg/zlog"

	"github.com/gin-gonic/gin"
)

func Login(c *gin.Context) {
	var loginReq request.LoginRequest
	if err := c.BindJSON(&loginReq); err != nil {
		zlog.Error(err.Error())
		c.JSON(200, gin.H{
			"code": 500,
			"msg":  constants.SYSTEM_ERROR,
		})
		return
	}
	message, userInfo, ret := gorm.UserInfoService.Login(loginReq)
	JsonBack(c, message, ret, userInfo)
}

func Register(c *gin.Context) {
	var registerReq request.RegisterRequest
	if err := c.BindJSON(&registerReq); err != nil {
		zlog.Error(err.Error())
		c.JSON(200, gin.H{
			"code": 500,
			"msg":  constants.SYSTEM_ERROR,
		})
		return
	}
	message, userInfo, ret := gorm.UserInfoService.Register(registerReq)
	JsonBack(c, message, ret, userInfo)
}

func SendSmsCode(c *gin.Context) {
	var smscCodeReq request.SendSmsCodeRequest
	if err := c.BindJSON(&smscCodeReq); err != nil {
		zlog.Error(err.Error())
		JsonBack(c, constants.SYSTEM_ERROR, 500, nil)
		return
	}
	message, ret := gorm.UserInfoService.SendSmsCode(smscCodeReq.Telephone)
	JsonBack(c, message, ret, nil)
}

func GetNewContactList(c *gin.Context) {
	var req request.OwnlistRequest
	if err := c.BindJSON(&req); err != nil {
		zlog.Error(err.Error())
		JsonBack(c, constants.SYSTEM_ERROR, 500, nil)
		return
	}
	message, data, ret := gorm.UserContactService.GetNewContactList(req.OwnerId)
	JsonBack(c, message, ret, data)
}

func ApplyContact(c *gin.Context) {
	var applyContactReq request.ApplyContactRequest
	if err := c.BindJSON(&applyContactReq); err != nil {
		zlog.Error(err.Error())
		JsonBack(c, constants.SYSTEM_ERROR, 500, nil)
		return
	}
	message, ret := gorm.UserContactService.ApplyContact(applyContactReq)
	zlog.Debug(message)
	JsonBack(c, message, ret, nil)
}

func PassContactApply(c *gin.Context) {
	var passContactReq request.ContactApplyRequest
	if err := c.BindJSON(&passContactReq); err != nil {
		zlog.Error(err.Error())
		JsonBack(c, constants.SYSTEM_ERROR, 500, nil)
		return
	}
	message, ret := gorm.UserContactService.PassContactApply(passContactReq)
	JsonBack(c, message, ret, nil)
}

func GetUserList(c *gin.Context) {
	var userListReq request.UserInfoRequest
	if err := c.BindJSON(&userListReq); err != nil {
		zlog.Error(err.Error())
		JsonBack(c, constants.SYSTEM_ERROR, 500, nil)
		return
	}
	message, data, ret := gorm.UserInfoService.GetUserInfo(userListReq)
	JsonBack(c, message, ret, data)
}
