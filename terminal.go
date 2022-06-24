package main

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

type askStruct struct {
	char rune
	help string
}

func readKey() rune {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return 0
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)
	b := make([]byte, 1)
	os.Stdin.Read(b)
	return rune(b[0])
}

func ask(msg string, actions []askStruct) rune {
	chars := ""
	for _, a := range actions {
		chars = chars + string(a.char)
	}

	for {
		fmt.Printf("%s [%s?] ", msg, chars)
		r := readKey()
		fmt.Println(string(r))

		if strings.ContainsRune(chars, r) {
			return r
		}
		if r == '?' {
			fmt.Printf("Commands:\n")
			for _, a := range actions {
				fmt.Printf("  %c -- %s\n", a.char, a.help)
			}
			fmt.Printf("  ? -- this help\n")
			continue
		}
		fmt.Printf("Please enter one of [%s?]\n", chars)
		fmt.Printf("  (Type '?' for help.)\n")
	}
}

func askError() bool {
	c := ask(`What now?`, []askStruct{
		{'e', "open editor again"},
		{'Q', "discard changes and quit"},
	})
	if c == 'e' {
		return true
	}
	return false
}
