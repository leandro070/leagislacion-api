package user

import (
	"legislacion/db"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type Errors struct {
	Errors []string `json:"errors"`
}

// User export interface
type User struct {
	ID           int8   `db:"id, omitempty" json:"id"`
	UserName     string `db:"username" json:"username"`
	FullName     string `db:"fullname, omitempty" json:"fullname,omitempty"`
	Email        string `db:"email" json:"email"`
	PasswordHash string `db:"passwordhash" json:"-"`
	PasswordSalt string `db:"-" json:"-"`
	IsDisabled   bool   `db:"isdisabled" json:"-"`
}

// Session export interface
type Session struct {
	SessionKEY   string    `db:"sessionkey"`
	UserID       int8      `db:"userid"`
	LoginTime    time.Time `db:"logintime"`
	LastSeenTime time.Time `db:"lastseentime"`
}

// CreateUserHandler handles the user creation
func CreateUserHandler(c *gin.Context) {
	var user User
	user.UserName = c.PostForm("username")
	user.PasswordSalt = c.PostForm("password")
	user.FullName = c.PostForm("fullname")
	user.Email = c.PostForm("email")

	errors := []string{}
	if len(user.UserName) == 0 {
		errors = append(errors, "Username required")
	}
	if len(user.PasswordSalt) == 0 {
		errors = append(errors, "Password required")
	}
	if len(user.Email) == 0 {
		errors = append(errors, "Email required")
	}
	if len(errors) > 0 {
		jsonErrors := Errors{Errors: errors}
		c.JSON(http.StatusBadRequest, jsonErrors)
		return
	}

	err := createUser(&user)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
	} else {
		c.JSON(http.StatusOK, user)
	}
}

// LoginHandler handle the users login
func LoginHandler(c *gin.Context) {
	var user User
	user.UserName = c.PostForm("username")
	user.PasswordSalt = c.PostForm("password")

	errors := []string{}
	if len(user.UserName) == 0 {
		errors = append(errors, "Username required")
	}
	if len(user.PasswordSalt) == 0 {
		errors = append(errors, "Password required")
	}
	if len(errors) > 0 {
		jsonErrors := Errors{Errors: errors}
		c.JSON(http.StatusBadRequest, jsonErrors)
		return
	}
	pq := db.GetDB()
	query := "SELECT * FROM users WHERE username = $1"
	row := pq.Db.QueryRow(query, user.UserName)

	err := row.Scan(&user.ID, &user.UserName, &user.FullName, &user.PasswordHash, &user.IsDisabled, &user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	isValid := verifyPassword(user.PasswordSalt, user.PasswordHash)

	if isValid == false {
		c.JSON(http.StatusUnprocessableEntity, err)
		return
	}
	c.JSON(http.StatusOK, user)
	return
}

func createUser(user *User) (err error) {
	user.PasswordHash, err = hashPassword(user.PasswordSalt)
	if err != nil {
		log.Fatal("ERROR HASHING:", err)
		return err
	}
	pq := db.GetDB()
	query := "INSERT INTO users (id, username, fullname, email, passwordhash, isdisabled) VALUES (nextval('users_seq'),$1, $2, $3, $4, false) RETURNING id;"
	row := pq.Db.QueryRow(query, user.UserName, user.FullName, user.Email, user.PasswordHash)

	row.Scan(&user.ID)
	return nil
}

func verifyPassword(password, originalHashed string) (isEqual bool) {
	passBytes := []byte(password)
	hashBytes := []byte(originalHashed)

	err := bcrypt.CompareHashAndPassword(hashBytes, passBytes)
	if err != nil {
		return false
	}
	return true
}

func hashPassword(password string) (passwordHashed string, err error) {
	passBytes := []byte(password)
	passBytes, err = bcrypt.GenerateFromPassword(passBytes, 10)
	if err != nil {
		return "", err
	}
	passwordHashed = string(passBytes)
	return passwordHashed, nil
}
