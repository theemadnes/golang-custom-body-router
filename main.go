package main

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

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

	callResponseChan := make(chan string) // Create a channel
	errorChan := make(chan error)         // Channel for errors

	// Attempt to call the downstream service
	go func() {
		fmt.Println("Calling downstream service")
		resp, err := http.Get("https://golang-hello-word-603904278888.us-central1.run.app")
		if err != nil {
			// Handle error (e.g., log the error)
			//fmt.Println("Error making GET request:", err)
			errorChan <- err // Send the error to the channel
			return           // Exit the goroutine if there's an error
		}
		defer resp.Body.Close()

		// Read the response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading response body:", err)
			return
		}

		// Create a reader for the JSON response
		reader := strings.NewReader(string(body))

		// Decode the JSON response
		var result map[string]interface{}
		if err := json.NewDecoder(reader).Decode(&result); err != nil {
			fmt.Println("Error decoding JSON:", err)
			return
		}
		fmt.Println("Response body:", string(body))

		fmt.Println("Message", result["message"])

		// Update callResponse if the GET request is successful
		callResponseChan <- result["message"].(string)
	}()

	// Respond with the evaluation result
	select {
	case callResponse := <-callResponseChan:
		// ... (Respond with the evaluation result) ...
		response := map[string]interface{}{
			"compressed": contentEncoding == "gzip",
			"data":       data,
			"call":       callResponse,
		}
		json.NewEncoder(w).Encode(response)
	case err := <-errorChan:
		// Handle the error from the downstream service
		fmt.Println("Error from downstream service:", err)
		//http.Error(w, "Error calling downstream service", http.StatusInternalServerError)
		response := map[string]interface{}{
			"compressed": contentEncoding == "gzip",
			"data":       data,
			"call":       "error",
		}
		json.NewEncoder(w).Encode(response)
	}

}
