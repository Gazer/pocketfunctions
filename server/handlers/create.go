package handlers

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/Gazer/pocketfunctions/models"
	"github.com/gin-gonic/gin"
)

type createRequest struct {
	Name string `json:"name"`
	Lang string `json:"lang"`
}

func Create(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request createRequest
		if err := c.BindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid JSON",
			})
			return
		}

		log.Printf("Creating function in %s", request.Name)
		id, err := models.CreateFunction(db, request.Name, request.Lang)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Unable to create function",
			})
			return
		}

		log.Printf("Function %s with id %d\n", request.Name, id)
		c.JSON(http.StatusOK, gin.H{
			"id": id,
		})
	}
}
