package main

import (
	"log"

	"github.com/davidwalter0/go-u2f/cfg"
)

func init() {
	if err := cfg.Setup(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	log.Println("main")
}
