package gorm

import (
	"IMChat/internal/dao"
	"IMChat/internal/dto/request"
	"IMChat/internal/dto/response"
	"IMChat/internal/model"
	myredis "IMChat/internal/service/redis"
	"IMChat/internal/service/sms"
	"IMChat/pkg/constants"
	"IMChat/pkg/enum/user_info/user_status_enum"
	"IMChat/pkg/zlog"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/utils"

	"regexp"
)

type userInfoService struct {
}

var UserInfoService = new(userInfoService)

func (u *userInfoService) checkTelephoneValid(telephone string) bool {
	pattern := `^1[3-9]\d{9}$`
	match, err := regexp.MatchString(pattern, telephone)
	if err != nil {
		zlog.Error(err.Error())
	}
	return match
}

func (u *userInfoService) checkTelephoneExist(telephone string) (string, int) {
	var userInfo model.UserInfo
	if res := dao.GormDB.First(&userInfo, "telephone = ?", telephone); res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			zlog.Info("该手机号未注册")
			return "", 0
		}
		zlog.Error(res.Error.Error())
		return constants.SYSTEM_ERROR, -1
	}
	message := "该手机号已注册"
	zlog.Info(message)
	return message, -2
}

func (u *userInfoService) checkEmailValid(email string) bool {
	pattern := `^[^\s@]+@[^\s@]+\.[^\s@]+$`
	match, err := regexp.MatchString(pattern, email)
	if err != nil {
		zlog.Error(err.Error())
	}
	return match
}

//func (u *userInfoService) checkUserIsAdminOrNot(user model.UserInfo) (string, *response.LoginRespond, int) {
//	return user.IsAdmin
//}

func (u *userInfoService) SendSmsCode(telephone string) (string, int) {
	return sms.VerificationCode(telephone)
}

func (u *userInfoService) Login(loginReq request.LoginRequest) (string, *response.LoginResponse, int) {
	password := loginReq.Password
	var userInfo model.UserInfo
	res := dao.GormDB.First(&userInfo, "telephone = ?", loginReq.Telephone)
	if res.Error != nil {
		messsage := "用户不存在"
		zlog.Error(messsage)
		return messsage, nil, -2
	}
	if userInfo.Password != password {
		message := "密码错误，请重试"
		zlog.Error(message)
		return message, nil, -1
	}

	loginResp := &response.LoginResponse{
		Uuid:      userInfo.Uuid,
		Telephone: userInfo.Telephone,
		Nickname:  userInfo.Nickname,
		Email:     userInfo.Email,
		Avatar:    userInfo.Avatar,
		Gender:    userInfo.Gender,
		Birthday:  userInfo.Birthday,
		Signature: userInfo.Signature,
		IsAdmin:   userInfo.IsAdmin,
		Status:    userInfo.Status,
	}
	year, month, day := userInfo.CreatedAt.Date()
	loginResp.CreatedAt = fmt.Sprintf("%d-%02d-%02d", year, month, day)

	return "登录成功", loginResp, 0
}

