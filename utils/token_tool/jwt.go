package token_tool

import (
	"encoding/json"
	"fmt"
	"github.com/SATA260/Doge-Token/constant"
	"github.com/SATA260/Doge-Token/model"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

func ParseToken(tokenString string) (*model.LoginInfo, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(constant.SECRET_KEY), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		loginInfoBytes, ok := claims["login_info"].(string)
		if !ok {
			return nil, fmt.Errorf("Invalid login_info in token_tool")
		}

		var loginInfo model.LoginInfo
		err = json.Unmarshal([]byte(loginInfoBytes), &loginInfo)
		if err != nil {
			return nil, err
		}
		return &loginInfo, nil
	}

	return nil, fmt.Errorf("token Invalid")
}

func GenerateToken(loginInfo *model.LoginInfo) (string, error) {
	loginInfoBytes, err := json.Marshal(loginInfo)
	if err != nil {
		return "", err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"login_info": string(loginInfoBytes),
		"exp":        time.Now().Unix() + constant.DEFAULT_EXPIRE_TIME,
	})

	return token.SignedString([]byte(constant.SECRET_KEY))
}
