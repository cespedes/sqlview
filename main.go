package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cespedes/tableview"
	"github.com/gdamore/tcell/v2"
	"github.com/jmoiron/sqlx"
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
	pageName   string
	pageArgs   []string
	db         *sqlx.DB
	result     SQLResult
	table      *tableview.TableView
	config
}

func sliceStringToAny(in []string) []interface{} {
	out := make([]interface{}, len(in))
	for i, v := range in {
		out[i] = v
	}
	return out
}

func (a *app) changePage(page string, args []string) error {
	var err error

	fields := strings.Fields(page)
	if len(fields) == 0 {
		return fmt.Errorf("empty page %q", page)
	}
	a.pageName = fields[0]
	a.pageArgs = fields[1:]

	// Bind dollars in pageArgs:
	for i, arg := range a.pageArgs {
		res := ""
		for j := strings.Index(arg, "$"); j != -1; j = strings.Index(arg, "$") {
			res += arg[:j]
			arg = arg[j+1:]
			argNum := 0
			for len(arg) > 0 && arg[0] >= '0' && arg[0] <= '9' {
				argNum *= 10
				argNum += int(arg[0]) - '0'
				arg = arg[1:]
			}
			res += args[argNum-1]
		}
		res += arg
		a.pageArgs[i] = res
	}

	query, bindArgs := sqlBind(a.db, a.Pages[a.pageName].Select, a.pageArgs)
	//	if a.table != nil && len(a.pageArgs) > 0 {
	//		a.table.Suspend(func() {
	//			fmt.Printf("QUERY=%q ARGS=%q\n", query, args)
	//			time.Sleep(5 * time.Second)
	//		})
	//	}
	a.result, err = sqlQuery(a.db, query, bindArgs...)
	if err != nil {
		err = fmt.Errorf("changePage(%s): %v <%q,%q>", page, err, query, bindArgs)
	}
	return err
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
	var err error
	app := app{}

	flags := flag.NewFlagSet(args[0], flag.ExitOnError)
	flags.BoolVar(&app.Debug, "debug", false, "Debugging")
	flags.StringVar(&app.ConfigFile, "config", filepath.Join(os.Getenv("HOME"), ".sqlview.yaml"), "Config file")
	flags.StringVar(&app.Editor, "editor", "", "Editor to use")
	flags.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: sqlvi [options] [<page from config file>]")
		fmt.Fprintln(os.Stderr, "Options:")
		flags.PrintDefaults()
	}

	if err = flags.Parse(args[1:]); err != nil {
		return err
	}
	if len(flags.Args()) > 1 {
		return fmt.Errorf("too many arguments")
	}
	// pageArgs := []string{}
	if len(flags.Args()) == 1 {
		app.pageName = flags.Args()[0]
	}

	if err = app.readConfig(); err != nil {
		return err
	}
	if app.pageName == "" {
		app.pageName = app.DefaultPage
	}

	if app.Pages[app.pageName].Select == "" {
		return fmt.Errorf("no query specified")
	}

	app.db, err = sqlConnect(app.Connect)
	if err != nil {
		return err
	}
	err = app.changePage(app.pageName, nil)
	if err != nil {
		return err
	}

	app.table = tableview.NewTableView()
	app.table.FillTable(app.result.Columns, app.result.Strings)
	app.table.SetInputCapture(func(key tableview.Key, r rune, row int) bool {
		for k, action := range app.Pages[app.pageName].Keys {
			if keyStringMatch(k, key, r) {
				app.table.Suspend(func() {
					fmt.Printf(">>> page=%q,key=%q: switching to page %q\n", app.pageName, k, action)
				})
				err = app.changePage(action, app.result.Strings[row])
				if err != nil {
					app.table.Suspend(func() {
						fmt.Printf("Error: %s\n", err.Error())
						os.Exit(1)
					})
				}
				app.table.FillTable(app.result.Columns, app.result.Strings)
				return false
			}
		}
		for k, sw := range app.Pages[app.pageName].SwitchKeys {
			if keyStringMatch(k, key, r) {
				id := app.result.Strings[row][0]
				action := sw[id]
				if action == "" {
					break
				}
				app.table.Suspend(func() {
					fmt.Printf(">>> page=%q,key=%q,id=%q: switching to page %q\n", app.pageName, k, id, action)
				})
				err = app.changePage(action, app.result.Strings[row])
				if err != nil {
					app.table.Suspend(func() {
						fmt.Printf("Error: %s\n", err.Error())
						os.Exit(1)
					})
				}
				app.table.FillTable(app.result.Columns, app.result.Strings)
				return false
			}
		}
		if key == tcell.KeyTAB || key == tcell.KeyCR {
			app.table.Suspend(func() {
				// fmt.Printf("keys = %+v\n", app.Pages[app.pageName].Keys)
				// fmt.Printf("switch-keys = %+v\n", app.Pages[app.pageName].SwitchKeys)
				// fmt.Printf("input: key=%d rune=%d row=%d\n", key, r, row)
				if key == tcell.KeyTAB {
					fmt.Println("(TAB)")
				}
				if key == tcell.KeyCR {
					fmt.Println("(ENTER)")
				}
			})
		}
		return true
	})
	app.table.SetSelectedFunc(func(row int) {
		// t.SetAlign(1, tableview.AlignRight)
		app.table.Suspend(func() {
			fmt.Printf("selected row: %d\n", row)
		})
	})
	app.table.NewCommand('N', "new", func(row int) {
		app.table.Suspend(func() {
			fmt.Printf("creating new page (TODO)\n")
			time.Sleep(time.Second)
		})
	})
	app.table.NewCommand('E', "edit", func(row int) {
		app.table.Suspend(func() {
			fmt.Printf("editing entry (TODO): row=%d\n", row)
			time.Sleep(time.Second)
		})
	})
	app.table.NewCommand('D', "delete", func(row int) {
		app.table.Suspend(func() {
			fmt.Printf("deleting entry (TODO): row=%d\n", row)
			time.Sleep(time.Second)
		})
	})
	app.table.Run()

	return nil
}
