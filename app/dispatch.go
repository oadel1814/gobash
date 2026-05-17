package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func isExecutable(name string) (bool, string) {
	for _, dir := range strings.Split(os.Getenv("PATH"), ":") {
		fullPath := dir + "/" + name
		if info, err := os.Stat(fullPath); err == nil {
			if info.Mode().Perm()&0111 != 0 {
				return true, fullPath
			}
		}
	}
	return false, ""
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
