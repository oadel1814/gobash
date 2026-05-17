package main

import (
	"io"
	"os"
	"strings"

	"github.com/chzyer/readline"
)

var rl *readline.Instance

type ShellCompleter struct{}

type Executable struct {
	Name string
	Path string
}

var executableCache map[string]string

func getExecutables() map[string]string {
	if executableCache != nil {
		return executableCache
	}

	executableCache = make(map[string]string)
	for _, dir := range strings.Split(os.Getenv("PATH"), ":") {
		files, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, file := range files {
			if info, err := file.Info(); err == nil && info.Mode().Perm()&0111 != 0 {
				executableCache[file.Name()] = dir + "/" + file.Name()
			}
		}
	}

	return executableCache
}

func (sc *ShellCompleter) Do(line []rune, pos int) ([][]rune, int) {
	prefix := string(line[:pos])
	var matches [][]rune
	for name := range builtins {
		if strings.HasPrefix(name, prefix) {
			matches = append(matches, []rune(name[len(prefix):]+" "))
		}
	}

	if len(matches) == 0 {
		for name := range getExecutables() {
			if strings.HasPrefix(name, prefix) {
				matches = append(matches, []rune(name[len(prefix):]+" "))
			}
		}
	}

	if len(matches) == 0 {
		io.WriteString(rl.Stdout(), "\a")
		return nil, 0
	}
	return matches, len(prefix)
}

func initReadline() {
	var err error
	rl, err = readline.NewEx(&readline.Config{
		Prompt:       "$ ",
		AutoComplete: &ShellCompleter{},
	})
	if err != nil {
		panic(err)
	}
}

func prompt() string {
	line, err := rl.Readline()
	if err != nil {
		os.Exit(0)
	}
	return strings.TrimSpace(line)
}
