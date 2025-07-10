package local_info_manager

import (
	"github.com/SATA260/Doge-Token/InfoManager"
	"github.com/SATA260/Doge-Token/model"
	"github.com/SATA260/Doge-Token/response"
	"github.com/SATA260/Doge-Token/utils/string_tool"
	"github.com/SATA260/Doge-Token/utils/token_tool"
	"sync"
	"time"
)

// 本地存储登录数据
type LocalLoginStore struct {
	InfoManager.LoginStore
	mutex *sync.RWMutex
	data  map[string]*model.LoginInfo
}

func Init() *LocalLoginStore {
	return &LocalLoginStore{
		mutex: &sync.RWMutex{},
		data:  make(map[string]*model.LoginInfo),
	}
}

func (l *LocalLoginStore) Set(token string, loginInfo *model.LoginInfo, tokenTimeout int64) *response.ResponseMsg {
	tokenInfo, err := token_tool.ParseToken(token)
	if err != nil {
		return response.FailMsg("token Invalid")
	}

	if string_tool.IsBlank(loginInfo.UserId) || string_tool.IsBlank(loginInfo.Username) {
		return response.FailMsg("login info Invalid")
	}

	if string_tool.IsBlank(tokenInfo.UserId) || tokenInfo.UserId != loginInfo.UserId {
		return response.FailMsg("user loginInfo in token_tool Invalid ")
	}

	expireTime := time.Now().Unix() + tokenTimeout
	if expireTime < time.Now().Unix() {
		return response.FailMsg("expire time invalid")
	}

	loginInfo.ExpireTime = expireTime

	l.mutex.Lock()
	l.data[tokenInfo.UserId] = loginInfo
	l.mutex.Unlock()

	return response.SuccessMsg(token)
}

func (l *LocalLoginStore) Get(token string) *model.LoginInfo {
	tokenInfo, err := token_tool.ParseToken(token)
	if err != nil {
		return nil
	}

	l.mutex.RLock()
	loginInfo, exists := l.data[tokenInfo.UserId]
	l.mutex.RUnlock()

	if !exists || loginInfo.Version != tokenInfo.Version {
		return nil
	}

	if loginInfo.ExpireTime < time.Now().Unix() {
		l.mutex.Lock()
		delete(l.data, tokenInfo.UserId)
		l.mutex.Unlock()

		return nil
	}

	return loginInfo
}

func (l *LocalLoginStore) Remove(token string) *response.ResponseMsg {
	tokenInfo, err := token_tool.ParseToken(token)
	if err != nil {
		return response.FailMsg("token Invalid")
	}

	l.mutex.Lock()
	delete(l.data, tokenInfo.UserId)
	l.mutex.Unlock()

	return response.SuccessMsg("Token removed successfully")
}
