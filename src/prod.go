//go:build prod
// +build prod

package main

import "github.com/gin-gonic/gin"

func init() {
	GinMode = gin.ReleaseMode
}
