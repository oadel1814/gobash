package main

import (
	"bufio"
	"fmt"
	"os"
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
		} else {
			fmt.Printf("%s: command not found\n", cmd)
		}
	}
}
