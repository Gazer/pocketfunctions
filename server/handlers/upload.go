package handlers

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strings"

	"github.com/Gazer/pocketfunctions/languages"
	"github.com/Gazer/pocketfunctions/models"
	"github.com/gin-gonic/gin"
)

func Upload(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := findFunction(db, c); err != nil {
			log.Println("Bad request")
			c.String(http.StatusBadRequest, "Invalid Data")
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
	file, fileHeader, err := c.Request.FormFile("file")
	if err != nil {
		return err
	}
	function.Code = strings.TrimRight(fileHeader.Filename, ".zip")
	return copyFile(db, function, file)
}

func copyFile(db *sql.DB, function *models.PocketFunction, file multipart.File) error {
	dst, err := os.Create(fmt.Sprintf("../dist/function_repository/%s.zip", function.Code))
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		return err
	}

	return saveAndDeploy(db, function)
}

func saveAndDeploy(db *sql.DB, function *models.PocketFunction) error {
	models.UpdateFunction(db, function)

	log.Print("Deploying ... \n")
	languages.DeployDart(function)
	return nil
}
