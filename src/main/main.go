package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

var (
	nextChan   = make(chan struct{})
	currentCmd *exec.Cmd
	cmdMu      sync.Mutex

	currentTrack string
	mu           sync.Mutex
	musicDir     = "./music"
	lastErr      error
)

func main() {
	go playbackLoop()
	next()

	http.HandleFunc("/status", statusHandler)
	http.HandleFunc("/next", nextHandler)
	http.HandleFunc("/add", addHandler)
	http.HandleFunc("/list", listHandler)

	log.Println("API listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func playbackLoop() {
	for {
		// <-nextChan // Warte auf /next

		file := getRandomFile()
		if file == "" {
			log.Println("No music files found.")
			time.Sleep(10 * time.Second)
			continue
		}

		log.Printf("Playing: %s", file)
		cmd := exec.Command("ffplay", "-nodisp", "-autoexit", file)
		cmdMu.Lock()
		currentCmd = cmd
		cmdMu.Unlock()

		err := cmd.Run() // blockiert bis Lied zu Ende
		if err != nil {
			log.Printf("Error playing file: %v", err)
		}
	}
}

func getRandomFile() string {
	files, err := filepath.Glob(filepath.Join(musicDir, "*.mp3"))
	if err != nil || len(files) == 0 {
		return ""
	}
	return files[rand.Intn(len(files))]
}

func next() {
	go func() {
		// Beende aktuelles Lied (optional)
		cmdMu.Lock()
		if currentCmd != nil && currentCmd.Process != nil {
			_ = currentCmd.Process.Kill()
		}
		cmdMu.Unlock()

		// Signal für nächstes Lied
		nextChan <- struct{}{}
	}()

}

func nextHandler(w http.ResponseWriter, r *http.Request) {

	next()
	w.Write([]byte("Next song requested.\n"))
}

func addHandler(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Query().Get("url")
	if url == "" {
		http.Error(w, "Missing 'url' query parameter", http.StatusBadRequest)
		return
	}

	go func(url string) {
		cmd := exec.Command("yt-dlp",
			"--extract-audio",
			"--audio-format", "mp3",
			"--output", filepath.Join(musicDir, "%(title)s.%(ext)s"),
			url,
		)
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("Error downloading YouTube audio: %v\nOutput: %s\n", err, string(output))
		} else {
			log.Printf("Downloaded: %s", url)
		}
	}(url)

	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("Download started\n"))
}

func listHandler(w http.ResponseWriter, r *http.Request) {
	files, err := filepath.Glob(filepath.Join(musicDir, "*.mp3"))
	if err != nil {
		http.Error(w, "Fehler beim Lesen des Musikverzeichnisses", http.StatusInternalServerError)
		return
	}

	var list []string
	for _, f := range files {
		list = append(list, filepath.Base(f))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()
	if lastErr != nil {
		json.NewEncoder(w).Encode(map[string]string{"currentTrack": currentTrack, "lastError": lastErr.Error()})
	} else {
		json.NewEncoder(w).Encode(map[string]string{"currentTrack": currentTrack, "lastError": ""})
	}
}
