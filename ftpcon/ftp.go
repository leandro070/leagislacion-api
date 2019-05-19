package ftpcon

import (
	"bufio"
	"fmt"
	"log"
	"mime/multipart"
	"strings"
	"time"

	"github.com/jlaffaye/ftp"
)

// FTPManager design the singleton model for an FTP connection
type FTPManager struct {
	FtpPublic  *ftp.ServerConn
	ftpPrivate *ftp.ServerConn
}

var ftpInstance = new()

// GetInstanceFTP return FTP instance
func GetInstanceFTP(isPublic bool) *ftp.ServerConn {
	c := ftpInstance.ftpPrivate
	if isPublic == true {
		c = ftpInstance.FtpPublic
	}
	return c
}

func new() FTPManager {
	//init host private
	cpriv, err := ftp.Dial(fmt.Sprintf(`%s:%d`, hostprivate, port), ftp.DialWithTimeout(5*time.Second))
	if err != nil {
		log.Fatal(err)
	}

	err = cpriv.Login(user, password)
	if err != nil {
		log.Fatal(err)
	}
	log.Print("FTP Private Connection established")

	cpub, err := ftp.Dial(fmt.Sprintf(`%s:%d`, hostpublic, port), ftp.DialWithTimeout(5*time.Second))
	if err != nil {
		log.Fatal(err)
	}

	err = cpub.Login(user, password)
	if err != nil {
		log.Fatal(err)
	}
	log.Print("FTP Public Connection established")
	return FTPManager{cpub, cpriv}
}

// MoveLegislacionDirectory try move dir to path /Legislacion
func MoveLegislacionDirectory(isPublic bool) (err error) {
	c := GetInstanceFTP(isPublic)
	err = c.ChangeDirToParent()
	if err != nil {
		log.Print("Error al cambiar al directorio raiz")
		return err
	}
	err = c.ChangeDir(path)
	if err != nil {
		log.Print("Error al cambiar al directorio Legislacion, tal vez no existe")
		err = createLegislacionDirectory(isPublic)
		if err != nil {
			log.Print("Error al cambiar al crear directorio Legislacion")
			return err
		}
		err = c.ChangeDir(path)
		if err != nil {
			log.Print("Error al cambiar al crear directorio Legislacion creado")
			return err
		}
	}

	return nil
}

// CreateLegislacionDirectory create the necessary directories to save the files
func createLegislacionDirectory(isPublic bool) (err error) {
	c := GetInstanceFTP(isPublic)
	err = c.ChangeDirToParent()
	if err != nil {
		return err
	}
	err = c.MakeDir(path)
	if err != nil {
		return err
	}
	return nil
}

func GetListFiles(isPublic bool) {
	return
}

// SaveFile save a file on FTP Storage
func SaveFile(header *multipart.FileHeader) (err error) {
	c := ftpInstance.ftpPrivate
	header.Filename = changeFileName(header.Filename)
	file, err := header.Open()
	if err != nil {
		log.Printf(`Ocurrió un error al abrir el archivo: %s`, err.Error())
		return err
	}
	bytes := make([]byte, header.Size)
	buffer := bufio.NewReader(file)
	_, err = buffer.Read(bytes)
	if err != nil {
		log.Printf(`Ocurrió un error al leer el archivo: %s`, err.Error())
		return err
	}
	err = c.Stor(path, buffer)
	if err != nil {
		log.Printf(`Ocurrió un error al guardar el archivo: %s`, err.Error())
		return err
	}
	return nil
}

func changeFileName(filename string) string {
	filename = strings.Replace(filename, " ", "_", -1)
	filename = strings.ToLower(filename)
	t := time.Now()
	filename = filename + t.Format("20060102150405")
	return filename
}
