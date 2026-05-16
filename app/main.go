package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

type Command struct {
	Name   string
	Args   []string
	Stdout string
	Stderr string
	Append bool
}

type HandlerFunc func(cmd Command) error

var builtins map[string]HandlerFunc

func init() {
	builtins = map[string]HandlerFunc{
		"echo": handleEcho,
		"cd":   handleCd,
		"pwd":  handlePwd,
		"type": handleType,
		"exit": handleExit,
	}
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
	if stdout != os.Stdout {
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
	if len(cmd.Args) == 0 {
		return os.Chdir(os.Getenv("HOME"))
	}

	if cmd.Args[0] == "~" {
		return os.Chdir(os.Getenv("HOME"))
	}

	err := os.Chdir(cmd.Args[0])
	if err != nil {
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
	if _, ok := builtins[cmd.Args[0]]; ok {
		fmt.Printf("%s is a shell builtin\n", cmd.Args[0])
		// return handler(Command{})
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

func prompt() string {
	fmt.Print("$ ")
	reader := bufio.NewReader(os.Stdin)
	line, _ := reader.ReadString('\n')
	return strings.TrimSpace(line)
}

func parse(input string) Command {
	tokens := strings.Fields(input)
	cmd := Command{}

	if len(tokens) == 0 {
		return cmd
	}

	cmd.Name = tokens[0]
	args := tokens[1:]

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case ">", "1>":
			if i+1 < len(args) {
				cmd.Stdout = args[i+1]
				args = append(args[:i], args[i+2:]...)
				i--
			}
		case ">>", "1>>":
			if i+1 < len(args) {
				cmd.Stdout = args[i+1]
				cmd.Append = true
				args = append(args[:i], args[i+2:]...)
				i--
			}
		case "2>":
			if i+1 < len(args) {
				cmd.Stderr = args[i+1]
				args = append(args[:i], args[i+2:]...)
				i--
			}
		case "2>>":
			if i+1 < len(args) {
				cmd.Stderr = args[i+1]
				cmd.Append = true
				args = append(args[:i], args[i+2:]...)
				i--
			}
		}
	}

	cmd.Args = args
	return cmd
}

func resolveStdout(cmd Command) (*os.File, error) {
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

func main() {
	for {
		input := prompt()
		cmd := parse(input)
		if err := dispatch(cmd); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}
}
