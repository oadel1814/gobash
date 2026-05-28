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

func executePipeline(cmds []Command) error {
	var waits []func() error
	var prevReader *os.File

	for i := 0; i < len(cmds); i++ {
		cmd := cmds[i]
		isLast := (i == len(cmds)-1)

		var nextReader, currentWriter *os.File

		if !isLast {
			r, w, err := os.Pipe()
			if err != nil {
				return err
			}
			nextReader = r
			currentWriter = w
		}

		cmd.StdinOverride = prevReader
		cmd.StdoutOverride = currentWriter

		if builtin, ok := builtins[cmd.Name]; ok {
			done := make(chan error, 1)
			pr := prevReader
			cw := currentWriter

			go func(c Command) {
				err := builtin(c)
				if pr != nil {
					pr.Close()
				}
				if cw != nil {
					cw.Close()
				}
				done <- err
			}(cmd)

			waits = append(waits, func() error { return <-done })

		} else {
			proc := exec.Command(cmd.Name, cmd.Args...)

			if prevReader != nil {
				proc.Stdin = prevReader
			} else {
				proc.Stdin = os.Stdin
			}

			if currentWriter != nil {
				proc.Stdout = currentWriter
			} else {
				stdout, err := resolveStdout(cmd)
				if err != nil {
					return err
				}
				if stdout != os.Stdout {
					defer stdout.Close()
				}
				proc.Stdout = stdout
			}

			stderr, err := resolveStderr(cmd)
			if err != nil {
				return err
			}
			if stderr != os.Stderr {
				defer stderr.Close()
			}
			proc.Stderr = stderr

			if err := proc.Start(); err != nil {
				return err
			}

			if prevReader != nil {
				prevReader.Close()
			}
			if currentWriter != nil {
				currentWriter.Close()
			}

			waits = append(waits, func() error { return proc.Wait() })
		}

		prevReader = nextReader
	}

	var lastErr error
	for _, w := range waits {
		err := w()
		if err != nil {
			if _, isExit := err.(*exec.ExitError); !isExit {
				lastErr = err
			}
		}
	}

	return lastErr
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

	return executePipeline(cmds)

}
