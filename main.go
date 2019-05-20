package main

import (
	"legislacion/files"
	"legislacion/user"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func homePage(c *gin.Context) {
	name := c.Query("name")
	lastname := c.DefaultQuery("lastname", "anonymous")
	if len(name) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name empty"})
		return
	} else if len(lastname) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Lastname empty"})
		return
	} else {
		c.String(http.StatusOK, "Hello %s %s", name, lastname)
		return
	}
}

func handlerFunctions() {
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	r.GET("/home", homePage)
	r.POST("/newUser", user.CreateUserHandler)
	r.POST("/login", user.LoginHandler)
	r.POST("/files/send", files.SendFileHandler)
	r.GET("/files", files.ListFilesHandler)

	err := r.Run(":3000")
	if err != nil {
		log.Panic(err)
	}
}

func main() {
	handlerFunctions()
}
