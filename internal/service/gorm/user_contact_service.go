package gorm

import (
	"IMChat/internal/dao"
	"IMChat/internal/dto/request"
	"IMChat/internal/dto/response"
	"IMChat/internal/model"
	myredis "IMChat/internal/service/redis"
	"IMChat/pkg/constants"
	"IMChat/pkg/enum/contact/contact_status_enum"
	"IMChat/pkg/enum/contact/contact_type_enum"
	"IMChat/pkg/enum/contact_apply/contact_apply_status_enum"
	"IMChat/pkg/enum/group_info/group_status_enum"
	"IMChat/pkg/enum/user_info/user_status_enum"
	"IMChat/pkg/util/random"
	"IMChat/pkg/zlog"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	//"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type userContactService struct {
}

var UserContactService = new(userContactService)

func (u *userContactService) ApplyContact(req request.ApplyContactRequest) (string, int) {
	if req.ContactId[0] == 'U' || req.ContactId[0] == 'u' {
		var user model.UserInfo
		req.ContactId = "U" + req.ContactId[1:]

		if res := dao.GormDB.First(&user, "uuid = ?", req.ContactId); res.Error != nil {
			if errors.Is(res.Error, gorm.ErrRecordNotFound) {
				zlog.Error("用户不存在")
				return "用户不存在", -2
			} else {
				zlog.Error(res.Error.Error())
				return constants.SYSTEM_ERROR, -1
			}
		}

		if user.Status == user_status_enum.DISABLE {
			zlog.Info("用户已被禁用")
			return "用户已被禁用", -2
		}
		var contactApply model.ContactApply
		if res := dao.GormDB.Where("user_id = ? AND contact_id = ?", req.OwnerId, req.ContactId).First(&contactApply); res.Error != nil {
			if errors.Is(res.Error, gorm.ErrRecordNotFound) {
				contactApply = model.ContactApply{
					Uuid:        fmt.Sprintf("A%s", random.GetNowAndLenRandomString(11)),
					UserId:      req.OwnerId,
					ContactId:   req.ContactId,
					ContactType: contact_type_enum.USER,
					Status:      contact_apply_status_enum.PENDING,
					Message:     req.Message,
					LastApplyAt: time.Now(),
				}
				if res := dao.GormDB.Create(&contactApply); res.Error != nil {
					zlog.Error(res.Error.Error())
					return constants.SYSTEM_ERROR, -1
				}
				return "申请成功", 0
			} else {
				zlog.Error(res.Error.Error())
				return constants.SYSTEM_ERROR, -1
			}
		}
		// 如果存在申请记录，先看看有没有被拉黑
		if contactApply.Status == contact_apply_status_enum.BLACK {
			return "对方已将你拉黑", -2
		}
		contactApply.LastApplyAt = time.Now()
		contactApply.Status = contact_apply_status_enum.PENDING

		if res := dao.GormDB.Save(&contactApply); res.Error != nil {
			zlog.Error(res.Error.Error())
			return constants.SYSTEM_ERROR, -1
		}
		return "申请成功", 0
	} else if req.ContactId[0] == 'G' || req.ContactId[0] == 'g' {
		var group model.GroupInfo
		req.ContactId = "G" + req.ContactId[1:]
		if res := dao.GormDB.First(&group, "uuid = ?", req.ContactId); res.Error != nil {
			if errors.Is(res.Error, gorm.ErrRecordNotFound) {
				zlog.Error("群聊不存在")
				return "群聊不存在", -2
			} else {
				zlog.Error(res.Error.Error())
				return constants.SYSTEM_ERROR, -1
			}
		}
		if group.Status == group_status_enum.DISABLE {
			zlog.Info("群聊已被禁用")
			return "群聊已被禁用", -2
		}
		var contactApply model.ContactApply
		if res := dao.GormDB.Where("user_id = ? AND contact_id = ?", req.OwnerId, req.ContactId).First(&contactApply); res.Error != nil {
			if errors.Is(res.Error, gorm.ErrRecordNotFound) {
				contactApply = model.ContactApply{
					Uuid:        fmt.Sprintf("A%s", random.GetNowAndLenRandomString(11)),
					UserId:      req.OwnerId,
					ContactId:   req.ContactId,
					ContactType: contact_type_enum.GROUP,
					Status:      contact_apply_status_enum.PENDING,
					Message:     req.Message,
					LastApplyAt: time.Now(),
				}
				if res := dao.GormDB.Create(&contactApply); res.Error != nil {
					zlog.Error(res.Error.Error())
					return constants.SYSTEM_ERROR, -1
				}
			} else {
				zlog.Error(res.Error.Error())
				return constants.SYSTEM_ERROR, -1
			}
		}
		contactApply.LastApplyAt = time.Now()

		if res := dao.GormDB.Save(&contactApply); res.Error != nil {
			zlog.Error(res.Error.Error())
			return constants.SYSTEM_ERROR, -1
		}
		return "申请成功", 0
	} else {
		return "用户/群聊不存在", -2
	}

}

