package main

import (
	"fmt"
	"os"
	"os/exec"

	"gopkg.in/yaml.v3"
)

// Editor is the representation of a YAML file
// to be edited by the user, using a text editor,
// in order to create or modify info from SQL.
type Editor struct {
	file    *os.File
	name    string
	columns []string
	Results []interface{} // each value will always be a string or a []string
}

// Close frees the resources referenced by an Editor
func (e *Editor) Close() error {
	e.file = nil
	return os.Remove(e.name)
}

// NewEditor creates an empty YAML file,
// and prepares a Editor to be run.
func NewEditor(columns []string) (*Editor, error) {
	values := make([]interface{}, len(columns))
	for i := range values {
		values[i] = ""
	}
	return NewEditorData(columns, values)
}

// NewEditorData fills a YAML file with some values,
// each of them must be a string or []string
// and prepares a Editor to be run.
func NewEditorData(columns []string, values []interface{}) (*Editor, error) {
	var e Editor
	var err error
	e.columns = columns
	e.file, err = os.CreateTemp("", "sqlview.*.yaml")
	if err != nil {
		return nil, err
	}
	e.name = e.file.Name()

	for i := range columns {
		if arr, ok := values[i].([]string); ok {
			fmt.Fprintf(e.file, "%s:\n", columns[i])
			for _, elem := range arr {
				fmt.Fprintf(e.file, "- %s\n", elem)
			}
			if len(arr) == 0 {
				fmt.Fprintf(e.file, "- \n")
			}
		} else if values[i] == nil {
			fmt.Fprintf(e.file, "%s: \n", columns[i])
		} else {
			fmt.Fprintf(e.file, "%s: %v\n", columns[i], values[i])
		}
	}
	if err = e.file.Close(); err != nil {
		return nil, err
	}
	return &e, nil
}

// Edit runs a text editor with the info in a Editor,
// and returns its result.
func (e *Editor) Edit(execs ...string) error {
	var err error
	paths := append(execs, os.Getenv("VISUAL"),
		os.Getenv("EDITOR"),
		"editor",
		"vim",
		"vi",
	)
	paths = append(paths, execs...)
	for _, str := range paths {
		err = runCommand(str, e.name)
		if err == nil {
			break
		}
		if _, ok := err.(*exec.ExitError); ok {
			return err
		}
	}
	if err != nil {
		return err
	}

	data, err := os.ReadFile(e.name)
	if err != nil {
		return err
	}

	out := make(map[string]interface{})

	if err = yaml.Unmarshal(data, &out); err != nil {
		return err
	}
	e.Results = make([]interface{}, len(e.columns))
	for key, value := range out {
		idx := -1
		for i, k := range e.columns {
			if k == key {
				idx = i
				break
			}
		}
		if idx == -1 {
			return fmt.Errorf("unexpected key %s", key)
		}
		if arr, ok := value.([]interface{}); ok {
			strs := make([]string, len(arr))
			for i := range arr {
				strs[i] = fmt.Sprint(arr[i])
			}
			e.Results[idx] = strs
		} else if value == nil {
			e.Results[idx] = ""
		} else {
			e.Results[idx] = fmt.Sprint(value)
		}
	}
	return nil
}

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
