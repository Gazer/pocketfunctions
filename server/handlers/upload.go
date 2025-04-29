package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"

	"github.com/Gazer/pocketfunctions/languages"
	"github.com/Gazer/pocketfunctions/models"
	"github.com/gin-gonic/gin"
)

func Upload(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := findFunction(db, c); err != nil {
			log.Println("Bad request")
			log.Println(err.Error())
			c.String(http.StatusBadRequest, err.Error())
		} else {
			c.String(http.StatusOK, "Ok")
		}
	}
}

func findFunction(db *sql.DB, c *gin.Context) error {
	function, err := models.GetFunctionByID(db, c.Param("id"))
	if err != nil {
		return fmt.Errorf("ID not found %s", c.Param("id"))
	}

	return loadFile(db, c, function)
}

func loadFile(db *sql.DB, c *gin.Context, function *models.PocketFunction) error {
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		return err
	}
	return copyFile(db, function, file)
}

func copyFile(db *sql.DB, function *models.PocketFunction, file multipart.File) error {
    lang, err := languages.Create(function)

    if err != nil {
        return err
    }

    lang.CopyFile(file)

    dockerId, err := saveAndDeploy(lang)
    function.DockerId = dockerId

	if err := models.UpdateFunction(db, function); err != nil {
		return err
	}
    return nil
}

func saveAndDeploy(lang languages.Language) (string, error) {
	log.Print("Deploying ... \n")
	dockerId, err := lang.Deploy()
	if err != nil {
		log.Println("Deploy failed")
		return "", err
	}

	log.Printf("Container started at %s\n", dockerId)

	return dockerId, nil
}
