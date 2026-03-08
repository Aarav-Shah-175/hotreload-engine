package watcher

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
)

type Event struct {
	Path string
	Op   fsnotify.Op
}

type Watcher struct {
	root string
	fsw  *fsnotify.Watcher
}

func New(root string) (*Watcher, error) {
	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("create fsnotify watcher: %w", err)
	}

	w := &Watcher{root: root, fsw: fsw}
	if err := w.addRecursive(root); err != nil {
		_ = fsw.Close()
		return nil, err
	}
	return w, nil
}

func (w *Watcher) Start(ctx context.Context) (<-chan Event, <-chan error) {
	events := make(chan Event, 128)
	errs := make(chan error, 16)

	go func() {
		defer close(events)
		defer close(errs)
		defer w.fsw.Close()

		for {
			select {
			case <-ctx.Done():
				return
			case err, ok := <-w.fsw.Errors:
				if !ok {
					return
				}
				errs <- err
			case raw, ok := <-w.fsw.Events:
				if !ok {
					return
				}

				cleanPath := filepath.Clean(raw.Name)
				if ShouldIgnore(cleanPath, w.root) {
					continue
				}

				isDir := false
				if stat, err := os.Stat(cleanPath); err == nil {
					isDir = stat.IsDir()
				}

				if raw.Has(fsnotify.Create) && isDir {
					if err := w.addRecursive(cleanPath); err != nil && !errors.Is(err, fs.ErrNotExist) {
						errs <- err
					}
					continue
				}

				if shouldForwardEvent(raw.Op, isDir) {
					events <- Event{Path: cleanPath, Op: raw.Op}
				}
			}
		}
	}()

	return events, errs
}

func shouldForwardEvent(op fsnotify.Op, isDir bool) bool {
	if isDir {
		return false
	}
	return op.Has(fsnotify.Write) || op.Has(fsnotify.Create) || op.Has(fsnotify.Remove) || op.Has(fsnotify.Rename)
}

func (w *Watcher) addRecursive(root string) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				return nil
			}
			return err
		}
		if !d.IsDir() {
			return nil
		}
		if ShouldIgnore(path, w.root) {
			if path == w.root {
				return nil
			}
			return filepath.SkipDir
		}
		if err := w.fsw.Add(path); err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				return nil
			}
			return fmt.Errorf("watch %q: %w", path, err)
		}
		return nil
	})
}

func ShouldIgnore(path, root string) bool {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return true
	}
	if rel == "." {
		return false
	}

	parts := strings.Split(rel, string(filepath.Separator))
	for _, part := range parts {
		if part == ".git" || part == "node_modules" || part == "bin" || part == "build" || part == "dist" || part == "tmp" {
			return true
		}
	}

	base := filepath.Base(path)
	if strings.HasSuffix(base, ".log") {
		return true
	}

	if strings.HasPrefix(base, ".#") || strings.HasPrefix(base, "#") || strings.HasSuffix(base, "#") || strings.HasSuffix(base, "~") {
		return true
	}

	if strings.HasSuffix(base, ".swp") || strings.HasSuffix(base, ".swo") || strings.HasSuffix(base, ".swx") || strings.HasSuffix(base, ".tmp") || strings.HasSuffix(base, ".temp") {
		return true
	}

	return false
}
