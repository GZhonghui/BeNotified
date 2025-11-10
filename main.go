package main

import (
	"log"

	"BeNotified.local/linebc"
)

func main() {
	if err := linebc.InitFromEnv(); err != nil {
		log.Fatal(err)
	}
	if err := linebc.BroadcastText("Hello everyone!"); err != nil {
		log.Fatal(err)
	}
}
