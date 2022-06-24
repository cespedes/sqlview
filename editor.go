package main

import (
	"os"
	"os/exec"
)

func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	return err
}

func (app *app) callEditor(args ...string) error {
	var err error
	strs := []string{
		app.Editor,
		os.Getenv("VISUAL"),
		os.Getenv("EDITOR"),
		"editor",
		"vim",
		"vi",
	}
	for _, str := range strs {
		err = runCommand(str, args...)
		if err == nil {
			return nil
		}
		if _, ok := err.(*exec.ExitError); ok {
			return err
		}
	}
	return err
}
