package main

import (
	"fmt"
	"os"
	"os/exec"
)

var backgroundCounter int = 1
var backgroundJobs = make(map[int]*exec.Cmd)

func isExecutable(name string) (bool, string) {
	path, ok := getExecutables()[name]
	return ok, path
}

func executeExternal(cmd Command) error {
	stdout, err := resolveStdout(cmd)
	if err != nil {
		return err
	}
	if stdout != os.Stdout {
		defer stdout.Close()
	}

	stderr, err := resolveStderr(cmd)
	if err != nil {
		return err
	}
	if stderr != os.Stderr {
		defer stderr.Close()
	}

	if cmd.background {
		proc := exec.Command(cmd.Name, cmd.Args...)
		proc.Stdin = os.Stdin
		proc.Stdout = stdout
		proc.Stderr = stderr

		backgroundJobs[backgroundCounter] = proc

		if err := proc.Start(); err != nil {
			return err
		}

		pid := proc.Process.Pid
		fmt.Printf("[%d] %d\n", backgroundCounter, pid)
		backgroundCounter++
		go func() {
			proc.Wait()
		}()

		return nil
	}

	proc := exec.Command(cmd.Name, cmd.Args...)
	proc.Stdin = os.Stdin
	proc.Stdout = stdout
	proc.Stderr = stderr

	if err := proc.Run(); err != nil {
		if _, isExit := err.(*exec.ExitError); !isExit {
			return err
		}
	}
	return nil
}

func dispatch(cmd Command) error {
	if cmd.Name == "" {
		return nil
	}
	if handler, ok := builtins[cmd.Name]; ok {
		return handler(cmd)
	}
	if ok, _ := isExecutable(cmd.Name); ok {
		return executeExternal(cmd)
	}
	return fmt.Errorf("%s: command not found", cmd.Name)
}
