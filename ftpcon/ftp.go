package ftpcon

import (
	"fmt"
	"log"
	"time"

	"github.com/jlaffaye/ftp"
)

// FTPManager estrablece el modelo singleton para una conexion FTP
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
	return FTPManager{c, true}
}

// MoveLegislacionDirectory intenta moverse al path de Legislacion
func MoveLegislacionDirectory() (err error) {
	c := ftpInstance.Ftp
	err = c.ChangeDirToParent()
	if err != nil {
		log.Fatal("Error al cambiar al directorio raiz")
		return err
	}
	err = c.ChangeDir(path)
	if err != nil {
		log.Fatal("Error al cambiar al directorio Legislacion, tal vez no existe")
		err = CreateLegislacionDirectory()
		if err != nil {
			return err
		}
		err = c.ChangeDir(path)
		if err != nil {
			return err
		}
	}

	return nil
}

// CreateLegislacionDirectory crea los directorios necesarios para guardar los archivos
func CreateLegislacionDirectory() (err error) {
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
