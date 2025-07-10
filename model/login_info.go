package model

import "encoding/json"

type LoginInfo struct {
	// user information
	UserId    string `json:"userId"`
	Username  string `json:"username"`
	ExtraInfo string `json:"extraInfo"`

	// role and permission
	RoleList       []string `json:"roleList"`
	PermissionList []string `json:"permissionList"`

	// settings
	ExpireTime int64  `json:"expireTime"`
	AutoRenew  bool   `json:"autoRenew"`
	Version    string `json:"version"`
}

func New(userId string, username string, version string, expireTime int64) *LoginInfo {
	return &LoginInfo{
		UserId:         userId,
		Username:       username,
		ExtraInfo:      "",
		RoleList:       []string{},
		PermissionList: []string{},
		ExpireTime:     expireTime,
		AutoRenew:      false,
		Version:        version,
	}
}

func NewWithPerm(userId string, username string, version string, expireTime int64, permissionList []string) *LoginInfo {
	return &LoginInfo{
		UserId:         userId,
		Username:       username,
		ExtraInfo:      "",
		RoleList:       []string{},
		PermissionList: permissionList,
		ExpireTime:     expireTime,
		AutoRenew:      false,
		Version:        version,
	}
}

func (l *LoginInfo) Serialize() ([]byte, error) {
	marshal, err := json.Marshal(l)
	if err != nil {
		return nil, err
	}

	return marshal, nil
}

func (l *LoginInfo) Deserialize(data []byte) error {
	err := json.Unmarshal(data, &l)
	if err != nil {
		return err
	}
	return nil
}
