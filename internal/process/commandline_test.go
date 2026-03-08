package process

import (
	"reflect"
	"testing"
)

func TestSplitCommandLine(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []string
		wantErr bool
	}{
		{name: "simple", input: ".\\bin\\server.exe", want: []string{".\\bin\\server.exe"}},
		{name: "with args", input: ".\\bin\\server.exe -port 8080", want: []string{".\\bin\\server.exe", "-port", "8080"}},
		{name: "quoted path", input: "\"C:\\Program Files\\server.exe\" --debug", want: []string{"C:\\Program Files\\server.exe", "--debug"}},
		{name: "single quoted", input: "'./bin/server' --flag", want: []string{"./bin/server", "--flag"}},
		{name: "unterminated quote", input: "\"./bin/server", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := splitCommandLine(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("splitCommandLine() = %#v, want %#v", got, tt.want)
			}
		})
	}
}
