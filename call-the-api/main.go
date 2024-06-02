package main

import (
	"io"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	router := mux.NewRouter()

	// Define a route for downloading the file
	router.HandleFunc("/download", func(w http.ResponseWriter, r *http.Request) {
		fileURL := "http://localhost:8035/download/ubuntu-22.04.4-live-server-amd64.iso"

		// Get the file
		resp, err := http.Get(fileURL)
		if err != nil {
			http.Error(w, "Error downloading file: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		// Check server response
		if resp.StatusCode != http.StatusOK {
			http.Error(w, "Error downloading file: "+resp.Status, http.StatusInternalServerError)
			return
		}

		// Set the headers
		w.Header().Set("Content-Disposition", "attachment; filename=ubuntu-22.04.4-live-server-amd64.iso")
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Length", resp.Header.Get("Content-Length"))

		// Stream the file content to the response
		if _, err := io.Copy(w, resp.Body); err != nil {
			http.Error(w, "Error copying file: "+err.Error(), http.StatusInternalServerError)
			return
		}
	})

	// Start the HTTP server
	http.ListenAndServe(":8081", router)
}
