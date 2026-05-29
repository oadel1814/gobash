package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
)

type HandlerFunc func(cmd Command) error

var builtins map[string]HandlerFunc

func init() {
	builtins = map[string]HandlerFunc{
		"echo":     handleEcho,
		"cd":       handleCd,
		"pwd":      handlePwd,
		"type":     handleType,
		"exit":     handleExit,
		"complete": handleComplete,
		"jobs":     handleJobs,
		"history":  handleHistory,
		"help":     handleHelp,
		"declare":  handleDeclare,
	}
}

func validateEnvVar(name string) error {
	if name == "" {
		return errors.New("variable name cannot be empty")
	}

	if name[0] >= '0' && name[0] <= '9' || name[0] == '-' || strings.Contains(name, "=") || strings.Contains(name, "-") {
		return errors.New("variable name cannot start with a digit or hyphen")
	}

	return nil
}

func handleDeclare(cmd Command) error {
	if len(cmd.Args) == 0 {
		return nil
	}

	if cmd.Args[0] == "-p" {
		if len(cmd.Args) < 2 {
			return errors.New("declare: usage: declare -p <var>")
		}

		varName := cmd.Args[1]
		value, exists := os.LookupEnv(varName)
		if !exists {
			return fmt.Errorf("declare: %s: not found", varName)
		}
		fmt.Printf("declare -- %s=%q\n", varName, value)
		return nil
	}

	// handle simple NAME=VALUE case
	args := strings.SplitN(cmd.Args[0], "=", 2)
	if len(args) != 2 {
		return errors.New("declare: usage: declare NAME=VALUE")
	}

	name := args[0]
	value := args[1]

	if err := validateEnvVar(name); err != nil {
		return fmt.Errorf("declare: `%s=%s': not a valid identifier", name, value)
	}

	os.Setenv(name, value)

	return nil
}

func handleHelp(cmd Command) error {
	const (
		reset  = "\033[0m"
		bold   = "\033[1m"
		dim    = "\033[2m"
		cyan   = "\033[96m"
		white  = "\033[97m"
		gray   = "\033[90m"
		yellow = "\033[33m"
		green  = "\033[32m"
	)

	type entry struct {
		cmd, args, desc string
	}

	sections := []struct {
		title   string
		entries []entry
	}{
		{
			"Navigation",
			[]entry{
				{"cd", "[dir]", "Change directory. cd ~ goes home, cd - returns to previous"},
				{"pwd", "", "Print current working directory"},
				{"ls", "[dir]", "List directory contents (delegates to system ls)"},
			},
		},
		{
			"Jobs & Processes",
			[]entry{
				{"jobs", "", "List background jobs with their status and PID"},
			},
		},
		{
			"Shell",
			[]entry{
				{"help", "[command]", "Show this screen, or detail for a specific command"},
				{"exit", "[code]", "Exit the shell with optional status code"},
				{"echo", "[args]", "Print arguments to stdout"},
				{"history", "", "Show command history"},
			},
		},
	}

	// If a specific command was requested, show its detail
	if len(cmd.Args) > 0 {
		target := cmd.Args[0]
		for _, sec := range sections {
			for _, e := range sec.entries {
				if strings.Fields(e.cmd)[0] == target {
					fmt.Printf("\n %s%s%s  %s%s%s\n",
						bold+white, e.cmd, reset,
						dim+gray, e.args, reset,
					)
					fmt.Printf(" %s%s%s\n\n", gray, e.desc, reset)
					return nil
				}
			}
		}
		fmt.Printf(" %sNo help entry for %q%s\n\n", gray, target, reset)
		return nil
	}

	// Full help screen
	fmt.Printf("\n %s%sGobash Help%s  %s%s%s\n\n",
		bold, white, reset,
		dim+gray, version, reset,
	)

	for _, sec := range sections {
		fmt.Printf(" %s%s%s\n", yellow, sec.title, reset)
		fmt.Printf(" %s%s%s\n", dim+gray, strings.Repeat("─", 52), reset)

		for _, e := range sec.entries {
			col := fmt.Sprintf("%s%-16s%s%s%-10s%s",
				cyan, e.cmd, reset,
				dim+gray, e.args, reset,
			)
			fmt.Printf("   %s  %s%s%s\n", col, gray, e.desc, reset)
		}
		fmt.Println()
	}

	fmt.Printf(" %s✦ Tip:%s %shelp <command>%s for more detail on any command.\n\n",
		yellow, reset, cyan, reset,
	)

	return nil
}

