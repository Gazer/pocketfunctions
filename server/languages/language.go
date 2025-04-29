package languages

import (
	"fmt"
	"mime/multipart"

	"github.com/Gazer/pocketfunctions/models"
)

type Language interface {
    CopyFile(file multipart.File) (error)
	Start() (string, error)
	Deploy() (string, error)
}

func Create(f *models.PocketFunction) (Language, error) {
    if f.Lang == "dart" {
        return newDart(f), nil
    }
    return nil, fmt.Errorf("Language not supported")
}
