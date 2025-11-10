package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"time"

	"chat_app/internal/hub"
)

var regExp = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,20}$`)

func main() {
	addr := flag.String("addr", ":8000", "Address of the app")
	flag.Parse()

	mux := http.NewServeMux()

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join("templates", "index.html"))
	})
	mux.HandleFunc("/rooms", func(w http.ResponseWriter, r *http.Request) {
		list := hub.GetRoomNames()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(list)
	})
	mux.HandleFunc("/chat", func(w http.ResponseWriter, r *http.Request) {
		room := r.URL.Query().Get("room")
		nick := r.URL.Query().Get("nick")
		existing := r.URL.Query().Get("existing")

		// Если пользователь выбрал комнату из списка — она имеет приоритет
		if existing != "" {
			room = existing
		}

		if room == "" || !regExp.MatchString(room) {
			http.Error(w, "Invalid room name", http.StatusBadRequest)
			return
		}

		if nick == "" || !regExp.MatchString(nick) {
			http.Error(w, "Invalid nickname", http.StatusBadRequest)
			return
		}

		// Просто отдать шаблон — chat.js сам из URL возьмёт параметры
		http.ServeFile(w, r, filepath.Join("templates", "chat.html"))
	})

	mux.HandleFunc("/room", func(w http.ResponseWriter, r *http.Request) {
		roomName := r.URL.Query().Get("room")
		nick := r.URL.Query().Get("nick")

		if roomName == "" || !regExp.MatchString(roomName) {
			http.Error(w, "Invalid room name", http.StatusBadRequest)
			return
		}

		if nick == "" || !regExp.MatchString(nick) {
			http.Error(w, "Invalid nickname", http.StatusBadRequest)
			return
		}

		hub.ServeRoomHTTP(w, r, roomName, nick)
	})

	server := &http.Server{
		Addr:    *addr,
		Handler: mux,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	go func() {
		log.Println("Starting server on", *addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("ListenAndServe error:", err)
		}
	}()

	<-stop
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown failed:", err)
	}

	log.Println("Server exited gracefully")
}
