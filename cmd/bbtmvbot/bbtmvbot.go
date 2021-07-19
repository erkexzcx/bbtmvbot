package main

import (
	"bbtmvbot"
	"flag"
	"log"

	"bbtmvbot/config"
	_ "bbtmvbot/website/all"
)

var configPath = flag.String("config", "config.yml", "path to config file")
var dbPath = flag.String("database", "bbtmvbot.db", "path to database file")

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	flag.Parse()

	c, err := config.New(*configPath)
	if err != nil {
		log.Fatalln(err)
	}

	bbtmvbot.Start(c, dbPath)
}