func (u *userContactService) PassContactApply(applyRequest request.ContactApplyRequest) (string, int) {
	var contactApply model.ContactApply
	ownerId := applyRequest.OwnerId
	contactId := applyRequest.ContactId

	if res := dao.GormDB.Where("contact_id = ? AND user_id = ?", ownerId, contactId).First(&contactApply); res.Error != nil {
		zlog.Error(res.Error.Error())
		return constants.SYSTEM_ERROR, -1
	}
	if ownerId[0] == 'U' {
		var user model.UserInfo
		if res := dao.GormDB.Where("uuid = ?", contactId).First(&user); res.Error != nil {
			zlog.Error(res.Error.Error())
			return constants.SYSTEM_ERROR, -1
		}
		if user.Status == user_status_enum.DISABLE {
			zlog.Error("用户已被禁用")
			return "用户已被禁用", -2
		}

		contactApply.Status = contact_apply_status_enum.AGREE
		if res := dao.GormDB.Save(&contactApply); res.Error != nil {
			zlog.Error(res.Error.Error())
			return constants.SYSTEM_ERROR, -1
		}

		newContact := model.UserContact{
			UserId:      ownerId,
			ContactId:   contactId,
			ContactType: contact_type_enum.USER,
			Status:      contact_status_enum.NORMAL,
			CreatedAt:   time.Now(),
			UpdateAt:    time.Now(),
		}
		if res := dao.GormDB.Create(&newContact); res.Error != nil {
			zlog.Error(res.Error.Error())
			return constants.SYSTEM_ERROR, -1
		}

		anotherContact := model.UserContact{
			UserId:      contactId,
			ContactId:   ownerId,
			ContactType: contact_type_enum.USER,     // 用户
			Status:      contact_status_enum.NORMAL, // 正常
			CreatedAt:   newContact.CreatedAt,
			UpdateAt:    newContact.UpdateAt,
		}
		if res := dao.GormDB.Create(&anotherContact); res.Error != nil {
			zlog.Error(res.Error.Error())
			return constants.SYSTEM_ERROR, -1
		}

		if err := myredis.SetKeyWithSets(constants.REDIS_USER_CONTACT_PREFIX+ownerId, contactId, 5*time.Second); err != nil {
			zlog.Error(err.Error())
			return constants.SYSTEM_ERROR, -1
		}

		return "已添加该联系人", 0

	} else if ownerId[0] == 'G' {
		var group model.GroupInfo
		if res := dao.GormDB.Where("uuid = ?", ownerId).Find(&group); res.Error != nil {
			zlog.Error(res.Error.Error())
		}
		if group.Status == group_status_enum.DISABLE {
			zlog.Error("群聊已被禁用")
			return "群聊已被禁用", -2
		}
		contactApply.Status = contact_apply_status_enum.AGREE
		if res := dao.GormDB.Save(&contactApply); res.Error != nil {
			zlog.Error(res.Error.Error())
			return constants.SYSTEM_ERROR, -1
		}

		// 群聊就只用创建一个UserContact
		newContact := model.UserContact{
			UserId:      contactId,
			ContactId:   ownerId,
			ContactType: contact_type_enum.GROUP,
			Status:      contact_status_enum.NORMAL,
			CreatedAt:   time.Now(),
			UpdateAt:    time.Now(),
		}
		if res := dao.GormDB.Create(&newContact); res.Error != nil {
			zlog.Error(res.Error.Error())
			return constants.SYSTEM_ERROR, -1
		}
		var members []string
		if err := json.Unmarshal(group.Members, &members); err != nil {
			zlog.Error(err.Error())
			return constants.SYSTEM_ERROR, -1
		}
		members = append(members, contactId)
		group.MemberCnt = len(members)
		group.Members, _ = json.Marshal(members)
		if res := dao.GormDB.Save(&group); res.Error != nil {
			zlog.Error(res.Error.Error())
			return constants.SYSTEM_ERROR, -1
		}

		if err := myredis.SetKeyWithSets(constants.REDIS_GROUP_CONTACT_PREFIX+ownerId, contactId, 5*time.Second); err != nil {
			zlog.Error(err.Error())
			return constants.SYSTEM_ERROR, -1
		}

		return "已通过加群申请", 0
	}

	return constants.SYSTEM_ERROR, -1
}

