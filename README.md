<img width="900" height="460" alt="image" src="https://github.com/user-attachments/assets/ff1b345e-1f47-4511-ab8a-f022d2148f23" />

# Gobash

A POSIX-like shell implementation written in Go as part of go studying and the journey of improving my skill in it by doing real projects rather than falling in the tutorial hell.

This project implements command parsing, built-in commands, program execution, job control, pipelines, redirections, command completion, history management, parameter expansion, and more.

---

## Features

### Core Shell Features

- Interactive shell prompt
- REPL (Read-Eval-Print Loop)
- Invalid command handling
- Command parsing
- Command execution
- Built-in command support

### Built-in Commands

- `exit`
- `echo`
- `type`
- `pwd`
- `cd`
- `history`
- `declare`
- `jobs`
- `complete`

---

## Executable Discovery & Program Execution

- Locate executables using `$PATH`
- Execute external programs
- Execute quoted executables
- Support command arguments

---

## Navigation

### `pwd`

Prints the current working directory.

### `cd`

Supports:

- Absolute paths
- Relative paths
- Home directory (`~`)

---

## Redirection

### Standard Output

```bash
command > file.txt
command >> file.txt
```

### Standard Error

```bash
command 2> error.txt
command 2>> error.txt
```

Implemented:

- Redirect stdout
- Redirect stderr
- Append stdout
- Append stderr

---

## Command Completion

### Builtin Completion

Tab completion for:

- Built-in commands
- Executables found in `$PATH`

### Advanced Completion

Supports:

- Partial completions
- Multiple matches
- Missing completions
- Completion with arguments
- Longest common prefix generation

---

## Filename Completion

Tab completion for filesystem paths.

Supports:

- Files
- Directories
- Nested paths
- Multi-argument completion
- Partial matches
- Multiple matches
- Missing completions

---

## Programmable Completion

Bash-style completion registration.

### Supported Features

- Register completions
- Unregister completions
- List registered completions
- Handle missing specifications
- Pass command-line arguments
- Pass environment variables
- Multiple completion candidates
- Longest common prefix resolution

---

## Background Jobs

Run commands in the background:

```bash
sleep 10 &
```

### Features

- `jobs` builtin
- Job tracking
- Background output handling
- Reaping completed jobs
- Multiple concurrent jobs
- Job number recycling

Examples:

```bash
jobs
jobs 1
```

---

## Pipelines

Supports Unix pipelines:

```bash
ls | grep go
```

### Features

- Dual-command pipelines
- Multi-command pipelines
- Pipelines with built-in commands

Examples:

```bash
cat file.txt | grep hello

ps aux | grep chrome | wc -l
```

---

## History

Command history support.

### Features

- `history` builtin
- List history entries
- Limit history entries
- Up-arrow navigation
- Down-arrow navigation
- Execute commands from history

Examples:

```bash
history
history 10
```

---

## History Persistence

History survives shell restarts.

### Features

- Read history from file
- Write history to file
- Append history to file
- Load history on startup
- Save history on exit

---

## Parameter Expansion

Shell variables and expansions.

### `declare`

Create shell variables:

```bash
declare NAME=Omar
```

### Expansion

```bash
echo $NAME
echo ${NAME}
```

### Features

- Variable creation
- Variable lookup
- Empty variable handling
- Variable name validation
- Brace expansion syntax

---

## Quoting

Supports standard shell quoting rules.

### Single Quotes

```bash
echo 'hello world'
```

### Double Quotes

```bash
echo "hello $USER"
```

### Escaping

```bash
echo hello\ world
```

Supported:

- Single quotes
- Double quotes
- Backslashes outside quotes
- Backslashes inside double quotes
- Correct handling inside single quotes
- Quoted executable execution

---

## Project Structure

```text
app/
├── builtins.go      # Built-in command implementations
├── dispatch.go      # Command dispatching
├── main.go          # Entry point
├── parser.go        # Shell parser
└── prompt.go        # Prompt and readline handling

go.mod
go.sum
run_gobash.sh
```

---

## Local Setup

### Prerequisites

- Go 1.26+
- Linux/macOS (recommended)

Verify installation:

```bash
go version
```

---

## Clone Repository

```bash
HTTPS
git clone git@github.com:oadel1814/gobash.git

SSH
git clone git@github.com:oadel1814/gobash.git
cd <repo_directory>
```

---

## Install Dependencies

```bash
go mod download
```

---

## Run the Shell

Using Go:

```bash
go run ./app
```

Or using the helper script:

```bash
./run_gobash.sh
```

---

## Example Session

```bash
$ echo Hello World
Hello World

$ pwd
/home/omar

$ declare NAME=Omar

$ echo $NAME
Omar

$ ls | grep go
main.go
parser.go

$ sleep 5 &
[1] 12345

$ jobs
[1] Running sleep 5
```

---

## Acknowledgements

Built as part of the CodeCrafters:

**Build Your Own Shell Challenge**

https://codecrafters.io
