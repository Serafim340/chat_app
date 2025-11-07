package main

import (
	"flag"
	"log"
	"net/http"
	"path/filepath"
	"sync"
	"text/template"
)

type templateHandler struct {
	once     sync.Once
	filename string
	templ    *template.Template
}

func (t *templateHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	t.once.Do(func() {
		t.templ = template.Must(template.ParseFiles(filepath.Join("templates", t.filename)))
	})
	t.templ.Execute(w, req)
}

func main() {

	var addr = flag.String("addr", ":8000", "Addr of the app")
	flag.Parse()

	r := newRoom()
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.Handle("/", &templateHandler{filename: "index.html"})
	http.Handle("/chat", &templateHandler{filename: "chat.html"})
	http.HandleFunc("/room", func(w http.ResponseWriter, r *http.Request) {
		roomName := r.URL.Query().Get("room")
		if roomName == "" {
			http.Error(w, "No room specified", http.StatusBadRequest)
			return
		}
		realRoom := getRoom(roomName)
		realRoom.ServeHTTP(w, r)
	})

	go r.run()

	log.Println("Starting web server on", *addr)

	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
		return
	}
}
