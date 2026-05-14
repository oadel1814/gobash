package main

import (
	"fmt"
	"strings"
)

// Ensures gofmt doesn't remove the "fmt" import in stage 1 (feel free to remove this!)
var _ = fmt.Print

func prompt() string {
	fmt.Print("$ ")
	var cmd string
	fmt.Scanln(&cmd)
	return cmd
}

func main() {
	// TODO: Uncomment the code below to pass the first stage
	for {
		var cmd string
		cmd = strings.ToLower(prompt())
		if cmd == "exit" {
			break
		}

		fmt.Printf("%s: command not found\n", cmd)
	}
}
