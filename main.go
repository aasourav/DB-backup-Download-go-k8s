package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
)

func main() {
	router := mux.NewRouter()

	// Define a GET route for downloading files

	router.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("api is working well"))
	})

	router.HandleFunc("/download/{filename}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		filename := vars["filename"]
		filePath := filepath.Join("/tmp", filename) // Construct the full file path

		fmt.Println("Attempting to serve file:", filePath) // Debug print

		// Check if the file exists
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			fmt.Println("File not found:", filePath) // Debug print
			http.Error(w, "file not found", http.StatusNotFound)
			return
		} else if err != nil {
			fmt.Println("Error checking file:", err) // Debug print
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		// Serve the file
		http.ServeFile(w, r, filePath)
	}).Methods("GET")

	// Print the current working directory
	if cwd, err := os.Getwd(); err == nil {
		fmt.Println("Current working directory:", cwd)
	} else {
		fmt.Println("Error getting working directory:", err)
	}

	// Listen and serve on 0.0.0.0:8030
	fmt.Println("Server is running on port 8030")
	if err := http.ListenAndServe(":8035", router); err != nil {
		fmt.Println("Failed to start server:", err)
	}
}
