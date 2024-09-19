package handlers

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/Gazer/pocketfunctions/languages"
	"github.com/Gazer/pocketfunctions/models"
	"github.com/gin-gonic/gin"
)

func Upload(db *sql.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		function, err := models.GetFunctionByID(db, c.Param("id"))
		if err != nil {
			log.Println("Bad request")
			c.String(http.StatusBadRequest, "Invalid Data")
			return
		}

		file, fileHeader, err := c.Request.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		function.Code = strings.TrimRight(fileHeader.Filename, ".zip")

		dst, err := os.Create(fmt.Sprintf("../dist/function_repository/%s.zip", function.Code))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer dst.Close()

		_, err = io.Copy(dst, file)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		models.UpdateFunction(db, function)

		log.Print("Deploying ... \n")
		languages.DeployDart(function)

		c.String(http.StatusOK, "Ok")
	}
}
