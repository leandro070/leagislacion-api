package user

import (
	"crypto/rand"
	"fmt"
	"legislacion/db"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// User export interface
type User struct {
	ID           int8   `db:"id, omitempty" json:"id"`
	UserName     string `db:"username" json:"username"`
	FullName     string `db:"fullname, omitempty" json:"fullname,omitempty"`
	Email        string `db:"email" json:"email"`
	PasswordHash string `db:"passwordhash" json:"-"`
	PasswordSalt string `db:"-" json:"-"`
	IsDisabled   bool   `db:"isdisabled" json:"is_disabled"`
	Token        string `db:"token" json:"token"`
	CreatedAt    string `db:"created_at" json:"-"`
	UpdatedAt    string `db:"updated_at" json:"-"`
}

// CreateUserHandler handles the user creation
func CreateUserHandler(c *gin.Context) {
	var user User
	user.UserName = c.PostForm("username")
	user.PasswordSalt = c.PostForm("password")
	user.FullName = c.PostForm("fullname")
	user.Email = c.PostForm("email")

	if len(user.UserName) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Usuario requerido"})
		return
	}
	if len(user.PasswordSalt) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Contraseña requerida"})
		return
	}
	if len(user.Email) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email requerido"})
		return
	}

	err := createUser(&user)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
	}
	user.PasswordSalt = ""
	c.JSON(http.StatusOK, user)
}

// LoginHandler handle the users login
func LoginHandler(c *gin.Context) {
	log.Print("LoginHandler")
	var userLogin struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	err := c.BindJSON(&userLogin)
	if err != nil {
		c.Error(err)
	}
	if len(userLogin.Username) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Usuario requerido"})
		return
	}
	if len(userLogin.Password) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Contraseña requerida"})
		return
	}

	var user User

	pq := db.GetDB()
	query := "SELECT * FROM users WHERE username = $1"
	row := pq.Db.QueryRow(query, userLogin.Username)

	err = row.Scan(&user.ID, &user.UserName, &user.FullName, &user.PasswordHash, &user.IsDisabled, &user.Email, &user.Token, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Usuario o contraseña incorrecto"})
		return
	}

	isValid := verifyPassword(userLogin.Password, user.PasswordHash)

	if isValid == false {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Usuario o contraseña incorrecto"})
		return
	}

	err = updateToken(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
	return
}

func updateToken(user *User) error {
	user.Token = tokenGenerator()
	query := `UPDATE users SET token = $1 WHERE id = $2;`
	pq := db.GetDB()
	_, err := pq.Db.Exec(query, user.Token, user.ID)
	if err != nil {
		return fmt.Errorf("El token no se pudo actualizar")
	}
	return nil
}

func createUser(user *User) (err error) {
	user.PasswordHash, err = hashPassword(user.PasswordSalt)
	if err != nil {
		log.Print("ERROR HASHING:", err)
		return err
	}
	user.Token = tokenGenerator()
	pq := db.GetDB()
	query := "INSERT INTO users (id, username, fullname, email, passwordhash, isdisabled, token) VALUES (nextval('users_seq'),$1, $2, $3, $4, false, $5) RETURNING id;"
	row := pq.Db.QueryRow(query, user.UserName, user.FullName, user.Email, user.PasswordHash, user.Token)

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

func tokenGenerator() string {
	b := make([]byte, 64)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

func UserByToken(c *gin.Context) {
	var u User
	token := c.GetHeader("Authorization")

	token, err := extractToken(token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	pq := db.GetDB()
	query := "SELECT * FROM users WHERE token = $1"
	row := pq.Db.QueryRow(query, token)

	err = row.Scan(&u.ID, &u.UserName, &u.FullName, &u.PasswordHash, &u.IsDisabled, &u.Email, &u.Token)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Token invalido"})
		return
	}
	log.Print(u)
	c.JSON(http.StatusOK, u)
}

func extractToken(bearer string) (string, error) {
	splitToken := strings.Split(bearer, "Beaver")
	if len(splitToken) != 2 {
		return "", fmt.Errorf("Formato de token invalido")
	}
	token := strings.TrimSpace(splitToken[1])
	return token, nil
}

func ValidateToken(token string) bool {
	token, err := extractToken(token)
	if err != nil {
		return false
	}
	pq := db.GetDB()
	query := "SELECT token FROM users WHERE token = $1"
	rows, err := pq.Db.Query(query, token)
	if err != nil {
		return false
	}
	return rows.Next()
}
