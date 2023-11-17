package handler

import (
	"github.com/gin-gonic/gin"
)

func RegisterRouters(r *gin.Engine) {
	configRoute(r)
}
func configRoute(r *gin.Engine) {
	//init
	hello := r.Group("/ping")
	{
		hello.GET("", func(c *gin.Context) {
			c.JSON(200, "pong")
		})
	}
}
