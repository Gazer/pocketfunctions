package main

import (
	"fmt"
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
		c.String(http.StatusOK, "Ok")
	})
	router.NoRoute(func(c *gin.Context) {
		var path = c.Request.URL.Path
		fmt.Println(path)
		if function, exists := functions[path]; exists {
			var response, headers, error = l.RunDart(function)
			if error != nil {
				fmt.Println("Execution failed")
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Execution failed",
				})
				return
			}

			fmt.Println(headers)
			fmt.Println(response)

			for key, value := range headers {
				c.Header(key, value)
			}
			c.String(http.StatusOK, response)

			// if value, exists := headers["content-type"]; exists && value == "application/json" {
			// 	var jsonData map[string]interface{}
			// 	err := json.Unmarshal([]byte(response), &jsonData)
			// 	if err != nil {
			// 		c.JSON(http.StatusInternalServerError, gin.H{
			// 			"error": "Invalid response JSON",
			// 		})
			// 		return
			// 	}
			// 	c.String(http.StatusOK, response)
			// } else {
			// 	c.String(http.StatusOK, response)
			// }
		}
	})
	router.Run("localhost:8080")
}
