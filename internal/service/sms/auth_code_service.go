package sms

import (
	"IMChat/internal/service/redis"
	"IMChat/pkg/constants"
	"IMChat/pkg/util/random"
	"IMChat/pkg/zlog"
	"fmt"
	"strconv"
	"time"
)

func VerificationCode(telephone string) (string, int) {
	key := "auth_code_" + telephone
	code, err := redis.GetKey(key)
	if err != nil {
		zlog.Error(err.Error())
		return constants.SYSTEM_ERROR, -1
	}
	if code != "" {
		message := "目前还不能发送验证码，请输入已发送的验证码"
		zlog.Info(message)
		return message, -2
	}

	code = strconv.Itoa(random.GetRandomInt(6))

	err = redis.SetKeyEx(key, code, 5*time.Minute)
	if err != nil {
		zlog.Error(err.Error())
		return constants.SYSTEM_ERROR, -1
	}

	// 控制台打印替代短信验证
	message := fmt.Sprintf("注册短信测试: %d", code)
	fmt.Println(message)
	zlog.Info(message)

	return "验证码发送成功，请及时在对应电话查收短信", 0
}
