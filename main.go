package main

import (
	"github.com/SATA260/Doge-Token/Doge"
	"github.com/SATA260/Doge-Token/InfoManager/local_info_manager"
	"github.com/SATA260/Doge-Token/doge_helper"
	"github.com/SATA260/Doge-Token/model"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	doge_helper.Init(local_info_manager.Init(), "", 0)
	doge := Doge.Init(doge_helper.GetInstance(), "http://localhost:8080", "/login", "/ping")

	r.Use(doge.Middleware())

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjg2NDAxNzUyMDY4MTgyLCJsb2dpbl9pbmZvIjoie1widXNlcklkXCI6XCIxMjM0NVwiLFwidXNlcm5hbWVcIjpcIuaIkOmDveS_oeaBr-W3peeoi-Wkp-WtplwiLFwiZXh0cmFJbmZvXCI6XCLmtYvor5XnlKjmiLfmlbDmja5cIixcInJvbGVMaXN0XCI6bnVsbCxcInBlcm1pc3Npb25MaXN0XCI6bnVsbCxcImV4cGlyZVRpbWVcIjowLFwiYXV0b1JlbmV3XCI6ZmFsc2UsXCJ2ZXJzaW9uXCI6XCJcIn0ifQ.rDrf2emg_a7heyhZfWpVHepDTqYjBN0Eexqfw0bsHck"
	loginInfo := model.LoginInfo{
		UserId:    "12345",
		Username:  "成都信息工程大学",
		ExtraInfo: "测试用户数据",
	}

	r.GET("/login", func(c *gin.Context) {
		info := doge_helper.GetInstance().Login(token, &loginInfo)
		c.JSON(200, info)
	})

	r.GET("/logout", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "logout",
		})
	})

	r.Run(":8080")
}
