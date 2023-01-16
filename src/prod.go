//go:build prod
// +build prod

package main

import "github.com/gin-gonic/gin"

// This file is only applicable when the executable is build with the 'prod' tag.

func init() {
	GinMode = gin.ReleaseMode
}
