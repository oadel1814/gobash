package main

import (
	"fmt"
	"os"
	"os/exec"
)

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

		if err := proc.Start(); err != nil {
			return err
		}

		jobID := 1
		for {
			if _, exists := backgroundJobs[jobID]; !exists {
				break
			}
			jobID++
		}

		backgroundJobs[jobID] = proc

		secondMostRecentJob = mostRecentJob
		mostRecentJob = jobID

		pid := proc.Process.Pid
		fmt.Printf("[%d] %d\n", jobID, pid)

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

func executePipeline(leftCmd, rightCmd Command) error {
	reader, writer, err := os.Pipe()
	if err != nil {
		return err
	}

	leftCmd.StdoutOverride = writer
	rightCmd.StdinOverride = reader

	var leftWait func() error
	var rightWait func() error

	if builtin, ok := builtins[leftCmd.Name]; ok {
		done := make(chan error)
		go func() {
			err := builtin(leftCmd)
			writer.Close()
			done <- err
		}()
		leftWait = func() error { return <-done }
	} else {
		leftProc := exec.Command(leftCmd.Name, leftCmd.Args...)
		leftProc.Stdin = os.Stdin
		leftProc.Stdout = writer
		leftProc.Stderr = os.Stderr
		if err := leftProc.Start(); err != nil {
			return err
		}
		writer.Close()
		leftWait = func() error { return leftProc.Wait() }
	}

	if builtin, ok := builtins[rightCmd.Name]; ok {
		done := make(chan error)
		go func() {
			err := builtin(rightCmd)
			reader.Close()
			done <- err
		}()
		rightWait = func() error { return <-done }
	} else {
		rightProc := exec.Command(rightCmd.Name, rightCmd.Args...)
		rightProc.Stdin = reader
		rightProc.Stdout = os.Stdout
		rightProc.Stderr = os.Stderr
		if err := rightProc.Start(); err != nil {
			return err
		}
		reader.Close()
		rightWait = func() error { return rightProc.Wait() }
	}

	errLeft := leftWait()
	errRight := rightWait()

	if errRight != nil {
		if _, isExit := errRight.(*exec.ExitError); !isExit {
			return errRight
		}
	} else if errLeft != nil {
		if _, isExit := errLeft.(*exec.ExitError); !isExit {
			return errLeft
		}
	}

	return nil
}

func dispatch(cmds []Command) error {
	if len(cmds) == 0 {
		return nil
	}

	if len(cmds) == 1 {
		cmd := cmds[0]
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

	if len(cmds) == 2 {
		return executePipeline(cmds[0], cmds[1])
	}

	return fmt.Errorf("pipelines with more than 2 commands not yet supported")
}