func (u *userInfoService) Register(registerReq request.RegisterRequest) (string, *response.LoginResponse, int) {
	key := "auth_code_" + registerReq.Telephone
	code, err := myredis.GetKey(key)
	if err != nil {
		zlog.Error(err.Error())
		return constants.SYSTEM_ERROR, nil, -1
	}
	if code != registerReq.SmScode {
		message := "验证码错误，请重试"
		zlog.Info(message)
		return message, nil, -2
	} else {
		if err := myredis.DelKeyIfExists(key); err != nil {
			zlog.Error(err.Error())
			return constants.SYSTEM_ERROR, nil, -1
		}
	}

	message, ret := u.checkTelephoneExist(registerReq.Telephone)
	if ret != 0 {
		return message, nil, ret
	}

	var newUserInfo model.UserInfo

	int_uuid, redis_err := myredis.IncrRedisWithDefault("global_user_id", 100001)
	if redis_err != nil {
		zlog.Error(redis_err.Error())
		return constants.SYSTEM_ERROR, nil, -1
	}

	newUserInfo.Uuid = "U" + utils.ToString(int_uuid)
	newUserInfo.Telephone = registerReq.Telephone
	newUserInfo.Nickname = registerReq.Nickname
	newUserInfo.Password = registerReq.Password
	newUserInfo.CreatedAt = time.Now()
	newUserInfo.IsAdmin = 0
	newUserInfo.Status = user_status_enum.NORMAL
	newUserInfo.Avatar = "https://cube.elemecdn.com/0/88/03b0d39583f48206768a7534e55bcpng.png\n"

	res := dao.GormDB.Create(&newUserInfo)
	if res.Error != nil {
		zlog.Error(res.Error.Error())
		return constants.SYSTEM_ERROR, nil, -1
	}

	registerResp := &response.LoginResponse{
		Uuid:      newUserInfo.Uuid,
		Telephone: newUserInfo.Telephone,
		Nickname:  newUserInfo.Nickname,
		Email:     newUserInfo.Email,
		Avatar:    newUserInfo.Avatar,
		Gender:    newUserInfo.Gender,
		Birthday:  newUserInfo.Birthday,
		Signature: newUserInfo.Signature,
		IsAdmin:   newUserInfo.IsAdmin,
		Status:    user_status_enum.NORMAL,
	}
	year, month, day := newUserInfo.CreatedAt.Date()
	registerResp.CreatedAt = fmt.Sprintf("%d-%02d-%02d", year, month, day)
	return "注册成功", registerResp, 0
}

func (u *userInfoService) GetUserInfo(userInfoReq request.UserInfoRequest) (string, *[]response.GetUserInfoResponse, int) {
	var rsp []response.GetUserInfoResponse
	cacheKey := constants.USER_INFO_PREFIX + userInfoReq.OwnerId

	// 1. 尝试从 Redis 读取
	redisData, err := myredis.FindKeyWithSets(cacheKey)

	if err == nil && len(redisData) > 0 {
		// --- 关键点：将 Redis 的字符串(JSON) 转化回 Go 结构体 ---
		for _, str := range redisData {
			var item response.GetUserInfoResponse
			// 每一个 str 是一个独立用户的 JSON
			if err := json.Unmarshal([]byte(str), &item); err == nil {
				rsp = append(rsp, item)
			} else {
				zlog.Error(err.Error())
				return constants.SYSTEM_ERROR, nil, -1
			}
		}
		zlog.Debug("从 Redis 获取用户列表成功")
	} else if err == nil {
		// 2. 如果 Redis 没有，则查询数据库
		var users []model.UserInfo
		db := dao.GormDB.Unscoped().Where("uuid = ?", userInfoReq.OwnerId).Find(&users)
		if db.Error != nil {
			zlog.Error(db.Error.Error())
			return constants.SYSTEM_ERROR, nil, -1
		}

		for _, user := range users {
			rp := response.GetUserInfoResponse{
				Uuid:      user.Uuid,
				Telephone: user.Telephone,
				Nickname:  user.Nickname,
				Status:    user.Status,
				IsAdmin:   user.IsAdmin,
				IsDeleted: user.DeletedAt.Valid,
				Avatar:    user.Avatar,
			}
			rsp = append(rsp, rp)
		}
		zlog.Debug("从数据库获取用户列表成功")
		for _, rp := range rsp {
			individualJson, _ := json.Marshal(rp)
			// 每一个用户作为集合的一个成员
			err = myredis.SetKeyWithSets(cacheKey, string(individualJson), 24*time.Hour)
		}
		if err != nil {
			zlog.Error("写入缓存失败")
		}
		zlog.Debug("将用户列表写入 Redis 成功")
	}

	fmt.Println("用户列表: ", rsp)
	return "获取用户列表成功", &rsp, 0
}