func (u *userContactService) RefuseContactApply(ownerId string, contactId string) (string, int) {
	// ownerId 如果是用户的话就是登录用户，如果是群聊的话就是群聊id
	var contactApply model.ContactApply
	if res := dao.GormDB.Where("contact_id = ? AND user_id = ?", ownerId, contactId).First(&contactApply); res.Error != nil {
		zlog.Error(res.Error.Error())
		return constants.SYSTEM_ERROR, -1
	}
	contactApply.Status = contact_apply_status_enum.REFUSE
	if res := dao.GormDB.Save(&contactApply); res.Error != nil {
		zlog.Error(res.Error.Error())
		return constants.SYSTEM_ERROR, -1
	}
	if ownerId[0] == 'U' {
		return "已拒绝该联系人申请", 0
	} else {
		return "已拒绝该加群申请", 0
	}

}

// BlackContact 拉黑联系人
func (u *userContactService) BlackContact(ownerId string, contactId string) (string, int) {
	// 拉黑
	if res := dao.GormDB.Model(&model.UserContact{}).Where("user_id = ? AND contact_id = ?", ownerId, contactId).Updates(map[string]interface{}{
		"status":    contact_status_enum.BLACK,
		"update_at": time.Now(),
	}); res.Error != nil {
		zlog.Error(res.Error.Error())
		return constants.SYSTEM_ERROR, -1
	}
	// 被拉黑
	if res := dao.GormDB.Model(&model.UserContact{}).Where("user_id = ? AND contact_id = ?", contactId, ownerId).Updates(map[string]interface{}{
		"status":    contact_status_enum.BE_BLACK,
		"update_at": time.Now(),
	}); res.Error != nil {
		zlog.Error(res.Error.Error())
		return constants.SYSTEM_ERROR, -1
	}
	// 删除会话
	var deletedAt gorm.DeletedAt
	deletedAt.Time = time.Now()
	deletedAt.Valid = true
	if res := dao.GormDB.Model(&model.Session{}).Where("send_id = ? AND receive_id = ?", ownerId, contactId).Update("deleted_at", deletedAt); res.Error != nil {
		zlog.Error(res.Error.Error())
		return constants.SYSTEM_ERROR, -1
	}
	return "已拉黑该联系人", 0
}

