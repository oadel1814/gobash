package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"syscall"
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
	}
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
		proc := backgroundJobs[id]

		var ws syscall.WaitStatus
		pid, err := syscall.Wait4(proc.Process.Pid, &ws, syscall.WNOHANG, nil)

		if pid == proc.Process.Pid || err != nil {
			var marker string
			switch id {
			case mostRecentJob:
				marker = "+"
			case secondMostRecentJob:
				marker = "-"
			default:
				marker = " "
			}

			processString := strings.Join(proc.Args, " ")
			fmt.Printf("[%d]%s  %-24s%s\n", id, marker, "Done", processString)
			reaped = append(reaped, id)
		}
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
		proc := backgroundJobs[id]

		var ws syscall.WaitStatus
		pid, err := syscall.Wait4(proc.Process.Pid, &ws, syscall.WNOHANG, nil)

		var marker string
		switch id {
		case mostRecentJob:
			marker = "+"
		case secondMostRecentJob:
			marker = "-"
		default:
			marker = " "
		}

		processString := strings.Join(proc.Args, " ")

		if pid == proc.Process.Pid || err != nil {
			fmt.Printf("[%d]%s  %-24s%s\n", id, marker, "Done", processString)
			reaped = append(reaped, id)
		} else {
			fmt.Printf("[%d]%s  %-24s%s &\n", id, marker, "Running", processString)
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
	os.Exit(0)
	return nil
}