func resolveStdout(cmd Command) (*os.File, error) {
	if cmd.StdoutOverride != nil {
		return cmd.StdoutOverride, nil
	}
	if cmd.Stdout == "" {
		return os.Stdout, nil
	}
	flags := os.O_WRONLY | os.O_CREATE
	if cmd.Append {
		flags |= os.O_APPEND
	} else {
		flags |= os.O_TRUNC
	}
	return os.OpenFile(cmd.Stdout, flags, 0644)
}

func resolveStderr(cmd Command) (*os.File, error) {
	if cmd.Stderr == "" {
		return os.Stderr, nil
	}
	flags := os.O_WRONLY | os.O_CREATE
	if cmd.Append {
		flags |= os.O_APPEND
	} else {
		flags |= os.O_TRUNC
	}
	return os.OpenFile(cmd.Stderr, flags, 0644)
}

var completions = map[string]string{}

var mostRecentJob int
var secondMostRecentJob int

func recomputeMarkers() {
	currentIds := make([]int, 0, len(backgroundJobs))
	for k := range backgroundJobs {
		currentIds = append(currentIds, k)
	}
	sort.Ints(currentIds)

	switch len(currentIds) {
	case 0:
		mostRecentJob, secondMostRecentJob = 0, 0
	case 1:
		mostRecentJob, secondMostRecentJob = currentIds[0], 0
	default:
		mostRecentJob = currentIds[len(currentIds)-1]
		secondMostRecentJob = currentIds[len(currentIds)-2]
	}
}

func reapJobs() {
	ids := make([]int, 0, len(backgroundJobs))
	for id := range backgroundJobs {
		ids = append(ids, id)
	}
	sort.Ints(ids)

	var reaped []int

	for _, id := range ids {
		job := backgroundJobs[id]

		if !job.Done {
			continue
		}

		marker := " "
		switch id {
		case mostRecentJob:
			marker = "+"
		case secondMostRecentJob:
			marker = "-"
		}

		fmt.Printf("[%d]%s  %-24s%s\n", id, marker, "Done", job.Args)
		reaped = append(reaped, id)
	}

	for _, id := range reaped {
		delete(backgroundJobs, id)
	}
	if len(reaped) > 0 {
		recomputeMarkers()
	}
}

func handleJobs(cmd Command) error {
	ids := make([]int, 0, len(backgroundJobs))
	for id := range backgroundJobs {
		ids = append(ids, id)
	}
	sort.Ints(ids)

	var reaped []int

	for _, id := range ids {
		job := backgroundJobs[id]

		marker := " "
		switch id {
		case mostRecentJob:
			marker = "+"
		case secondMostRecentJob:
			marker = "-"
		}

		if job.Done {
			fmt.Printf("[%d]%s  %-24s%s\n", id, marker, "Done", job.Args)
			reaped = append(reaped, id)
		} else {
			fmt.Printf("[%d]%s  %-24s%s\n", id, marker, "Running", job.Args)
		}
	}

	for _, id := range reaped {
		delete(backgroundJobs, id)
	}
	if len(reaped) > 0 {
		recomputeMarkers()
	}

	return nil
}

var history []Command

var historyLastAppended int

func loadHistory(filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		tokens := tokenize(line)
		if len(tokens) > 0 {
			history = append(history, Command{
				Name: tokens[0],
				Args: tokens[1:],
			})
		}
	}
	historyLastAppended = len(history)
	return nil
}

func writeHistory(filePath string) error {
	var builder strings.Builder
	for _, cmd := range history {
		builder.WriteString(cmd.Name)
		if len(cmd.Args) > 0 {
			builder.WriteString(" " + strings.Join(cmd.Args, " "))
		}
		builder.WriteString("\n")
	}

	return os.WriteFile(filePath, []byte(builder.String()), 0644)
}

func appendHistory(filePath string) error {
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("history: cannot open file for appending: %w", err)
	}
	defer f.Close()

	for i := historyLastAppended; i < len(history); i++ {
		line := history[i].Name
		if len(history[i].Args) > 0 {
			line += " " + strings.Join(history[i].Args, " ")
		}
		line += "\n"
		if _, err := f.WriteString(line); err != nil {
			return fmt.Errorf("history: cannot write to file: %w", err)
		}
	}
	historyLastAppended = len(history)
	return nil
}