func (u *userContactService) GetNewContactList(ownerId string) (string, []response.NewContactResponse, int) {
	var contactApplyList []model.ContactApply
	if res := dao.GormDB.Where("contact_id = ? AND status = ?", ownerId, contact_apply_status_enum.PENDING).Find(&contactApplyList); res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			message := "没有新的好友申请"
			zlog.Info(message)
			return message, nil, -1
		} else {
			zlog.Error(res.Error.Error())
			return constants.SYSTEM_ERROR, nil, -1
		}
	}

	var resp []response.NewContactResponse
	for _, contactApply := range contactApplyList {
		var message string
		if contactApply.Message == "" {
			message = "申请理由：无"
		} else {
			message = "申请理由：" + contactApply.Message
		}
		newContact := response.NewContactResponse{
			ContactId: contactApply.Uuid,
			Message:   message,
		}
		var user model.UserInfo
		if res := dao.GormDB.First(&user, "uuid = ?", contactApply.UserId); res.Error != nil {
			return constants.SYSTEM_ERROR, nil, -1
		}
		newContact.ContactId = user.Uuid
		newContact.ContactName = user.Nickname
		newContact.ContactAvatar = user.Avatar
		resp = append(resp, newContact)
	}
	return "获取成功", resp, 0
}

func (u *userContactService) GetContactInfo(contactId string) (string, response.GetContactInfoResponse, int) {
	if contactId[0] == 'G' {
		var group model.GroupInfo
		if res := dao.GormDB.First(&group, "uuid = ?", contactId); res.Error != nil {
			zlog.Error(res.Error.Error())
			return constants.SYSTEM_ERROR, response.GetContactInfoResponse{}, -1
		}
		// 没被禁用
		if group.Status != group_status_enum.DISABLE {
			return "获取联系人信息成功", response.GetContactInfoResponse{
				ContactId:        group.Uuid,
				ContactName:      group.Name,
				ContactAvatar:    group.Avatar,
				ContactNotice:    group.Notice,
				ContactAddMode:   group.AddMode,
				ContactMembers:   group.Members,
				ContactMemberCnt: group.MemberCnt,
				ContactOwnerId:   group.OwnerId,
			}, 0
		} else {
			zlog.Error("该群聊处于禁用状态")
			return "该群聊处于禁用状态", response.GetContactInfoResponse{}, -2
		}
	} else {
		var user model.UserInfo
		if res := dao.GormDB.First(&user, "uuid = ?", contactId); res.Error != nil {
			zlog.Error(res.Error.Error())
			return constants.SYSTEM_ERROR, response.GetContactInfoResponse{}, -1
		}
		log.Println(user)
		if user.Status != user_status_enum.DISABLE {
			return "获取联系人信息成功", response.GetContactInfoResponse{
				ContactId:        user.Uuid,
				ContactName:      user.Nickname,
				ContactAvatar:    user.Avatar,
				ContactBirthday:  user.Birthday,
				ContactEmail:     user.Email,
				ContactPhone:     user.Telephone,
				ContactGender:    user.Gender,
				ContactSignature: user.Signature,
			}, 0
		} else {
			zlog.Info("该用户处于禁用状态")
			return "该用户处于禁用状态", response.GetContactInfoResponse{}, -2
		}
	}
}

// CancelBlackContact 取消拉黑联系人
func (u *userContactService) CancelBlackContact(ownerId string, contactId string) (string, int) {
	// 因为前端的设定，这里需要判断一下ownerId和contactId是不是有拉黑和被拉黑的状态
	var blackContact model.UserContact
	if res := dao.GormDB.Where("user_id = ? AND contact_id = ?", ownerId, contactId).First(&blackContact); res.Error != nil {
		zlog.Error(res.Error.Error())
		return constants.SYSTEM_ERROR, -1
	}
	if blackContact.Status != contact_status_enum.BLACK {
		return "未拉黑该联系人，无需解除拉黑", -2
	}
	var beBlackContact model.UserContact
	if res := dao.GormDB.Where("user_id = ? AND contact_id = ?", contactId, ownerId).First(&beBlackContact); res.Error != nil {
		zlog.Error(res.Error.Error())
		return constants.SYSTEM_ERROR, -1
	}
	if beBlackContact.Status != contact_status_enum.BE_BLACK {
		return "该联系人未被拉黑，无需解除拉黑", -2
	}

	// 取消拉黑
	blackContact.Status = contact_status_enum.NORMAL
	beBlackContact.Status = contact_status_enum.NORMAL
	if res := dao.GormDB.Save(&blackContact); res.Error != nil {
		zlog.Error(res.Error.Error())
		return constants.SYSTEM_ERROR, -1
	}
	if res := dao.GormDB.Save(&beBlackContact); res.Error != nil {
		zlog.Error(res.Error.Error())
		return constants.SYSTEM_ERROR, -1
	}
	return "已解除拉黑该联系人", 0
}

