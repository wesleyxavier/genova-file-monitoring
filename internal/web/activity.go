package web

import (
	"sync"
	"time"
)

const maxActivityHistory = 1000

// ActivityEvent representa um evento em tempo real do monitoramento para o shell da UI.
type ActivityEvent struct {
	ID        int64  `json:"id"`
	Timestamp string `json:"timestamp"`
	Type      string `json:"type"` // detected | analyzing | clean | finding | error
	FilePath  string `json:"file_path"`
	Message   string `json:"message"`
}

// ActivityHub mantém um ring buffer de eventos recentes e distribui via SSE.
type ActivityHub struct {
	mu     sync.Mutex
	nextID int64
	events []ActivityEvent
	subs   map[chan ActivityEvent]struct{}
}

// NewActivityHub cria um hub de eventos de atividade.
func NewActivityHub() *ActivityHub {
	return &ActivityHub{
		events: make([]ActivityEvent, 0, 128),
		subs:   make(map[chan ActivityEvent]struct{}),
	}
}

// Publish registra um novo evento e notifica todos os clientes SSE conectados.
func (h *ActivityHub) Publish(evType, filePath, message string) {
	h.mu.Lock()
	h.nextID++
	ev := ActivityEvent{
		ID:        h.nextID,
		Timestamp: time.Now().Format("15:04:05"),
		Type:      evType,
		FilePath:  filePath,
		Message:   message,
	}
	if len(h.events) >= maxActivityHistory {
		copy(h.events, h.events[1:])
		h.events[len(h.events)-1] = ev
	} else {
		h.events = append(h.events, ev)
	}
	for ch := range h.subs {
		select {
		case ch <- ev:
		default:
		}
	}
	h.mu.Unlock()
}

// Snapshot retorna uma cópia dos eventos recentes armazenados.
func (h *ActivityHub) Snapshot() []ActivityEvent {
	h.mu.Lock()
	defer h.mu.Unlock()
	out := make([]ActivityEvent, len(h.events))
	copy(out, h.events)
	return out
}

// Subscribe registra um canal para receber novos eventos em tempo real.
func (h *ActivityHub) Subscribe() (chan ActivityEvent, func()) {
	ch := make(chan ActivityEvent, 64)
	h.mu.Lock()
	h.subs[ch] = struct{}{}
	h.mu.Unlock()
	return ch, func() {
		h.mu.Lock()
		delete(h.subs, ch)
		close(ch)
		h.mu.Unlock()
	}
}
