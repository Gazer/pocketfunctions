package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Gazer/pocketfunctions/languages"
	"github.com/Gazer/pocketfunctions/models"
	"github.com/gin-gonic/gin"
)

func (api *PocketAPI) Execute() gin.HandlerFunc {
	return func(c *gin.Context) {
		var path = c.Request.URL.Path

		fmt.Printf("Requested Uri: %s\n", path)

		function, err := models.GetFunctionByUri(api.Db, path)
		if err != nil {
			fmt.Printf("Error: %s\n", err.Error())
			c.String(http.StatusNotFound, err.Error())
			models.RegisterExecuted(api.Db, 0, 0, http.StatusNotFound)
			return
		}

		filePath := createDataFile(c)
		defer os.Remove(filePath)

		startTime := time.Now()
		var response, headers, error = languages.RunDart(function, filePath)
		elapsed := time.Since(startTime)

		if error != nil {
			c.String(http.StatusInternalServerError, response)
			models.RegisterExecuted(api.Db, function.Id, elapsed, http.StatusInternalServerError)
			return
		}

		models.RegisterExecuted(api.Db, function.Id, elapsed, http.StatusOK)

		for key, value := range headers {
			c.Header(key, value)
		}
		c.String(http.StatusOK, response)
	}
}

func createDataFile(c *gin.Context) string {
	f, err := os.CreateTemp("/tmp", "tmpfile-")
	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	// write data to the temporary file
	f.Write([]byte(c.Request.URL.Path))
	f.Write([]byte("\n"))
	f.Write([]byte(c.Request.URL.RawQuery))
	f.Write([]byte("\n"))
	f.Write([]byte(c.Request.Method))
	f.Write([]byte("\n"))
	f.Write([]byte(c.GetHeader("Content-Type")))
	f.Write([]byte("\n"))
	body, err := io.ReadAll(c.Request.Body)
	if err == nil {
		f.Write([]byte(body))
	}
	return f.Name()
}