// BlackApply 拉黑申请
func (u *userContactService) BlackApply(ownerId string, contactId string) (string, int) {
	var contactApply model.ContactApply
	if res := dao.GormDB.Where("contact_id = ? AND user_id = ?", ownerId, contactId).First(&contactApply); res.Error != nil {
		zlog.Error(res.Error.Error())
		return constants.SYSTEM_ERROR, -1
	}
	contactApply.Status = contact_apply_status_enum.BLACK
	if res := dao.GormDB.Save(&contactApply); res.Error != nil {
		zlog.Error(res.Error.Error())
		return constants.SYSTEM_ERROR, -1
	}
	return "已拉黑该申请", 0
}

func (u *userContactService) DeleteContact(ownerId, contactId string) (string, int) {
	// status改变为删除
	var deletedAt gorm.DeletedAt
	deletedAt.Time = time.Now()
	deletedAt.Valid = true
	if res := dao.GormDB.Model(&model.UserContact{}).Where("user_id = ? AND contact_id = ?", ownerId, contactId).Updates(map[string]interface{}{
		"deleted_at": deletedAt,
		"status":     contact_status_enum.DELETE,
	}); res.Error != nil {
		zlog.Error(res.Error.Error())
		return constants.SYSTEM_ERROR, -1
	}

	if res := dao.GormDB.Model(&model.UserContact{}).Where("user_id = ? AND contact_id = ?", contactId, ownerId).Updates(map[string]interface{}{
		"deleted_at": deletedAt,
		"status":     contact_status_enum.BE_DELETE,
	}); res.Error != nil {
		zlog.Error(res.Error.Error())
		return constants.SYSTEM_ERROR, -1
	}

	if res := dao.GormDB.Model(&model.Session{}).Where("send_id = ? AND receive_id = ?", ownerId, contactId).Update("deleted_at", deletedAt); res.Error != nil {
		zlog.Error(res.Error.Error())
		return constants.SYSTEM_ERROR, -1
	}

	if res := dao.GormDB.Model(&model.Session{}).Where("send_id = ? AND receive_id = ?", contactId, ownerId).Update("deleted_at", deletedAt); res.Error != nil {
		zlog.Error(res.Error.Error())
		return constants.SYSTEM_ERROR, -1
	}
	// 联系人添加的记录得删，这样之后再添加就看新的申请记录，如果申请记录结果是拉黑就没法再添加，如果是拒绝可以再添加
	if res := dao.GormDB.Model(&model.ContactApply{}).Where("contact_id = ? AND user_id = ?", ownerId, contactId).Update("deleted_at", deletedAt); res.Error != nil {
		zlog.Error(res.Error.Error())
		return constants.SYSTEM_ERROR, -1
	}
	if res := dao.GormDB.Model(&model.ContactApply{}).Where("contact_id = ? AND user_id = ?", contactId, ownerId).Update("deleted_at", deletedAt); res.Error != nil {
		zlog.Error(res.Error.Error())
		return constants.SYSTEM_ERROR, -1
	}
	if err := myredis.DelKeysWithPattern("contact_user_list_" + ownerId); err != nil {
		zlog.Error(err.Error())
	}
	return "删除联系人成功", 0
}

