package main

import (
	"legislacion/files"
	"legislacion/user"
	"legislacion/utils"
	"log"
	"net/http"
	"os"

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

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		errors := utils.Errors{}
		if len(token) == 0 {
			errors.Errors = append(errors.Errors, "Token required")
		}
		userValid := user.ValidateToken(token)
		if userValid == false {
			errors.Errors = append(errors.Errors, "Token invalid")
		}
		if len(errors.Errors) > 0 {
			c.JSON(http.StatusUnauthorized, errors)
			c.Abort()
			return
		}
	}
}

func handlerFunctions() {
	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("$PORT must be set")
	}

	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(CORS())

	r.GET("/home", homePage)
	r.POST("/newUser", user.CreateUserHandler)
	r.POST("/login", user.LoginHandler)
	authorized := r.Group("/")

	authorized.Use(AuthRequired())
	{
		authorized.POST("/files", files.SendFileHandler)
		authorized.PUT("/files/:id", files.UpdateFileHandler)
		authorized.DELETE("/files/:id", files.DeleteFileHandler)
		authorized.GET("/userByToken", user.UserByToken)

	}
	r.GET("/files/:id", files.FindFileByIDHandler)
	r.GET("/files", files.ListFilesHandler)
	r.GET("/download/:id", files.DownloadFileHandler)

	err := r.Run(":" + port)
	if err != nil {
		log.Panic(err)
	}
}

func main() {
	handlerFunctions()
}
