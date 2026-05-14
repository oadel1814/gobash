package main

import (
	"fmt"
)

// Ensures gofmt doesn't remove the "fmt" import in stage 1 (feel free to remove this!)
var _ = fmt.Print

func main() {
	// TODO: Uncomment the code below to pass the first stage
	for {
		fmt.Print("$ ")
		var cmd string

		// blocking the code until the user inputs a line of text and presses enter
		// read a line of input and store it in the address of the "cmd" variable
		fmt.Scanln(&cmd)
		fmt.Printf("%s: command not found\n", cmd)
	}
}
