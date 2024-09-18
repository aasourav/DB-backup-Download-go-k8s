package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

func main() {
	// Create a new Gin router
	router := gin.Default()

	// Define a GET route for testing
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "API is working well")
	})

	// Define a GET route for downloading files
	router.GET("/download/:filename", func(c *gin.Context) {
		filename := c.Param("filename")

		// Sanitize filename to prevent directory traversal attacks
		if strings.Contains(filename, "..") {
			c.String(http.StatusBadRequest, "invalid filename")
			return
		}

		// Define the file path (in /tmp directory)
		filePath := filepath.Join("/tmp", filename)

		fmt.Println("Attempting to serve file:", filePath)

		// Check if the file exists
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			c.String(http.StatusNotFound, "file not found")
			return
		} else if err != nil {
			c.String(http.StatusInternalServerError, "internal server error")
			return
		}

		// Set Content-Disposition to force download
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
		c.File(filePath)
	})

	// Run the Gin server (blocking call)
	fmt.Println("Server is running on port 8035")
	if err := router.Run(":8035"); err != nil {
		fmt.Println("Failed to start server:", err)
	}
}
