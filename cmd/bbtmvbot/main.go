package main

import (
	"bbtmvbot"
	"flag"
	"fmt"
	"log"

	"bbtmvbot/config"
	_ "bbtmvbot/website/all"
)

var (
	version string

	configPath  = flag.String("config", "config.yml", "path to config file")
	flagVersion = flag.Bool("version", false, "prints version of the application")
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	flag.Parse()

	if *flagVersion {
		fmt.Println("Version:", version)
		return
	}

	c, err := config.New(*configPath)
	if err != nil {
		log.Fatalln("Configuration error:", err)
	}

	bbtmvbot.Start(c)
}
