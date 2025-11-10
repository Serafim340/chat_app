package hub

import (
	"net/http"
	"sync"
)

var (
	rooms = make(map[string]*room)
	mu    sync.Mutex
)

func ServeRoomHTTP(w http.ResponseWriter, req *http.Request, roomName string, nick string) {
	mu.Lock()
	r, ok := rooms[roomName]
	if !ok {
		r = newRoom(roomName)
		rooms[roomName] = r
		go r.run()
	}
	mu.Unlock()

	r.serveHTTP(w, req, nick)
}

// GetRoomNames — список всех комнат
func GetRoomNames() []string {
	mu.Lock()
	defer mu.Unlock()
	names := []string{}
	for n := range rooms {
		names = append(names, n)
	}
	return names
}
