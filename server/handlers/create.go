package handlers

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/Gazer/pocketfunctions/models"
	"github.com/gin-gonic/gin"
)

type CreateRequest struct {
	Uri string `json:"uri"`
}

func Create(db *sql.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		var request CreateRequest
		if err := c.BindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalida JSON",
			})
			return
		}

		log.Printf("Creating function in %s", request.Uri)
		// TODO: Function may exists
		id, err := models.CreateFunction(db, request.Uri)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Unable to create function",
			})
			return
		}

		log.Printf("New function registered at %s\n", request.Uri)
		c.JSON(http.StatusOK, gin.H{
			"id": id,
		})
	}
}
