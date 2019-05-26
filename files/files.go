package files

import (
	"bytes"
	"fmt"
	"legislacion/db"
	"legislacion/utils"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// File define interface of File between DB and server
type File struct {
	ID        int8   `db:"id" json:"id"`
	MimeType  string `db:"mime_type" json:"mime_type,omitempty"`
	Filename  string `db:"file_name" json:"filename"`
	FileLabel string `db:"file_label" json:"label"`
	Filedata  []byte `db:"file_data" json:"filedata,omitempty"`
}

// SendFileHandler save file in Postgres
func SendFileHandler(c *gin.Context) {
	log.Print("SendFileHandler")

	errors := utils.Errors{}
	label := c.PostForm("label")
	if len(label) == 0 {
		errors.Errors = append(errors.Errors, "Label required")
	}
	header, err := c.FormFile("file")
	if err != nil {
		errors.Errors = append(errors.Errors, err.Error())
		return
	}
	if len(errors.Errors) > 0 {
		c.JSON(http.StatusBadRequest, errors)
		return
	}

	header.Filename = changeFileName(header.Filename)

	f, err := header.Open()
	if err != nil {
		errors.Errors = append(errors.Errors, `There was an error opening the file: %s`, err.Error())
		c.JSON(http.StatusInternalServerError, errors)
		return
	}
	defer f.Close()

	bytes := make([]byte, header.Size)
	f.Read(bytes)
	file := File{Filename: header.Filename, Filedata: bytes, MimeType: header.Header.Get("Content-Type"), FileLabel: label}

	pq := db.GetDB()
	query := "INSERT INTO files (mime_type, file_name, file_data, file_label) VALUES ($1, $2, $3, $4) RETURNING id"
	row := pq.Db.QueryRow(query, file.MimeType, file.Filename, file.Filedata, file.FileLabel)
	err = row.Scan(&file.ID)
	if err != nil {
		errors.Errors = append(errors.Errors, `There was an error saving the file: %s`, err.Error())
		c.JSON(http.StatusInternalServerError, errors)
		return
	}

	c.JSON(http.StatusOK, file)
	return
}

// ListFilesHandler get files info
func ListFilesHandler(c *gin.Context) {
	log.Print("ListFilesHandler")
	errors := utils.Errors{}

	pq := db.GetDB()
	query := "SELECT id, file_name, file_label FROM files"
	rows, err := pq.Db.Query(query)
	if err != nil {
		errors.Errors = append(errors.Errors, `There was an error saving the file: %s`, err.Error())
		c.JSON(http.StatusInternalServerError, errors)
		return
	}
	files := make([]File, 0)
	for rows.Next() {
		var file File
		if err := rows.Scan(&file.ID, &file.Filename, &file.FileLabel); err != nil {
			log.Print("ERROR SCAN FILES", err)
		}
		files = append(files, file)
	}
	if err != nil {
		errors.Errors = append(errors.Errors, "The files were not read correctly")
		c.JSON(http.StatusInternalServerError, errors)
	}
	c.JSON(http.StatusOK, files)
	return
}

// FindFileByID search a file
func FindFileByIDHandler(c *gin.Context) {
	log.Print("FindFileByID")

	id := c.Param("id")
	pq := db.GetDB()

	query := "SELECT id, mime_type, file_name, file_label FROM files WHERE id=$1"
	row := pq.Db.QueryRow(query, id)

	file := File{0, "MimeType", "FileName", "FileLabel", make([]byte, 0)}
	row.Scan(&file.ID, &file.MimeType, &file.Filename, &file.FileLabel)

	c.JSON(http.StatusOK, file)
	return
}

func DownloadFileHandler(c *gin.Context) {
	log.Print("DownloadFile")

	id := c.Param("id")
	pq := db.GetDB()

	query := "SELECT file_data, file_name, mime_type FROM files WHERE id=$1"
	row := pq.Db.QueryRow(query, id)

	filedata := make([]byte, 0)
	filename := ""
	contentType := ""
	row.Scan(&filedata, &filename, &contentType)

	reader := bytes.NewReader(filedata)
	contentLength := int64(len(filedata))
	extraHeaders := map[string]string{
		"Content-Disposition": fmt.Sprintf(`attachment; filename="%s"`, filename),
	}

	c.DataFromReader(http.StatusOK, contentLength, contentType, reader, extraHeaders)
}

func UpdateFileHandler(c *gin.Context) {
	log.Print("UpdateFileHandler")
	errors := utils.Errors{}

	idStr := c.Param("id")
	if len(idStr) == 0 {
		errors.Errors = append(errors.Errors, "ID not defined")
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		errors.Errors = append(errors.Errors, "ID not number")
	}
	if len(errors.Errors) > 0 {
		c.JSON(http.StatusBadRequest, errors)
		return
	}
	pq := db.GetDB()

	query := "SELECT id, file_label, file_name FROM files WHERE id=$1 LIMIT 1;"
	var file File
	row := pq.Db.QueryRow(query, id)
	err = row.Scan(&file.ID, &file.FileLabel, &file.Filename)
	if err != nil {
		log.Print(err)
		errors.Errors = append(errors.Errors, "File not found")
		c.JSON(http.StatusNotFound, errors)
		return
	}

	label := c.PostForm("label")
	if len(label) == 0 {
		errors.Errors = append(errors.Errors, "Nothing to modify")
		c.JSON(http.StatusNotModified, errors)
		return
	}
	query = "UPDATE files SET file_label = $1 WHERE id = $2"
	res, err := pq.Db.Exec(query, label, id)
	if err != nil {
		log.Print(err)
		errors.Errors = append(errors.Errors, "The file could not be modified")
		c.JSON(http.StatusInternalServerError, errors)
		return
	}
	numDeleted, err := res.RowsAffected()
	if err != nil {
		log.Print(err)
	}
	if numDeleted == 0 {
		errors.Errors = append(errors.Errors, "The file was not modified")
		c.JSON(http.StatusNotModified, errors)
		return
	}
	file.FileLabel = label
	c.JSON(http.StatusOK, file)
	return
}

func DeleteFileHandler(c *gin.Context) {
	log.Print("UpdateFileHandler")
	errors := utils.Errors{}

	idStr := c.Param("id")
	if len(idStr) == 0 {
		errors.Errors = append(errors.Errors, "ID not defined")
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		errors.Errors = append(errors.Errors, "ID not number")
	}
	if len(errors.Errors) > 0 {
		c.JSON(http.StatusBadRequest, errors)
		return
	}
	pq := db.GetDB()
	query := "DELETE FROM files WHERE id=$1;"
	res, err := pq.Db.Exec(query, id)
	if err != nil {
		log.Print(err)
		errors.Errors = append(errors.Errors, "File not found")
		c.JSON(http.StatusNotFound, errors)
		return
	}

	numberRowsDeleted, err := res.RowsAffected()
	if err != nil {
		log.Print(err)
	}
	if numberRowsDeleted == 0 {
		errors.Errors = append(errors.Errors, "File not found")
		c.JSON(http.StatusNotFound, errors)
		return
	}

	c.JSON(http.StatusOK, "Deleted")
}

func changeFileName(filename string) string {
	filename = strings.Replace(filename, " ", "_", -1)
	filename = strings.ToLower(filename)
	t := time.Now()
	filename = t.Format("20060102150405") + "_" + filename
	return filename
}
