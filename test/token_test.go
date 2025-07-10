package test

import (
	"fmt"
	"github.com/SATA260/Doge-Token/model"
	"github.com/SATA260/Doge-Token/utils/token_tool"
	"testing"
)

func TestToken(t *testing.T) {
	loginInfo := model.LoginInfo{
		UserId:    "12345",
		Username:  "成都信息工程大学",
		ExtraInfo: "测试用户数据",
	}

	fmt.Println("开始测试")
	fmt.Println("原始数据:", loginInfo)

	token, err := token_tool.GenerateToken(&loginInfo)
	if err != nil {
		t.Errorf("生成token失败: %v", err)
		return
	}

	fmt.Println("生成的token:", token)

	parsedLoginInfo, err := token_tool.ParseToken(token)
	if err != nil {
		t.Errorf("解析token失败: %v", err)
		return
	}

	fmt.Println("解析后的登录信息:", parsedLoginInfo)
}
