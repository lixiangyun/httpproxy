package main

import (
	"flag"
	"log"
	"time"
)

var (
	CONFIG_FILE string
	DEBUG       bool
	HELP        bool
)

func init() {
	flag.StringVar(&CONFIG_FILE, "conf", "config.yaml", "http proxy global config file.")
	flag.BoolVar(&DEBUG, "debug", false, "debug mode.")
	flag.BoolVar(&HELP, "h", false, "this help.")
}

func main() {
	flag.Parse()

	_, err := LoadConfig("./config.yaml")
	if err != nil {
		log.Fatalln(err.Error())
		return
	}

	RouterInit()
	ClusterInit()
	ListenerInit()

	for {
		time.Sleep(time.Second)
	}
}
