package main

import (
	"bufio"
	"fmt"
	"os"
	"slices"
	"strings"
)

// Ensures gofmt doesn't remove the "fmt" import in stage 1 (feel free to remove this!)
var _ = fmt.Print

func prompt() string {
	fmt.Print("$ ")
	reader := bufio.NewReader(os.Stdin)
	cmd, _ := reader.ReadString('\n')
	return strings.TrimSpace(cmd)
}

func main() {
	for {
		cmd := prompt()
		words := strings.Fields(cmd)
		if strings.ToLower(words[0]) == "exit" {
			break
		} else if strings.ToLower(words[0]) == "echo" {
			fmt.Println(strings.Join(words[1:], " "))
		} else if strings.ToLower(words[0]) == "type" {
			if slices.Contains([]string{"echo", "exit", "type"}, strings.ToLower(words[1])) == true {
				fmt.Printf("%s is a shell builtin\n", strings.ToLower(words[1]))
			} else {
				fmt.Printf("%s: not found\n", strings.ToLower(words[1]))
			}
		} else {
			fmt.Printf("%s: command not found\n", cmd)
		}
	}
}
