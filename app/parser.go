package main

import (
	"os"
	"strings"
)

type Command struct {
	Name           string
	Args           []string
	Stdout         string
	Stderr         string
	Append         bool
	background     bool
	StdinOverride  *os.File
	StdoutOverride *os.File
}

func expandEnvVars(args string) string {

	if strings.Contains(args, "{") {
		// handle ${VAR} syntax
		var result strings.Builder
		i := 0
		for i < len(args) {
			if args[i] == '$' && i+1 < len(args) && args[i+1] == '{' {
				end := strings.Index(args[i:], "}")
				if end != -1 {
					varName := args[i+2 : i+end]
					value := os.Getenv(varName)
					result.WriteString(value)
					i += end + 1
				} else {
					result.WriteByte(args[i])
					i++
				}
			} else {
				result.WriteByte(args[i])
				i++
			}
		}
		return result.String()
	}

	parts := strings.Split(args, "$")
	if len(parts) == 1 {
		return args
	}

	var result strings.Builder
	result.WriteString(parts[0])

	for _, part := range parts[1:] {

		varName := ""
		for i, r := range part {
			if r == '_' || (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
				varName += string(r)
			} else {
				result.WriteString(os.Getenv(varName))
				result.WriteString(part[i:])
				break
			}
		}

		if varName != "" {
			result.WriteString(os.Getenv(varName))
		}
	}

	return result.String()
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

			if strings.Contains(args[i], "$") {
				expanded := expandEnvVars(args[i])
				if expanded == "" {
					args = append(args[:i], args[i+1:]...)
					i--
				} else {
					args[i] = expanded
				}
				continue
			}

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
	var tokens []string
	var current strings.Builder
	inSingle := false
	inDouble := false

	runes := []rune(input)

	for i := 0; i < len(runes); i++ {
		r := runes[i]

		if r == '\\' && inDouble {
			if i+1 < len(runes) {
				if runes[i+1] == '"' || runes[i+1] == '\\' || runes[i+1] == '$' || runes[i+1] == '`' || runes[i+1] == '\n' {
					current.WriteRune(runes[i+1])
					i++
				}
				continue
			}
		}

		// handle backslash outside quotes
		if r == '\\' && !inSingle && !inDouble {
			if i+1 < len(runes) {
				current.WriteRune(runes[i+1])
				i++
			}
			continue
		}

		switch r {
		case '\'':
			if !inDouble {
				inSingle = !inSingle
				continue
			}
		case '"':
			if !inSingle {
				inDouble = !inDouble
				continue
			}

		case ' ', '\t':
			if !inSingle && !inDouble {
				if current.Len() > 0 {
					tokens = append(tokens, current.String())
					current.Reset()
				}
				continue
			}
		}

		current.WriteRune(r)
	}

	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}

	return tokens
}
