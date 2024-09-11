package main

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file:", err)
	}

	// Now you can access variables like before
	port := os.Getenv("PORT")
	if port == "" {
		fmt.Println("PORT environment variable not set; setting to default of 8080")
		port = "8080"
	}
	fmt.Println("PORT:", port)

	// Start the server

	http.HandleFunc("/", evaluateHandler)
	fmt.Println("Server listening on port ", port)
	http.ListenAndServe(":8080", nil)
}

func evaluateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Determine if the body is compressed
	var reader io.Reader = r.Body
	contentEncoding := r.Header.Get("Content-Encoding")
	if contentEncoding == "gzip" {
		gzipReader, err := gzip.NewReader(r.Body)
		if err != nil {
			http.Error(w, "Invalid gzip encoding", http.StatusBadRequest)
			return
		}
		defer gzipReader.Close()
		reader = gzipReader
	}

	// Attempt to decode the JSON body
	var data map[string]interface{}
	if err := json.NewDecoder(reader).Decode(&data); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	// Respond with the evaluation result
	response := map[string]interface{}{
		"compressed": contentEncoding == "gzip",
		"data":       data,
	}
	json.NewEncoder(w).Encode(response)
}
