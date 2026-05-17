package main

import (
	"io"
	"os"
	"sort"
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

var lastPrefix string
var tabPressedOnce bool

func (sc *ShellCompleter) Do(line []rune, pos int) ([][]rune, int) {
	prefix := string(line[:pos])

	matchSet := make(map[string]struct{})

	for name := range builtins {
		if strings.HasPrefix(name, prefix) {
			matchSet[name] = struct{}{}
		}
	}

	for name := range getExecutables() {
		if strings.HasPrefix(name, prefix) {
			matchSet[name] = struct{}{}
		}
	}

	var names []string
	for name := range matchSet {
		names = append(names, name)
	}

	if len(names) == 0 {
		io.WriteString(rl.Stdout(), "\a")
		tabPressedOnce = false
		lastPrefix = ""
		return nil, 0
	}

	sort.Strings(names)

	if len(names) == 1 {
		tabPressedOnce = false
		lastPrefix = ""
		return [][]rune{[]rune(names[0][len(prefix):] + " ")}, 0
	}

	if !tabPressedOnce || prefix != lastPrefix {
		tabPressedOnce = true
		lastPrefix = prefix
		io.WriteString(rl.Stdout(), "\a")
		return nil, 0
	}

	tabPressedOnce = false
	lastPrefix = ""

	output := "\r\n" + strings.Join(names, "  ") + "\r\n" + rl.Config.Prompt + prefix
	os.Stdout.WriteString(output)

	return nil, 0
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
