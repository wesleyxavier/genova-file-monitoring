package web

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/wesleyxavier/genova-file-monitoring/internal/store"
)

//go:embed static/index.html
var staticFS embed.FS

// HealthProvider fornece o estado atual dos subsistemas locais.
type HealthProvider interface {
	HealthStatus(ctx context.Context) HealthResponse
}

// HealthResponse representa a resposta de GET /api/health.
type HealthResponse struct {
	Status        string `json:"status"`
	WatcherActive bool   `json:"watcher_active"`
	WatchPath     string `json:"watch_path"`
	LLMReady      bool   `json:"llm_ready"`
	LLMPort       int    `json:"llm_port"`
	FalkorDBReady bool   `json:"falkordb_ready"`
}

// Server expõe a UI e API local exclusivamente em 127.0.0.1.
type Server struct {
	addr           string
	store          store.Store
	health         HealthProvider
	activity       *ActivityHub
	onConfigChange func(store.MonitoringConfig) error
	httpServer     *http.Server
}

// NewServer cria o servidor HTTP vinculado estritamente a 127.0.0.1:<port>.
func NewServer(port int, st store.Store, health HealthProvider, activity *ActivityHub, onConfigChange func(store.MonitoringConfig) error) *Server {
	s := &Server{
		addr:           fmt.Sprintf("127.0.0.1:%d", port),
		store:          st,
		health:         health,
		activity:       activity,
		onConfigChange: onConfigChange,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", s.handleIndex)
	mux.HandleFunc("GET /api/health", s.handleHealth)
	mux.HandleFunc("GET /api/config", s.handleGetConfig)
	mux.HandleFunc("PUT /api/config", s.handlePutConfig)
	mux.HandleFunc("GET /api/findings", s.handleListFindings)
	mux.HandleFunc("GET /api/activity", s.handleActivitySnapshot)
	mux.HandleFunc("GET /api/events", s.handleEventsSSE)
	mux.HandleFunc("POST /api/pick-folder", s.handlePickFolder)

	s.httpServer = &http.Server{
		Addr:              s.addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	return s
}

// Addr retorna o endereço de bind (sempre 127.0.0.1:<port>).
func (s *Server) Addr() string {
	return s.addr
}

// Handler retorna o http.Handler interno (útil para testes com httptest).
func (s *Server) Handler() http.Handler {
	return s.httpServer.Handler
}

// Start inicia o listener HTTP em 127.0.0.1.
func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

// Shutdown encerra o servidor HTTP de forma limpa.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	data, err := staticFS.ReadFile("static/index.html")
	if err != nil {
		http.Error(w, "UI indisponível", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(data)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	var resp HealthResponse
	if s.health != nil {
		resp = s.health.HealthStatus(r.Context())
	} else {
		resp = HealthResponse{Status: "ok"}
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	cfg, err := s.store.LoadConfig(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	writeJSON(w, http.StatusOK, cfg)
}

func (s *Server) handlePutConfig(w http.ResponseWriter, r *http.Request) {
	var cfg store.MonitoringConfig
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}
	if err := store.ValidateConfig(cfg); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := s.store.SaveConfig(r.Context(), cfg); err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	if s.onConfigChange != nil {
		if err := s.onConfigChange(cfg); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	writeJSON(w, http.StatusOK, cfg)
}

func (s *Server) handleListFindings(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	items, err := s.store.ListFindings(r.Context(), store.FindingFilter{
		Severity: store.Severity(q.Get("severity")),
		File:     q.Get("file"),
		Limit:    limit,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func (s *Server) handleActivitySnapshot(w http.ResponseWriter, r *http.Request) {
	if s.activity == nil {
		writeJSON(w, http.StatusOK, map[string]any{"items": []ActivityEvent{}})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": s.activity.Snapshot()})
}

func (s *Server) handleEventsSSE(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok || s.activity == nil {
		http.Error(w, "SSE indisponível", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ch, unsub := s.activity.Subscribe()
	defer unsub()

	for {
		select {
		case <-r.Context().Done():
			return
		case ev, open := <-ch:
			if !open {
				return
			}
			data, err := json.Marshal(ev)
			if err != nil {
				continue
			}
			_, _ = fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		}
	}
}

func (s *Server) handlePickFolder(w http.ResponseWriter, r *http.Request) {
	path, err := PickFolderNative(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"path": path})
}