func (u *userContactService) GetAddGroupList(groupId string) (string, []response.AddGroupListResponse, int) {
	var contactApplyList []model.ContactApply
	if res := dao.GormDB.Where("contact_id = ? AND status = ?", groupId, contact_apply_status_enum.PENDING).Find(&contactApplyList); res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			zlog.Info("没有在申请的联系人")
			return "没有在申请的联系人", nil, 0
		} else {
			zlog.Error(res.Error.Error())
			return constants.SYSTEM_ERROR, nil, -1
		}
	}
	var rsp []response.AddGroupListResponse
	for _, contactApply := range contactApplyList {
		var message string
		if contactApply.Message == "" {
			message = "申请理由：无"
		} else {
			message = "申请理由：" + contactApply.Message
		}
		newContact := response.AddGroupListResponse{
			ContactId: contactApply.Uuid,
			Message:   message,
		}
		var user model.UserInfo
		if res := dao.GormDB.First(&user, "uuid = ?", contactApply.UserId); res.Error != nil {
			return constants.SYSTEM_ERROR, nil, -1
		}
		newContact.ContactId = user.Uuid
		newContact.ContactName = user.Nickname
		newContact.ContactAvatar = user.Avatar
		rsp = append(rsp, newContact)
	}
	return "获取成功", rsp, 0
}

func (u *userContactService) LoadMyJoinedGroup(ownerId string) (string, []response.LoadMyJoinedGroupResponse, int) {
	rspString, err := myredis.GetKeyNilError("my_joined_group_list_" + ownerId)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			var contactList []model.UserContact
			// 没有退群，也没有被踢出群聊
			if res := dao.GormDB.Order("created_at DESC").Where("user_id = ? AND status != 6 AND status != 7", ownerId).Find(&contactList); res.Error != nil {
				// 不存在不是业务问题，用Info，return 0
				if errors.Is(res.Error, gorm.ErrRecordNotFound) {
					message := "目前不存在加入的群聊"
					zlog.Info(message)
					return message, nil, 0
				} else {
					zlog.Error(res.Error.Error())
					return constants.SYSTEM_ERROR, nil, -1
				}
			}
			var groupList []model.GroupInfo
			for _, contact := range contactList {
				if contact.ContactId[0] == 'G' {
					// 获取群聊信息
					var group model.GroupInfo
					if res := dao.GormDB.First(&group, "uuid = ?", contact.ContactId); res.Error != nil {
						zlog.Error(res.Error.Error())
						return constants.SYSTEM_ERROR, nil, -1
					}
					// 群没被删除，同时群主不是自己
					// 群主删除或admin删除群聊，status为7，即被踢出群聊，所以不用判断群是否被删除，删除了到不了这步
					if group.OwnerId != ownerId {
						groupList = append(groupList, group)
					}
				}
			}
			var groupListRsp []response.LoadMyJoinedGroupResponse
			for _, group := range groupList {
				groupListRsp = append(groupListRsp, response.LoadMyJoinedGroupResponse{
					GroupId:   group.Uuid,
					GroupName: group.Name,
					Avatar:    group.Avatar,
				})
			}
			rspString, err := json.Marshal(groupListRsp)
			if err != nil {
				zlog.Error(err.Error())
			}
			if err := myredis.SetKeyEx("my_joined_group_list_"+ownerId, string(rspString), time.Minute*constants.REDIS_TIMEOUT); err != nil {
				zlog.Error(err.Error())
			}
			return "获取加入群成功", groupListRsp, 0
		} else {
			zlog.Error(err.Error())
			return constants.SYSTEM_ERROR, nil, -1
		}
	}
	var rsp []response.LoadMyJoinedGroupResponse
	if err := json.Unmarshal([]byte(rspString), &rsp); err != nil {
		zlog.Error(err.Error())
	}
	return "获取加入群成功", rsp, 0
}
