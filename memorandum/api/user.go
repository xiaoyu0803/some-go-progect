package api

import (
	"todo_list/service"

	"github.com/gin-gonic/gin"
)

func UserRegister(c *gin.Context) {
	var userRegister service.UsersService
	if err := c.ShouldBind(&userRegister); err == nil {
		res := userRegister.Register()
		c.JSON(200, res)
	} else {
		c.JSON(400, err)
	}
}

func UserLogin(c *gin.Context) {
	var userLogin service.UsersService
	if err := c.ShouldBind(&userLogin); err == nil {
		res := userLogin.Login()
		c.JSON(200, res)
	} else {
		c.JSON(400, err)
	}
}
