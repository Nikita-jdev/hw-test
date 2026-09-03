package main

import (
	"os"
	"testing"
)

// TestHelperProcess запускается как дочерний процесс и проверяет,
// что переменные окружения применены корректно.
func TestHelperProcess(_ *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	foo := os.Getenv("FOO")
	bar := os.Getenv("BAR")
	removed := os.Getenv("REMOVE_ME")
	keep := os.Getenv("KEEP_ME")

	if foo != "new" || bar != "value" || removed != "" || keep != "original" {
		os.Exit(1)
	}
	os.Exit(0)
}

// TestRunCmdEnv проверяет, что переменные из env переопределяют
// существующие, задаются новые, а помеченные удаляются.
func TestRunCmdEnv(t *testing.T) {
	t.Setenv("FOO", "old")
	t.Setenv("REMOVE_ME", "x")
	t.Setenv("KEEP_ME", "original")
	t.Setenv("GO_WANT_HELPER_PROCESS", "1")

	env := Environment{
		"FOO":       {Value: "new"},
		"BAR":       {Value: "value"},
		"REMOVE_ME": {NeedRemove: true},
	}

	code := RunCmd([]string{os.Args[0], "-test.run=TestHelperProcess"}, env)
	if code != 0 {
		t.Errorf("RunCmd code = %d, want 0 (переменные окружения применены неверно)", code)
	}
}

// TestRunCmdExitCode проверяет, что код выхода совпадает с кодом команды.
func TestRunCmdExitCode(t *testing.T) {
	t.Setenv("GO_WANT_HELPER_PROCESS", "1")
	if code := RunCmd([]string{os.Args[0], "-test.run=TestHelperExit"}, Environment{}); code != 7 {
		t.Errorf("RunCmd code = %d, want 7", code)
	}
}

// TestHelperExit всегда завершается с кодом 7.
func TestHelperExit(_ *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	os.Exit(7)
}
