package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Gazer/pocketfunctions/handlers"
)

func main() {
	log.Print("Let's go\n")

	api := handlers.New(os.Args[1])
	api.InitAdminUI()

	fmt.Printf("├─ REST API: http://127.0.0.1:%d/api/\n", 8080)
	fmt.Printf("└─ Admin UI: http://127.0.0.1:%d/_/\n", 8080)

	api.Router.Run("localhost:8080")
}
