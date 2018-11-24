package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/davidwalter0/go-u2f/cfg"
	"github.com/davidwalter0/go-u2f/u2f"
)

func init() {
	if err := cfg.Setup(); err != nil {
		log.Fatal(err)
	}

	CfgDump()

	go func() {
		for {
			select {
			case text := <-Message:
				log.Println("Saw Message", text)
			}
		}
	}()
}

func CfgDump() {
	text, err := json.MarshalIndent(cfg.Env, "", "  ")
	if err != nil {
		return
	}
	log.Printf("\n%s\n", string(text))
}

var Message = make(chan string)

func main() {
	log.Println(u2f.InsertSecurityKeyMessage)
	log.Println(u2f.PressKeyToAuthenticate)
	for _, Action := range []string{u2f.Register, u2f.Authenticate} {
		log.Println("Action", Action)
		Message <- Action
		switch Action {
		case u2f.Register:
			{
				done := false
				for i := 0; i < 3 && !done; i++ {
					Message <- fmt.Sprintf("%s: %s", Action, u2f.PressKeyToAuthenticate)
					if err := u2f.U2FAction(Action, Message); err != nil {
						Message <- err.Error()
						time.Sleep(5 * time.Second)
					} else {
						done = true
					}
				}
			}
		case u2f.Authenticate:
			{
				done := false

				for i := 0; i < 3 && !done; i++ {
					Message <- fmt.Sprintf("%s: %s", Action, u2f.PressKeyToAuthenticate)
					if err := u2f.U2FAction(Action, Message); err != nil {
						Message <- err.Error()
						time.Sleep(5 * time.Second)
					} else {
						done = true
					}
				}
			}
		default:
		}
	}
	log.Printf("Exiting in 3 ")
	for i := 0; i < 3; i++ {
		time.Sleep(time.Second)
		fmt.Printf(".")
	}
	fmt.Println()
	close(Message)
}
