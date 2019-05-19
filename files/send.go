package files

import (
	"legislacion/ftpcon"
	"net/http"

	"github.com/gin-gonic/gin"
)

// SendFileHandler handles file saving
func SendFileHandler(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	defer file.Close()
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	err = ftpcon.MoveLegislacionDirectory()
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	err = ftpcon.SaveFile(file, header)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	return
}
