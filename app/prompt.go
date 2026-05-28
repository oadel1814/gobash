package main

import (
	"io"
	"log"
	"os"
	"os/exec"
	"sort"
	"strconv"
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

func longestCommonPrefix(strs []string) string {
	if len(strs) == 0 {
		return ""
	}

	if len(strs) == 1 {
		return strs[0]
	}

	minLen := len(strs[0])

	for i := range strs {
		minLen = min(minLen, len(strs[i]))
	}

	for i := 0; i < minLen; i++ {
		ch := strs[0][i]
		for _, s := range strs[1:] {
			if s[i] != ch {
				return strs[0][:i]
			}
		}
	}

	return strs[0][:minLen]
}

/*

    cd di--<TAB>r1/
	cd dir1/dir2/

*/

func (sc *ShellCompleter) Do(line []rune, pos int) ([][]rune, int) {
	fullLine := string(line[:pos])

	words := strings.Fields(fullLine)

	var prefix string
	if len(words) == 0 {
		return nil, 0
	}

	if fullLine[len(fullLine)-1] == ' ' {
		prefix = ""
	} else {
		prefix = words[len(words)-1]
	}

	isFirstWord := len(words) == 1 && prefix != ""

	if !isFirstWord {
		commandName := words[0]
		if completerPath, ok := completions[commandName]; ok {
			var prevWord string
			if prefix == "" {
				prevWord = words[len(words)-1]
			} else if len(words) >= 2 {
				prevWord = words[len(words)-2]
			}
			candidates := runCompletionScript(completerPath, commandName, prefix, prevWord, fullLine)

			if len(candidates) == 0 {
				io.WriteString(rl.Stdout(), "\a")
				return nil, 0
			}

			if len(candidates) == 1 {
				completion := candidates[0][len(prefix):]
				return [][]rune{[]rune(completion + " ")}, 0
			}

			lcp := longestCommonPrefix(candidates)
			if len(lcp) > len(prefix) {
				completion := lcp[len(prefix):]
				return [][]rune{[]rune(completion)}, 0
			}

			sort.Strings(candidates)
			if !tabPressedOnce || prefix != lastPrefix {
				tabPressedOnce = true
				lastPrefix = prefix
				io.WriteString(rl.Stdout(), "\a")
				return nil, 0
			}

			tabPressedOnce = false
			lastPrefix = ""
			output := "\r\n" + strings.Join(candidates, "  ") + "\r\n" + rl.Config.Prompt + fullLine
			os.Stdout.WriteString(output)
			return nil, 0
		}
	}

	matchSet := make(map[string]struct{})

	var dir string

	if isFirstWord {
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
	} else {

		if strings.Contains(prefix, "/") {
			dir = prefix[:strings.LastIndex(prefix, "/")]
			prefix = prefix[strings.LastIndex(prefix, "/")+1:]
		} else {
			dir = "."
		}

		files, err := os.ReadDir(dir)
		if err != nil {
			log.Fatal(err)
		}
		for _, file := range files {
			if strings.HasPrefix(file.Name(), prefix) {
				matchSet[file.Name()] = struct{}{}
			}
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

		suffix := " "

		var path string
		if dir != "" {
			path = dir + "/" + names[0]
		}
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			suffix = "/"
		}

		return [][]rune{[]rune(names[0][len(prefix):] + suffix)}, 0
	}

	lcp := longestCommonPrefix(names)
	if len(lcp) > len(prefix) {
		tabPressedOnce = false
		lastPrefix = ""
		return [][]rune{[]rune(lcp[len(prefix):])}, 0
	}

	if !tabPressedOnce || prefix != lastPrefix {
		tabPressedOnce = true
		lastPrefix = prefix
		io.WriteString(rl.Stdout(), "\a")
		return nil, 0
	}

	tabPressedOnce = false
	lastPrefix = ""

	for i, name := range names {
		suffix := " "

		var path string
		if dir != "" {
			path = dir + "/" + name
		}
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			suffix = "/"
		}

		names[i] = name + suffix
	}

	output := "\r\n" + strings.Join(names, "  ") + "\r\n" + rl.Config.Prompt + fullLine
	os.Stdout.WriteString(output)

	return nil, 0
}

func runCompletionScript(path, commandName, prefix, prevWord, fullLine string) []string {
	cmd := exec.Command(path, commandName, prefix, prevWord)
	cmd.Env = append(os.Environ(),
		"COMP_LINE="+fullLine,
		"COMP_POINT="+strconv.Itoa(len(fullLine)),
	)
	out, err := cmd.Output()
	if err != nil {
		return nil
	}

	var candidates []string
	for _, line := range strings.Split(strings.TrimRight(string(out), "\r\n"), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			candidates = append(candidates, line)
		}
	}
	return candidates
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
	reapJobs()

	line, err := rl.Readline()
	if err != nil {
		os.Exit(0)
	}
	return strings.TrimSpace(line)
}
