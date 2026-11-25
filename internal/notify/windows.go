package notify

import (
	"fmt"

	"github.com/gen2brain/beeep"
	"github.com/wesleyxavier/genova-file-monitoring/internal/store"
)

// Dispatcher abstrai o envio da notificação nativa no sistema operacional.
type Dispatcher interface {
	Send(f store.Finding) error
}

// WindowsDispatcher usa beeep para disparar toasts no Windows 10/11.
type WindowsDispatcher struct{}

// Send exibe Alert (com som) para severidade Crítica e Notify para Alta/Média.
func (d *WindowsDispatcher) Send(f store.Finding) error {
	title := fmt.Sprintf("[%s] %s", f.Severity, f.RuleID)
	body := fmt.Sprintf("%s (L%d): %s", f.FilePath, f.LineStart, f.Message)
	if f.Severity == store.SeverityCritica {
		return beeep.Alert(title, body, "")
	}
	return beeep.Notify(title, body, "")
}
