package main

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"sync/atomic"

	cfg "github.com/davidwalter0/go-cfg"
)

type Application struct {
	Filename string `json:"filename" required:"true" doc:"filename to persist state to"` // default:"registration.json"`
	Debug    bool
	Tracing  bool
}

func (app *Application) Trace(f string) func() {
	if app.Tracing {
		called := atomic.AddInt64(Semaphore, 0)
		text := fmt.Sprintf("Item[%d] LastAction[%s] Action[%s]",
			called,
			LastAction,
			Action,
		)
		log.Println("Entering", f, text)
	}
	return func() {
		if app.Tracing {
			called := atomic.AddInt64(Semaphore, 0)
			text := fmt.Sprintf("Item[%d] LastAction[%s] Action[%s]",
				called,
				LastAction,
				Action,
			)
			log.Println("Leaving", f, text)
		}
	}
}

func (app *Application) String() string {
	text, err := json.Marshal(app)
	if err != nil {
		return ""
	}
	return string(text)
}

var app = &Application{}

func init() {
	if err := cfg.Env(app); err != nil {
		cfg.Usage()
		log.Fatal(err)
	}
}

var Semaphore = new(int64)
var mutex = &sync.Mutex{}
