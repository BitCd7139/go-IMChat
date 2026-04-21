package v1

import (
	"IMChat/internal/dto/request"
	"IMChat/internal/service/gorm"
	"IMChat/pkg/constants"
	"IMChat/pkg/zlog"
	"log"
	"net/http"

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
	message, data, ret := gorm.UserInfoService.GetUserList(userListReq)
	JsonBack(c, message, ret, data)
}

func GetContactInfo(c *gin.Context) {
	var getContactReq request.GetContactInfoRequest
	if err := c.BindJSON(&getContactReq); err != nil {
		zlog.Error(err.Error())
		JsonBack(c, constants.SYSTEM_ERROR, 500, nil)
		return
	}
	log.Println(getContactReq)
	message, contactInfo, ret := gorm.UserContactService.GetContactInfo(getContactReq.ContactId)
	JsonBack(c, message, ret, contactInfo)
}

// LoadMyJoinedGroup 获取我加入的群聊
func LoadMyJoinedGroup(c *gin.Context) {
	var loadMyJoinedGroupReq request.OwnlistRequest
	if err := c.BindJSON(&loadMyJoinedGroupReq); err != nil {
		zlog.Error(err.Error())
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": constants.SYSTEM_ERROR,
		})
		return
	}
	message, groupList, ret := gorm.UserContactService.LoadMyJoinedGroup(loadMyJoinedGroupReq.OwnerId)
	JsonBack(c, message, ret, groupList)
}

func DeleteContact(c *gin.Context) {
	var deleteContactReq request.DeleteContactRequest
	if err := c.BindJSON(&deleteContactReq); err != nil {
		zlog.Error(err.Error())
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": constants.SYSTEM_ERROR,
		})
		return
	}
	message, ret := gorm.UserContactService.DeleteContact(deleteContactReq.OwnerId, deleteContactReq.ContactId)
	JsonBack(c, message, ret, nil)
}

// RefuseContactApply 拒绝联系人申请
func RefuseContactApply(c *gin.Context) {
	var passContactApplyReq request.PassContactApplyRequest
	if err := c.BindJSON(&passContactApplyReq); err != nil {
		zlog.Error(err.Error())
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": constants.SYSTEM_ERROR,
		})
		return
	}
	message, ret := gorm.UserContactService.RefuseContactApply(passContactApplyReq.OwnerId, passContactApplyReq.ContactId)
	JsonBack(c, message, ret, nil)
}

func BlackContact(c *gin.Context) {
	var req request.BlackContactRequest
	if err := c.BindJSON(&req); err != nil {
		zlog.Error(err.Error())
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": constants.SYSTEM_ERROR,
		})
		return
	}
	message, ret := gorm.UserContactService.BlackContact(req.OwnerId, req.ContactId)
	JsonBack(c, message, ret, nil)
}

// CancelBlackContact 解除拉黑联系人
func CancelBlackContact(c *gin.Context) {
	var req request.BlackContactRequest
	if err := c.BindJSON(&req); err != nil {
		zlog.Error(err.Error())
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": constants.SYSTEM_ERROR,
		})
		return
	}
	message, ret := gorm.UserContactService.CancelBlackContact(req.OwnerId, req.ContactId)
	JsonBack(c, message, ret, nil)
}

// GetAddGroupList 获取新的群聊申请列表
func GetAddGroupList(c *gin.Context) {
	var req request.AddGroupListRequest
	if err := c.BindJSON(&req); err != nil {
		zlog.Error(err.Error())
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": constants.SYSTEM_ERROR,
		})
		return
	}
	message, data, ret := gorm.UserContactService.GetAddGroupList(req.GroupId)
	JsonBack(c, message, ret, data)
}

// BlackApply 拉黑申请
func BlackApply(c *gin.Context) {
	var req request.BlackApplyRequest
	if err := c.BindJSON(&req); err != nil {
		zlog.Error(err.Error())
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": constants.SYSTEM_ERROR,
		})
		return
	}
	message, ret := gorm.UserContactService.BlackApply(req.OwnerId, req.ContactId)
	JsonBack(c, message, ret, nil)
}
