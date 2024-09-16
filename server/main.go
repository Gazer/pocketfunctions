package main

import (
	"io"
	"log"
	"net/http"

	"github.com/Gazer/pocketfunctions/languages"
	"github.com/Gazer/pocketfunctions/models"
	"github.com/gin-gonic/gin"
)

var functions = map[string]*models.PocketFunction{}

func main() {
	log.Print("Let's go\n")

	router := gin.Default()
	router.POST("/_/create/*path", func(c *gin.Context) {
		// nil
		newFunction := models.PocketFunctionFromRequest(c)
		if newFunction == nil {
			log.Println("Bad request")
			c.String(http.StatusBadRequest, "Invalid Data")
			return
		}

		log.Printf("New function registered %s id=%s\n", newFunction.Uri, newFunction.Id)
		functions[newFunction.Uri] = newFunction

		log.Print("Deploying ... \n")
		languages.DeployDart(newFunction)

		c.String(http.StatusOK, "Ok")
	})

	router.NoRoute(func(c *gin.Context) {
		var path = c.Request.URL.Path

		if function, exists := functions[path]; exists {
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
	})
	router.Run("localhost:8080")
}