func handleHistory(cmd Command) error {
	if len(cmd.Args) >= 2 {
		switch cmd.Args[0] {
		case "-r":
			filePath := cmd.Args[1]
			return loadHistory(filePath)
		case "-w":
			filePath := cmd.Args[1]
			return writeHistory(filePath)
		case "-a":
			filePath := cmd.Args[1]
			return appendHistory(filePath)
		}
		return nil
	}

	n := len(history)
	if len(cmd.Args) > 0 {
		var err error
		n, err = strconv.Atoi(cmd.Args[0])
		if err != nil {
			return errors.New("history: argument must be a number")
		}
		if n < 0 {
			return errors.New("history: argument must be non-negative")
		}
		if n > len(history) {
			n = len(history)
		}
	}

	stdout, err := resolveStdout(cmd)
	if err != nil {
		return err
	}
	if stdout != os.Stdout && cmd.StdoutOverride == nil {
		defer stdout.Close()
	}

	for i := len(history) - n; i < len(history); i++ {
		cmdString := history[i].Name
		if len(history[i].Args) > 0 {
			cmdString += " " + strings.Join(history[i].Args, " ")
		}
		fmt.Fprintf(stdout, "%5d  %s\n", i+1, cmdString)
	}

	return nil
}

func handleComplete(cmd Command) error {
	if len(cmd.Args) == 0 {
		return nil
	}

	flag := cmd.Args[0]

	switch flag {
	case "-p":
		if len(cmd.Args) < 2 {
			return errors.New("complete: usage: complete -p <command>")
		}

		command := cmd.Args[1]
		path, ok := completions[command]
		if !ok {
			return errors.New("complete: " + command + ": no completion specification")
		}
		fmt.Printf("complete -C '%s' %s\n", path, command)

	case "-C":
		if len(cmd.Args) < 3 {
			return errors.New("complete: usage: complete -C <path> <command>")
		}
		path := cmd.Args[1]
		command := cmd.Args[2]
		completions[command] = path
	case "-r":
		if len(cmd.Args) < 2 {
			return errors.New("complete: usage: complete -r <command>")
		}
		command := cmd.Args[1]
		delete(completions, command)
	}

	return nil
}

func handleEcho(cmd Command) error {
	if cmd.Name == "" {
		return nil
	}
	output := strings.Join(cmd.Args, " ")
	stdout, err := resolveStdout(cmd)
	if err != nil {
		return err
	}
	if stdout != os.Stdout && cmd.StdoutOverride == nil {
		defer stdout.Close()
	}
	_, err = io.WriteString(stdout, output+"\n")

	stderr, err2 := resolveStderr(cmd)
	if err2 != nil {
		return err2
	}
	if stderr != os.Stderr {
		defer stderr.Close()
	}
	if err != nil {
		_, _ = io.WriteString(stderr, err.Error()+"\n")
		return err
	}
	return nil
}

func handleCd(cmd Command) error {
	if cmd.Name == "" {
		return nil
	}
	if len(cmd.Args) == 0 || cmd.Args[0] == "~" {
		return os.Chdir(os.Getenv("HOME"))
	}
	if err := os.Chdir(cmd.Args[0]); err != nil {
		return fmt.Errorf("cd: %s: No such file or directory", cmd.Args[0])
	}
	return nil
}

func handlePwd(cmd Command) error {
	if cmd.Name == "" {
		return nil
	}
	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	fmt.Println(dir)
	return nil
}

func handleType(cmd Command) error {
	if cmd.Name == "" {
		return nil
	}

	stdout, err := resolveStdout(cmd)
	if err != nil {
		return err
	}

	if stdout != os.Stdout && cmd.StdoutOverride == nil {
		defer stdout.Close()
	}

	if _, ok := builtins[cmd.Args[0]]; ok {
		fmt.Printf("%s is a shell builtin\n", cmd.Args[0])
		return nil
	}
	if ok, path := isExecutable(cmd.Args[0]); ok {
		fmt.Printf("%s is %s\n", cmd.Args[0], path)
		return nil
	}
	fmt.Printf("%s not found\n", cmd.Args[0])
	return nil
}

func handleExit(cmd Command) error {
	appendHistory(os.Getenv("HISTFILE"))
	os.Exit(0)
	return nil
}
