package main

import (
	"bbtmvbot"
	"flag"
	"fmt"
	"log"
	"os"

	"bbtmvbot/config"
	_ "bbtmvbot/website/all"
)

var (
	version string

	configPath  = flag.String("config", "config.yml", "path to config file")
	flagVersion = flag.Bool("version", false, "prints version of the application")
)

func main() {
	if *flagVersion {
		fmt.Println("Version:", version)
		return
	}

	// Read config
	c, err := config.New(*configPath)
	if err != nil {
		log.Fatalln("Configuration error:", err)
	}

	// Ensure data dir exists
	if _, err := os.Stat(c.DataDir); os.IsNotExist(err) {
		err = os.MkdirAll(c.DataDir, 0755)
		if err != nil {
			log.Fatalln("Could not create data dir:", err)
		}
	}

	bbtmvbot.Start(c)
}
