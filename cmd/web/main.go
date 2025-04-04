package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"path/filepath"
	"strconv"
)

var TILE_CNT int = 12
var IMG_DIR string = "static/img"
var tiles []string
var found []bool
var previousIndex int = -1

func initializeGame() {
	// Read the list of images from the static/img directory
	files, err := filepath.Glob(IMG_DIR + "/A*.svg")
	if err != nil {
		log.Fatalf("Failed to read images from static/img directory: %v", err)
	}

	// Shuffle the list of images
	rand.Shuffle(len(files), func(i, j int) { files[i], files[j] = files[j], files[i] })

	// Pick the first images
	if len(files) < TILE_CNT/2 {
		log.Fatalf("Not enough images in static/img directory")
	}

	tiles = make([]string, TILE_CNT)
	// Assign each image to two random positions in the tiles array
	for i := range TILE_CNT / 2 {
		for range 2 {
			for {
				pos := rand.Intn(TILE_CNT)
				if tiles[pos] == "" {
					tiles[pos] = files[i]
					break
				}
			}
		}
	}

	found = make([]bool, TILE_CNT)
	log.Println("Tiles:", tiles)
}

func init() {
	initializeGame()
}

func tileClickHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	index, err := strconv.Atoi(id)
	if err != nil || index < 0 || index >= len(tiles) {
		http.Error(w, "Invalid id parameter", http.StatusBadRequest)
		log.Printf("Error: Invalid id parameter: %v", err)
		return
	}

	var response string
	if previousIndex == -1 {
		response = ""
		for i := range len(tiles) {
			response += fmt.Sprintf(`<div class="tile" id="tile%d" hx-get="/api/tile-content?id=%d" hx-trigger="click" hx-swap="innerHTML">`, i, i)
			if found[i] || i == index {
				response += fmt.Sprintf(`<img src="%s" alt="IMG">`, tiles[i])
			} else {
				response += `<img src="static/img/question-mark.svg" alt="IMG">`
			}
			response += `</div>`
		}
		previousIndex = index
	} else {
		if index != previousIndex {
			if tiles[index] == tiles[previousIndex] {
				found[index] = true
				found[previousIndex] = true
			}
		}
		response = ""
		for i := range len(tiles) {
			response += fmt.Sprintf(`<div class="tile" id="tile%d" hx-get="/api/tile-content?id=%d" hx-trigger="click" hx-swap="innerHTML">`, i, i)
			if found[i] || i == index || i == previousIndex {
				response += fmt.Sprintf(`<img src="%s" alt="IMG">`, tiles[i])
			} else {
				response += `<img src="static/img/question-mark.svg" alt="IMG">`
			}
			response += `</div>`
		}
		previousIndex = -1
	}
	w.Header().Set("HX-Target", ".grid")
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte(response))
	if err != nil {
		log.Printf("Error writing response: %v", err)
	}
}

func resetHandler(w http.ResponseWriter, r *http.Request) {
	initializeGame()
	previousIndex = -1
	var response string
	w.Header().Set("HX-Target", "updateTiles")
	for i := range TILE_CNT {
		response += fmt.Sprintf(`<div class="tile" id="tile%d" hx-get="/api/tile-content?id=%d" hx-trigger="click" hx-swap="innerHTML">`, i, i)
		response += `<img src="static/img/question-mark.svg" alt="IMG">`
		response += `</div>`
	}
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(response))
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
	http.HandleFunc("/api/tile-content", tileClickHandler)
	// API endpoint for reset
	http.HandleFunc("/api/reset", resetHandler)

	// Start the server
	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
