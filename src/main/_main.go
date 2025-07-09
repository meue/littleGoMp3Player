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
	currentTrack string
	cmd          *exec.Cmd
	mu           sync.Mutex
	musicDir     = "./music"
	lastErr      error
)

func main() {
	rand.Seed(time.Now().UnixNano())
	go playRandomTrack()

	http.HandleFunc("/status", statusHandler)
	http.HandleFunc("/next", nextHandler)
	http.HandleFunc("/add", addHandler)

	log.Println("API listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func playRandomTrack() {
	files, err := filepath.Glob(filepath.Join(musicDir, "*.mp3"))
	if err != nil || len(files) == 0 {
		log.Fatal("No MP3 files found in music directory.")
	}

	track := files[rand.Intn(len(files))]

	mu.Lock()
	currentTrack = filepath.Base(track)
	cmd = exec.Command("mpg123", track)
	mu.Unlock()

	log.Println("Playing:", currentTrack)
	err = cmd.Run()
	if err != nil {
		lastErr = err
		log.Println("Error playing track:", err)
	}
}

func stopCurrentTrack() {
	mu.Lock()
	defer mu.Unlock()
	if cmd != nil && cmd.Process != nil {
		_ = cmd.Process.Kill()
	}
}

func nextHandler(w http.ResponseWriter, r *http.Request) {
	stopCurrentTrack()
	go playRandomTrack()
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Playing next track\n"))
}

func addHandler(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Query().Get("url")
	if url == "" {
		http.Error(w, "Missing 'url' query parameter", http.StatusBadRequest)
		return
	}

	go func(url string) {
		cmd := exec.Command("/usr/local/bin/yt-dlp",
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

func statusHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()
	if lastErr != nil {
		json.NewEncoder(w).Encode(map[string]string{"currentTrack": currentTrack, "lastError": lastErr.Error()})
	} else {
		json.NewEncoder(w).Encode(map[string]string{"currentTrack": currentTrack, "lastError": ""})
	}
}
