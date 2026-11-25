package queue

import (
	"context"
	"sync"
	"time"
)

// Queue aplica debounce por arquivo e processa jobs sequencialmente (1 por vez).
type Queue struct {
	mu       sync.Mutex
	debounce time.Duration
	timers   map[string]*time.Timer
	jobs     chan string
	handler  func(ctx context.Context, filePath string)
	closed   bool
}

// New cria uma fila sequencial com janela de debounce configurável.
func New(debounce time.Duration, bufferSize int, handler func(ctx context.Context, filePath string)) *Queue {
	if debounce <= 0 {
		debounce = 500 * time.Millisecond
	}
	if bufferSize <= 0 {
		bufferSize = 128
	}
	return &Queue{
		debounce: debounce,
		timers:   make(map[string]*time.Timer),
		jobs:     make(chan string, bufferSize),
		handler:  handler,
	}
}

// SetDebounce atualiza a janela de debounce em tempo de execução.
func (q *Queue) SetDebounce(d time.Duration) {
	if d <= 0 {
		return
	}
	q.mu.Lock()
	q.debounce = d
	q.mu.Unlock()
}

// Enqueue agenda o arquivo após a janela de debounce, reiniciando o timer em saves repetidos.
func (q *Queue) Enqueue(filePath string) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.closed {
		return
	}
	if t, exists := q.timers[filePath]; exists {
		t.Stop()
	}
	q.timers[filePath] = time.AfterFunc(q.debounce, func() {
		q.mu.Lock()
		delete(q.timers, filePath)
		closed := q.closed
		q.mu.Unlock()
		if closed {
			return
		}
		select {
		case q.jobs <- filePath:
		default:
			// Evita bloquear o timer caso o buffer esteja saturado.
		}
	})
}

// Run consome a fila sequencialmente (1 job por vez) até ctx ser cancelado.
func (q *Queue) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			q.shutdownTimers()
			return
		case filePath := <-q.jobs:
			if q.handler != nil {
				q.handler(ctx, filePath)
			}
		}
	}
}

func (q *Queue) shutdownTimers() {
	q.mu.Lock()
	q.closed = true
	for k, t := range q.timers {
		t.Stop()
		delete(q.timers, k)
	}
	q.mu.Unlock()
}
