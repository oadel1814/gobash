package main

import (
	"fmt"
	"os"
	"os/exec"
	"time"
)

const version = "1.0.0"

func printBanner() {
	const (
		reset      = "\033[0m"
		bold       = "\033[1m"
		dim        = "\033[2m"
		cyan       = "\033[36m"
		brightCyan = "\033[96m"
		white      = "\033[97m"
		gray       = "\033[90m"
		green      = "\033[32m"
		yellow     = "\033[33m"
	)

	cmd := exec.Command("npx", "oh-my-logo", "Gobash",
		"--palette-colors", `["#00ADD8","#00ADD8","#00ADD8"]`,
		"--filled",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()

	fmt.Printf("\n %s%s%s  %s%s%s\n",
		bold+white, "Gobash", reset,
		dim+gray, version, reset,
	)
	fmt.Printf(" %sA shell written in Go.  Type %s%shelp%s%s for available commands.%s\n\n",
		gray,
		reset, cyan, reset, gray,
		reset,
	)

	printBannerRow("  Author", "Omar Adel", cyan, gray, reset)
	printBannerRow("  Source", "github.com/oadel1814/gobash", cyan, gray, reset)

	fmt.Println()

	hints := []struct{ key, desc string }{
		{"TAB", "complete"},
		{"↑ ↓", "history"},
		{"Ctrl+C", "cancel"},
		{"Ctrl+D", "exit"},
		{"Ctrl+L", "clear"},
	}
	fmt.Printf(" ")
	for i, h := range hints {
		fmt.Printf("%s%s%s %s%s%s", brightCyan, h.key, reset, dim+gray, h.desc, reset)
		if i < len(hints)-1 {
			fmt.Printf("  %s│%s  ", gray, reset)
		}
	}
	fmt.Println()

	tips := []string{
		"Use && to chain commands: make && ./bin/gobash",
		"Redirect output with > file.txt or >> to append",
		"Run jobs in background with & — check them with jobs",
		"Use $? to inspect the last exit code",
		"Pipe commands together: ls | grep .go | wc -l",
	}
	tip := tips[time.Now().UnixNano()%int64(len(tips))]
	fmt.Printf("\n %s✦ Tip:%s %s%s%s\n\n", yellow, reset, dim+gray, tip, reset)
}

func printBannerRow(label, value, labelColor, valueColor, reset string) {
	fmt.Printf(" %s%s%-10s%s  %s%s%s\n",
		labelColor, label, "", reset,
		valueColor, value, reset,
	)
}

func main() {
	initReadline()
	defer rl.Close()
	// printBanner()
	loadHistory(os.Getenv("HISTFILE"))
	for {
		input := prompt()
		cmd := parse(input)
		if err := dispatch(cmd); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}
}
