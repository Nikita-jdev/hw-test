package main

import (
	"os"
	"path/filepath"
	"testing"
)

// makeEnvDir создаёт временную директорию с переданными файлами
// и гарантирует её удаление после теста.
func makeEnvDir(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func TestReadDir(t *testing.T) {
	dir := makeEnvDir(t, map[string]string{
		"FOO":        "123",
		"BAR":        "value",
		"EMPTY_FILE": "",
		"FIRST_LINE": "line1\nline2",
		"TRIM":       "  spaced  \t\t ",
		"BIN":        "a\x00b",
	})

	env, err := ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir error: %v", err)
	}

	tests := []struct {
		name    string
		want    string
		needRem bool
	}{
		{"FOO", "123", false},
		{"BAR", "value", false},
		{"EMPTY_FILE", "", true},
		{"FIRST_LINE", "line1", false},
		{"TRIM", "  spaced", false},
		{"BIN", "a\nb", false},
	}

	for _, tc := range tests {
		got, ok := env[tc.name]
		if !ok {
			t.Errorf("ReadDir: переменная %s не найдена", tc.name)
			continue
		}
		if got.NeedRemove != tc.needRem {
			t.Errorf("ReadDir(%s): NeedRemove = %v, want %v", tc.name, got.NeedRemove, tc.needRem)
		}
		if got.Value != tc.want {
			t.Errorf("ReadDir(%s): Value = %q, want %q", tc.name, got.Value, tc.want)
		}
	}
}

func TestReadDirError(t *testing.T) {
	if _, err := ReadDir(filepath.Join(t.TempDir(), "no_such_dir")); err == nil {
		t.Error("ReadDir: ожидалась ошибка для несуществующей директории")
	}
}
