package main

import (
	"log"
	"os"

	"github.com/davidwalter0/go-u2f/cfg"
)

func init() {
	os.Setenv("FILENAME", "registration.json")
	if err := cfg.Setup(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	log.Println("main")
	log.Println(cfg.Env.Format())
}
