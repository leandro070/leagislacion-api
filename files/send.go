package files

import (
	"legislacion/db"
	"legislacion/user"
	"legislacion/utils"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// File define interface of File between DB and server
type File struct {
	ID       int8           `db:"id" json:"-"`
	MimeType string         `db:"mime_type" json:"mime_type,omitempty"`
	Filename string         `db:"file_name" json:"filename"`
	Filedata []byte         `db:"file_data" json:"-"`
	FileObj  multipart.File `db:"-" json:"-"`
}

// SendFileHandler save file in Postgres
func SendFileHandler(c *gin.Context) {
	token := c.GetHeader("token")
	errors := utils.Errors{}
	if len(token) == 0 {
		errors.Errors = append(errors.Errors, "Token required")
		c.JSON(http.StatusBadRequest, errors)
	}
	userValid := user.ValidateToken(token)
	if userValid == false {
		errors.Errors = append(errors.Errors, "Token invalid")
		c.JSON(http.StatusBadRequest, errors)
	}
	header, err := c.FormFile("file")
	if err != nil {
		errors.Errors = append(errors.Errors, err.Error())
		c.JSON(http.StatusBadRequest, errors)
		return
	}
	header.Filename = changeFileName(header.Filename)

	f, err := header.Open()
	if err != nil {
		errors.Errors = append(errors.Errors, `Ocurrió un error al abrir el archivo: %s`, err.Error())
		c.JSON(http.StatusInternalServerError, errors)
		return
	}
	defer f.Close()

	bytes := make([]byte, header.Size)
	f.Read(bytes)
	file := File{Filename: header.Filename, FileObj: f, Filedata: bytes, MimeType: header.Header.Get("Content-Type")}

	pq := db.GetDB()
	query := "INSERT INTO files (mime_type, file_name, file_data) VALUES ($1, $2, $3) RETURNING id"
	row := pq.Db.QueryRow(query, file.MimeType, file.Filename, file.Filedata)
	err = row.Scan(&file.ID)
	if err != nil {
		errors.Errors = append(errors.Errors, `Ocurrió un error al guardar el archivo: %s`, err.Error())
		c.JSON(http.StatusInternalServerError, errors)
		return
	}
	c.JSON(http.StatusOK, file)
	return
}

func changeFileName(filename string) string {
	filename = strings.Replace(filename, " ", "_", -1)
	filename = strings.ToLower(filename)
	t := time.Now()
	filename = t.Format("20060102150405") + "_" + filename
	return filename
}
