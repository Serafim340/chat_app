package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"text/template"
	"time"

	"chat_app/internal/hub"
)

var roomNamePattern = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,20}$`)

type templateHandler struct {
	filename string
	templ    *template.Template
}

func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if t.templ == nil {
		path := filepath.Join("templates", t.filename)
		t.templ = template.Must(template.ParseFiles(path))
	}
	if err := t.templ.Execute(w, nil); err != nil {
		http.Error(w, "Template rendering error", http.StatusInternalServerError)
	}
}

func main() {
	addr := flag.String("addr", ":8000", "Address of the app")
	flag.Parse()

	mux := http.NewServeMux()

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	mux.Handle("/", &templateHandler{filename: "index.html"})
	mux.Handle("/chat", &templateHandler{filename: "chat.html"})

	mux.HandleFunc("/room", func(w http.ResponseWriter, r *http.Request) {
		roomName := r.URL.Query().Get("room")
		if !roomNamePattern.MatchString(roomName) {
			http.Error(w, "Invalid room name", http.StatusBadRequest)
			return
		}
		hub.ServeRoomHTTP(w, r, roomName)
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
