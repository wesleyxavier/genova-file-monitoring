package notify

import (
	"context"
	"log"
	"sync"

	"github.com/wesleyxavier/genova-file-monitoring/internal/store"
)

// Service coordena avaliação de ruído, envio de notificação Windows e persistência no FalkorDB.
type Service struct {
	mu         sync.RWMutex
	cfg        store.MonitoringConfig
	policy     *Policy
	dispatcher Dispatcher
	store      store.Store
}

// NewService cria o serviço de notificações e histórico.
func NewService(cfg store.MonitoringConfig, dispatcher Dispatcher, st store.Store) *Service {
	return &Service{
		cfg:        cfg,
		policy:     NewPolicy(),
		dispatcher: dispatcher,
		store:      st,
	}
}

// UpdateConfig atualiza a configuração ativa (severidade mínima e regras silenciadas).
func (s *Service) UpdateConfig(cfg store.MonitoringConfig) {
	s.mu.Lock()
	s.cfg = cfg
	s.mu.Unlock()
}

// ProcessFindings avalia cada achado, dispara toast quando elegível e persiste todos no store.
func (s *Service) ProcessFindings(ctx context.Context, findings []store.Finding) {
	s.mu.RLock()
	cfg := s.cfg
	s.mu.RUnlock()

	for _, f := range findings {
		notified, reason := s.policy.Evaluate(f, cfg)
		f.Notified = notified
		f.SuppressionReason = reason

		if notified && s.dispatcher != nil {
			if err := s.dispatcher.Send(f); err != nil {
				log.Printf("notify: falha ao exibir notificação para %s (%s): %v", f.FilePath, f.RuleID, err)
			}
		}

		if s.store != nil {
			if err := s.store.SaveFinding(ctx, f); err != nil {
				log.Printf("notify: falha ao persistir achado %s (%s) no FalkorDB: %v", f.FilePath, f.RuleID, err)
			}
		}
	}
}
