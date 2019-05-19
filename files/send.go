package files

import (
	"legislacion/ftpcon"
	"net/http"

	"github.com/gin-gonic/gin"
)

// SendFileHandler handles file saving
func SendFileHandler(c *gin.Context) {
	header, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	err = ftpcon.MoveLegislacionDirectory(false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	err = ftpcon.SaveFile(header)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	return
}

// ListFilesHandler return list of files in FTP
func ListFilesHandler(c *gin.Context) {
	header, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	err = ftpcon.MoveLegislacionDirectory(false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	err = ftpcon.SaveFile(header)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	return
}
