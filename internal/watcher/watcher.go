package watcher

import (
	"context"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/fsnotify/fsnotify"
)

// Watcher observa alterações de escrita em uma pasta de projeto em modo somente leitura.
type Watcher struct {
	mu      sync.RWMutex
	filter  *Filter
	fsw     *fsnotify.Watcher
	onEvent func(filePath string)
}

// New cria um Watcher recursivo para root, invocando onEvent para arquivos elegíveis salvos.
func New(root string, extensions []string, onEvent func(filePath string)) (*Watcher, error) {
	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	w := &Watcher{
		filter:  NewFilter(root, extensions),
		fsw:     fsw,
		onEvent: onEvent,
	}
	if err := w.addRecursive(root); err != nil {
		_ = fsw.Close()
		return nil, err
	}
	return w, nil
}

func (w *Watcher) addRecursive(root string) error {
	return filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() {
			return nil
		}
		if w.filter.ShouldIgnoreDir(p) {
			return filepath.SkipDir
		}
		return w.fsw.Add(p)
	})
}

// Run inicia o loop de eventos até que ctx seja cancelado.
func (w *Watcher) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case ev, ok := <-w.fsw.Events:
			if !ok {
				return
			}
			w.handleEvent(ev)
		case err, ok := <-w.fsw.Errors:
			if !ok {
				return
			}
			log.Printf("watcher: erro do fsnotify: %v", err)
		}
	}
}

func (w *Watcher) handleEvent(ev fsnotify.Event) {
	if ev.Has(fsnotify.Create) {
		if info, err := os.Stat(ev.Name); err == nil && info.IsDir() {
			w.mu.RLock()
			ignored := w.filter.ShouldIgnoreDir(ev.Name)
			w.mu.RUnlock()
			if !ignored {
				_ = w.addRecursive(ev.Name)
			}
			return
		}
	}

	if !ev.Has(fsnotify.Write) && !ev.Has(fsnotify.Create) {
		return
	}

	w.mu.RLock()
	eligible := w.filter.ShouldAnalyzePath(ev.Name)
	w.mu.RUnlock()
	if eligible && w.onEvent != nil {
		w.onEvent(ev.Name)
	}
}

// Close encerra o fsnotify watcher.
func (w *Watcher) Close() error {
	return w.fsw.Close()
}
