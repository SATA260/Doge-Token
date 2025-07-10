package redis_info_manager

import (
	"context"
	"errors"
	"fmt"
	"github.com/SATA260/Doge-Token/InfoManager"
	"github.com/SATA260/Doge-Token/constant"
	"github.com/SATA260/Doge-Token/model"
	"github.com/SATA260/Doge-Token/response"
	"github.com/SATA260/Doge-Token/utils/string_tool"
	"github.com/SATA260/Doge-Token/utils/token_tool"
	"github.com/redis/go-redis/v9"
	"sync"
	"time"
)

// Redis存储登录数据
type RedisLoginStore struct {
	InfoManager.LoginStore
	mutex       sync.RWMutex
	Nodes       string
	Username    string
	Password    string
	keyPrefix   string
	redisClient *redis.Client
}

// 初始化redis,返回info_manager对象
func Init(nodes string, username string, password string, keyPrefix string) *RedisLoginStore {
	rdb := redis.NewClient(&redis.Options{
		Addr:         nodes,
		Username:     username,
		Password:     password,
		DB:           0,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		MaxRetries:   3,
	})

	if string_tool.IsBlank(keyPrefix) {
		keyPrefix = constant.DOGE_USER_STORE_PREFIX
	}

	return &RedisLoginStore{
		Nodes:       nodes,
		Username:    username,
		Password:    password,
		keyPrefix:   keyPrefix,
		redisClient: rdb,
	}
}

func (r *RedisLoginStore) parseStoreKey(loginInfo *model.LoginInfo) string {
	if loginInfo.UserId == "" {
		return ""
	}

	return fmt.Sprintf("%s%s", r.keyPrefix, loginInfo.UserId)
}

func (r *RedisLoginStore) Set(token string, loginInfo *model.LoginInfo, tokenTimeout int64) *response.ResponseMsg {
	tokenInfo, err := token_tool.ParseToken(token)
	if err != nil {
		return response.FailMsg("token Invalid")
	}

	storeKey := r.parseStoreKey(tokenInfo)
	if string_tool.IsBlank(storeKey) {
		return response.FailMsg("token Invalid")
	}

	expireTime := time.Now().Unix() + tokenTimeout
	if expireTime < time.Now().Unix() {
		return response.FailMsg("expire time invalid")
	}
	loginInfo.ExpireTime = expireTime

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	loginInfoSer, err := loginInfo.Serialize()
	if err != nil {
		// 出错时返回错误信息
		return response.FailMsg(fmt.Sprintf("serialize loginInfo failed: %v", err))
	}

	set := r.redisClient.Set(ctx, storeKey, loginInfoSer, time.Duration(tokenTimeout)*time.Second)
	if set.Err() != nil {
		if errors.Is(set.Err(), context.DeadlineExceeded) {
			return response.FailMsg("redis set token out of time")
		}
		return response.FailMsg("redis set token failed")
	}

	return response.SuccessMsg(token)

}

func (r *RedisLoginStore) Get(token string) *model.LoginInfo {
	tokenInfo, err := token_tool.ParseToken(token)
	if err != nil {
		return nil
	}

	storeKey := r.parseStoreKey(tokenInfo)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	redisInfo, err := r.redisClient.Get(ctx, storeKey).Result()
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil
		}
		return nil
	}

	var loginInfo model.LoginInfo
	err = loginInfo.Deserialize([]byte(redisInfo))
	if err != nil {
		return nil
	}

	return &loginInfo
}

func (r *RedisLoginStore) Remove(token string) *response.ResponseMsg {
	tokenInfo, err := token_tool.ParseToken(token)
	if err != nil {
		return response.FailMsg("token Invalid")
	}

	storeKey := r.parseStoreKey(tokenInfo)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	del := r.redisClient.Del(ctx, storeKey)
	if del.Err() != nil {
		if errors.Is(del.Err(), context.DeadlineExceeded) {
			return response.FailMsg("redis del token out of time")
		}
		return response.FailMsg("redis del token failed")
	}

	return response.SuccessMsg("success")
}
