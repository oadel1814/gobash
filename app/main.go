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

func tokenize(cmd *string) []string {
	return strings.Fields(*cmd)
}

func main() {
	for {
		cmd := prompt()
		tokens := tokenize(&cmd)
		if strings.ToLower(tokens[0]) == "exit" {
			break
		} else if strings.ToLower(tokens[0]) == "echo" {
			fmt.Println(strings.Join(tokens[1:], " "))
		} else if strings.ToLower(tokens[0]) == "type" {
			if slices.Contains([]string{"echo", "exit", "type"}, strings.ToLower(tokens[1])) == true {
				fmt.Printf("%s is a shell builtin\n", strings.ToLower(tokens[1]))
			} else {
				found := false

				for _, path := range strings.Split(os.Getenv("PATH"), ":") {
					fullPath := path + "/" + tokens[1]
					if fileInfo, err := os.Stat(fullPath); err == nil {
						if fileInfo.Mode().Perm()&0111 != 0 {
							fmt.Printf("%s is %s\n", strings.ToLower(tokens[1]), fullPath)
							found = true
							break
						}
					}
				}

				if found == false {
					fmt.Printf("%s: not found\n", strings.ToLower(tokens[1]))
				}

			}
		} else {
			fmt.Printf("%s: command not found\n", cmd)
		}
	}
}
