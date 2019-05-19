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
	Ftp           *ftp.ServerConn
	IsInitialized bool
}

var ftpInstance = new()

// GetFTP return FTP instance
func GetFTP() FTPManager {
	return ftpInstance
}

func new() FTPManager {
	c, err := ftp.Dial(fmt.Sprintf(`%s:%d`, host, port), ftp.DialWithTimeout(5*time.Second))
	if err != nil {
		log.Fatal(err)
	}

	err = c.Login(user, password)
	if err != nil {
		log.Fatal(err)
	}

	log.Print("FTP Connection established")
	current, err := c.CurrentDir()
	if err != nil {
		log.Fatal(err)
	}
	log.Print(current)
	return FTPManager{c, true}
}

// MoveLegislacionDirectory try move dir to path /Legislacion
func MoveLegislacionDirectory() (err error) {
	c := ftpInstance.Ftp
	err = c.ChangeDirToParent()
	if err != nil {
		log.Print("Error al cambiar al directorio raiz")
		return err
	}
	err = c.ChangeDir(path)
	if err != nil {
		log.Print("Error al cambiar al directorio Legislacion, tal vez no existe")
		err = createLegislacionDirectory()
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
func createLegislacionDirectory() (err error) {
	c := ftpInstance.Ftp
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

// SaveFile save a file on FTP Storage
func SaveFile(file multipart.File, header *multipart.FileHeader) (err error) {
	c := ftpInstance.Ftp
	header.Filename = changeFileName(header.Filename)
	bytes := make([]byte, header.Size)
	buffer := bufio.NewReader(file)
	_, err = buffer.Read(bytes)
	if err != nil {
		log.Printf(`Ocurrió un error al leer el archivo: %s`, err.Error())
		return err
	}
	dir, err := c.CurrentDir()
	log.Print(dir)
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
