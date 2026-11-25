package analyzer

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// LlamaProcess gerencia o ciclo de vida do processo filho llama-server.
type LlamaProcess struct {
	mu      sync.Mutex
	cmd     *exec.Cmd
	port    int
	ready   atomic.Bool
	baseURL string
}

// NewLlamaProcess cria um gerenciador para a porta validada (11440-11450).
func NewLlamaProcess(port int) (*LlamaProcess, error) {
	if err := ValidateLlamaPort(port); err != nil {
		return nil, err
	}
	return &LlamaProcess{
		port:    port,
		baseURL: fmt.Sprintf("http://127.0.0.1:%d", port),
	}, nil
}

// Start inicia o llama-server como processo filho e aguarda /health responder 200 OK.
func (p *LlamaProcess) Start(ctx context.Context, cmdStr string, readyTimeout time.Duration) error {
	args := splitCommandArgs(cmdStr)
	if len(args) == 0 {
		return errors.New("LLAMA_SERVER_CMD não configurado no .env")
	}

	p.mu.Lock()
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	if err := cmd.Start(); err != nil {
		p.mu.Unlock()
		return fmt.Errorf("falha ao iniciar llama-server: %w", err)
	}
	p.cmd = cmd
	p.mu.Unlock()

	if readyTimeout <= 0 {
		readyTimeout = 45 * time.Second
	}
	deadline := time.Now().Add(readyTimeout)
	client := &http.Client{Timeout: 2 * time.Second}

	for time.Now().Before(deadline) {
		if ctx.Err() != nil {
			_ = p.Stop()
			return ctx.Err()
		}
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, p.baseURL+"/health", nil)
		resp, err := client.Do(req)
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				p.ready.Store(true)
				return nil
			}
		}
		time.Sleep(500 * time.Millisecond)
	}

	_ = p.Stop()
	return fmt.Errorf("timeout aguardando /health do llama-server na porta %d", p.port)
}

// IsReady informa se o llama-server respondeu ao healthcheck.
func (p *LlamaProcess) IsReady() bool {
	return p.ready.Load()
}

// Port retorna a porta configurada para o llama-server.
func (p *LlamaProcess) Port() int {
	return p.port
}

// Stop encerra o processo filho do llama-server de forma limpa.
func (p *LlamaProcess) Stop() error {
	p.ready.Store(false)
	p.mu.Lock()
	cmd := p.cmd
	p.cmd = nil
	p.mu.Unlock()

	if cmd == nil || cmd.Process == nil {
		return nil
	}
	_ = cmd.Process.Kill()
	_, _ = cmd.Process.Wait()
	return nil
}

func splitCommandArgs(s string) []string {
	var args []string
	var cur strings.Builder
	inQuote := byte(0)

	for i := 0; i < len(s); i++ {
		ch := s[i]
		switch {
		case inQuote != 0:
			if ch == inQuote {
				inQuote = 0
			} else {
				cur.WriteByte(ch)
			}
		case ch == '"' || ch == '\'':
			inQuote = ch
		case ch == ' ' || ch == '\t':
			if cur.Len() > 0 {
				args = append(args, cur.String())
				cur.Reset()
			}
		default:
			cur.WriteByte(ch)
		}
	}
	if cur.Len() > 0 {
		args = append(args, cur.String())
	}
	return args
}
