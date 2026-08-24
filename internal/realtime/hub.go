package realtime

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
)

type Hub struct {
	mu        sync.RWMutex
	clients   map[chan []byte]struct{}
	broadcast chan []byte
}

func NewHub() *Hub {
	return &Hub{clients: map[chan []byte]struct{}{}, broadcast: make(chan []byte, 32)}
}
func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client <- message:
				default:
				}
			}
			h.mu.RUnlock()
		case <-ctx.Done():
			return
		}
	}
}
func (h *Hub) Publish(event string, payload any) {
	message, _ := json.Marshal(map[string]any{"event": event, "data": payload})
	select {
	case h.broadcast <- message:
	default:
	}
}
func (h *Hub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "stream unsupported", 500)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	client := make(chan []byte, 8)
	h.mu.Lock()
	h.clients[client] = struct{}{}
	h.mu.Unlock()
	defer func() { h.mu.Lock(); delete(h.clients, client); close(client); h.mu.Unlock() }()
	for {
		select {
		case message := <-client:
			_, _ = w.Write(append(append([]byte("data: "), message...), '\n', '\n'))
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}
