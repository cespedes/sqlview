package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cespedes/tableview"
	"github.com/gdamore/tcell/v2"
)

func main() {
	err := run(os.Args[0:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		os.Exit(1)
	}
}

type app struct {
	Debug      bool
	ConfigFile string
	config
}

func keyStringMatch(str string, k tcell.Key, r rune) bool {
	switch str {
	case "enter":
		return k == tcell.KeyCR
	case "tab":
		return k == tcell.KeyTAB
	case "esc":
		return k == tcell.KeyESC
	}
	runes := []rune(str)
	if len(runes) == 1 && k == tcell.KeyRune {
		return r == runes[0]
	}
	return false
}

func run(args []string) error {
	app := app{}

	flags := flag.NewFlagSet(args[0], flag.ExitOnError)
	flags.BoolVar(&app.Debug, "debug", false, "Debugging")
	flags.StringVar(&app.ConfigFile, "config", filepath.Join(os.Getenv("HOME"), ".sqlview.yaml"), "Config file")
	flags.StringVar(&app.Format, "format", "", "Output format to use (default \"org\")")
	flags.StringVar(&app.Editor, "editor", "", "Editor to use")
	flags.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: sqlvi [options] [<page from config file>]")
		fmt.Fprintln(os.Stderr, "Options:")
		flags.PrintDefaults()
	}

	if err := flags.Parse(args[1:]); err != nil {
		return err
	}
	if len(flags.Args()) > 1 {
		return fmt.Errorf("too many arguments")
	}
	pageName := ""
	if len(flags.Args()) == 1 {
		pageName = flags.Args()[0]
	}

	err := app.readConfig()
	if err != nil {
		return err
	}
	if pageName == "" {
		pageName = app.Default
	}

	if app.Pages[pageName].Select == "" {
		return fmt.Errorf("no query specified")
	}

	db, err := sqlConnect(app.Connect)
	if err != nil {
		return err
	}
	result, err := sqlGenericQuery(db, app.Pages[pageName].Select)
	if err != nil {
		return err
	}

	t := tableview.NewTableView()
	t.FillTable(result.Columns, result.Strings)
	t.SetInputCapture(func(key tableview.Key, r rune, row int) bool {
		for k, action := range app.Pages[pageName].Keys {
			if keyStringMatch(k, key, r) {
				pageName = action
				result, err = sqlGenericQuery(db, app.Pages[pageName].Select)
				if err != nil {
					t.Suspend(func() {
						fmt.Printf("Error: %s\n", err.Error())
						os.Exit(1)
					})
				}
				t.Suspend(func() {
					fmt.Printf(">>> switching to page %q\n", pageName)
				})
				t.FillTable(result.Columns, result.Strings)
				return false
			}
		}
		for k, sw := range app.Pages[pageName].SwitchKeys {
			if keyStringMatch(k, key, r) {
				id := result.Strings[row][0]
				action := sw[id]
				if action == "" {
					break
				}
				pageName = action
				result, err = sqlGenericQuery(db, app.Pages[pageName].Select)
				if err != nil {
					t.Suspend(func() {
						fmt.Printf("Error: %s\n", err.Error())
						os.Exit(1)
					})
				}
				t.Suspend(func() {
					fmt.Printf(">>> switching to page %q\n", pageName)
				})
				t.FillTable(result.Columns, result.Strings)
				return false
			}
		}
		t.Suspend(func() {
			// fmt.Printf("keys = %+v\n", app.Pages[pageName].Keys)
			// fmt.Printf("switch-keys = %+v\n", app.Pages[pageName].SwitchKeys)
			fmt.Printf("input: key=%d rune=%d row=%d\n", key, r, row)
			if key == tcell.KeyTAB {
				fmt.Println("(TAB)")
			}
			if key == tcell.KeyCR {
				fmt.Println("(ENTER)")
			}
		})
		return true
	})
	t.SetSelectedFunc(func(row int) {
		// t.SetAlign(1, tableview.AlignRight)
		t.Suspend(func() {
			fmt.Printf("selected row: %d\n", row)
		})
	})
	t.NewCommand('F', "foo", func(row int) {
		t.Suspend(func() {
			fmt.Printf("command f: row: %d\n", row)
		})
	})
	t.Run()

	return nil
}
