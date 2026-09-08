package main

import (
	"errors"
	"os"
	"os/exec"
	"strings"
)

// RunCmd runs a command + arguments (cmd) with environment variables from env.
func RunCmd(cmd []string, env Environment) (returnCode int) {
	// Команда и её аргументы приходят из командной строки и запускаются намеренно.
	command := exec.Command(cmd[0], cmd[1:]...) //nolint:gosec

	// Переносим окружение, заменяя/удаляя переменные из env.
	for _, kv := range os.Environ() {
		name := strings.SplitN(kv, "=", 2)[0]
		if _, ok := env[name]; ok {
			continue
		}
		command.Env = append(command.Env, kv)
	}
	for name, val := range env {
		if val.NeedRemove {
			continue
		}
		command.Env = append(command.Env, name+"="+val.Value)
	}

	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	if err := command.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return exitErr.ExitCode()
		}
		return 1
	}
	return 0
}
