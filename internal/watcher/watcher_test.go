package watcher

import (
	"path/filepath"
	"testing"

	"github.com/fsnotify/fsnotify"
)

func TestShouldIgnore(t *testing.T) {
	root := filepath.Join("tmp", "project")

	tests := []struct {
		name string
		path string
		want bool
	}{
		{name: "root", path: root, want: false},
		{name: "go file", path: filepath.Join(root, "cmd", "main.go"), want: false},
		{name: "git", path: filepath.Join(root, ".git", "config"), want: true},
		{name: "node modules", path: filepath.Join(root, "node_modules", "x.js"), want: true},
		{name: "bin", path: filepath.Join(root, "bin", "server"), want: true},
		{name: "build", path: filepath.Join(root, "build", "server"), want: true},
		{name: "dist", path: filepath.Join(root, "dist", "asset.js"), want: true},
		{name: "tmp dir", path: filepath.Join(root, "tmp", "cache.txt"), want: true},
		{name: "log file", path: filepath.Join(root, "app.log"), want: true},
		{name: "editor swap", path: filepath.Join(root, "main.go.swp"), want: true},
		{name: "editor swap swo", path: filepath.Join(root, "main.go.swo"), want: true},
		{name: "editor backup", path: filepath.Join(root, "main.go~"), want: true},
		{name: "editor temp", path: filepath.Join(root, ".#main.go"), want: true},
		{name: "editor hash temp", path: filepath.Join(root, "#main.go#"), want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ShouldIgnore(tt.path, root); got != tt.want {
				t.Fatalf("ShouldIgnore(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestShouldForwardEvent(t *testing.T) {
	tests := []struct {
		name  string
		op    fsnotify.Op
		isDir bool
		want  bool
	}{
		{name: "write file", op: fsnotify.Write, isDir: false, want: true},
		{name: "create file", op: fsnotify.Create, isDir: false, want: true},
		{name: "rename file", op: fsnotify.Rename, isDir: false, want: true},
		{name: "remove file", op: fsnotify.Remove, isDir: false, want: true},
		{name: "chmod ignored", op: fsnotify.Chmod, isDir: false, want: false},
		{name: "create dir", op: fsnotify.Create, isDir: true, want: false},
		{name: "write dir", op: fsnotify.Write, isDir: true, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldForwardEvent(tt.op, tt.isDir); got != tt.want {
				t.Fatalf("shouldForwardEvent(%v, %v) = %v, want %v", tt.op, tt.isDir, got, tt.want)
			}
		})
	}
}
