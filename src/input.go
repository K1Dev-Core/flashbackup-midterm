package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
)

var commands = []string{"/help", "/source", "/dest", "/set", "/add", "/settings", "/list", "/move", "/check", "/delete", "/clean", "/exit"}

func interactiveTerminal() bool {
	return term.IsTerminal(int(os.Stdin.Fd())) && term.IsTerminal(int(os.Stdout.Fd()))
}

func (a *App) runInteractive() {
	fd := int(os.Stdin.Fd())
	state, err := term.MakeRaw(fd)
	if err != nil {
		a.runScanner()
		return
	}
	defer term.Restore(fd, state)

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		line, ok, err := readInteractiveLine(reader)
		if err != nil {
			fmt.Fprintln(os.Stderr, "input error:", err)
			return
		}
		if !ok {
			fmt.Print("\r\n")
			return
		}
		if line != "" && a.command(line) {
			return
		}
	}
}

func readInteractiveLine(reader *bufio.Reader) (string, bool, error) {
	line := make([]rune, 0, 64)
	for {
		r, _, err := reader.ReadRune()
		if err != nil {
			if err == io.EOF {
				return "", false, nil
			}
			return "", false, err
		}
		switch r {
		case '\r', '\n':
			fmt.Print("\r\n")
			return string(line), true, nil
		case '\t':
			next, matches := completeCommand(string(line))
			if len(matches) > 1 {
				fmt.Printf("\r\nคำสั่งที่ตรงกัน: %s\r\n> %s", strings.Join(matches, "  "), string(line))
			} else if next != string(line) {
				fmt.Print(next[len(string(line)):])
				line = []rune(next)
			}
		case 3:
			fmt.Print("^C\r\n")
			return "", true, nil
		case 4:
			if len(line) == 0 {
				return "", false, nil
			}
		case 8, 127:
			if len(line) > 0 {
				line = line[:len(line)-1]
				fmt.Print("\b \b")
			}
		default:
			if r >= 32 {
				line = append(line, r)
				fmt.Printf("%c", r)
			}
		}
	}
}

func completeCommand(line string) (string, []string) {
	if strings.ContainsAny(line, " \t") {
		return line, nil
	}
	matches := make([]string, 0, len(commands))
	for _, command := range commands {
		if strings.HasPrefix(command, line) {
			matches = append(matches, command)
		}
	}
	if len(matches) == 1 {
		return matches[0] + " ", nil
	}
	return line, matches
}
