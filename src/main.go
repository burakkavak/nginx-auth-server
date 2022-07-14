package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

var GinMode = gin.DebugMode

func main() {
	gin.SetMode(GinMode)

	router := gin.Default()
	router.LoadHTMLGlob("public/*.html")
	router.Static("/css", "public/css")

	router.GET("/login", login)

	address := GetListenAddress() + ":" + strconv.Itoa(GetListenPort())

	err := router.Run(address)

	if err != nil {
		panic(fmt.Errorf("fatal error trying to launch the webserver: %w", err))
	}
}

func login(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", nil)
}
