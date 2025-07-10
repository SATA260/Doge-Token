package doge_helper

import (
	"github.com/SATA260/Doge-Token/InfoManager"
	"github.com/SATA260/Doge-Token/constant"
	"github.com/SATA260/Doge-Token/model"
	"github.com/SATA260/Doge-Token/response"
	"github.com/SATA260/Doge-Token/utils/cookie_tool"
	"github.com/SATA260/Doge-Token/utils/string_tool"
	"github.com/gin-gonic/gin"
)

// 读取浏览器cookie，param，attr等的工具
type DogeHelper struct {
	LoginStore   InfoManager.LoginStore
	tokenKey     string
	tokenTimeout int64
}

var DogeHelperInstance *DogeHelper

func Init(loginStore InfoManager.LoginStore, tokenKey string, tokenTimeout int64) {
	if string_tool.IsBlank(tokenKey) {
		tokenKey = constant.DOGE_TOKEN
	}

	if tokenTimeout <= 0 {
		tokenTimeout = constant.EXPIRE_TIME_FOR_10_YEAR
	}

	DogeHelperInstance = &DogeHelper{
		LoginStore:   loginStore,
		tokenKey:     tokenKey,
		tokenTimeout: tokenTimeout,
	}
}

func GetInstance() *DogeHelper {
	return DogeHelperInstance
}

// ---------------------- login tool ----------------------
func (instance *DogeHelper) Login(token string, logininfo *model.LoginInfo) *response.ResponseMsg {
	return instance.LoginStore.Set(token, logininfo, GetInstance().tokenTimeout)
}

func (instance *DogeHelper) Logout(token string) *response.ResponseMsg {
	return instance.LoginStore.Remove(token)
}

func (instance *DogeHelper) LoginCheck(token string) *model.LoginInfo {
	loginInfo := instance.LoginStore.Get(token)
	return loginInfo
}

func (instance *DogeHelper) LoginCheckWithAttr(c *gin.Context) *model.LoginInfo {
	value, _ := c.Get(constant.DOGE_USER)
	return value.(*model.LoginInfo)
}

func (instance *DogeHelper) LoginCheckWithHeader(c *gin.Context) *model.LoginInfo {
	token := c.GetHeader(GetInstance().tokenKey)
	return instance.LoginCheck(token)
}

// ---------------------- login cookie tool ----------------------
func (instance *DogeHelper) LoginWithCookie(c *gin.Context, token string, loginInfo *model.LoginInfo, ifRemember bool) *response.ResponseMsg {
	loginResult := instance.Login(token, loginInfo)

	if loginResult.IsSuccess() {
		cookie_tool.SetWithRemember(c, GetInstance().tokenKey, token, ifRemember)
	}

	return loginResult
}

func (instance *DogeHelper) LogoutWithCookie(c *gin.Context) *response.ResponseMsg {
	cookie := cookie_tool.Get(c, instance.tokenKey)
	if string_tool.IsBlank(cookie) {
		return response.SuccessMsg("")
	}

	loginResult := instance.Logout(cookie)
	cookie_tool.Remove(c, instance.tokenKey)

	return loginResult
}

func (instance *DogeHelper) LoginCheckWithCookie(c *gin.Context) *model.LoginInfo {
	token := cookie_tool.Get(c, instance.tokenKey)

	loginInfo := instance.LoginCheck(token)
	if loginInfo == nil {
		cookie_tool.Remove(c, instance.tokenKey)
	}

	return loginInfo
}

func (instance *DogeHelper) LoginCheckWithCookieOrParam(c *gin.Context) *model.LoginInfo {
	loginInfo := instance.LoginCheckWithCookie(c)
	if loginInfo != nil {
		return loginInfo
	}

	token := c.Param(constant.DOGE_TOKEN)
	if string_tool.IsBlank(token) {
		loginInfo = instance.LoginCheck(token)
	}

	if loginInfo != nil {
		cookie_tool.SetWithRemember(c, instance.tokenKey, token, true)
		return loginInfo
	}

	return nil
}

func (instance *DogeHelper) getTokenWithCookie(c *gin.Context) string {
	return cookie_tool.Get(c, instance.tokenKey)
}
