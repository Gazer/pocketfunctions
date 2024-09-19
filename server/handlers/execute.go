package handlers

import (
	"database/sql"
	"fmt"
	"io"
	"net/http"

	"github.com/Gazer/pocketfunctions/languages"
	"github.com/Gazer/pocketfunctions/models"
	"github.com/gin-gonic/gin"
)

func Execute(db *sql.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		var path = c.Request.URL.Path

		fmt.Printf("Requested Uri: %s\n", path)

		function, err := models.GetFunctionByUri(db, path)
		if err != nil {
			fmt.Printf("Error: %s\n", err.Error())
			c.String(http.StatusNotFound, err.Error())
			return
		}

		env := make(map[string]string)
		env["pf_path"] = path
		env["pf_query"] = c.Request.URL.RawQuery
		env["pf_method"] = c.Request.Method
		body, err := io.ReadAll(c.Request.Body)
		if err == nil {
			env["pf_body"] = string(body)
		}
		env["pf_content_type"] = c.GetHeader("Content-Type")

		var response, headers, error = languages.RunDart(function, env)
		if error != nil {
			c.String(http.StatusInternalServerError, response)
			return
		}

		for key, value := range headers {
			c.Header(key, value)
		}
		c.String(http.StatusOK, response)
	}
}
