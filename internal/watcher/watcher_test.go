package watcher

import (
	"path/filepath"
	"testing"
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
		{name: "build bin", path: filepath.Join(root, "bin", "server"), want: true},
		{name: "editor swap", path: filepath.Join(root, "main.go.swp"), want: true},
		{name: "editor backup", path: filepath.Join(root, "main.go~"), want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ShouldIgnore(tt.path, root); got != tt.want {
				t.Fatalf("ShouldIgnore(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}
