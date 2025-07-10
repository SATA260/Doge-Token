package cookie_tool

import (
	"github.com/SATA260/Doge-Token/constant"
	"github.com/gin-gonic/gin"
	"net/url"
)

func Set(c *gin.Context, key, value, domain string, path string, maxAge int, isSecure, isHttpOnly bool) {
	// 编码 value
	encodedValue := url.QueryEscape(value)

	/**
	name：Cookie 的名称。
	value：Cookie 的值。
	maxAge：Cookie 的最大存活时间（以秒为单位）。
	path：Cookie 的路径，默认为当前路径。
	domain：Cookie 的域名。
	secure：一个布尔值，若为 true，则 Cookie 仅通过 HTTPS 连接发送。
	httpOnly：一个布尔值，若为 true，则 JavaScript 无法访问该 Cookie。
	*/
	c.SetCookie(key, encodedValue, maxAge, path, domain, isSecure, isHttpOnly)
}

func SetWithRemember(c *gin.Context, key, value string, ifRemember bool) {
	age := -1
	if ifRemember {
		age = constant.COOKIE_MAX_AGE
	}
	Set(c, key, value, "", constant.COOKIE_PATH, age, false, true)
}

func SetWithMaxAge(c *gin.Context, key, value string, maxAge int) {
	Set(c, key, value, "", constant.COOKIE_PATH, maxAge, false, true)
}

func Remove(c *gin.Context, key string) {
	c.SetCookie(key, "", 0, constant.COOKIE_PATH, "", false, true)
}

func Get(c *gin.Context, key string) string {
	value, err := c.Cookie(key)
	if err != nil {
		return ""
	}
	// 解码 value
	decodedValue, err := url.QueryUnescape(value)
	if err != nil {
		return ""
	}
	return decodedValue
}
