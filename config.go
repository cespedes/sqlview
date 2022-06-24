package main

import (
	"fmt"
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
)

/*
   editor: vim  # optional
   default: countries
   pages:
     countries:
       connect: postgres://user:pass@host/database
       select: SELECT id,country,capital,population FROM countries
       insert: INSERT INTO countries (country,capital,population) VALUES ($2,$3,$4)
       update: UPDATE countries SET country=$2,capital=$3,population=$4 WHERE id=$1
       delete: DELETE FROM countries WHERE id=$1
     products:
       connect: postgres://user:pass@host/database
       select: SELECT id,name,price FROM products
       insert: INSERT INTO products (name,price) VALUES ($2,$3)
       update: UPDATE products SET name=$2,price=$3 WHERE id=$1
       delete: DELETE FROM product WHERE id=$1
*/

type configPage struct {
	Select     string
	Insert     string
	Update     string
	Delete     string
	Keys       map[string]string
	SwitchKeys map[string]map[string]string `yaml:"switch-keys"`
}

type config struct {
	Editor  string
	Default string
	Format  string
	Connect string
	Pages   map[string]configPage
}

func (app *app) readConfig() error {
	var config config

	if app.Debug {
		log.Printf("Reading config file %s", app.ConfigFile)
	}
	data, err := ioutil.ReadFile(app.ConfigFile)
	if err != nil {
		if app.Debug {
			log.Printf("Warning: cannot read %s: %s", app.ConfigFile, err.Error())
		}
		return nil
	}
	if err := yaml.UnmarshalStrict(data, &config); err != nil {
		return fmt.Errorf("parsing %s: %w", app.ConfigFile, err)
	}

	app.Default = config.Default

	if app.Format == "" {
		app.Format = config.Format
		if app.Format == "" {
			app.Format = "org"
		}
	}
	if app.Connect == "" {
		app.Connect = config.Connect
	}
	app.Pages = config.Pages

	return nil
}
