package InfoManager

import (
	"github.com/SATA260/Doge-Token/model"
	"github.com/SATA260/Doge-Token/response"
)

type LoginStore interface {
	Set(token string, loginInfo *model.LoginInfo, tokenTimeout int64) *response.ResponseMsg
	Get(token string) *model.LoginInfo
	Remove(token string) *response.ResponseMsg
}
