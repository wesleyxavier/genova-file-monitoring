package queue

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestQueueDebounceAndSequentialExecution(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var count atomic.Int32
	var active atomic.Int32
	var maxActive atomic.Int32

	q := New(80*time.Millisecond, 16, func(ctx context.Context, filePath string) {
		cur := active.Add(1)
		for {
			prev := maxActive.Load()
			if cur <= prev || maxActive.CompareAndSwap(prev, cur) {
				break
			}
		}
		time.Sleep(20 * time.Millisecond)
		active.Add(-1)
		count.Add(1)
	})
	go q.Run(ctx)

	// Múltiplos saves rápidos no mesmo arquivo devem resultar em apenas 1 job.
	q.Enqueue("a.go")
	q.Enqueue("a.go")
	q.Enqueue("a.go")
	// Arquivo distinto gera outro job.
	q.Enqueue("b.go")

	deadline := time.Now().Add(2 * time.Second)
	for count.Load() < 2 && time.Now().Before(deadline) {
		time.Sleep(20 * time.Millisecond)
	}

	if got := count.Load(); got != 2 {
		t.Fatalf("esperava 2 execuções (1 para a.go + 1 para b.go), obteve %d", got)
	}
	if maxActive.Load() != 1 {
		t.Fatalf("esperava execução estritamente sequencial (1 por vez), obteve %d", maxActive.Load())
	}
}
