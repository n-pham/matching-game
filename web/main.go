package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"path/filepath"
	"strconv"
	"time"
)

var tiles [8]string

func init() {
	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())

	// Read the list of images from the static/img directory
	files, err := filepath.Glob("static/img/*.svg")
	if err != nil {
		log.Fatalf("Failed to read images from static/img directory: %v", err)
	}

	// Shuffle the list of images
	rand.Shuffle(len(files), func(i, j int) { files[i], files[j] = files[j], files[i] })

	// Pick the first 4 images
	if len(files) < 4 {
		log.Fatalf("Not enough images in static/img directory")
	}
	selectedImages := files[:4]

	// Assign each image to two random positions in the tiles array
	assigned := make(map[int]bool)
	for _, img := range selectedImages {
		for i := 0; i < 2; i++ {
			for {
				pos := rand.Intn(len(tiles))
				if !assigned[pos] {
					tiles[pos] = img
					assigned[pos] = true
					break
				}
			}
		}
	}
}

func tileContentHandler(w http.ResponseWriter, r *http.Request) {
	// Get the id parameter from the query
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Missing id parameter", http.StatusBadRequest)
		log.Printf("Error: Missing id parameter")
		return
	}

	// Convert id to integer
	index, err := strconv.Atoi(id)
	if err != nil || index < 0 || index >= len(tiles) {
		http.Error(w, "Invalid id parameter", http.StatusBadRequest)
		log.Printf("Error: Invalid id parameter: %v", err)
		return
	}

	// Get the image URL for the given id
	imageURL := tiles[index-1]

	// Return the image URL as JSON
	response := fmt.Sprintf(`<img src="%s" alt="Default Tile Image" style="display: block;">`, imageURL)
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte(response))
	if err != nil {
		log.Printf("Error writing response: %v", err)
	}
}

func main() {
	// Serve the index.html file at the root
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/index.html")
	})
	// Serve image files from the static/img directory
	http.Handle("/static/img/", http.StripPrefix("/static/img/", http.FileServer(http.Dir("static/img"))))

	// API endpoint for tile content
	http.HandleFunc("/api/tile-content", tileContentHandler)

	// Start the server
	log.Println("Starting server on :8080")
	log.Println("Tiles:", tiles)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
