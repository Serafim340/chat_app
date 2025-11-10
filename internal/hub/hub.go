package hub

import (
	"net/http"
	"sync"
)

var (
	rooms = make(map[string]*room)
	mu    sync.Mutex
)

// ServeRoomHTTP — единственная экспортируемая функция
func ServeRoomHTTP(w http.ResponseWriter, req *http.Request, name string) {
	mu.Lock()
	r, ok := rooms[name]
	if !ok {
		r = newRoom()
		rooms[name] = r
		go r.run()
	}
	mu.Unlock()

	r.serveHTTP(w, req)
}
