package main

import "strings"

// Command holds a parsed shell command.
type Command struct {
	Name       string
	Args       []string
	Stdout     string
	Stderr     string
	Append     bool
	background bool
}

func parse(input string) []Command {

	// parse for pipes here and return a slice of Commands, one per pipe segment
	pipeSegments := strings.Split(input, "|")
	commands := make([]Command, 0, len(pipeSegments))

	for _, segment := range pipeSegments {
		tokens := tokenize(strings.TrimSpace(segment))
		cmd := Command{}

		if len(tokens) == 0 {
			commands = append(commands, cmd)
			continue
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
			case "&":
				cmd.background = true
				args = append(args[:i], args[i+1:]...)
				i--
			default:
				continue
			}
		}

		cmd.Args = args
		commands = append(commands, cmd)
	}
	return commands
}

func tokenize(input string) []string {
	// thin wrapper kept separate so you can later
	// swap in a smarter tokenizer (e.g. quoting support)
	return strings.Fields(input)
}
