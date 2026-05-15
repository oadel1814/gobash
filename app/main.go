package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"
)

func prompt() string {
	fmt.Print("$ ")
	reader := bufio.NewReader(os.Stdin)
	cmd, _ := reader.ReadString('\n')
	return strings.TrimSpace(cmd)
}

func tokenize(cmd *string) []string {
	return strings.Fields(*cmd)
}

func is_executable(cmd *string) (bool, string) {

	for _, path := range strings.Split(os.Getenv("PATH"), ":") {
		fullPath := path + "/" + *cmd
		if fileInfo, err := os.Stat(fullPath); err == nil {
			if fileInfo.Mode().Perm()&0111 != 0 {
				return true, fullPath
			}
		}
	}

	return false, ""
}

func execute_external(cmd *string, args []string) error {
	proc := exec.Command(*cmd, args...)
	proc.Stdin = os.Stdin
	proc.Stdout = os.Stdout
	proc.Stderr = os.Stderr
	return proc.Run()
}

func main() {
	for {
		cmd := prompt()
		tokens := tokenize(&cmd)
		executable, _ := is_executable(&tokens[0])

		if executable {
			args := tokens[1:]
			err := execute_external(&tokens[0], args)
			if err != nil {
				fmt.Printf("%s: %s\n", tokens[0], err.Error())
			}
		} else if strings.ToLower(tokens[0]) == "exit" {
			break
		} else if strings.ToLower(tokens[0]) == "echo" {
			fmt.Println(strings.Join(tokens[1:], " "))
		} else if strings.ToLower(tokens[0]) == "type" {
			if slices.Contains([]string{"echo", "exit", "type"}, strings.ToLower(tokens[1])) == true {
				fmt.Printf("%s is a shell builtin\n", strings.ToLower(tokens[1]))
			} else {
				found, fullPath := is_executable(&tokens[1])
				if found {
					fmt.Printf("%s is %s\n", strings.ToLower(tokens[1]), fullPath)
				} else {
					fmt.Printf("%s: not found\n", strings.ToLower(tokens[1]))
				}
			}
		} else {
			fmt.Printf("%s: command not found\n", cmd)
		}
	}
}
