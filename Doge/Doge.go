package Doge

import (
	"github.com/SATA260/Doge-Token/constant"
	"github.com/SATA260/Doge-Token/doge_helper"
	"github.com/SATA260/Doge-Token/path/path_matcher_impl"
	"github.com/SATA260/Doge-Token/utils/string_tool"
	"github.com/gin-gonic/gin"
	"strings"
)

type Doge struct {
	helper         *doge_helper.DogeHelper
	antPathMatcher *path_matcher_impl.AntPathMatcher
	ServerAddr     string `json:"server_addr"`
	LoginPath      string `json:"login_path"`
	ExcludePaths   string `json:"exclude_path"`
}

func Init(dogeHelper *doge_helper.DogeHelper, serverAddr string, loginPath string, excludePaths string) *Doge {
	return &Doge{
		helper:         dogeHelper,
		antPathMatcher: path_matcher_impl.NewAntPathMatcher(),
		ServerAddr:     serverAddr,
		LoginPath:      loginPath,
		ExcludePaths:   excludePaths,
	}
}

func (d *Doge) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		servletPath := c.Request.URL.Path

		// 排除重定向路径以及白名单
		if !string_tool.IsBlank(servletPath) {
			for _, excludePath := range strings.Split(d.ExcludePaths, ",") {
				if string_tool.IsBlank(excludePath) {
					continue
				}

				uriPattern := strings.TrimSpace(excludePath)
				if d.antPathMatcher.Match(uriPattern, servletPath) {
					c.Next()
				}

				if d.antPathMatcher.Match(servletPath, constant.CLIENT_REDIRECT_URL) {
					c.Next()
				}
			}
		}

		// 检查token是否存在
		loginInfo := d.helper.LoginCheckWithCookieOrParam(c)
		if loginInfo == nil {
			// 4、login fail message
			loginFailMsg := gin.H{
				"code": constant.CODE_LOGIN_FAILED,
				"msg":  "not login.",
			}

			// isJson
			header := c.GetHeader("Content-Type")
			isJson := strings.Contains(strings.ToLower(header), "json")
			if isJson {
				// write response
				c.JSON(200, loginFailMsg)
				return
			} else {
				// redirect login-path
				finalLoginPath := d.ServerAddr + constant.CLIENT_REDIRECT_URL

				c.Redirect(302, finalLoginPath)
				return
			}
		}

		c.Next()
	}
}
