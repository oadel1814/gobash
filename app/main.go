package main

import (
	"fmt"
	"os"
)

func main() {
	initReadline()
	defer rl.Close()

	for {
		input := prompt()
		cmd := parse(input)
		if err := dispatch(cmd); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}
}
