package main

import (
	"log"

	"github.com/vishal1132/cafebucks/config"
	eb "github.com/vishal1132/cafebucks/eventbus"
)

var coffees []eb.Coffee
var orderMap map[int]eb.Order

func main() {
	seedCoffees()
	cfg, err := config.LoadEnv()
	if err != nil {
		log.Println(err)
	}
	l := config.DefaultLogger(cfg)
	if err := runserver(cfg, l); err != nil {
		l.Fatal().Err(err).Msg("Failed to run order service server")
	}
}

func seedCoffees() {
	orderMap = make(map[int]eb.Order, 10)
	coffees = make([]eb.Coffee, 0, 10)
	coffees = []eb.Coffee{
		{"cappuccino", 1.2},
		{"frappuccino", 1.5},
		{"espresso", 1.8},
		{"americano", 1.9},
		{"indiano", 2.5},
	}
}
