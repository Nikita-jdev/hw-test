package main

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestCopy(t *testing.T) {
	tmpDir := "tmp"
	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name          string
		from, to      string
		offset, limit int64
	}{
		{
			name:   "copy full file",
			from:   "testdata/input.txt",
			to:     filepath.Join(tmpDir, "out_offset0_limit0.txt"),
			offset: 0,
			limit:  0,
		},
		{
			name:   "copy first 10 bytes",
			from:   "testdata/input.txt",
			to:     filepath.Join(tmpDir, "out_offset0_limit10.txt"),
			offset: 0,
			limit:  10,
		},
		{
			name:   "copy first 1000 bytes",
			from:   "testdata/input.txt",
			to:     filepath.Join(tmpDir, "out_offset0_limit1000.txt"),
			offset: 0,
			limit:  1000,
		},
		{
			name:   "copy all (limit > file size)",
			from:   "testdata/input.txt",
			to:     filepath.Join(tmpDir, "out_offset0_limit10000.txt"),
			offset: 0,
			limit:  10000,
		},
		{
			name:   "copy from offset 100",
			from:   "testdata/input.txt",
			to:     filepath.Join(tmpDir, "out_offset100_limit1000.txt"),
			offset: 100,
			limit:  1000,
		},
		{
			name:   "copy from offset 6000",
			from:   "testdata/input.txt",
			to:     filepath.Join(tmpDir, "out_offset6000_limit1000.txt"),
			offset: 6000,
			limit:  1000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Copy(tt.from, tt.to, tt.offset, tt.limit)
			if err != nil {
				t.Fatalf("Copy() error = %v", err)
			}

			verifyCopyResult(t, tt.from, tt.to, tt.offset, tt.limit)
		})
	}
}

func verifyCopyResult(t *testing.T, from, to string, offset, limit int64) {
	t.Helper()

	// Проверяем, что выходной файл создан
	info, err := os.Stat(to)
	if err != nil {
		t.Fatalf("Cannot stat output file: %v", err)
	}
	if info.Size() == 0 {
		t.Logf("Warning: output file is empty")
	}

	// Проверяем содержимое
	expected, err := os.Open(from)
	if err != nil {
		t.Fatalf("Cannot open source: %v", err)
	}
	defer expected.Close()

	if _, err := expected.Seek(offset, io.SeekStart); err != nil {
		t.Fatalf("Cannot seek in source: %v", err)
	}

	expectedData := readExpectedData(t, expected, limit)
	actualData := readActualData(t, to)

	if !bytes.Equal(expectedData, actualData) {
		t.Errorf("Content mismatch: expected %d bytes, got %d bytes",
			len(expectedData), len(actualData))
	}
}

func readExpectedData(t *testing.T, file *os.File, limit int64) []byte {
	t.Helper()

	if limit == 0 {
		data, err := io.ReadAll(file)
		if err != nil {
			t.Fatalf("Cannot read source: %v", err)
		}
		return data
	}

	data := make([]byte, limit)
	n, err := file.Read(data)
	if err != nil && !errors.Is(err, io.EOF) {
		t.Fatalf("Cannot read source: %v", err)
	}
	return data[:n]
}

func readActualData(t *testing.T, path string) []byte {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Cannot read output: %v", err)
	}
	return data
}

// Дополнительные тесты для проверки ошибок.
func TestCopyErrors(t *testing.T) {
	tests := []struct {
		name      string
		from, to  string
		offset    int64
		limit     int64
		wantError error
	}{
		{
			name:      "empty source path",
			from:      "",
			to:        "tmp/output.txt",
			offset:    0,
			limit:     0,
			wantError: ErrEmptyPath,
		},
		{
			name:      "empty destination path",
			from:      "testdata/input.txt",
			to:        "",
			offset:    0,
			limit:     0,
			wantError: ErrEmptyPath,
		},
		{
			name:      "offset exceeds file size",
			from:      "testdata/input.txt",
			to:        "tmp/output.txt",
			offset:    999999,
			limit:     0,
			wantError: ErrOffsetExceedsFileSize,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Copy(tt.from, tt.to, tt.offset, tt.limit)
			if !errors.Is(err, tt.wantError) {
				t.Errorf("Copy() error = %v, want %v", err, tt.wantError)
			}
		})
	}

	os.RemoveAll("tmp")
}

// Бенчмарк для проверки производительности.
func BenchmarkCopy(b *testing.B) {
	from := "testdata/input.txt"
	to := "tmp/bench_output.txt"

	// Создаем tmp директорию если её нет
	if err := os.MkdirAll("tmp", 0o755); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Copy(from, to, 0, 0)
		os.Remove(to)
	}
}
