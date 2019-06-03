package files

import (
	"bytes"
	"fmt"
	"legislacion/db"
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

	label := c.PostForm("label")
	if len(label) == 0 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Nombre de archivo requerido"})
	}
	header, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Archivo no adjunto. Por favor adjunte un archivo"})
		return
	}

	header.Filename = changeFileName(header.Filename)

	f, err := header.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ocurrio un error abriendo el archivo"})
		log.Printf(`There was an error opening the file: %s`, err.Error())
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ocurrió un error al guardar un archivo"})
		log.Printf(`There was an error saving the file: %s`, err.Error())
		return
	}

	c.JSON(http.StatusOK, file)
	return
}

// ListFilesHandler get files info
func ListFilesHandler(c *gin.Context) {

	pq := db.GetDB()
	query := "SELECT id, file_name, file_label FROM files ORDER BY created_at DESC;"
	rows, err := pq.Db.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ocurrio un error al buscar los archivos"})
		log.Print("There was an error saving the file: %s", err.Error())
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ocurrio un error, los archivos no se pudieron leer correctamente"})
		log.Println("The files were not read correctly")
		return
	}
	c.JSON(http.StatusOK, files)
	return
}

// FindFileByID search a file
func FindFileByIDHandler(c *gin.Context) {
	id := c.Param("id")
	pq := db.GetDB()

	query := "SELECT id, mime_type, file_name, file_label FROM files WHERE id=$1"
	row := pq.Db.QueryRow(query, id)

	file := File{0, "MimeType", "FileName", "FileLabel", make([]byte, 0)}
	err := row.Scan(&file.ID, &file.MimeType, &file.Filename, &file.FileLabel)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ocurrio un error en el servidor"})
		log.Println(err.Error())
		return
	}
	c.JSON(http.StatusOK, file)
	return
}

func DownloadFileHandler(c *gin.Context) {
	id := c.Param("id")
	pq := db.GetDB()

	query := "SELECT file_data, file_name, mime_type FROM files WHERE id=$1"
	row := pq.Db.QueryRow(query, id)

	filedata := make([]byte, 0)
	filename := ""
	contentType := ""
	err := row.Scan(&filedata, &filename, &contentType)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Archivo no encontrado"})
		log.Printf("File not found: %s", err.Error())
		return
	}
	reader := bytes.NewReader(filedata)
	contentLength := int64(len(filedata))
	extraHeaders := map[string]string{
		"Content-Disposition": fmt.Sprintf(`attachment; filename="%s"`, filename),
	}

	c.DataFromReader(http.StatusOK, contentLength, contentType, reader, extraHeaders)
}

func UpdateFileHandler(c *gin.Context) {
	log.Print("UpdateFileHandler")

	idStr := c.Param("id")
	if len(idStr) == 0 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Archivo imposible de identificar. Envie el ID del archivo a modificar"})
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "El ID enviado no tiene formato válido. Asegurese de que sea un número"})
		return
	}
	pq := db.GetDB()

	var file File
	query := "SELECT id, file_label, file_name FROM files WHERE id=$1 LIMIT 1;"
	row := pq.Db.QueryRow(query, id)
	err = row.Scan(&file.ID, &file.FileLabel, &file.Filename)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Archivo no encontrado"})
		log.Printf("File not found: %s", err.Error())
		return
	}
	var structLabel struct {
		FileLabel string `db:"file_label" json:"label"`
	}
	err = c.BindJSON(&structLabel)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if len(structLabel.FileLabel) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nombre de archivo requerido"})
		return
	}
	query = "UPDATE files SET file_label = $1 WHERE id = $2"
	res, err := pq.Db.Exec(query, structLabel.FileLabel, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "El archivo no pudo ser modificado"})
		log.Printf("The file could not be modified: %s", err.Error())
		return
	}
	numDeleted, err := res.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ocurrio un error inesperado"})
		log.Printf("Ocurrio un error inesperado: %s", err.Error())
		return
	}

	if numDeleted == 0 {
		c.JSON(http.StatusNotModified, gin.H{"error": "Archivo no modificado"})
		return
	}
	file.FileLabel = structLabel.FileLabel
	c.JSON(http.StatusOK, file)
	return
}

func DeleteFileHandler(c *gin.Context) {

	idStr := c.Param("id")
	if len(idStr) == 0 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Archivo imposible de identificar. Envie el ID del archivo a borrar"})
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "El ID enviado no tiene formato válido. Asegurese de que sea un número"})
		return
	}

	pq := db.GetDB()
	query := "DELETE FROM files WHERE id=$1;"
	res, err := pq.Db.Exec(query, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Archivo no encontrado"})
		log.Printf("File not found: %s", err.Error())
		return
	}

	numberRowsDeleted, err := res.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Archivo no eliminado"})
		log.Printf("Error: %s", err.Error())
		return
	}
	if numberRowsDeleted == 0 {
		c.JSON(http.StatusNotModified, gin.H{"error": "Archivo no eliminado"})
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
