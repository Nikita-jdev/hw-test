package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
)

type Environment map[string]EnvValue

// EnvValue helps to distinguish between empty files and files with the first empty line.
type EnvValue struct {
	Value      string
	NeedRemove bool
}

// ReadDir reads a specified directory and returns map of env variables.
// Variables represented as files where filename is name of variable, file first line is a value.
func ReadDir(dir string) (Environment, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	env := make(Environment, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || strings.Contains(entry.Name(), "=") {
			continue
		}

		data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			return nil, err
		}

		if len(data) == 0 {
			env[entry.Name()] = EnvValue{NeedRemove: true}
			continue
		}

		// Берём только первую строку файла.
		if i := bytes.IndexByte(data, '\n'); i >= 0 {
			data = data[:i]
		}

		value := strings.TrimRight(string(data), " \t")
		value = strings.ReplaceAll(value, "\x00", "\n")
		env[entry.Name()] = EnvValue{Value: value}
	}
	return env, nil
}
