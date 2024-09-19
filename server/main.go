package main

import (
	"log"

	"github.com/Gazer/pocketfunctions/handlers"
	"github.com/Gazer/pocketfunctions/models"
	"github.com/gin-gonic/gin"
)

func main() {
	log.Print("Let's go\n")

	db := models.InitDB()

	router := gin.Default()

	router.POST("/_/create", handlers.Create(db))
	router.POST("/_/upload/:id", handlers.Upload(db))
	router.NoRoute(handlers.Execute(db))

	router.Run("localhost:8080")
}
