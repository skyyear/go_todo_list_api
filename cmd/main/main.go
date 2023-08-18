package main

// Import packages
import (
	"github.com/gin-gonic/gin"
)

// Initialize database connection in main function
func main() {
	router := gin.Default()

	router.Run("localhost:8080")
}
