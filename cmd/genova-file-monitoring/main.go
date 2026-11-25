package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/wesleyxavier/genova-file-monitoring/internal/analyzer"
	"github.com/wesleyxavier/genova-file-monitoring/internal/notify"
	"github.com/wesleyxavier/genova-file-monitoring/internal/queue"
	"github.com/wesleyxavier/genova-file-monitoring/internal/store"
	"github.com/wesleyxavier/genova-file-monitoring/internal/watcher"
	"github.com/wesleyxavier/genova-file-monitoring/internal/web"
)

type appRuntime struct {
	mu            sync.Mutex
	cfg           store.MonitoringConfig
	store         store.Store
	llama         *analyzer.LlamaProcess
	llmClient     *analyzer.LLMClient
	queue         *queue.Queue
	notifySvc     *notify.Service
	activity      *web.ActivityHub
	watcher       *watcher.Watcher
	watcherCancel context.CancelFunc
}

func (a *appRuntime) HealthStatus(ctx context.Context) web.HealthResponse {
	a.mu.Lock()
	watchPath := a.cfg.WatchPath
	watcherActive := a.watcher != nil
	a.mu.Unlock()

	falkorOK := a.store.Ping(ctx) == nil
	return web.HealthResponse{
		Status:        "ok",
		WatcherActive: watcherActive,
		WatchPath:     watchPath,
		LLMReady:      a.llama != nil && a.llama.IsReady(),
		LLMPort:       a.llama.Port(),
		FalkorDBReady: falkorOK,
	}
}

func (a *appRuntime) applyConfig(parentCtx context.Context, cfg store.MonitoringConfig) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.watcherCancel != nil {
		a.watcherCancel()
		a.watcherCancel = nil
	}
	if a.watcher != nil {
		_ = a.watcher.Close()
		a.watcher = nil
	}

	a.cfg = cfg
	a.queue.SetDebounce(time.Duration(cfg.DebounceMS) * time.Millisecond)
	a.notifySvc.UpdateConfig(cfg)

	if cfg.WatchPath == "" {
		return nil
	}

	w, err := watcher.New(cfg.WatchPath, cfg.Extensions, func(filePath string) {
		rel := filePath
		if r, err := filepath.Rel(cfg.WatchPath, filePath); err == nil {
			rel = filepath.ToSlash(r)
		}
		if a.activity != nil {
			a.activity.Publish("detected", rel, "Arquivo salvo/modificado; aguardando debounce...")
		}
		a.queue.Enqueue(filePath)
	})
	if err != nil {
		return err
	}
	wCtx, wCancel := context.WithCancel(parentCtx)
	a.watcher = w
	a.watcherCancel = wCancel
	go w.Run(wCtx)
	log.Printf("watcher: observando %s", cfg.WatchPath)
	return nil
}

func (a *appRuntime) analyzeJob(ctx context.Context, fullPath string) {
	data, err := os.ReadFile(fullPath)
	if err != nil {
		log.Printf("analyzer: arquivo indisponível %s: %v", fullPath, err)
		return
	}
	if !watcher.IsEligibleContent(data) {
		return
	}

	a.mu.Lock()
	root := a.cfg.WatchPath
	a.mu.Unlock()

	relPath, err := filepath.Rel(root, fullPath)
	if err != nil {
		relPath = fullPath
	}
	relPath = filepath.ToSlash(relPath)
	content := string(data)

	if a.activity != nil {
		a.activity.Publish("analyzing", relPath, "Analisando arquivo (regras determinísticas + LLM)...")
	}

	// Princípio IX: regras determinísticas rodam antes do LLM.
	findings := analyzer.RunDeterministic(relPath, content)

	if a.llama != nil && a.llama.IsReady() && a.llmClient != nil {
		llmFindings, err := a.llmClient.AnalyzeFile(ctx, relPath, content)
		if err != nil {
			log.Printf("analyzer: erro ao consultar LLM para %s: %v", relPath, err)
			if a.activity != nil {
				a.activity.Publish("error", relPath, "Erro ao consultar LLM: "+err.Error())
			}
		} else {
			findings = append(findings, llmFindings...)
		}
	}

	if len(findings) > 0 {
		a.notifySvc.ProcessFindings(ctx, findings)
		if a.activity != nil {
			a.activity.Publish("finding", relPath, fmt.Sprintf("%d achado(s) identificado(s).", len(findings)))
		}
	} else if a.activity != nil {
		a.activity.Publish("clean", relPath, "Análise concluída: nenhum problema relevante encontrado.")
	}
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	envCfg, err := analyzer.LoadEnvConfig(".env", ".env.development")
	if err != nil {
		log.Fatalf("erro de configuração: %v", err)
	}

	st := store.NewFalkorStore(envCfg.FalkorDBAddr, envCfg.FalkorDBUser, envCfg.FalkorDBPassword)
	defer st.Close()

	if err := st.Ping(ctx); err != nil {
		log.Printf("aviso: FalkorDB indisponível em %s (%v); usando configuração padrão em memória até reconectar", envCfg.FalkorDBAddr, err)
	}

	cfg, err := st.LoadConfig(ctx)
	if err != nil {
		log.Printf("aviso: falha ao carregar config do FalkorDB (%v); usando padrão", err)
		cfg = store.DefaultConfig()
	}

	llamaProc, err := analyzer.NewLlamaProcess(envCfg.LlamaPort)
	if err != nil {
		log.Fatalf("erro de porta do llama-server: %v", err)
	}
	llmClient, err := analyzer.NewLLMClient(envCfg.LlamaPort)
	if err != nil {
		log.Fatalf("erro ao criar cliente LLM: %v", err)
	}

	activityHub := web.NewActivityHub()
	rt := &appRuntime{
		cfg:       cfg,
		store:     st,
		llama:     llamaProc,
		llmClient: llmClient,
		notifySvc: notify.NewService(cfg, &notify.WindowsDispatcher{}, st),
		activity:  activityHub,
	}
	rt.queue = queue.New(time.Duration(cfg.DebounceMS)*time.Millisecond, 128, rt.analyzeJob)
	go rt.queue.Run(ctx)

	if envCfg.LlamaServerCmd != "" {
		go func() {
			log.Printf("llm: iniciando llama-server na porta %d...", envCfg.LlamaPort)
			if err := llamaProc.Start(ctx, envCfg.LlamaServerCmd, 60*time.Second); err != nil && !errors.Is(err, context.Canceled) {
				log.Printf("llm: falha ao iniciar llama-server: %v", err)
			} else if llamaProc.IsReady() {
				log.Printf("llm: llama-server pronto em 127.0.0.1:%d", envCfg.LlamaPort)
			}
		}()
	} else {
		log.Printf("llm: LLAMA_SERVER_CMD não definido; apenas regras determinísticas ativas")
	}

	if err := rt.applyConfig(ctx, cfg); err != nil {
		log.Printf("watcher: erro ao iniciar observação: %v", err)
	}

	srv := web.NewServer(envCfg.WebPort, st, rt, activityHub, func(updated store.MonitoringConfig) error {
		return rt.applyConfig(ctx, updated)
	})
	go func() {
		log.Printf("web: interface local em http://%s", srv.Addr())
		if err := srv.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("web: erro no servidor HTTP: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("encerrando genova-file-monitoring...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
	_ = llamaProc.Stop()
}
