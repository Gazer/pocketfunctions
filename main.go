package main

import (
	"fmt"
	"io"
	"net/http"

	l "gin/languages"
	"gin/models"

	"github.com/gin-gonic/gin"
)

var functions = map[string]*models.PocketFunction{}

func main() {
	fmt.Print("Let's go\n")

	router := gin.Default()
	router.POST("/_/create", func(c *gin.Context) {
		// nil
		newFunction := models.PocketFunctionFromRequest(c)
		if newFunction == nil {
			fmt.Println("Bad request")
			c.String(http.StatusBadRequest, "Invalid JSON")
			return
		}

		fmt.Printf("New function registered %s id=%s\n", newFunction.Uri, newFunction.Id)
		functions[newFunction.Uri] = newFunction

		fmt.Print("Deploying ... \n")
		l.DeployDart(newFunction)

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
			env["pf_content-type"] = c.GetHeader("Content-Type")

			var response, headers, error = l.RunDart(function, env)
			if error != nil {
				fmt.Println(response)
				fmt.Println(error)
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Execution failed",
				})
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
