package main

import (
	_ "legislacion/db"
	"legislacion/user"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Article struct {
	Title   string `json:"title"`
	Desc    string `json:"description,omitempty"`
	Content string `json:"content"`
}

type Articles []Article

func allArticles(c *gin.Context) {
	articles := Articles{
		Article{"Albertito", "Se fue a la guerra", "que dolor, que pena"},
		Article{Title: "Albertito", Content: "Se fue a la guerra"},
	}
	c.JSON(http.StatusOK, gin.H{
		"data":    articles,
		"message": nil,
	})
}

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
	r.GET("/articles", allArticles)
	r.POST("/newUser", user.CreateUserHandler)
	r.POST("/login", user.LoginHandler)
	err := r.Run(":3000")
	if err != nil {
		log.Panic(err)
	}
}

func main() {
	handlerFunctions()
}
