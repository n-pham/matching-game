package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"path/filepath"
	"strconv"
)

var tiles [8]string
var IMG_DIR string = "static/img"
var selectedIndex int

func init() {
	// Read the list of images from the static/img directory
	files, err := filepath.Glob(IMG_DIR + "/A*.svg")
	if err != nil {
		log.Fatalf("Failed to read images from static/img directory: %v", err)
	}

	// Shuffle the list of images
	rand.Shuffle(len(files), func(i, j int) { files[i], files[j] = files[j], files[i] })

	// Pick the first 4 images
	if len(files) < 4 {
		log.Fatalf("Not enough images in static/img directory")
	}

	// Assign each image to two random positions in the tiles array
	assigned := make(map[int]bool)
	for _, img := range files[:4] {
		for range 2 {
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

func tileClickHandler(w http.ResponseWriter, r *http.Request) {
	//
	id := r.URL.Query().Get("id")
	log.Print(id)
	index, err := strconv.Atoi(id)
	if err != nil || index < 0 || index >= len(tiles) {
		http.Error(w, "Invalid id parameter", http.StatusBadRequest)
		log.Printf("Error: Invalid id parameter: %v", err)
		return
	}

	var response string
	if selectedIndex == -1 {
		w.Header().Set("HX-Trigger", fmt.Sprintf(`update%d`, index))
		response = fmt.Sprintf(`<img src="%s" alt="IMG">`, tiles[index])
		selectedIndex = index
	} else if index != selectedIndex {
		w.Header().Set("HX-Trigger", fmt.Sprintf(`update%d,update%d`, index, selectedIndex))
		if tiles[index] == tiles[selectedIndex] {
			response = fmt.Sprintf(`<img src="%s" alt="IMG">`, tiles[index])
		} else {
			response = fmt.Sprintf(`<img src="%s" alt="IMG">`, "static/img/question-mark.svg")
		}
		selectedIndex = -1
	}

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
		http.ServeFile(w, r, "static/test.html")
	})
	// Serve image files from the static/img directory
	http.Handle("/static/img/", http.StripPrefix("/static/img/", http.FileServer(http.Dir("static/img"))))

	// API endpoint for tile content
	http.HandleFunc("/api/tile-content", tileClickHandler)
	// Start the server
	log.Println("Starting server on :8080")
	log.Println("tiles:", tiles)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
